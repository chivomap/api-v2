package models

// ErrorResponse representa la estructura de una respuesta de error
type ErrorResponse struct {
	Error     string `json:"error" example:"Mensaje de error"`
	Timestamp string `json:"timestamp" example:"2023-05-25T12:34:56Z"`
}

// HealthResponse representa la respuesta del endpoint de salud
type HealthResponse struct {
	Status  string `json:"status" example:"UP"`
	Version string `json:"version" example:"1.0.0"`
}

// SismoResponse representa la respuesta del endpoint de sismos
type SismoResponse struct {
	TotalSismos int     `json:"totalSismos" example:"10"`
	Data        []Sismo `json:"data"`
}

// Sismo representa un evento sísmico
type Sismo struct {
	Fecha        string `json:"fecha" example:"2023-05-25 10:30:00"`
	Fases        string `json:"fases" example:"P,S"`
	Latitud      string `json:"latitud" example:"13.6894"`
	Longitud     string `json:"longitud" example:"-89.1872"`
	Profundidad  string `json:"profundidad" example:"5.5"`
	Magnitud     string `json:"magnitud" example:"4.2"`
	Localizacion string `json:"localizacion" example:"5 km al Este de San Salvador"`
	RMS          string `json:"rms" example:"0.3"`
	Estado       string `json:"estado" example:"Revisado"`
}

// GeoDataResponse representa la respuesta del endpoint de datos geográficos
type GeoDataResponse struct {
	Departamentos []string `json:"departamentos" example:"['San Salvador', 'La Libertad', 'Santa Ana']"`
	Municipios    []string `json:"municipios" example:"['San Salvador', 'Santa Tecla', 'Mejicanos']"`
	Distritos     []string `json:"distritos" example:"['Centro', 'Norte', 'Sur']"`
}

// ScrapedDataResponse representa la respuesta del endpoint de scraping
type ScrapedDataResponse struct {
	TotalItems int           `json:"totalItems" example:"5"`
	Items      []ScrapedData `json:"items"`
}

// StandardResponse representa una estructura de respuesta estándar
type StandardResponse struct {
	Timestamp string      `json:"timestamp" example:"2023-05-25T12:34:56Z"`
	Data      interface{} `json:"data"`
}
