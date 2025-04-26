package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// SendResponse envía una respuesta JSON estandarizada
func SendResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      data,
	})
}

// RespondWithError envía una respuesta de error estandarizada
func RespondWithError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
