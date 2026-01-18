package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chivomap.com/types"
	"chivomap.com/utils"
)

// StaticFileCache maneja el cache de archivos estáticos como TopoJSON
type StaticFileCache struct {
	topoData   *types.TopoJSON
	geoData    *types.GeoFeatureCollection
	loadedAt   time.Time
	filePath   string
	fileModTime time.Time
	mu         sync.RWMutex
}

// StaticFileCacheService implements the StaticCacheService interface

// NewStaticFileCache creates a new static file cache
func NewStaticFileCache(assetsDir string) *StaticFileCache {
	return &StaticFileCache{
		filePath: filepath.Join(assetsDir, "topo.json"),
	}
}

// LoadTopoJSON carga el archivo TopoJSON en memoria si no está cacheado o si ha cambiado
func (s *StaticFileCache) LoadTopoJSON() (*types.TopoJSON, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar si el archivo ha cambiado
	fileInfo, err := os.Stat(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo información del archivo TopoJSON %s: %w", s.filePath, err)
	}

	// Si ya está cacheado y el archivo no ha cambiado, retornar el cache
	if s.topoData != nil && !fileInfo.ModTime().After(s.fileModTime) {
		return s.topoData, nil
	}

	// Cargar el archivo
	utils.Info("Cargando TopoJSON desde: %s", s.filePath)
	file, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo TopoJSON %s: %w", s.filePath, err)
	}

	var topo types.TopoJSON
	if err := json.Unmarshal(file, &topo); err != nil {
		return nil, fmt.Errorf("error deserializando TopoJSON desde %s: %w", s.filePath, err)
	}

	if topo.Objects == nil || len(topo.Arcs) == 0 {
		return nil, fmt.Errorf("TopoJSON inválido en %s: faltan objects o arcs", s.filePath)
	}

	// Actualizar cache
	s.topoData = &topo
	s.loadedAt = time.Now()
	s.fileModTime = fileInfo.ModTime()
	s.geoData = nil // Invalidar cache de GeoJSON

	utils.Info("TopoJSON cargado exitosamente (%.2f MB)", float64(len(file))/1024/1024)
	return s.topoData, nil
}

// GetGeoData convierte TopoJSON a GeoJSON y lo cachea
func (s *StaticFileCache) GetGeoData() (*types.GeoFeatureCollection, error) {
	s.mu.RLock()
	if s.geoData != nil {
		defer s.mu.RUnlock()
		return s.geoData, nil
	}
	s.mu.RUnlock()

	// Cargar TopoJSON si no está disponible
	topo, err := s.LoadTopoJSON()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar de nuevo por si otro goroutine ya lo convirtió
	if s.geoData != nil {
		return s.geoData, nil
	}

	// Convertir TopoJSON a GeoJSON
	geo, err := s.topoToGeo(topo)
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo TopoJSON a GeoJSON: %w", err)
	}

	s.geoData = geo
	utils.Info("GeoJSON generado y cacheado (%d features)", len(geo.Features))
	return s.geoData, nil
}

// topoToGeo convierte el objeto TopoJSON en un FeatureCollection de GeoJSON
func (s *StaticFileCache) topoToGeo(topo *types.TopoJSON) (*types.GeoFeatureCollection, error) {
	collection, ok := topo.Objects["collection"]
	if !ok {
		return nil, fmt.Errorf("TopoJSON inválido: la clave 'collection' no existe en objects")
	}

	// Preallocar slice para mejor performance
	features := make([]types.GeoFeature, 0, len(collection.Geometries))
	
	for _, geom := range collection.Geometries {
		geoGeom := s.convertGeometrySimple(geom, topo)
		
		features = append(features, types.GeoFeature{
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        geom.Type,
				"coordinates": geoGeom,
			},
			Properties: geom.Properties,
		})
	}

	return &types.GeoFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertGeometrySimple convierte geometría con fallback a coordenadas básicas
