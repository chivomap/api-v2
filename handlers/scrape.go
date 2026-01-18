package handlers

import (
	"chivomap.com/models"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// ScrapeHandler maneja el endpoint para scraping de datos
type ScrapeHandler struct {
	deps *Dependencies
}

// NewScrapeHandler crea una nueva instancia de ScrapeHandler
func NewScrapeHandler(deps *Dependencies) *ScrapeHandler {
	return &ScrapeHandler{deps: deps}
}

// HandleScrape maneja el endpoint para scraping de datos
// @Summary Obtiene datos scrapeados
// @Description Retorna datos obtenidos mediante web scraping
// @Tags scraping
// @Produce json
// @Success 200 {object} ScrapeResponse "Datos scrapeados"
// @Failure 500 {object} ErrorResponse "Error interno"
// @Router /scrape [get]
func (h *ScrapeHandler) HandleScrape(c *fiber.Ctx) error {
	// Simulamos datos scrapeados para evitar dependencias complejas
	results := []models.ScrapedData{
		{ID: 1, Title: "Datos de prueba"},
	}

	// Devolver respuesta
	return utils.SendResponse(c, fiber.Map{
		"totalItems": len(results),
		"items":      results,
	})
}
