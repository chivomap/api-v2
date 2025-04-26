package handlers

import (
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// HealthCheck maneja el endpoint para verificar el estado de la API
// @Summary Verificación del estado de la API
// @Description Retorna el estado actual de la API
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Estado de la API"
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {
	return utils.SendResponse(c, fiber.Map{
		"status":  "UP",
		"version": "1.0.0",
	})
}
