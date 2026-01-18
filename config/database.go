package config

import (
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/go-libsql"
)

// DB es la conexión global a la base de datos
var DB *sql.DB

// ConnectDB establece la conexión con la base de datos Turso
func ConnectDB() error {
	// Asegurarse de que la configuración esté cargada
	if AppConfig.DatabaseURL == "" {
		if err := LoadConfig(); err != nil {
			return fmt.Errorf("error cargando configuración: %w", err)
		}
	}

	dbURL := AppConfig.DatabaseURL
	authToken := AppConfig.DatabaseToken

	if dbURL == "" {
		return fmt.Errorf("TURSO_DATABASE_URL es requerida")
	}

	// Para SQLite local, el token no es necesario
	var url string
	if authToken != "" {
		url = fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	} else {
		url = dbURL
	}

	var err error
	DB, err = sql.Open("libsql", url)
	if err != nil {
		return fmt.Errorf("error al conectar a la base de datos: %w", err)
	}

	// Verificar la conexión
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error al verificar la conexión a la base de datos: %w", err)
	}

	// Crear tabla si no existe
	createTable := `
    CREATE TABLE IF NOT EXISTS scraped_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL
    );`
	_, err = DB.Exec(createTable)
	if err != nil {
		return fmt.Errorf("error creando la tabla: %w", err)
	}

	return nil
}
