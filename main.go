package main

import (
	"log"

	"chivomap.com/config"
	"chivomap.com/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	config.ConnectDB()
	handlers.SetupRoutes(app)

	log.Println("🚀 Servidor corriendo en http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
