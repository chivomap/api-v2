package geospatial

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	
	"chivomap.com/config"
)

// Transform contiene la escala y traslación para reconstruir las coordenadas.
type Transform struct {
	Scale     []float64 `json:"scale"`
	Translate []float64 `json:"translate"`
}

// TopoJSON representa el archivo TopoJSON.
type TopoJSON struct {
	Type      string                `json:"type"`
	Objects   map[string]TopoObject `json:"objects"`
	Arcs      [][][]float64         `json:"arcs"`
	Transform Transform             `json:"transform"`
}

// TopoObject agrupa las geometrías.
type TopoObject struct {
	Type       string     `json:"type"`
	Geometries []Geometry `json:"geometries"`
}

// Geometry ahora incluye un campo "Coordinates" para puntos y multipuntos.
type Geometry struct {
	Type        string          `json:"type"`
	Arcs        json.RawMessage `json:"arcs"`
	Coordinates json.RawMessage `json:"coordinates,omitempty"`
	Properties  map[string]any  `json:"properties"`
}

// GeoFeature representa un Feature de GeoJSON.
type GeoFeature struct {
	Type       string                 `json:"type"`
	Geometry   any                    `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoFeatureCollection representa una colección de Features (GeoJSON).
type GeoFeatureCollection struct {
	Type     string       `json:"type"`
	Features []GeoFeature `json:"features"`
}

// GeoData contiene listas de departamentos, municipios y distritos.
type GeoData struct {
	Departamentos []string `json:"departamentos"`
	Municipios    []string `json:"municipios"`
	Distritos     []string `json:"distritos"`
}

// readTopoJSON lee y decodifica un archivo TopoJSON.
func readTopoJSON(path string) (*TopoJSON, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var topo TopoJSON
	if err := json.Unmarshal(file, &topo); err != nil {
		return nil, err
	}
	if topo.Objects == nil || len(topo.Arcs) == 0 {
		return nil, errors.New("topojson inválido: faltan objects o arcs")
	}
	return &topo, nil
}

// convertArc reconstruye un arco aplicándole la transformación.
// Se verifica que transform.Scale y transform.Translate tengan al menos dos elementos.
func convertArc(arc [][]float64, transform Transform) [][]float64 {
	if len(transform.Scale) < 2 || len(transform.Translate) < 2 {
		return [][]float64{}
	}
	coords := make([][]float64, 0, len(arc))
	var x, y float64
	for _, delta := range arc {
		if len(delta) < 2 {
			continue
		}
		x += delta[0]
		y += delta[1]
		coords = append(coords, []float64{
			transform.Scale[0]*x + transform.Translate[0],
			transform.Scale[1]*y + transform.Translate[1],
		})
	}
	return coords
}

// reverseCoords invierte el orden de las coordenadas.
func reverseCoords(coords [][]float64) [][]float64 {
	for i, j := 0, len(coords)-1; i < j; i, j = i+1, j-1 {
		coords[i], coords[j] = coords[j], coords[i]
	}
	return coords
}

// mergeArcs concatena los arcos usando un slice de índices (enteros).
func mergeArcs(arcIndices []int, topoArcs [][][]float64, transform Transform) [][]float64 {
	var coords [][]float64
	for _, ai := range arcIndices {
		var arcCoords [][]float64
		index := ai
		reverse := false
		if ai < 0 {
			index = -ai - 1
			reverse = true
		}
		if index < 0 || index >= len(topoArcs) {
			continue
		}
		arcCoords = convertArc(topoArcs[index], transform)
		if reverse {
			arcCoords = reverseCoords(arcCoords)
		}
		// Evitar duplicados: descartar el primer punto de cada segmento adicional.
		if len(coords) > 0 && len(arcCoords) > 0 {
			arcCoords = arcCoords[1:]
		}
		coords = append(coords, arcCoords...)
	}
	return coords
}

// convertGeometry convierte una geometría TopoJSON a GeoJSON según su tipo.
func convertGeometry(geom Geometry, topo *TopoJSON) (any, error) {
	switch geom.Type {
	case "Point":
		var coords []float64
		if err := json.Unmarshal(geom.Coordinates, &coords); err != nil {
			return nil, err
		}
		return coords, nil

	case "MultiPoint":
		var coords [][]float64
		if err := json.Unmarshal(geom.Coordinates, &coords); err != nil {
			return nil, err
		}
		return coords, nil

	case "LineString":
		var arcIndices []int
		if err := json.Unmarshal(geom.Arcs, &arcIndices); err != nil {
			return nil, err
		}
		return mergeArcs(arcIndices, topo.Arcs, topo.Transform), nil

	case "Polygon":
		var rings [][]int
		if err := json.Unmarshal(geom.Arcs, &rings); err != nil {
			return nil, err
		}
		var polygon [][][]float64
		for _, ring := range rings {
			line := mergeArcs(ring, topo.Arcs, topo.Transform)
			polygon = append(polygon, line)
		}
		return polygon, nil

	case "MultiLineString":
		var multiLine [][]int
		if err := json.Unmarshal(geom.Arcs, &multiLine); err != nil {
			return nil, err
		}
		var lines [][][]float64
		for _, lineArcs := range multiLine {
			line := mergeArcs(lineArcs, topo.Arcs, topo.Transform)
			lines = append(lines, line)
		}
		return lines, nil

	case "MultiPolygon":
		var multiPoly [][][]int
		if err := json.Unmarshal(geom.Arcs, &multiPoly); err != nil {
			return nil, err
		}
		var polys [][][][]float64
		for _, poly := range multiPoly {
			var polygon [][][]float64
			for _, ring := range poly {
				line := mergeArcs(ring, topo.Arcs, topo.Transform)
				polygon = append(polygon, line)
			}
			polys = append(polys, polygon)
		}
		return polys, nil

	default:
		return nil, errors.New("tipo de geometría no soportada: " + geom.Type)
	}
}

// topoToGeo convierte el objeto TopoJSON en un FeatureCollection de GeoJSON.
func topoToGeo(topo *TopoJSON) (*GeoFeatureCollection, error) {
	collection, ok := topo.Objects["collection"]
	if !ok {
		return nil, errors.New("la clave 'collection' no existe")
	}
	var features []GeoFeature
	for _, geom := range collection.Geometries {
		geoGeom, err := convertGeometry(geom, topo)
		if err != nil {
			// Se omite la geometría si hay error en la conversión.
			continue
		}
		features = append(features, GeoFeature{
			Type:       "Feature",
			Geometry:   geoGeom,
			Properties: geom.Properties,
		})
	}
	return &GeoFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

// mapToSlice convierte las claves de un mapa a un slice de strings.
func mapToSlice(m map[string]struct{}) []string {
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

// GetMunicipios filtra las features por el valor exacto en la propiedad especificada ("D", "M" o "NAM").
func GetMunicipios(query, whatIs string) (*GeoFeatureCollection, error) {
	if whatIs != "D" && whatIs != "M" && whatIs != "NAM" {
		return nil, errors.New("el segundo parámetro debe ser 'M', 'D' o 'NAM'")
	}
	topoPath := filepath.Join(config.AppConfig.AssetsDir, "topo.json")
	topo, err := readTopoJSON(topoPath)
	if err != nil {
		return nil, err
	}
	geo, err := topoToGeo(topo)
	if err != nil {
		return nil, err
	}
	var filteredFeatures []GeoFeature
	for _, feat := range geo.Features {
		if feat.Properties == nil {
			continue
		}
		propVal, ok := feat.Properties[whatIs].(string)
		if !ok {
			continue
		}
		if propVal == query {
			filteredFeatures = append(filteredFeatures, feat)
		}
	}
	return &GeoFeatureCollection{
		Type:     "FeatureCollection",
		Features: filteredFeatures,
	}, nil
}

// GetGeoData extrae nombres únicos de departamentos, municipios y distritos a partir del TopoJSON.
func GetGeoData() (*GeoData, error) {
	topoPath := filepath.Join(config.AppConfig.AssetsDir, "topo.json")
	topo, err := readTopoJSON(topoPath)
	if err != nil {
		return nil, err
	}
	geo, err := topoToGeo(topo)
	if err != nil {
		return nil, err
	}
	departamentos := make(map[string]struct{})
	municipios := make(map[string]struct{})
	distritos := make(map[string]struct{})
	for _, feat := range geo.Features {
		if feat.Properties == nil {
			continue
		}
		if d, ok := feat.Properties["D"].(string); ok && d != "" {
			departamentos[d] = struct{}{}
		}
		if m, ok := feat.Properties["M"].(string); ok && m != "" {
			municipios[m] = struct{}{}
		}
		if nam, ok := feat.Properties["NAM"].(string); ok && nam != "" {
			distritos[nam] = struct{}{}
		}
	}
	return &GeoData{
		Departamentos: mapToSlice(departamentos),
		Municipios:    mapToSlice(municipios),
		Distritos:     mapToSlice(distritos),
	}, nil
}
