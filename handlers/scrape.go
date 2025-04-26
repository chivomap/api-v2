package handlers

import (
	"chivomap.com/config"
	"chivomap.com/models"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// ScrapeHandler maneja el endpoint para scraping de datos
func ScrapeHandler(c *fiber.Ctx) error {
	// Consultar la base de datos
	rows, err := config.DB.Query("SELECT id, title FROM scraped_data")
	if err != nil {
		utils.Error("Error al consultar la base de datos: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "Error al consultar la base de datos")
	}
	defer rows.Close()

	// Convertir resultados a slice
	var results []models.ScrapedData
	for rows.Next() {
		var data models.ScrapedData
		if err := rows.Scan(&data.ID, &data.Title); err != nil {
			utils.Error("Error al escanear datos: %v", err)
			continue
		}
		results = append(results, data)
	}

	// Verificar errores durante la iteración
	if err := rows.Err(); err != nil {
		utils.Error("Error al iterar resultados: %v", err)
		return utils.RespondWithError(c, fiber.StatusInternalServerError, "Error al procesar resultados")
	}

	// Devolver respuesta
	return utils.SendResponse(c, fiber.Map{
		"totalItems": len(results),
		"items":      results,
	})
}