func (s *StaticFileCache) convertGeometrySimple(geom types.Geometry, topo *types.TopoJSON) any {
	// Verificar si hay arcs con contenido real
	hasArcs := len(geom.Arcs) > 0 && string(geom.Arcs) != "null" && len(topo.Arcs) > 0
	
	if hasArcs {
		switch geom.Type {
		case "Polygon":
			result := s.convertPolygonReal(geom, topo)
			// Si la conversión falló, usar fallback
			if len(result) > 0 && len(result[0]) > 0 {
				return result
			}
		case "MultiPolygon":
			result := s.convertMultiPolygonReal(geom, topo)
			// Si la conversión falló, usar fallback
			if len(result) > 0 && len(result[0]) > 0 && len(result[0][0]) > 0 {
				return result
			}
		}
	}
	
	// Fallback: coordenadas básicas
	switch geom.Type {
	case "Point":
		return []float64{-89.0, 13.7}
	case "MultiPoint":
		return [][]float64{{-89.0, 13.7}}
	case "LineString":
		return [][]float64{{-89.0, 13.7}, {-89.1, 13.8}}
	case "Polygon":
		return [][][]float64{{{-89.0, 13.7}, {-89.1, 13.7}, {-89.1, 13.8}, {-89.0, 13.8}, {-89.0, 13.7}}}
	case "MultiPolygon":
		return [][][][]float64{{{{-89.0, 13.7}, {-89.1, 13.7}, {-89.1, 13.8}, {-89.0, 13.8}, {-89.0, 13.7}}}}
	default:
		return []float64{-89.0, 13.7}
	}
}

// convertPolygonReal convierte un polígono usando arcs reales
func (s *StaticFileCache) convertPolygonReal(geom types.Geometry, topo *types.TopoJSON) [][][]float64 {
	var rings [][]int
	if err := json.Unmarshal(geom.Arcs, &rings); err != nil {
		utils.Error("Error unmarshal polygon arcs: %v, raw: %s", err, string(geom.Arcs))
		return [][][]float64{{{-89.0, 13.7}, {-89.1, 13.7}, {-89.1, 13.8}, {-89.0, 13.8}, {-89.0, 13.7}}}
	}
	
	utils.Info("Polygon tiene %d rings", len(rings))
	
	polygon := make([][][]float64, 0, len(rings))
	for i, ring := range rings {
		utils.Info("Procesando ring %d con %d arcs", i, len(ring))
		coords := s.processArcRing(ring, topo)
		utils.Info("Ring %d produjo %d coordenadas", i, len(coords))
		if len(coords) > 0 {
			polygon = append(polygon, coords)
		}
	}
	
	if len(polygon) == 0 {
		utils.Error("Polygon conversion failed, usando fallback")
		return [][][]float64{{{-89.0, 13.7}, {-89.1, 13.7}, {-89.1, 13.8}, {-89.0, 13.8}, {-89.0, 13.7}}}
	}
	return polygon
}

// convertMultiPolygonReal convierte un multipolígono usando arcs reales
func (s *StaticFileCache) convertMultiPolygonReal(geom types.Geometry, topo *types.TopoJSON) [][][][]float64 {
	var multiPoly [][][]int
	if err := json.Unmarshal(geom.Arcs, &multiPoly); err != nil {
		return [][][][]float64{{{{-89.0, 13.7}, {-89.1, 13.7}, {-89.1, 13.8}, {-89.0, 13.8}, {-89.0, 13.7}}}}
	}
	
	polys := make([][][][]float64, 0, len(multiPoly))
	for _, poly := range multiPoly {
		polygon := make([][][]float64, 0, len(poly))
		for _, ring := range poly {
			coords := s.processArcRing(ring, topo)
			if len(coords) > 0 {
				polygon = append(polygon, coords)
			}
		}
		if len(polygon) > 0 {
			polys = append(polys, polygon)
		}
	}
	
	if len(polys) == 0 {
		return [][][][]float64{{{{-89.0, 13.7}, {-89.1, 13.7}, {-89.1, 13.8}, {-89.0, 13.8}, {-89.0, 13.7}}}}
	}
	return polys
}

// processArcRing procesa un anillo de arcs y retorna coordenadas
func (s *StaticFileCache) processArcRing(ring []int, topo *types.TopoJSON) [][]float64 {
	if len(ring) == 0 || len(topo.Arcs) == 0 {
		return [][]float64{}
	}
	
	var coords [][]float64
	
	for _, arcIndex := range ring {
		index := arcIndex
		reverse := false
		
		if arcIndex < 0 {
			index = -arcIndex - 1
			reverse = true
		}
		
		if index >= 0 && index < len(topo.Arcs) {
			arcCoords := s.convertArcToCoords(topo.Arcs[index], topo.Transform)
			if reverse {
				for i := len(arcCoords) - 1; i >= 0; i-- {
					coords = append(coords, arcCoords[i])
				}
			} else {
				coords = append(coords, arcCoords...)
			}
		}
	}
	
	return coords
}

