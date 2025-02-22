package geospatial

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	geojson "github.com/paulmach/go.geojson"
)

// TopoJSON representa la estructura de un archivo TopoJSON.
type TopoJSON struct {
	Type    string                `json:"type"`
	Objects map[string]TopoObject `json:"objects"`
	Arcs    [][][]float64         `json:"arcs"`
}

// TopoObject representa un objeto TopoJSON con geometrías.
type TopoObject struct {
	Type       string     `json:"type"`
	Geometries []Geometry `json:"geometries"`
}

// Geometry representa la geometría de un objeto TopoJSON.
type Geometry struct {
	Type string `json:"type"`
	// Se asume que cada referencia de arco es un slice con un único entero.
	Arcs       [][][]int         `json:"arcs"`
	Properties map[string]string `json:"properties"`
}

// GeoData contiene la información agrupada de departamentos, municipios y distritos.
type GeoData struct {
	Departamentos []string
	Municipios    []string
	Distritos     []string
}

// reverseCoordinates invierte el orden de las coordenadas en un slice.
func reverseCoordinates(coords [][]float64) [][]float64 {
	n := len(coords)
	reversed := make([][]float64, n)
	for i, coord := range coords {
		reversed[n-1-i] = coord
	}
	return reversed
}

// convertToGeoJSON transforma un TopoJSON a una colección de features GeoJSON.
// Devuelve error si la clave "collection" no existe.
func convertToGeoJSON(topo *TopoJSON) (*geojson.FeatureCollection, error) {
	collection, ok := topo.Objects["collection"]
	if !ok {
		return nil, errors.New("la clave 'collection' no existe en topo.Objects")
	}

	features := make([]*geojson.Feature, 0, len(collection.Geometries))
	for _, geom := range collection.Geometries {
		coords := convertArcsToCoordinates(geom.Arcs, topo.Arcs)
		feature := geojson.NewFeature(coords)
		feature.Properties = make(map[string]interface{})
		for k, v := range geom.Properties {
			feature.Properties[k] = v
		}
		features = append(features, feature)
	}
	return &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

// convertArcsToCoordinates convierte las referencias de arcos a coordenadas reales,
// invirtiendo el orden en caso de índices negativos.
func convertArcsToCoordinates(arcs [][][]int, topoArcs [][][]float64) *geojson.Geometry {
	if len(arcs) == 0 {
		return &geojson.Geometry{
			Type:    "Polygon",
			Polygon: [][][]float64{},
		}
	}

	coords := make([][][]float64, 0, len(arcs))

	// Procesar cada anillo del polígono
	for _, ring := range arcs {
		ringCoords := make([][]float64, 0)

		// Procesar cada arco en el anillo
		for _, arcIndices := range ring {
			for _, idx := range arcIndices {
				// Manejar índices negativos
				actualIdx := idx
				if idx < 0 {
					actualIdx = len(topoArcs) + idx
				}

				if actualIdx >= 0 && actualIdx < len(topoArcs) {
					// Añadir todas las coordenadas del arco
					for _, coord := range topoArcs[actualIdx] {
						ringCoords = append(ringCoords, coord)
					}
				}
			}
		}

		// Solo añadir el anillo si tiene coordenadas
		if len(ringCoords) > 0 {
			coords = append(coords, ringCoords)
		}
	}

	return &geojson.Geometry{
		Type:    "Polygon",
		Polygon: coords,
	}
}

// mapToSlice convierte un mapa de cadenas en un slice.
func mapToSlice(m map[string]struct{}) []string {
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

// GetMunicipios filtra y devuelve las features cuyo valor de propiedad coincide parcialmente con el query.
// El parámetro whatIs debe ser "D", "M" o "NAM".
func GetMunicipios(query, whatIs string) (*geojson.FeatureCollection, error) {
	if whatIs != "D" && whatIs != "M" && whatIs != "NAM" {
		return nil, errors.New("el segundo parámetro debe ser 'M', 'D' o 'NAM'")
	}

	topo, err := readTopoJSON("./utils/assets/topo.json")
	if err != nil {
		return nil, err
	}

	geojsonFC, err := convertToGeoJSON(topo)
	if err != nil {
		return nil, err
	}

	filteredFeatures := make([]*geojson.Feature, 0)
	for _, feature := range geojsonFC.Features {
		if feature.Properties == nil {
			continue
		}

		propVal, ok := feature.Properties[whatIs].(string)
		if !ok {
			continue
		}

		// Comparación en minúsculas para permitir coincidencias parciales.
		if strings.Contains(strings.ToLower(propVal), strings.ToLower(query)) {
			filteredFeatures = append(filteredFeatures, feature)
		}
	}

	return &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: filteredFeatures,
	}, nil
}

// GetGeoData extrae la información de departamentos, municipios y distritos
// de las propiedades de las features.
func GetGeoData() (*GeoData, error) {
	topo, err := readTopoJSON("./utils/assets/topo.json")
	if err != nil {
		return nil, err
	}

	geojsonFC, err := convertToGeoJSON(topo)
	if err != nil {
		return nil, err
	}

	departamentos := make(map[string]struct{})
	municipios := make(map[string]struct{})
	distritos := make(map[string]struct{})

	for _, feature := range geojsonFC.Features {
		if feature.Properties == nil {
			continue
		}

		if d, ok := feature.Properties["D"].(string); ok && d != "" {
			departamentos[d] = struct{}{}
		}
		if m, ok := feature.Properties["M"].(string); ok && m != "" {
			municipios[m] = struct{}{}
		}
		if nam, ok := feature.Properties["NAM"].(string); ok && nam != "" {
			distritos[nam] = struct{}{}
		}
	}

	return &GeoData{
		Departamentos: mapToSlice(departamentos),
		Municipios:    mapToSlice(municipios),
		Distritos:     mapToSlice(distritos),
	}, nil
}

// readTopoJSON lee y deserializa un archivo TopoJSON desde la ruta especificada.
func readTopoJSON(path string) (*TopoJSON, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var topo TopoJSON
	if err := json.Unmarshal(file, &topo); err != nil {
		return nil, err
	}

	return &topo, nil
}
