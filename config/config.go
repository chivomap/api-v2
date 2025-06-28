package config

import (
	"log"
	"os"
	"path/filepath"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	// Puerto del servidor HTTP
	ServerPort string
	// URL de la base de datos Turso
	DatabaseURL string
	// Token de autenticación para la base de datos Turso
	DatabaseToken string
	// Directorio base de la aplicación
	BaseDir string
	// Directorio de assets
	AssetsDir string
}

// AppConfig es la configuración global de la aplicación
var AppConfig Config

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() {
	// Detectar directorio base de la aplicación
	execPath, err := os.Executable()
	var baseDir string
	if err != nil {
		// Fallback a directorio de trabajo actual
		baseDir, _ = os.Getwd()
	} else {
		baseDir = filepath.Dir(execPath)
	}
	
	AppConfig = Config{
		ServerPort:    getEnvOrDefault("PORT", "8080"),
		DatabaseURL:   getEnv("TURSO_DATABASE_URL"),
		DatabaseToken: getEnvOrDefault("TURSO_AUTH_TOKEN", ""), // Token opcional para SQLite local
		BaseDir:      baseDir,
		AssetsDir:    getEnvOrDefault("ASSETS_DIR", filepath.Join(baseDir, "utils", "assets")),
	}

	log.Printf("Configuración cargada: Puerto=%s, BaseDir=%s", AppConfig.ServerPort, AppConfig.BaseDir)
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
