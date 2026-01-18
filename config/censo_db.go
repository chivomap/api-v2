package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/tursodatabase/go-libsql"
)

// CensoDB es la conexión a la base de datos del censo
var CensoDB *sql.DB

// LoadCensoConfig carga la configuración de la base de datos del censo
func LoadCensoConfig() (string, string, error) {
	dbURL := os.Getenv("TURSO_DATABASE_URL_CENSO")
	authToken := os.Getenv("TURSO_AUTH_TOKEN_CENSO")

	if dbURL == "" || authToken == "" {
		return "", "", fmt.Errorf("faltan las credenciales de la base de datos del censo: configura TURSO_DATABASE_URL_CENSO y TURSO_AUTH_TOKEN_CENSO en .env")
	}

	return dbURL, authToken, nil
}

// ConnectCensoDB establece la conexión con la base de datos del censo
func ConnectCensoDB() error {
	dbURL, authToken, err := LoadCensoConfig()
	if err != nil {
		return fmt.Errorf("error cargando configuración del censo: %w", err)
	}
	
	url := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	CensoDB, err = sql.Open("libsql", url)
	if err != nil {
		return fmt.Errorf("error al conectar a la base de datos del censo: %w", err)
	}

	// Verificar la conexión
	if err = CensoDB.Ping(); err != nil {
		return fmt.Errorf("error al verificar la conexión a la base de datos del censo: %w", err)
	}

	return nil
}

// CloseCensoDB cierra la conexión a la base de datos del censo
func CloseCensoDB() error {
	if CensoDB != nil {
		if err := CensoDB.Close(); err != nil {
			return fmt.Errorf("error cerrando la conexión a la base de datos del censo: %w", err)
		}
	}
	return nil
}
