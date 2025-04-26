package handlers

import (
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(app *fiber.App) {
	// Middleware para logging de solicitudes
	app.Use(utils.RequestLogger())

	// Sismos
	sismosHandler := NewSismosHandler()
	app.Get("/sismos", sismosHandler.GetSismos)
	app.Get("/sismos/refresh", sismosHandler.ForceRefreshSismos)

	// Geo
	geoHandler := NewGeoHandler()
	app.Get("/geo/filter", geoHandler.GetMunicipios)
	app.Get("/geo/search-data", geoHandler.GetGeoData)

	// Scraping
	app.Get("/scrape", ScrapeHandler)

	// Otras rutas
	app.Get("/health", HealthCheck)
}
