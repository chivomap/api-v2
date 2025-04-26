package main

import (
	"os"
	"os/signal"
	"syscall"

	"chivomap.com/config"
	"chivomap.com/handlers"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

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

	utils.Info("✅ Servidor cerrado correctamente")
}
