package handlers

import (
	"log"
	"time"

	"chivomap.com/services/scraping"

	"github.com/gofiber/fiber/v2"
)

// Endpoint para obtener los datos de los sismos
func GetSismos(c *fiber.Ctx) error {
	sismos, err := scraping.ScrapeSismos()
	if err != nil {
		log.Println("❌ Error en el scraping:", err)
		return c.Status(500).JSON(fiber.Map{"error": "No se pudieron obtener los datos"})
	}

	return c.JSON(fiber.Map{
		"timestamp":   map[string]string{"time": time.Now().Format("2006-01-02 15:04:05")},
		"totalSismos": len(sismos),
		"data":        sismos,
	})
}
