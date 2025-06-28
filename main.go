package main

import (
	"os"
	"os/signal"
	"syscall"

	"chivomap.com/config"
	_ "chivomap.com/docs" // Importación necesaria para Swagger
	"chivomap.com/handlers"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title ChivoMap API
// @version 1.0
// @description API que proporciona datos geoespaciales y sísmicos de El Salvador.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url https://chivomap.com/support
// @contact.email support@chivomap.com
// @license.name MIT License
// @license.url https://github.com/oclazi/chivomap-api/blob/main/LICENSE.md
// @host localhost:8080
// @BasePath /
func main() {
	// Cargar configuración
	config.LoadConfig()

	// Inicializar Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.RespondWithError(c, fiber.StatusInternalServerError,
				"Error interno del servidor")
		},
	})

	// Middleware de recuperación de errores
	app.Use(recover.New())

	// Configurar CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, https://chivomap.com",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: false,
	}))

	// Conectar a la base de datos
	config.ConnectDB()

	// Configurar rutas
	handlers.SetupRoutes(app)

	// Configurar Swagger con tema oscuro y toggle
	utils.SetupSwagger(app)

	// Configurar canal para cierre graceful
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Iniciar el servidor en una goroutine
	go func() {
		serverPort := ":" + config.AppConfig.ServerPort
		if serverPort == ":" {
			serverPort = ":8080"
		}

		utils.Info("🚀 Servidor corriendo en http://localhost%s", serverPort)
		utils.Info("📚 Documentación Swagger disponible en http://localhost%s/docs/", serverPort)
		if err := app.Listen(serverPort); err != nil {
			utils.Fatal("Error al iniciar el servidor: %v", err)
		}
	}()

	// Esperar señal de cierre
	<-c
	utils.Info("Cerrando servidor gracefully...")

	// Cerrar servidor
	if err := app.Shutdown(); err != nil {
		utils.Error("Error al cerrar el servidor: %v", err)
	}

	// Cerrar conexiones de base de datos
	if config.DB != nil {
		if err := config.DB.Close(); err != nil {
			utils.Error("Error al cerrar la base de datos principal: %v", err)
		}
	}
	
	if config.CensoDB != nil {
		if err := config.CensoDB.Close(); err != nil {
			utils.Error("Error al cerrar la base de datos del censo: %v", err)
		}
	}

	utils.Info("✅ Servidor cerrado correctamente")
}