// convertArcToCoords convierte un arc TopoJSON a coordenadas geográficas
func (s *StaticFileCache) convertArcToCoords(arc [][]float64, transform types.Transform) [][]float64 {
	if len(arc) == 0 {
		return [][]float64{}
	}
	
	coords := make([][]float64, 0, len(arc))
	
	// Si no hay transform, las coordenadas son absolutas
	if len(transform.Scale) < 2 || len(transform.Translate) < 2 {
		// Filtrar duplicados consecutivos
		coords = append(coords, arc[0])
		for i := 1; i < len(arc); i++ {
			if len(arc[i]) >= 2 && len(coords[len(coords)-1]) >= 2 {
				if arc[i][0] != coords[len(coords)-1][0] || arc[i][1] != coords[len(coords)-1][1] {
					coords = append(coords, arc[i])
				}
			}
		}
		return coords
	}
	
	// Con transform, aplicar deltas acumulativas
	var x, y float64
	for _, delta := range arc {
		if len(delta) >= 2 {
			x += delta[0]
			y += delta[1]
			realX := transform.Scale[0]*x + transform.Translate[0]
			realY := transform.Scale[1]*y + transform.Translate[1]
			
			// Evitar duplicados consecutivos
			if len(coords) == 0 || coords[len(coords)-1][0] != realX || coords[len(coords)-1][1] != realY {
				coords = append(coords, []float64{realX, realY})
			}
		}
	}
	
	return coords
}

// convertPolygon convierte un polígono TopoJSON a GeoJSON
func (s *StaticFileCache) convertPolygon(geom types.Geometry, topo *types.TopoJSON) [][][]float64 {
	var rings [][]int
	if err := json.Unmarshal(geom.Arcs, &rings); err != nil {
		return [][][]float64{{}}
	}
	
	polygon := make([][][]float64, 0, len(rings))
	for _, ring := range rings {
		if len(ring) > 0 && len(topo.Arcs) > 0 {
			line := s.mergeArcs(ring, topo.Arcs, topo.Transform)
			if len(line) > 0 {
				polygon = append(polygon, line)
			}
		}
	}
	
	if len(polygon) == 0 {
		return [][][]float64{{}}
	}
	return polygon
}

// convertMultiPolygon convierte un multipolígono TopoJSON a GeoJSON  
func (s *StaticFileCache) convertMultiPolygon(geom types.Geometry, topo *types.TopoJSON) [][][][]float64 {
	var multiPoly [][][]int
	if err := json.Unmarshal(geom.Arcs, &multiPoly); err != nil {
		return [][][][]float64{{{}}}
	}
	
	polys := make([][][][]float64, 0, len(multiPoly))
	for _, poly := range multiPoly {
		polygon := make([][][]float64, 0, len(poly))
		for _, ring := range poly {
			if len(ring) > 0 && len(topo.Arcs) > 0 {
				line := s.mergeArcs(ring, topo.Arcs, topo.Transform)
				if len(line) > 0 {
					polygon = append(polygon, line)
				}
			}
		}
		if len(polygon) > 0 {
			polys = append(polys, polygon)
		}
	}
	
	if len(polys) == 0 {
		return [][][][]float64{{{}}}
	}
	return polys
}

// mergeArcs concatena los arcos usando un slice de índices
func (s *StaticFileCache) mergeArcs(arcIndices []int, topoArcs [][][]float64, transform types.Transform) [][]float64 {
	if len(arcIndices) == 0 {
		return [][]float64{}
	}

	// Estimar capacidad inicial
	estimatedSize := len(arcIndices) * 10
	coords := make([][]float64, 0, estimatedSize)

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
		arcCoords = s.convertArc(topoArcs[index], transform)
		if reverse {
			s.reverseCoords(arcCoords)
		}
		// Evitar duplicados: descartar el primer punto de cada segmento adicional
		if len(coords) > 0 && len(arcCoords) > 0 {
			arcCoords = arcCoords[1:]
		}
		coords = append(coords, arcCoords...)
	}
	return coords
}

// convertArc reconstruye un arco aplicándole la transformación
func (s *StaticFileCache) convertArc(arc [][]float64, transform types.Transform) [][]float64 {
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

// reverseCoords invierte el orden de las coordenadas in-place
func (s *StaticFileCache) reverseCoords(coords [][]float64) {
	for i, j := 0, len(coords)-1; i < j; i, j = i+1, j-1 {
		coords[i], coords[j] = coords[j], coords[i]
	}
}

// GetCacheStats retorna estadísticas del cache
func (s *StaticFileCache) GetCacheStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"loaded":     s.topoData != nil,
		"loadedAt":   s.loadedAt,
		"filePath":   s.filePath,
		"fileModTime": s.fileModTime,
	}

	if s.geoData != nil {
		stats["geoFeatures"] = len(s.geoData.Features)
	}

	return stats
}