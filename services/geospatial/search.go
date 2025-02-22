package geospatial

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	geojson "github.com/paulmach/go.geojson"
)

type TopoJSON struct {
	Type    string                `json:"type"`
	Objects map[string]TopoObject `json:"objects"`
	Arcs    [][][]float64         `json:"arcs"`
}

type TopoObject struct {
	Type       string     `json:"type"`
	Geometries []Geometry `json:"geometries"`
}

type Geometry struct {
	Type       string            `json:"type"`
	Arcs       [][][]int         `json:"arcs"`
	Properties map[string]string `json:"properties"`
}

type GeoData struct {
	Departamentos []string
	Municipios    []string
	Distritos     []string
}

func convertToGeoJSON(topo *TopoJSON) *geojson.FeatureCollection {
	features := make([]*geojson.Feature, 0)

	for _, geom := range topo.Objects["collection"].Geometries {
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
	}
}

func convertArcsToCoordinates(arcs [][][]int, topoArcs [][][]float64) *geojson.Geometry {
	// Validación de entrada
	if len(arcs) == 0 {
		return &geojson.Geometry{
			Type:    "Polygon",
			Polygon: [][][]float64{},
		}
	}

	coords := make([][][]float64, 0, len(arcs[0]))

	for _, arcRing := range arcs {
		ringCoords := make([][]float64, 0)

		for _, arcIndex := range arcRing {
			// Validar que el índice está en rango
			if len(arcIndex) == 0 {
				continue
			}

			index := arcIndex[0]
			// Convertir índices negativos a positivos si es necesario
			if index < 0 {
				index = len(topoArcs) + index
			}

			// Validar que el índice está en rango
			if index >= 0 && index < len(topoArcs) {
				lineCoords := topoArcs[index]
				ringCoords = append(ringCoords, lineCoords...)
			}
		}

		if len(ringCoords) > 0 {
			coords = append(coords, ringCoords)
		}
	}

	return &geojson.Geometry{
		Type:    "Polygon",
		Polygon: coords,
	}
}

func mapToSlice(m map[string]struct{}) []string {
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

func GetMunicipios(query, whatIs string) (*geojson.FeatureCollection, error) {
	if whatIs != "D" && whatIs != "M" && whatIs != "NAM" {
		return nil, errors.New("el segundo parámetro debe ser 'M', 'D' o 'NAM'")
	}

	data, err := readTopoJSON("./utils/assets/topo.json")
	if err != nil {
		return nil, err
	}

	geojsonFC := convertToGeoJSON(data)
	features := make([]*geojson.Feature, 0)

	for _, feature := range geojsonFC.Features {
		if feature.Properties == nil {
			continue
		}

		propertyValue, ok := feature.Properties[whatIs].(string)
		if !ok {
			continue
		}

		// Convertir ambos a minúsculas y buscar coincidencia parcial
		propertyLower := strings.ToLower(propertyValue)
		queryLower := strings.ToLower(query)

		if strings.Contains(propertyLower, queryLower) {
			features = append(features, feature)
		}
	}

	return &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

func GetGeoData() (*GeoData, error) {
	data, err := readTopoJSON("./utils/assets/topo.json")
	if err != nil {
		return nil, err
	}

	geojsonFC := convertToGeoJSON(data)

	departamentos := make(map[string]struct{})
	municipios := make(map[string]struct{})
	distritos := make(map[string]struct{})

	for _, feature := range geojsonFC.Features {
		// Validar que las propiedades existan
		if feature.Properties == nil {
			continue
		}

		// Validar y convertir cada propiedad de forma segura
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
