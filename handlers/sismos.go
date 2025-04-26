package handlers

import (
	"chivomap.com/services"
	"chivomap.com/services/scraping"
	"chivomap.com/utils"

	"github.com/gofiber/fiber/v2"
)

// SismosHandler maneja los endpoints relacionados con sismos
type SismosHandler struct {
	cache *services.CacheService[[]scraping.Sismo]
}

// NewSismosHandler crea una nueva instancia de SismosHandler
func NewSismosHandler() *SismosHandler {
	return &SismosHandler{
		cache: services.NewCacheService[[]scraping.Sismo](3), // 3 minutos TTL
	}
}

// updateCacheInBackground actualiza la caché en segundo plano
func (h *SismosHandler) updateCacheInBackground() {
	h.cache.SetUpdating(true)
	defer h.cache.SetUpdating(false)

	newData, err := scraping.ScrapeSismos()
	if err != nil {
		utils.Error("Error al actualizar caché: %v", err)
		return
	}
	h.cache.Set(newData)
}

// GetSismos maneja el endpoint GET /sismos
func (h *SismosHandler) GetSismos(c *fiber.Ctx) error {
	if data, ok := h.cache.Get(); ok {
		// Si la caché necesita actualización, se lanza en background
		if h.cache.NeedsUpdate() {
			go h.updateCacheInBackground()
		}
		return utils.SendResponse(c, fiber.Map{
			"totalSismos": len(data),
			"data":        data,
		})
	}

	// Primera carga: no hay datos en caché
	utils.Info("Primera carga, obteniendo datos...")
	data, err := scraping.ScrapeSismos()
	if err != nil {
		utils.Error("Error en el scraping: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "No se pudieron obtener los datos")
	}
	h.cache.Set(data)
	return utils.SendResponse(c, fiber.Map{
		"totalSismos": len(data),
		"data":        data,
	})
}

// ForceRefreshSismos permite forzar la actualización de la caché mediante el endpoint GET /sismos/refresh
func (h *SismosHandler) ForceRefreshSismos(c *fiber.Ctx) error {
	utils.Info("Forzando actualización del cache...")
	data, err := scraping.ScrapeSismos()
	if err != nil {
		utils.Error("Error al refrescar los datos: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "No se pudieron actualizar los datos")
	}
	h.cache.Set(data)
	return utils.SendResponse(c, fiber.Map{
		"message":     "Cache actualizada exitosamente",
		"totalSismos": len(data),
		"data":        data,
	})
}
