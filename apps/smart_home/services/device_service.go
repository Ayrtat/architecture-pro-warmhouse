package services

import (
	"context"
	"errors"
	"fmt"
	"smarthome/db"
	"smarthome/models"
)

var (
	ErrInvalidFilterFormat     = errors.New("invalid filter format. Expected [key=\"value\"]")
	ErrInvalidCommandArguments = errors.New("invalid command arguments")
	ErrInvalidCommand          = errors.New("invalid command")
	ErrUnimplemented           = errors.New("unimplemented")
)

// DeviceService handles device-related business logic
type DeviceService struct {
	sensorService *SensorService
}

// NewDeviceService creates a new DeviceService instance
func NewDeviceService(db *db.DB, temperatureService *TemperatureService) *DeviceService {
	return &DeviceService{
		sensorService: NewSensorService(db, temperatureService),
	}
}

// GetAllDevices returns all devices with optional filtering
func (s *DeviceService) GetAllDevices(ctx context.Context, filter string) ([]models.Device, error) {
	sensors, err := s.sensorService.GetAllSensors(ctx)
	if err != nil {
		return nil, NewDeviceError(ErrTypeInternal, "failed to get devices", err)
	}

	var devices []models.Device
	for _, sensor := range sensors {
		device := convertSensorToDevice(sensor)
		if filter != "" {
			// [key="value"] format
			if !isValidFilterFormat(filter) {
				return nil, NewDeviceError(ErrTypeInvalidInput, "invalid filter format. Expected [key=\"value\"]", nil)
			}

			key, value := parseFilter(filter)
			if device.Metadata[key] != value {
				continue
			}
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// GetDevice returns a specific device by ID
func (s *DeviceService) GetDevice(ctx context.Context, id int) (*models.Device, error) {
	sensor, err := s.sensorService.GetSensorByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), err)
		}
		return nil, NewDeviceError(ErrTypeInternal, "failed to get device", err)
	}

	device := convertSensorToDevice(sensor)
	return &device, nil
}

// RegisterDevice creates a new device
func (s *DeviceService) RegisterDevice(ctx context.Context, deviceReg models.DeviceRegistration) (*models.Device, error) {
	if deviceReg.Name == "" || deviceReg.Type == "" {
		return nil, NewDeviceError(ErrTypeInvalidInput, "name and type are required", nil)
	}

	sensorCreate := models.SensorCreate{
		Name:     deviceReg.Name,
		Type:     convertDeviceTypeToSensorType(deviceReg.Type),
		Location: deviceReg.Metadata["location"],
		Unit:     deviceReg.Metadata["unit"],
	}

	sensor, err := s.sensorService.CreateSensor(ctx, sensorCreate)
	if err != nil {
		return nil, NewDeviceError(ErrTypeInternal, "failed to register device", err)
	}

	device := convertSensorToDevice(sensor)
	return &device, nil
}

// UnregisterDevice deletes a device
func (s *DeviceService) UnregisterDevice(ctx context.Context, id uint64) error {
	err := s.sensorService.DeleteSensor(ctx, int(id))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), err)
		}
		return NewDeviceError(ErrTypeInternal, "failed to unregister device", err)
	}
	return nil
}

// GetDeviceState returns the current state of a device
func (s *DeviceService) GetDeviceState(ctx context.Context, id uint64) (*models.DeviceState, error) {
	sensor, err := s.sensorService.GetSensorByID(ctx, int(id))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), err)
		}
		return nil, NewDeviceError(ErrTypeInternal, "failed to get device state", err)
	}

	state := &models.DeviceState{
		Status:      models.DeviceStatus(sensor.Status),
		LastUpdated: sensor.LastUpdated,
	}
	return state, nil
}

// UpdateDeviceState updates the state of a device
func (s *DeviceService) UpdateDeviceState(ctx context.Context, id int, state models.DeviceState) error {
	temperature := 0.0
	if state.Temperature != nil {
		temperature = *state.Temperature
	}
	err := s.sensorService.UpdateSensorValue(ctx, id, temperature, string(state.Status))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), err)
		}
		return NewDeviceError(ErrTypeInternal, "failed to update device state", err)
	}
	return nil
}

// SendCommand sends a command to a device
func (s *DeviceService) SendCommand(ctx context.Context, id int, command models.DeviceCommand) error {
	return NewDeviceError(ErrTypeUnimplemented, "command not implemented", nil)
}

// Helpers

func convertSensorToDevice(sensor models.Sensor) models.Device {
	return models.Device{
		ID:            uint64(sensor.ID),
		Name:          sensor.Name,
		Type:          convertSensorTypeToDeviceType(sensor.Type),
		Functionality: []models.DeviceCommandDesc{},
		Metadata: map[string]string{
			"unit":     sensor.Unit,
			"location": sensor.Location,
		},
		LastSeen:  sensor.LastUpdated,
		CreatedAt: sensor.CreatedAt,
		UpdatedAt: sensor.LastUpdated,
	}
}

func convertSensorTypeToDeviceType(typ models.SensorType) models.DeviceType {
	switch typ {
	case models.Temperature:
		return models.Thermostat
	default:
		return models.Unknown
	}
}

func convertDeviceTypeToSensorType(typ models.DeviceType) models.SensorType {
	switch typ {
	case models.Thermostat:
		return models.Temperature
	default:
		return models.SensorType("unknown")
	}
}

func isValidFilterFormat(filter string) bool {
	return len(filter) >= 4 && filter[0] == '[' && filter[len(filter)-1] == ']'
}

func parseFilter(filter string) (key, value string) {
	content := filter[1 : len(filter)-1]
	parts := splitFilter(content)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func splitFilter(content string) []string {
	var parts []string
	var current string
	var inQuotes bool

	for i := 0; i < len(content); i++ {
		switch content[i] {
		case '"':
			inQuotes = !inQuotes
		case '=':
			if !inQuotes {
				parts = append(parts, current)
				current = ""
				continue
			}
		}
		current += string(content[i])
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
