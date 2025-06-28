package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/go-libsql"
)

// CensoDB es la conexión a la base de datos del censo
var CensoDB *sql.DB

// LoadCensoConfig carga la configuración de la base de datos del censo
func LoadCensoConfig() (string, string) {
	dbURL := os.Getenv("TURSO_DATABASE_URL_CENSO")
	authToken := os.Getenv("TURSO_AUTH_TOKEN_CENSO")

	if dbURL == "" || authToken == "" {
		log.Fatal("❌ Faltan las credenciales de la base de datos del censo. Configura TURSO_DATABASE_URL_CENSO y TURSO_AUTH_TOKEN_CENSO en .env")
	}

	return dbURL, authToken
}

// ConnectCensoDB establece la conexión con la base de datos del censo
func ConnectCensoDB() {
	dbURL, authToken := LoadCensoConfig()
	url := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	var err error
	CensoDB, err = sql.Open("libsql", url)
	if err != nil {
		log.Fatalf("❌ Error al conectar a la base de datos del censo: %v", err)
	}

	// Verificar la conexión
	if err = CensoDB.Ping(); err != nil {
		log.Fatalf("❌ Error al verificar la conexión a la base de datos del censo: %v", err)
	}

	log.Println("✅ Conectado a la base de datos del censo en Turso")
}

// CloseCensoDB cierra la conexión a la base de datos del censo
func CloseCensoDB() {
	if CensoDB != nil {
		CensoDB.Close()
		log.Println("Conexión a la base de datos del censo cerrada")
	}
}
