package utils

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// SetupSwagger configura la ruta para la documentación Swagger
func SetupSwagger(app *fiber.App) {
	// Configurar Swagger con opciones básicas
	swaggerHandler := fiberSwagger.FiberWrapHandler(
		fiberSwagger.DocExpansion("none"), // Mantener los endpoints colapsados por defecto
		fiberSwagger.DomID("swagger-ui"),  // ID DOM para Swagger UI
	)

	// Configurar ruta para la documentación
	app.Get("/docs/*", swaggerHandler)

	// Redirección para la ruta base
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html")
	})
}
