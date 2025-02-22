package handlers

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {
	app.Get("/scrape", ScrapeHandler)
	app.Get("/sismos", GetSismos)
	app.Get("/sismos/refresh", ForceRefreshSismos)

	app.Get("/geo/filter", NewGeoHandler().GetMunicipios)
	app.Get("/geo/search-data", NewGeoHandler().GetGeoData)
}
