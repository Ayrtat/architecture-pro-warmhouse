package handlers

import (
	"net/http"
	"strconv"
	"time"

	"smarthome/services"

	"github.com/gin-gonic/gin"
)

type TelemetryHandler struct {
	telemetryService *services.TelemetryService
}

func NewTelemetryHandler(telemetryService *services.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{
		telemetryService: telemetryService,
	}
}

// RegisterRoutes registers all device-related routes
func (h *TelemetryHandler) RegisterRoutes(router *gin.RouterGroup) {
	telemetry := router.Group("/telemetry")
	{
		telemetry.GET("/health", h.HealthCheck)

		telemetry.GET("/device/:id", h.GetDeviceTelemetry)
		telemetry.GET("/devices", h.GetAllDevicesTelemetry)
	}
}

// GetDeviceTelemetry godoc
// @Summary Get device telemetry
// @Description Get current telemetry data for a specific device
// @Tags telemetry
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Success 200 {object} models.TelemetryData
// @Failure 400 {object} Response
// @Router /telemetry/device/{id} [get]
func (h *TelemetryHandler) GetDeviceTelemetry(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	telemetry := h.telemetryService.GetDeviceTelemetry(uint64(deviceID))
	c.JSON(http.StatusOK, telemetry)
}

// GetAllDevicesTelemetry godoc
// @Summary Get all devices telemetry
// @Description Get current telemetry data for all devices
// @Tags telemetry
// @Accept json
// @Produce json
// @Success 200 {object} map[uint64]*models.TelemetryData
// @Router /telemetry/devices [get]
func (h *TelemetryHandler) GetAllDevicesTelemetry(c *gin.Context) {
	telemetry := h.telemetryService.GetAllDevicesTelemetry()
	c.JSON(http.StatusOK, gin.H{
		"total_devices": len(telemetry),
		"timestamp":     time.Now(),
		"devices":       telemetry,
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
