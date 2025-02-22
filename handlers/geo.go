package handlers

import (
	"log"
	"time"

	"chivomap.com/services"
	"chivomap.com/services/geospatial"
	"github.com/gofiber/fiber/v2"
	geojson "github.com/paulmach/go.geojson"
)

type GeoHandler struct {
	geoDataCache *services.CacheService[*geospatial.GeoData]
	municCache   *services.CacheService[map[string]*geojson.FeatureCollection]
}

func NewGeoHandler() *GeoHandler {
	return &GeoHandler{
		geoDataCache: services.NewCacheService[*geospatial.GeoData](60), // 1 hora
		municCache:   services.NewCacheService[map[string]*geojson.FeatureCollection](60),
	}
}

func (h *GeoHandler) GetMunicipios(c *fiber.Ctx) error {
	query := c.Query("query")
	whatIs := c.Query("whatIs")

	if query == "" || whatIs == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Se requieren los parámetros 'query' y 'whatIs'",
		})
	}

	cacheKey := whatIs + ":" + query

	if cached, ok := h.municCache.Get(); ok {
		if data, exists := cached[cacheKey]; exists {
			return c.JSON(fiber.Map{
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
				"data":      data,
			})
		}
	}

	data, err := geospatial.GetMunicipios(query, whatIs)
	if err != nil {
		log.Printf("❌ Error al obtener municipios: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	cached := make(map[string]*geojson.FeatureCollection)
	cached[cacheKey] = data
	h.municCache.Set(cached)

	return c.JSON(fiber.Map{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"data":      data,
	})
}

func (h *GeoHandler) GetGeoData(c *fiber.Ctx) error {
	if data, ok := h.geoDataCache.Get(); ok {
		if h.geoDataCache.NeedsUpdate() {
			go h.updateGeoDataCache()
		}
		return c.JSON(fiber.Map{
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"data":      data,
		})
	}

	data, err := geospatial.GetGeoData()
	if err != nil {
		log.Println("❌ Error al obtener geo data:", err)
		return c.Status(500).JSON(fiber.Map{"error": "No se pudieron obtener los datos"})
	}

	h.geoDataCache.Set(data)
	return c.JSON(fiber.Map{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"data":      data,
	})
}

func (h *GeoHandler) updateGeoDataCache() {
	h.geoDataCache.SetUpdating(true)
	defer h.geoDataCache.SetUpdating(false)

	newData, err := geospatial.GetGeoData()
	if err != nil {
		log.Println("❌ Error al actualizar caché geo:", err)
		return
	}

	h.geoDataCache.Set(newData)
}

func (h *GeoHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "UP"})
}
