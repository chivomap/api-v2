package utils

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// SetupSwagger configura la ruta para la documentación Swagger
func SetupSwagger(app *fiber.App) {
	// Configurar Swagger con handler que incluye assets estáticos
	app.Get("/docs/*", fiberSwagger.FiberWrapHandler(
		fiberSwagger.URL("doc.json"), // Especificar la URL del JSON
		fiberSwagger.DocExpansion("none"),
	))
	
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html", 301)
	})
}
