package services

import (
	"encoding/json"
	"maps"
	"net/http"
	devicesmodels "smarthome/devices/models"
	"smarthome/telemetry/models"
	"strconv"
	"sync"
	"time"
)

type TelemetryService struct {
	mtx                sync.RWMutex
	currentDevicesData map[uint64]*models.SensorTelemetry
	currentSensorsData map[uint64]*models.SensorTelemetry

	sensorsBaseURL     string
	temperatureService *TemperatureService

	httpClient *http.Client
}

func NewTelemetryService(sensorsBaseURL, temperatureBaseURL string) *TelemetryService {
	service := &TelemetryService{
		currentDevicesData: make(map[uint64]*models.SensorTelemetry),
		currentSensorsData: make(map[uint64]*models.SensorTelemetry),
		sensorsBaseURL:     sensorsBaseURL,
		temperatureService: NewTemperatureService(temperatureBaseURL),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	go service.startMetricsCollection()

	return service
}

// GetSensorTelemetry gets current sensor's telemetry
func (s *TelemetryService) GetSensorTelemetry(sensorID uint64) *models.SensorTelemetry {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if data, exists := s.currentDevicesData[sensorID]; exists {
		return data
	}

	return &models.SensorTelemetry{
		SensorID:  sensorID,
		Timestamp: time.Now(),
	}
}

// GetAllSensorsTelemetry gets current sensors' telemetry
func (s *TelemetryService) GetAllSensorsTelemetry() map[uint64]*models.SensorTelemetry {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	result := make(map[uint64]*models.SensorTelemetry)
	maps.Copy(result, s.currentDevicesData)
	return result
}

func (s *TelemetryService) startMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.collectSensorMetrics()
	}
}

func (s *TelemetryService) collectSensorMetrics() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Get all sensors from devices API
	resp, err := s.httpClient.Get(s.sensorsBaseURL)
	if err != nil {
		s.currentDevicesData = map[uint64]*models.SensorTelemetry{}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.currentDevicesData = map[uint64]*models.SensorTelemetry{}
		return
	}

	var sensors []devicesmodels.Sensor
	if err := json.NewDecoder(resp.Body).Decode(&sensors); err != nil {
		s.currentDevicesData = map[uint64]*models.SensorTelemetry{}
		return
	}

	for _, sensor := range sensors {
		// Get current temperature reading
		tempResp, err := s.temperatureService.GetTemperatureByID(strconv.Itoa(sensor.ID))
		if err != nil {
			data := &models.SensorTelemetry{
				SensorID:  uint64(sensor.ID),
				Timestamp: time.Now(),
			}
			s.currentDevicesData[uint64(sensor.ID)] = data
			continue
		}

		data := &models.SensorTelemetry{
			SensorID:   uint64(sensor.ID),
			Timestamp:  time.Now(),
			Temerature: &tempResp.Value,
		}

		s.currentDevicesData[uint64(sensor.ID)] = data
	}
}
