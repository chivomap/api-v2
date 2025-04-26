package config

import (
	"log"
	"os"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	ServerPort    string
	DatabaseURL   string
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
