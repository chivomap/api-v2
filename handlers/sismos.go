package handlers

import (
	"log"
	"time"

	"chivomap.com/services"
	"chivomap.com/services/scraping"

	"github.com/gofiber/fiber/v2"
)

// Creamos una instancia global de caché para un slice de sismos con TTL de 3 minutos.
var sismosCache = services.NewCacheService[[]scraping.Sismo](3)

// updateCacheInBackground actualiza la caché en segundo plano.
func updateCacheInBackground() {
	sismosCache.SetUpdating(true)
	defer sismosCache.SetUpdating(false)

	newData, err := scraping.ScrapeSismos()
	if err != nil {
		log.Println("❌ Error al actualizar caché:", err)
		return
	}
	sismosCache.Set(newData)
}

// GetSismos maneja el endpoint GET /sismos.
// Devuelve los datos en caché si están disponibles y, si la caché ha expirado, inicia una actualización en background.
func GetSismos(c *fiber.Ctx) error {
	if data, ok := sismosCache.Get(); ok {
		// Si la caché necesita actualización, se lanza en background.
		if sismosCache.NeedsUpdate() {
			go updateCacheInBackground()
		}
		return c.JSON(fiber.Map{
			"timestamp":   map[string]string{"time": time.Now().Format("2006-01-02 15:04:05")},
			"totalSismos": len(data),
			"data":        data,
		})
	}

	// Primera carga: no hay datos en caché.
	log.Println("⏳ Primera carga, obteniendo datos...")
	data, err := scraping.ScrapeSismos()
	if err != nil {
		log.Println("❌ Error en el scraping:", err)
		return c.Status(500).JSON(fiber.Map{"error": "No se pudieron obtener los datos"})
	}
	sismosCache.Set(data)
	return c.JSON(fiber.Map{
		"timestamp":   map[string]string{"time": time.Now().Format("2006-01-02 15:04:05")},
		"totalSismos": len(data),
		"data":        data,
	})
}

// ForceRefreshSismos permite forzar la actualización de la caché mediante el endpoint GET /sismos/refresh.
func ForceRefreshSismos(c *fiber.Ctx) error {
	log.Println("🔄 Forzando actualización del cache...")
	data, err := scraping.ScrapeSismos()
	if err != nil {
		log.Println("❌ Error al refrescar los datos:", err)
		return c.Status(500).JSON(fiber.Map{"error": "No se pudieron actualizar los datos"})
	}
	sismosCache.Set(data)
	return c.JSON(fiber.Map{
		"message":     "Cache actualizada exitosamente",
		"timestamp":   map[string]string{"time": time.Now().Format("2006-01-02 15:04:05")},
		"totalSismos": len(data),
		"data":        data,
	})
}
