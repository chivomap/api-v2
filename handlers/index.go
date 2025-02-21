package handlers

import "github.com/gofiber/fiber/v2"

// SetupRoutes registra los endpoints de scraping y sismos.
func SetupRoutes(app *fiber.App) {
	app.Get("/scrape", ScrapeHandler)
	app.Get("/sismos", GetSismos)
	app.Get("/sismos/refresh", ForceRefreshSismos)
}
