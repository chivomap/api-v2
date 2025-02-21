package main

import (
	"log"

	"chivomap.com/config"
	"chivomap.com/handlers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Conectar a la base de datos
	config.ConnectDB()

	handlers.SetupRoutes(app)

	log.Println("🚀 Servidor corriendo en http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
