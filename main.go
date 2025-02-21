package main

import (
	"log"

	"chivomap.com/config"
	"chivomap.com/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Conectar a la base de datos
	config.ConnectDB()

	// Registrar rutas
	routes.SetupRoutes(app)

	log.Println("🚀 Servidor corriendo en http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
