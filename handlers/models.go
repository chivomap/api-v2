package handlers

import (
	"chivomap.com/models"
	"chivomap.com/services/geospatial"
	"chivomap.com/services/scraping"
)

// ErrorResponse representa una respuesta de error estándar
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HealthResponse representa la respuesta del endpoint de salud
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// SismosResponse representa la respuesta del endpoint de sismos
type SismosResponse struct {
	TotalSismos int              `json:"totalSismos"`
	Data        []scraping.Sismo `json:"data"`
}

// SismosRefreshResponse representa la respuesta del endpoint de actualización de sismos
type SismosRefreshResponse struct {
	Message     string           `json:"message"`
	TotalSismos int              `json:"totalSismos"`
	Data        []scraping.Sismo `json:"data"`
}

// GeoDataResponse representa la respuesta para datos geográficos
type GeoDataResponse struct {
	GeoData *geospatial.GeoData `json:"geoData"`
}

// GeoFilterResponse representa la respuesta para datos filtrados
type GeoFilterResponse struct {
	Type     string                   `json:"type"`
	Features []map[string]interface{} `json:"features"`
}

// ScrapeResponse representa la respuesta del endpoint de scraping
type ScrapeResponse struct {
	TotalItems int                  `json:"totalItems"`
	Items      []models.ScrapedData `json:"items"`
}
