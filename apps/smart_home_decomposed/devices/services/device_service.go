package services

import (
	"context"
	"errors"
	"fmt"
	"smarthome/devices/models"
	"sync"
	"time"
)

var (
	ErrInvalidFilterFormat     = errors.New("invalid filter format. Expected [key=\"value\"]")
	ErrInvalidCommandArguments = errors.New("invalid command arguments")
	ErrInvalidCommand          = errors.New("invalid command")
	ErrUnimplemented           = errors.New("unimplemented")
)

// DeviceService handles device-related business logic
type DeviceService struct {
	devices map[uint64]models.Device
	mu      sync.RWMutex
}

// NewDeviceService creates a new DeviceService instance
func NewDeviceService() *DeviceService {
	return &DeviceService{
		devices: make(map[uint64]models.Device),
	}
}

// GetAllDevices returns all devices with optional filtering
func (s *DeviceService) GetAllDevices(ctx context.Context, filter string) ([]models.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var devices []models.Device
	for _, device := range s.devices {
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[uint64(id)]
	if !exists {
		return nil, NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), nil)
	}

	return &device, nil
}

// RegisterDevice creates a new device
func (s *DeviceService) RegisterDevice(ctx context.Context, deviceReg models.DeviceRegistration) (*models.Device, error) {
	if deviceReg.Name == "" || deviceReg.Type == "" {
		return nil, NewDeviceError(ErrTypeInvalidInput, "name and type are required", nil)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a new ID
	id := uint64(len(s.devices) + 1)
	now := time.Now()

	device := models.Device{
		ID:            id,
		Name:          deviceReg.Name,
		Type:          deviceReg.Type,
		Functionality: []models.DeviceCommandDesc{},
		Metadata:      deviceReg.Metadata,
		LastSeen:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.devices[id] = device
	return &device, nil
}

// UnregisterDevice deletes a device
func (s *DeviceService) UnregisterDevice(ctx context.Context, id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[id]; !exists {
		return NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), nil)
	}

	delete(s.devices, id)
	return nil
}

// GetDeviceState returns the current state of a device
func (s *DeviceService) GetDeviceState(ctx context.Context, id uint64) (*models.DeviceState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), nil)
	}

	status := models.StatusUnkown
	if s, ok := device.Metadata["status"]; ok {
		status = models.DeviceStatus(s)
	}

	state := &models.DeviceState{
		Status:      status,
		LastUpdated: device.LastSeen,
	}

	// Parse temperature if it exists
	if tempStr, ok := device.Metadata["temperature"]; ok {
		var temp float64
		if _, err := fmt.Sscanf(tempStr, "%f", &temp); err == nil {
			state.Temperature = &temp
		}
	}

	return state, nil
}

// UpdateDeviceState updates the state of a device
func (s *DeviceService) UpdateDeviceState(ctx context.Context, id int, state models.DeviceState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, exists := s.devices[uint64(id)]
	if !exists {
		return NewDeviceError(ErrTypeNotFound, fmt.Sprintf("device %d not found", id), nil)
	}

	if device.Metadata == nil {
		device.Metadata = make(map[string]string)
	}

	device.Metadata["status"] = string(state.Status)
	device.LastSeen = time.Now()

	if state.Temperature != nil {
		device.Metadata["temperature"] = fmt.Sprintf("%f", *state.Temperature)
	}

	s.devices[uint64(id)] = device
	return nil
}

// SendCommand sends a command to a device
func (s *DeviceService) SendCommand(ctx context.Context, id int, command models.DeviceCommand) error {
	return NewDeviceError(ErrTypeUnimplemented, "command not implemented", nil)
}

// Helpers

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
