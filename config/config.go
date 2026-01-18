package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/joho/godotenv"
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
func LoadConfig() error {
	// Cargar archivo .env si existe
	if err := godotenv.Load(); err != nil {
		// No es un error crítico si no existe .env
		fmt.Printf("Advertencia: No se pudo cargar .env: %v\n", err)
	}
	
	// Detectar directorio base de la aplicación
	execPath, err := os.Executable()
	var baseDir string
	if err != nil {
		// Fallback a directorio de trabajo actual
		baseDir, _ = os.Getwd()
	} else {
		baseDir = filepath.Dir(execPath)
	}
	
	databaseURL, err := getEnv("TURSO_DATABASE_URL")
	if err != nil {
		return fmt.Errorf("error cargando configuración: %w", err)
	}
	
	AppConfig = Config{
		ServerPort:    getEnvOrDefault("PORT", "8080"),
		DatabaseURL:   databaseURL,
		DatabaseToken: getEnvOrDefault("TURSO_AUTH_TOKEN", ""), // Token opcional para SQLite local
		BaseDir:      baseDir,
		AssetsDir:    getEnvOrDefault("ASSETS_DIR", filepath.Join(baseDir, "utils", "assets")),
	}

	return nil
}

// getEnv obtiene una variable de entorno requerida
func getEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("variable de entorno requerida no encontrada: %s", key)
	}
	return val, nil
}

// getEnvOrDefault obtiene una variable de entorno o devuelve un valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}
