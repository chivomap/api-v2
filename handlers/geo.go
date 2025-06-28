package handlers

import (
	"sync"
	
	"chivomap.com/services"
	"chivomap.com/services/geospatial"
	"chivomap.com/types"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// GeoHandler maneja los endpoints relacionados con datos geoespaciales
type GeoHandler struct {
	geoDataCache *services.CacheService[*types.GeoData]
	municCache   *services.CacheService[map[string]*types.GeoFeatureCollection]
	cacheMutex   sync.RWMutex // Protege operaciones de cache
}

// NewGeoHandler crea una nueva instancia de GeoHandler
func NewGeoHandler() *GeoHandler {
	return &GeoHandler{
		geoDataCache: services.NewCacheService[*types.GeoData](60), // 1 hora
		municCache:   services.NewCacheService[map[string]*types.GeoFeatureCollection](60),
	}
}

// GetMunicipios maneja el endpoint para filtrar municipios
// @Summary Filtrar datos geoespaciales
// @Description Filtra datos geoespaciales según parámetros
// @Tags geo
// @Produce json
// @Param query query string true "Cadena de búsqueda"
// @Param whatIs query string true "Tipo de filtro: D (departamentos), M (municipios), NAM (nombres/ubicaciones)"
// @Success 200 {object} GeoFilterResponse "Resultados filtrados"
// @Failure 400 {object} ErrorResponse "Parámetros inválidos"
// @Failure 500 {object} ErrorResponse "Error interno"
// @Router /geo/filter [get]
func (h *GeoHandler) GetMunicipios(c *fiber.Ctx) error {
	query := c.Query("query")
	whatIs := c.Query("whatIs")

	// Validar parámetros
	validatedQuery, validQuery := utils.ValidateQuery(query)
	validatedWhatIs, validWhatIs := utils.ValidateWhatIs(whatIs)

	if !validQuery || !validWhatIs {
		return utils.RespondWithError(c, fiber.StatusBadRequest,
			"Parámetros inválidos. 'query' debe ser una cadena válida (máx 100 chars) y 'whatIs' debe ser: D, M, o NAM")
	}

	// Usar valores validados
	cacheKey := validatedWhatIs + ":" + validatedQuery
	
	// Check cache with read lock
	h.cacheMutex.RLock()
	if cached, ok := h.municCache.Get(); ok {
		if data, exists := cached[cacheKey]; exists {
			h.cacheMutex.RUnlock()
			return utils.SendResponse(c, data)
		}
	}
	h.cacheMutex.RUnlock()

	// Los valores ya están validados y en el formato correcto (D, M, NAM)
	data, err := geospatial.GetMunicipios(validatedQuery, validatedWhatIs)
	if err != nil {
		utils.Error("Error al obtener municipios: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError, err.Error())
	}

	// Update cache with write lock
	h.cacheMutex.Lock()
	cached, _ := h.municCache.Get()
	if cached == nil {
		cached = make(map[string]*types.GeoFeatureCollection)
	}
	cached[cacheKey] = data
	h.municCache.Set(cached)
	h.cacheMutex.Unlock()

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
