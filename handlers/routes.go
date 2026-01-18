package handlers

import (
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(app *fiber.App, deps *Dependencies) {
	// Middleware para logging de solicitudes
	app.Use(utils.RequestLogger())

	// Sismos
	sismosHandler := NewSismosHandler(deps)
	app.Get("/sismos", sismosHandler.GetSismos)
	app.Get("/sismos/refresh", sismosHandler.ForceRefreshSismos)

	// Geo
	geoHandler := NewGeoHandler(deps)
	app.Get("/geo/filter", geoHandler.GetMunicipios)
	app.Get("/geo/search-data", geoHandler.GetGeoData)

	// Scraping
	app.Get("/scrape", NewScrapeHandler(deps))

	// Otras rutas
	app.Get("/health", NewHealthHandler(deps))
}
