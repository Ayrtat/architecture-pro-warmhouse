package services

import (
	"context"
	"maps"
	"smarthome/db"
	"smarthome/models"
	"strconv"
	"sync"
	"time"
)

type TelemetryService struct {
	currentDataMtx sync.RWMutex
	currentData    map[uint64]*models.TelemetryData

	deviceDB           *db.DB
	temperatureService *TemperatureService
}

func NewTelemetryService(deviceDB *db.DB, temperatureService *TemperatureService) *TelemetryService {
	service := &TelemetryService{
		currentData:        make(map[uint64]*models.TelemetryData),
		deviceDB:           deviceDB,
		temperatureService: temperatureService,
	}

	go service.startMetricsCollection()

	return service
}

// GetDeviceTelemetry gets current device's telemetry
func (s *TelemetryService) GetDeviceTelemetry(deviceID uint64) *models.TelemetryData {
	s.currentDataMtx.RLock()
	defer s.currentDataMtx.RUnlock()

	if data, exists := s.currentData[deviceID]; exists {
		return data
	}

	return &models.TelemetryData{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		State: models.DeviceState{
			Status: models.StatusUnkown,
		},
	}
}

// GetDeviceTelemetry gets current devices' telemetry
func (s *TelemetryService) GetAllDevicesTelemetry() map[uint64]*models.TelemetryData {
	s.currentDataMtx.RLock()
	defer s.currentDataMtx.RUnlock()

	result := make(map[uint64]*models.TelemetryData)
	maps.Copy(result, s.currentData)
	return result
}

func (s *TelemetryService) startMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.collectMetrics()
	}
}

func (s *TelemetryService) collectMetrics() {
	s.currentDataMtx.Lock()
	defer s.currentDataMtx.Unlock()

	// Get all temperature sensors
	sensors, err := s.deviceDB.GetSensors(context.Background())
	if err != nil {
		// no current data
		s.currentData = map[uint64]*models.TelemetryData{}
		return
	}

	for _, sensor := range sensors {
		// Get current temperature reading
		tempResp, err := s.temperatureService.GetTemperatureByID(strconv.Itoa(sensor.ID))
		if err != nil {
			data := &models.TelemetryData{
				DeviceID:  uint64(sensor.ID),
				Timestamp: time.Now(),
				State: models.DeviceState{
					Status: models.StatusUnkown,
				},
			}
			s.currentData[uint64(sensor.ID)] = data
			continue
		}

		data := &models.TelemetryData{
			DeviceID:  uint64(sensor.ID),
			Timestamp: time.Now(),
			State: models.DeviceState{
				Status:      models.DeviceStatus(tempResp.Status),
				LastUpdated: time.Now(),
				Temperature: &tempResp.Value,
			},
		}

		s.currentData[uint64(sensor.ID)] = data
	}
}
