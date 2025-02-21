package routes

import (
	"chivomap.com/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/scrape", handlers.ScrapeHandler)
	// app.Get("/geo", handlers.GeoHandler)
}
