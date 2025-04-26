package handlers

import (
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// HealthCheck maneja el endpoint para verificar el estado de la API
func HealthCheck(c *fiber.Ctx) error {
	return utils.SendResponse(c, fiber.Map{
		"status":  "UP",
		"version": "1.0.0",
	})
}
