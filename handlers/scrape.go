package handlers

import (
	"log"

	"chivomap.com/config"
	"chivomap.com/models"

	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
)

// Scrapea una página y almacena el título en la base de datos
func ScrapeHandler(c *fiber.Ctx) error {
	collector := colly.NewCollector()
	var title string

	collector.OnHTML("title", func(e *colly.HTMLElement) {
		title = e.Text
	})

	err := collector.Visit("https://example.com")
	if err != nil {
		log.Println("❌ Error en scraping:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Scraping fallido"})
	}

	_, err = config.DB.Exec("INSERT INTO scraped_data (title) VALUES (?)", title)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error guardando en DB"})
	}

	return c.JSON(models.ScrapedData{Title: title})
}
