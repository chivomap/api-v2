package geospatial

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type Transform struct {
	Scale     []float64 `json:"scale"`
	Translate []float64 `json:"translate"`
}

type TopoJSON struct {
	Type      string                `json:"type"`
	Objects   map[string]TopoObject `json:"objects"`
	Arcs      [][][]float64         `json:"arcs"`
	Transform Transform             `json:"transform"`
}

type TopoObject struct {
	Type       string     `json:"type"`
	Geometries []Geometry `json:"geometries"`
}

type Geometry struct {
	Type       string         `json:"type"`
	Arcs       [][][]int      `json:"arcs"`
	Properties map[string]any `json:"properties"`
}

type GeoData struct {
	Departamentos []string `json:"departamentos"`
	Municipios    []string `json:"municipios"`
	Distritos     []string `json:"distritos"`
}

func mapToSlice(m map[string]struct{}) []string {
	if len(m) == 0 {
		return []string{}
	}
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

func GetMunicipios(query, whatIs string) (*TopoJSON, error) {
	if whatIs != "D" && whatIs != "M" && whatIs != "NAM" {
		return nil, errors.New("el segundo parámetro debe ser 'M', 'D' o 'NAM'")
	}

	topo, err := readTopoJSON("./utils/assets/topo.json")
	if err != nil {
		return nil, err
	}

	collection, ok := topo.Objects["collection"]
	if !ok {
		return nil, errors.New("la clave 'collection' no existe")
	}

	filteredGeometries := make([]Geometry, 0)
	for _, geom := range collection.Geometries {
		if geom.Properties == nil {
			continue
		}
		propVal, ok := geom.Properties[whatIs].(string)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(propVal), strings.ToLower(query)) {
			filteredGeometries = append(filteredGeometries, geom)
		}
	}

	// Crear nuevo TopoJSON con las geometrías filtradas
	filteredTopo := &TopoJSON{
		Type: topo.Type,
		Objects: map[string]TopoObject{
			"collection": {
				Type:       collection.Type,
				Geometries: filteredGeometries,
			},
		},
		Arcs:      topo.Arcs,
		Transform: topo.Transform,
	}

	return filteredTopo, nil
}

func GetGeoData() (*GeoData, error) {
	topo, err := readTopoJSON("./utils/assets/topo.json")
	if err != nil {
		return nil, err
	}

	collection, ok := topo.Objects["collection"]
	if !ok {
		return nil, errors.New("la clave 'collection' no existe")
	}

	departamentos := make(map[string]struct{})
	municipios := make(map[string]struct{})
	distritos := make(map[string]struct{})

	for _, geom := range collection.Geometries {
		if geom.Properties == nil {
			continue
		}
		if d, ok := geom.Properties["D"].(string); ok && d != "" {
			departamentos[d] = struct{}{}
		}
		if m, ok := geom.Properties["M"].(string); ok && m != "" {
			municipios[m] = struct{}{}
		}
		if nam, ok := geom.Properties["NAM"].(string); ok && nam != "" {
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

	if topo.Objects == nil || len(topo.Arcs) == 0 {
		return nil, errors.New("topojson inválido: faltan objects o arcs")
	}

	return &topo, nil
}
