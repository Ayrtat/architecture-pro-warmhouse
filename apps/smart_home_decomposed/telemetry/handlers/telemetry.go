package handlers

import (
	"net/http"
	"strconv"
	"time"

	"smarthome/telemetry/services"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type TelemetryHandler struct {
	telemetryService *services.TelemetryService
}

func NewTelemetryHandler(telemetryService *services.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{
		telemetryService: telemetryService,
	}
}

// RegisterRoutes registers all telemetry-related routes
func (h *TelemetryHandler) RegisterRoutes(router *gin.RouterGroup) {
	telemetry := router.Group("/telemetry")
	{
		telemetry.GET("/health", h.HealthCheck)

		telemetry.GET("/sensors/:id", h.GetSensorTelemetry)
		telemetry.GET("/sensors", h.GetAllSensorsTelemetry)
	}
}

// GetSensorTelemetry godoc
// @Summary Get sensor telemetry
// @Description Get current telemetry data for a specific sensor
// @Tags telemetry
// @Accept json
// @Produce json
// @Param id path int true "Sensor ID"
// @Success 200 {object} models.TelemetryData
// @Failure 400 {object} Response
// @Router /telemetry/sensors/{id} [get]
func (h *TelemetryHandler) GetSensorTelemetry(c *gin.Context) {
	sensorIDStr := c.Param("id")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	telemetry := h.telemetryService.GetSensorTelemetry(uint64(sensorID))
	c.JSON(http.StatusOK, telemetry)
}

// GetAllSensorsTelemetry godoc
// @Summary Get all sensors telemetry
// @Description Get current telemetry data for all sensors
// @Tags telemetry
// @Accept json
// @Produce json
// @Success 200 {object} map[uint64]*models.TelemetryData
// @Router /telemetry/sensors [get]
func (h *TelemetryHandler) GetAllSensorsTelemetry(c *gin.Context) {
	telemetry := h.telemetryService.GetAllSensorsTelemetry()
	c.JSON(http.StatusOK, gin.H{
		"total":     len(telemetry),
		"timestamp": time.Now(),
		"sensors":   telemetry,
	})
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the telemetry service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=bool}
// @Router /telemetry/health [get]
func (h *TelemetryHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    true,
		Message: "Healthcheck status checked successfully",
	})
}
