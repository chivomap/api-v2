package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"chivomap.com/config"
	"chivomap.com/container"
	_ "chivomap.com/docs" // Importación necesaria para Swagger
	"chivomap.com/handlers"
	"chivomap.com/services"
	"chivomap.com/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
	if err := config.LoadConfig(); err != nil {
		utils.Fatal("Error cargando configuración: %v", err)
	}

	// Inicializar Fiber con límites de seguridad
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.RespondWithError(c, fiber.StatusInternalServerError,
				"Error interno del servidor")
		},
		// Límites de seguridad para prevenir ataques DoS
		BodyLimit:       10 * 1024 * 1024, // 10MB máximo
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		ReadBufferSize:  8192,
		WriteBufferSize: 8192,
	})

	// Middleware de recuperación de errores
	app.Use(recover.New())

	// Rate limiting para prevenir abuso
	app.Use(limiter.New(limiter.Config{
		Max:        100,               // 100 requests
		Expiration: 1 * time.Minute,   // por minuto
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Por IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.RespondWithError(c, fiber.StatusTooManyRequests,
				"Demasiadas solicitudes. Intente de nuevo más tarde.")
		},
	}))

	// Configurar CORS
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173,https://chivomap.com,https://www.chivomap.com"
	}
	
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: false,
	}))

	// Conectar a la base de datos
	if err := config.ConnectDB(); err != nil {
		utils.Fatal("Error conectando a la base de datos: %v", err)
	}

	// Crear contenedor de dependencias
	configService := services.NewConfigServiceFromGlobal()
	container, err := container.NewContainer(configService, config.DB, config.CensoDB)
	if err != nil {
		utils.Fatal("Error creando contenedor de dependencias: %v", err)
	}
	defer container.Close()

	// Crear dependencias para handlers
	deps := &handlers.Dependencies{
		Config:      container.Config,
		DB:          container.DB,
		CensoDB:     container.CensoDB,
		StaticCache: container.StaticCache,
		Logger:      container.Logger,
	}

	// Configurar rutas
	handlers.SetupRoutes(app, deps)

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
