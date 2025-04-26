package utils

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Info registra un mensaje informativo
func Info(format string, v ...interface{}) {
	logWithLevel("INFO", format, v...)
}

// Error registra un mensaje de error
func Error(format string, v ...interface{}) {
	logWithLevel("ERROR", format, v...)
}

// Fatal registra un mensaje de error fatal y termina el programa
func Fatal(format string, v ...interface{}) {
	logWithLevel("FATAL", format, v...)
	log.Fatal("Terminando aplicación después de error fatal")
}

// Debug registra un mensaje de depuración
func Debug(format string, v ...interface{}) {
	logWithLevel("DEBUG", format, v...)
}

// logWithLevel registra un mensaje con un nivel específico
func logWithLevel(level, format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s - "+format, append([]interface{}{level, timestamp}, v...)...)
}

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		log.Printf("%s %s %d %s",
			c.Method(), c.Path(), c.Response().StatusCode(), duration)

		return err
	}
}
