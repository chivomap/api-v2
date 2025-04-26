package config

import (
	"log"
	"os"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	// Puerto del servidor HTTP
	ServerPort string
	// URL de la base de datos Turso
	DatabaseURL string
	// Token de autenticación para la base de datos Turso
	DatabaseToken string
}

// AppConfig es la configuración global de la aplicación
var AppConfig Config

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() {
	AppConfig = Config{
		ServerPort:    getEnvOrDefault("PORT", "8080"),
		DatabaseURL:   getEnv("TURSO_DATABASE_URL"),
		DatabaseToken: getEnv("TURSO_AUTH_TOKEN"),
	}

	log.Printf("Configuración cargada: Puerto=%s", AppConfig.ServerPort)
}

// getEnv obtiene una variable de entorno requerida
func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ La variable de entorno %s es requerida", key)
	}
	return val
}

// getEnvOrDefault obtiene una variable de entorno o devuelve un valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}
