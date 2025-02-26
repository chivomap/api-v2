package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// --- Definición de estructuras TopoJSON ---

// Topology representa el objeto TopoJSON completo.
type Topology struct {
	Type      string                `json:"type"`
	Transform *Transform            `json:"transform,omitempty"`
	Arcs      [][][]float64         `json:"arcs"`
	Objects   map[string]TopoObject `json:"objects"`
}

// Transform define la transformación para convertir coordenadas enteras a reales.
type Transform struct {
	Scale     []float64 `json:"scale"`
	Translate []float64 `json:"translate"`
}

// TopoObject representa un objeto (usualmente una colección de geometrías).
type TopoObject struct {
	Type       string         `json:"type"`
	Geometries []TopoGeometry `json:"geometries"`
	// Se pueden incluir propiedades a nivel del objeto (opcional)
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TopoGeometry representa cada geometría contenida en un TopoObject.
type TopoGeometry struct {
	Type string `json:"type"`
	// El campo "arcs" se utiliza para geometrías que referencian arcos (Polygon, LineString, etc.)
	Arcs json.RawMessage `json:"arcs,omitempty"`
	// Para Point o MultiPoint se usa "coordinates"
	Coordinates json.RawMessage        `json:"coordinates,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// --- Estructuras GeoJSON ---

// GeoJSONGeometry representa la geometría en formato GeoJSON.
type GeoJSONGeometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

// Feature es una entidad GeoJSON con geometría y propiedades.
type Feature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// FeatureCollection es el conjunto de features en formato GeoJSON.
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

// --- Funciones de conversión ---

// applyTransform aplica la transformación (scale y translate) a un punto.
func applyTransform(point []float64, t *Transform) []float64 {
	if t == nil {
		return point
	}
	return []float64{
		point[0]*t.Scale[0] + t.Translate[0],
		point[1]*t.Scale[1] + t.Translate[1],
	}
}

// getArcCoordinates calcula las coordenadas de un arco.
// Si existe la transformación, asume que los puntos están delta‑codificados.
// Si no existe, se usan directamente.
func getArcCoordinates(index int, topo *Topology) ([][]float64, error) {
	reverse := false
	if index < 0 {
		reverse = true
		index = -index - 1
	}
	if index < 0 || index >= len(topo.Arcs) {
		return nil, errors.New("arc index out of range")
	}
	arc := topo.Arcs[index]
	coords := make([][]float64, len(arc))

	// Si existe transform, acumulamos los deltas.
	if topo.Transform != nil {
		cum := []float64{0, 0}
		for i, delta := range arc {
			cum[0] += delta[0]
			cum[1] += delta[1]
			point := []float64{cum[0], cum[1]}
			point = applyTransform(point, topo.Transform)
			coords[i] = point
		}
	} else {
		// Si no hay transform, se asume que cada punto ya es absoluto.
		for i, pt := range arc {
			// Se hace una copia del punto
			coords[i] = []float64{pt[0], pt[1]}
		}
	}

	if reverse {
		// Invertir el orden de las coordenadas si el índice era negativo.
		for i, j := 0, len(coords)-1; i < j; i, j = i+1, j-1 {
			coords[i], coords[j] = coords[j], coords[i]
		}
	}
	return coords, nil
}

// convertArcsToLine convierte un slice de índices de arcos en una línea (secuencia de coordenadas).
func convertArcsToLine(arcIndices []int, topo *Topology) ([][2]float64, error) {
	var line [][2]float64
	for i, arcIndex := range arcIndices {
		coords, err := getArcCoordinates(arcIndex, topo)
		if err != nil {
			return nil, err
		}
		// A partir del segundo arco, se elimina el primer punto para evitar duplicados.
		if i > 0 && len(coords) > 0 {
			coords = coords[1:]
		}
		for _, pt := range coords {
			if len(pt) < 2 {
				continue
			}
			line = append(line, [2]float64{pt[0], pt[1]})
		}
	}
	return line, nil
}

// convertTopoGeometry convierte una TopoGeometry a una GeoJSONGeometry.
// Se implementan los casos para Polygon, MultiPolygon, LineString, MultiLineString,
// Point y MultiPoint.
func convertTopoGeometry(tg TopoGeometry, topo *Topology) (GeoJSONGeometry, error) {
	var g GeoJSONGeometry
	g.Type = tg.Type
	switch tg.Type {
	case "Polygon":
		// Se espera que tg.Arcs sea un arreglo de anillos: [][]int
		var rings [][]int
		if err := json.Unmarshal(tg.Arcs, &rings); err != nil {
			return g, err
		}
		var polygon []([][2]float64)
		for _, ring := range rings {
			line, err := convertArcsToLine(ring, topo)
			if err != nil {
				return g, err
			}
			polygon = append(polygon, line)
		}
		g.Coordinates = polygon
		return g, nil
	case "MultiPolygon":
		// Se espera que tg.Arcs sea un arreglo de polígonos: [][][]int
		var multipolygon [][][]int
		if err := json.Unmarshal(tg.Arcs, &multipolygon); err != nil {
			return g, err
		}
		var multiPoly []([][][2]float64)
		for _, polygonIndices := range multipolygon {
			var polygon []([][2]float64)
			for _, ring := range polygonIndices {
				line, err := convertArcsToLine(ring, topo)
				if err != nil {
					return g, err
				}
				polygon = append(polygon, line)
			}
			multiPoly = append(multiPoly, polygon)
		}
		g.Coordinates = multiPoly
		return g, nil
	case "LineString":
		// Se espera que tg.Arcs sea un arreglo de enteros: []int
		var arcs []int
		if err := json.Unmarshal(tg.Arcs, &arcs); err != nil {
			return g, err
		}
		line, err := convertArcsToLine(arcs, topo)
		if err != nil {
			return g, err
		}
		g.Coordinates = line
		return g, nil
	case "MultiLineString":
		// Se espera que tg.Arcs sea un arreglo de arreglos de enteros: [][]int
		var multiArcs [][]int
		if err := json.Unmarshal(tg.Arcs, &multiArcs); err != nil {
			return g, err
		}
		var multiLine []([][2]float64)
		for _, arcs := range multiArcs {
			line, err := convertArcsToLine(arcs, topo)
			if err != nil {
				return g, err
			}
			multiLine = append(multiLine, line)
		}
		g.Coordinates = multiLine
		return g, nil
	case "Point":
		// Para Point se usan las coordenadas directamente
		var pt [2]float64
		if err := json.Unmarshal(tg.Coordinates, &pt); err != nil {
			return g, err
		}
		g.Coordinates = pt
		return g, nil
	case "MultiPoint":
		var pts [][2]float64
		if err := json.Unmarshal(tg.Coordinates, &pts); err != nil {
			return g, err
		}
		g.Coordinates = pts
		return g, nil
	default:
		return g, fmt.Errorf("unsupported geometry type: %s", tg.Type)
	}
}

// convertTopoObject convierte un TopoObject (se espera que sea una GeometryCollection)
// a un FeatureCollection en formato GeoJSON.
func convertTopoObject(obj TopoObject, topo *Topology) (FeatureCollection, error) {
	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{},
	}
	for _, tg := range obj.Geometries {
		geoGeom, err := convertTopoGeometry(tg, topo)
		if err != nil {
			return fc, err
		}
		feat := Feature{
			Type:       "Feature",
			Geometry:   geoGeom,
			Properties: tg.Properties,
		}
		fc.Features = append(fc.Features, feat)
	}
	return fc, nil
}

// FilterFeatureCollection filtra los features de un FeatureCollection,
// conservando aquellos en los que la propiedad (por ejemplo, "D") tenga el valor query.
func FilterFeatureCollection(fc FeatureCollection, query, whatIs string) FeatureCollection {
	filtered := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{},
	}
	for _, feat := range fc.Features {
		if val, ok := feat.Properties[whatIs]; ok {
			if strVal, ok := val.(string); ok && strVal == query {
				filtered.Features = append(filtered.Features, feat)
			}
		}
	}
	return filtered
}

// --- Función principal ---

func main() {
	// Leer el archivo TopoJSON
	data, err := ioutil.ReadFile("../../utils/assets/topo.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al leer el archivo: %v\n", err)
		os.Exit(1)
	}

	// Decodificar el JSON en la estructura Topology
	var topo Topology
	if err := json.Unmarshal(data, &topo); err != nil {
		fmt.Fprintf(os.Stderr, "Error al parsear el TopoJSON: %v\n", err)
		os.Exit(1)
	}

	// Verificar que exista el objeto "collection"
	obj, ok := topo.Objects["collection"]
	if !ok {
		fmt.Fprintln(os.Stderr, "No se encontró el objeto 'collection' en el TopoJSON")
		os.Exit(1)
	}

	// Convertir el objeto TopoJSON a un FeatureCollection GeoJSON
	fc, err := convertTopoObject(obj, &topo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al convertir TopoJSON a GeoJSON: %v\n", err)
		os.Exit(1)
	}

	// Filtrar los features según la propiedad deseada
	// Por ejemplo, filtrar donde la propiedad "D" sea igual a "nombre_del_departamento"
	query := "Nahuizalco" // Reemplaza por el valor buscado
	whatIs := "NAM"       // Puede ser "D", "M" o "NAM"
	filtered := FilterFeatureCollection(fc, query, whatIs)

	// Serializar el FeatureCollection filtrado a JSON
	out, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al serializar GeoJSON: %v\n", err)
		os.Exit(1)
	}

	// Guardar el JSON en un archivo
	outputFile := "./filtered_geojson.json"
	if err := ioutil.WriteFile(outputFile, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error al escribir el archivo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("GeoJSON filtrado guardado en: %s\n", outputFile)
}
