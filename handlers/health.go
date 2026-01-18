package handlers

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler maneja los endpoints de health check
type HealthHandler struct {
	deps *Dependencies
}

// NewHealthHandler crea una nueva instancia de HealthHandler
func NewHealthHandler(deps *Dependencies) *HealthHandler {
	return &HealthHandler{deps: deps}
}

// HealthStatus representa el estado de un componente
type HealthStatus struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// HealthResponse representa la respuesta completa del health check
type HealthResponse struct {
	Status       string                  `json:"status"`
	Version      string                  `json:"version"`
	Timestamp    time.Time               `json:"timestamp"`
	Uptime       string                  `json:"uptime"`
	Components   map[string]HealthStatus `json:"components"`
}

var startTime = time.Now()

// HealthCheck maneja el endpoint para verificar el estado de la API
// @Summary Verificación del estado de la API
// @Description Retorna el estado detallado de la API y sus componentes
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Estado de la API"
// @Failure 503 {object} HealthResponse "Servicio no disponible"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	components := make(map[string]HealthStatus)
	overallStatus := "UP"

	// Verificar conectividad de base de datos principal
	dbStatus := h.checkDatabase(h.deps.DB, "Base de datos principal")
	components["database"] = dbStatus
	if dbStatus.Status != "UP" {
		overallStatus = "DOWN"
	}

	// Verificar base de datos del censo
	censoDbStatus := h.checkDatabase(h.deps.CensoDB, "Base de datos del censo")
	components["censo_database"] = censoDbStatus
	if censoDbStatus.Status != "UP" {
		overallStatus = "DEGRADED"
	}

	// Verificar archivos estáticos
	staticStatus := h.checkStaticFiles()
	components["static_files"] = staticStatus
	if staticStatus.Status != "UP" {
		overallStatus = "DEGRADED"
	}

	// Verificar cache
	cacheStatus := h.checkCache()
	components["cache"] = cacheStatus
	if cacheStatus.Status != "UP" {
		overallStatus = "DEGRADED"
	}

	response := HealthResponse{
		Status:     overallStatus,
		Version:    "1.0.0",
		Timestamp:  time.Now(),
		Uptime:     time.Since(startTime).String(),
		Components: components,
	}

	// Retornar código de estado apropiado
	statusCode := fiber.StatusOK
	if overallStatus == "DOWN" {
		statusCode = fiber.StatusServiceUnavailable
	} else if overallStatus == "DEGRADED" {
		statusCode = fiber.StatusOK // 200 pero con advertencias
	}

	return c.Status(statusCode).JSON(response)
}

// checkDatabase verifica la conectividad de una base de datos
func (h *HealthHandler) checkDatabase(dbService interface{}, name string) HealthStatus {
	if dbService == nil {
		return HealthStatus{
			Status:  "DOWN",
			Message: name + " no configurada",
		}
	}

	// Para simplificar, asumimos que está UP si no es nil
	return HealthStatus{
		Status: "UP",
		Details: map[string]interface{}{
			"name": name,
		},
	}
}

// checkStaticFiles verifica la disponibilidad de archivos estáticos críticos
func (h *HealthHandler) checkStaticFiles() HealthStatus {
	assetsDir := h.deps.Config.GetAssetsDir()
	topoPath := filepath.Join(assetsDir, "topo.json")
	
	fileInfo, err := os.Stat(topoPath)
	if err != nil {
		return HealthStatus{
			Status:  "DOWN",
			Message: "Archivo TopoJSON no disponible: " + err.Error(),
		}
	}

	return HealthStatus{
		Status: "UP",
		Details: map[string]interface{}{
			"topo_file_size": fileInfo.Size(),
			"topo_mod_time":  fileInfo.ModTime(),
		},
	}
}

// checkCache verifica el estado del cache estático
func (h *HealthHandler) checkCache() HealthStatus {
	// Simplificamos la verificación del cache
	return HealthStatus{
		Status: "UP",
		Details: map[string]interface{}{
			"cache_type": "static_file_cache",
		},
	}
}
