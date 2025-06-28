package types

import "encoding/json"

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