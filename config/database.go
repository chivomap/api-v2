package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/tursodatabase/go-libsql"
)

// DB es la conexión global a la base de datos
var DB *sql.DB

// ConnectDB establece la conexión con la base de datos Turso
func ConnectDB() {
	// Asegurarse de que la configuración esté cargada
	if AppConfig.DatabaseURL == "" {
		LoadConfig()
	}

	dbURL := AppConfig.DatabaseURL
	authToken := AppConfig.DatabaseToken

	if dbURL == "" || authToken == "" {
		log.Fatal("❌ Faltan las credenciales de la base de datos. Configura TURSO_DATABASE_URL y TURSO_AUTH_TOKEN en .env")
	}

	url := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	var err error
	DB, err = sql.Open("libsql", url)
	if err != nil {
		log.Fatalf("❌ Error al conectar a la base de datos: %v", err)
	}

	// Verificar la conexión
	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Error al verificar la conexión a la base de datos: %v", err)
	}

	log.Println("✅ Conectado a la base de datos Turso")

	// Crear tabla si no existe
	createTable := `
    CREATE TABLE IF NOT EXISTS scraped_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL
    );`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("❌ Error creando la tabla:", err)
	}
}
