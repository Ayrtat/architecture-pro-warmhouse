package handlers

import (
	"errors"
	"net/http"
	"smarthome/devices/models"
	"smarthome/devices/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct {
	deviceService *services.DeviceService
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewDeviceHandler creates a new DeviceHandler instance
func NewDeviceHandler(deviceService *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// RegisterRoutes registers all device-related routes
func (h *DeviceHandler) RegisterRoutes(router *gin.RouterGroup) {
	devices := router.Group("/devices")
	{
		devices.GET("/health", h.HealthCheck)

		devices.GET("", h.GetAllDevices)
		devices.GET("/:id", h.GetDevice)
		devices.POST("", h.RegisterDevice)
		devices.DELETE("/:id", h.UnregisterDevice)

		devices.GET("/:id/state", h.GetDeviceState)
		devices.PUT("/:id/state", h.UpdateDeviceState)

		devices.POST("/:id/commands", h.SendCommand)
	}
}

// mapDeviceError maps device service errors to HTTP status codes
func mapDeviceError(err error) (int, string) {
	var deviceErr *services.DeviceError
	if errors.As(err, &deviceErr) {
		switch deviceErr.Type {
		case services.ErrTypeNotFound:
			return http.StatusNotFound, deviceErr.Error()
		case services.ErrTypeInvalidInput:
			return http.StatusBadRequest, deviceErr.Error()
		case services.ErrTypeInternal:
			return http.StatusInternalServerError, deviceErr.Error()
		case services.ErrTypeUnimplemented:
			return http.StatusUnprocessableEntity, deviceErr.Error()
		}
	}
	return http.StatusInternalServerError, "Internal server error"
}

// GetAllDevices godoc
// @Summary Get all devices
// @Description Get a list of all registered devices with optional filtering
// @Tags devices
// @Accept json
// @Produce json
// @Param filter query string false "Filter in format [key=\"value\"]"
// @Success 200 {object} Response{data=[]models.Device}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /devices [get]
func (h *DeviceHandler) GetAllDevices(c *gin.Context) {
	filter := c.Query("filter")
	devices, err := h.deviceService.GetAllDevices(c.Request.Context(), filter)
	if err != nil {
		status, message := mapDeviceError(err)
		c.JSON(status, Response{
			Success: false,
			Error:   message,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    devices,
		Message: "Devices retrieved successfully",
	})
}

// GetDevice godoc
// @Summary Get a specific device
// @Description Get details of a specific device by ID
// @Tags devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Success 200 {object} Response{data=models.Device}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid device ID format",
		})
		return
	}

	device, err := h.deviceService.GetDevice(c.Request.Context(), id)
	if err != nil {
		status, message := mapDeviceError(err)
		c.JSON(status, Response{
			Success: false,
			Error:   message,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    device,
		Message: "Device retrieved successfully",
	})
}

// RegisterDevice godoc
// @Summary Register a new device
// @Description Register a new device in the system
// @Tags devices
// @Accept json
// @Produce json
// @Param device body models.DeviceRegistration true "Device registration data"
// @Success 201 {object} Response{data=models.Device}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /devices [post]
func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var deviceReg models.DeviceRegistration
	if err := c.ShouldBindJSON(&deviceReg); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	device, err := h.deviceService.RegisterDevice(c.Request.Context(), deviceReg)
	if err != nil {
		status, message := mapDeviceError(err)
		c.JSON(status, Response{
			Success: false,
			Error:   message,
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    device,
		Message: "Device registered successfully",
	})
}

// UnregisterDevice godoc
// @Summary Unregister a device
// @Description Remove a device from the system
// @Tags devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /devices/{id} [delete]
func (h *DeviceHandler) UnregisterDevice(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid device ID",
		})
		return
	}

	err = h.deviceService.UnregisterDevice(c.Request.Context(), uint64(id))
	if err != nil {
		status, message := mapDeviceError(err)
		c.JSON(status, Response{
			Success: false,
			Error:   message,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Device unregistered successfully",
	})
}

// GetDeviceState godoc
// @Summary Get device state
// @Description Get the current state of a device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Success 200 {object} Response{data=models.DeviceState}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /devices/{id}/state [get]
func (h *DeviceHandler) GetDeviceState(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid device ID",
		})
		return
	}

	state, err := h.deviceService.GetDeviceState(c.Request.Context(), uint64(id))
	if err != nil {
		status, message := mapDeviceError(err)
		c.JSON(status, Response{
			Success: false,
			Error:   message,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    state,
		Message: "Device state retrieved successfully",
	})
}

// UpdateDeviceState godoc
// @Summary Update device state
// @Description Update the state of a device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Param state body models.DeviceState true "New device state"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /devices/{id}/state [put]
func (h *DeviceHandler) UpdateDeviceState(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid device ID",
		})
		return
	}

	var state models.DeviceState
	if err := c.ShouldBindJSON(&state); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	err = h.deviceService.UpdateDeviceState(c.Request.Context(), id, state)
	if err != nil {
		status, message := mapDeviceError(err)
		c.JSON(status, Response{
			Success: false,
			Error:   message,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Device state updated successfully",
	})
}

// SendCommand godoc
// @Summary Send command to device
// @Description Send a command to a device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Param command body models.DeviceCommand true "Command to execute"
// @Success 200 {object} Response{data=models.DeviceCommand}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /devices/{id}/commands [post]
func (h *DeviceHandler) SendCommand(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid device ID",
		})
		return
	}

	var command models.DeviceCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	err = h.deviceService.SendCommand(c.Request.Context(), id, command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    command,
		Message: "Command executed successfully",
	})
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=bool}
// @Router /devices/health [get]
func (h *DeviceHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    true,
		Message: "Healthcheck status checked successfully",
	})
}
