package handlers

import (
	"chivomap.com/services"
	"chivomap.com/services/geospatial"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// GeoHandler maneja los endpoints relacionados con datos geoespaciales
type GeoHandler struct {
	geoDataCache *services.CacheService[*geospatial.GeoData]
	municCache   *services.CacheService[map[string]*geospatial.GeoFeatureCollection]
}

// NewGeoHandler crea una nueva instancia de GeoHandler
func NewGeoHandler() *GeoHandler {
	return &GeoHandler{
		geoDataCache: services.NewCacheService[*geospatial.GeoData](60), // 1 hora
		municCache:   services.NewCacheService[map[string]*geospatial.GeoFeatureCollection](60),
	}
}

// GetMunicipios maneja el endpoint para filtrar municipios
// @Summary Filtrar datos geoespaciales
// @Description Filtra datos geoespaciales según parámetros
// @Tags geo
// @Produce json
// @Param query query string true "Cadena de búsqueda"
// @Param whatIs query string true "Tipo de filtro (departamento, municipio, etc.)"
// @Success 200 {object} GeoFilterResponse "Resultados filtrados"
// @Failure 400 {object} ErrorResponse "Parámetros inválidos"
// @Failure 500 {object} ErrorResponse "Error interno"
// @Router /geo/filter [get]
func (h *GeoHandler) GetMunicipios(c *fiber.Ctx) error {
	query := c.Query("query")
	whatIs := c.Query("whatIs")

	if query == "" || whatIs == "" {
		return utils.RespondWithError(c, fiber.StatusBadRequest,
			"Se requieren los parámetros 'query' y 'whatIs'")
	}

	cacheKey := whatIs + ":" + query
	if cached, ok := h.municCache.Get(); ok {
		if data, exists := cached[cacheKey]; exists {
			return utils.SendResponse(c, data)
		}
	}

	data, err := geospatial.GetMunicipios(query, whatIs)
	if err != nil {
		utils.Error("Error al obtener municipios: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError, err.Error())
	}

	cached := make(map[string]*geospatial.GeoFeatureCollection)
	cached[cacheKey] = data
	h.municCache.Set(cached)

	return utils.SendResponse(c, data)
}

// GetGeoData maneja el endpoint para obtener datos geográficos
// @Summary Obtiene datos geográficos
// @Description Retorna datos geográficos completos de El Salvador
// @Tags geo
// @Produce json
// @Success 200 {object} GeoDataResponse "Datos geográficos"
// @Failure 500 {object} ErrorResponse "Error interno"
// @Router /geo/search-data [get]
func (h *GeoHandler) GetGeoData(c *fiber.Ctx) error {
	if data, ok := h.geoDataCache.Get(); ok {
		if h.geoDataCache.NeedsUpdate() {
			go h.updateGeoDataCache()
		}
		return utils.SendResponse(c, data)
	}

	data, err := geospatial.GetGeoData()
	if err != nil {
		utils.Error("Error al obtener geo data: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError,
			"No se pudieron obtener los datos")
	}

	h.geoDataCache.Set(data)
	return utils.SendResponse(c, data)
}

// updateGeoDataCache actualiza la caché de datos geográficos en segundo plano
func (h *GeoHandler) updateGeoDataCache() {
	h.geoDataCache.SetUpdating(true)
	defer h.geoDataCache.SetUpdating(false)

	newData, err := geospatial.GetGeoData()
	if err != nil {
		utils.Error("Error al actualizar caché geo: %v", err)
		return
	}

	h.geoDataCache.Set(newData)
}
