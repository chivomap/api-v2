package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chivomap.com/config"
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

var (
	// Instancia global del cache de archivos estáticos
	staticCache *StaticFileCache
	cacheOnce   sync.Once
)

// GetStaticCache retorna la instancia singleton del cache estático
func GetStaticCache() *StaticFileCache {
	cacheOnce.Do(func() {
		staticCache = &StaticFileCache{
			filePath: filepath.Join(config.AppConfig.AssetsDir, "topo.json"),
		}
	})
	return staticCache
}

// LoadTopoJSON carga el archivo TopoJSON en memoria si no está cacheado o si ha cambiado
func (s *StaticFileCache) LoadTopoJSON() (*types.TopoJSON, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar si el archivo ha cambiado
	fileInfo, err := os.Stat(s.filePath)
	if err != nil {
		return nil, err
	}

	// Si ya está cacheado y el archivo no ha cambiado, retornar el cache
	if s.topoData != nil && !fileInfo.ModTime().After(s.fileModTime) {
		return s.topoData, nil
	}

	// Cargar el archivo
	utils.Info("Cargando TopoJSON desde: %s", s.filePath)
	file, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var topo types.TopoJSON
	if err := json.Unmarshal(file, &topo); err != nil {
		return nil, err
	}

	if topo.Objects == nil || len(topo.Arcs) == 0 {
		return nil, errors.New("topojson inválido: faltan objects o arcs")
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
		return nil, err
	}

	s.geoData = geo
	utils.Info("GeoJSON generado y cacheado (%d features)", len(geo.Features))
	return s.geoData, nil
}

// topoToGeo convierte el objeto TopoJSON en un FeatureCollection de GeoJSON
func (s *StaticFileCache) topoToGeo(topo *types.TopoJSON) (*types.GeoFeatureCollection, error) {
	collection, ok := topo.Objects["collection"]
	if !ok {
		return nil, errors.New("la clave 'collection' no existe")
	}

	// Preallocar slice para mejor performance
	features := make([]types.GeoFeature, 0, len(collection.Geometries))
	
	for _, geom := range collection.Geometries {
		geoGeom, err := s.convertGeometry(geom, topo)
		if err != nil {
			// Se omite la geometría si hay error en la conversión
			continue
		}
		features = append(features, types.GeoFeature{
			Type:       "Feature",
			Geometry:   geoGeom,
			Properties: geom.Properties,
		})
	}

	return &types.GeoFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

// convertGeometry convierte una geometría TopoJSON a GeoJSON según su tipo
func (s *StaticFileCache) convertGeometry(geom types.Geometry, topo *types.TopoJSON) (any, error) {
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
		return s.mergeArcs(arcIndices, topo.Arcs, topo.Transform), nil

	case "Polygon":
		var rings [][]int
		if err := json.Unmarshal(geom.Arcs, &rings); err != nil {
			return nil, err
		}
		polygon := make([][][]float64, 0, len(rings))
		for _, ring := range rings {
			line := s.mergeArcs(ring, topo.Arcs, topo.Transform)
			polygon = append(polygon, line)
		}
		return polygon, nil

	case "MultiLineString":
		var multiLine [][]int
		if err := json.Unmarshal(geom.Arcs, &multiLine); err != nil {
			return nil, err
		}
		lines := make([][][]float64, 0, len(multiLine))
		for _, lineArcs := range multiLine {
			line := s.mergeArcs(lineArcs, topo.Arcs, topo.Transform)
			lines = append(lines, line)
		}
		return lines, nil

	case "MultiPolygon":
		var multiPoly [][][]int
		if err := json.Unmarshal(geom.Arcs, &multiPoly); err != nil {
			return nil, err
		}
		polys := make([][][][]float64, 0, len(multiPoly))
		for _, poly := range multiPoly {
			polygon := make([][][]float64, 0, len(poly))
			for _, ring := range poly {
				line := s.mergeArcs(ring, topo.Arcs, topo.Transform)
				polygon = append(polygon, line)
			}
			polys = append(polys, polygon)
		}
		return polys, nil

	default:
		return nil, errors.New("tipo de geometría no soportada: " + geom.Type)
	}
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