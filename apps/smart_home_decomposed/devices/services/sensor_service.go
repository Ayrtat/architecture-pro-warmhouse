package services

import (
	"context"
	"fmt"
	"log"
	"smarthome/devices/db"
	"smarthome/devices/models"
)

// SensorService handles sensor business logic
type SensorService struct {
	db                 *db.DB
	temperatureService *TemperatureService
}

// NewSensorService creates a new SensorService
func NewSensorService(db *db.DB, temperatureService *TemperatureService) *SensorService {
	return &SensorService{
		db:                 db,
		temperatureService: temperatureService,
	}
}

// GetAllSensors returns all sensors with real-time data for temperature sensors
func (s *SensorService) GetAllSensors(ctx context.Context) ([]models.Sensor, error) {
	sensors, err := s.db.GetSensors(ctx)
	if err != nil {
		return nil, err
	}

	// Update temperature sensors with real-time data
	for i, sensor := range sensors {
		if sensor.Type == models.Temperature {
			tempData, err := s.temperatureService.GetTemperatureByID(fmt.Sprintf("%d", sensor.ID))
			if err == nil {
				sensors[i].Value = tempData.Value
				sensors[i].Status = tempData.Status
				sensors[i].LastUpdated = tempData.Timestamp
				log.Printf("Updated temperature data for sensor %d from external API", sensor.ID)
			} else {
				log.Printf("Failed to fetch temperature data for sensor %d: %v", sensor.ID, err)
			}
		}
	}

	return sensors, nil
}

// GetSensorByID returns a sensor by ID with real-time data for temperature sensors
func (s *SensorService) GetSensorByID(ctx context.Context, id int) (models.Sensor, error) {
	sensor, err := s.db.GetSensorByID(ctx, id)
	if err != nil {
		return models.Sensor{}, err
	}

	// If this is a temperature sensor, fetch real-time data
	if sensor.Type == models.Temperature {
		tempData, err := s.temperatureService.GetTemperatureByID(fmt.Sprintf("%d", sensor.ID))
		if err == nil {
			sensor.Value = tempData.Value
			sensor.Status = tempData.Status
			sensor.LastUpdated = tempData.Timestamp
			log.Printf("Updated temperature data for sensor %d from external API", sensor.ID)
		} else {
			log.Printf("Failed to fetch temperature data for sensor %d: %v", sensor.ID, err)
		}
	}

	return sensor, nil
}

// CreateSensor creates a new sensor
func (s *SensorService) CreateSensor(ctx context.Context, sensorCreate models.SensorCreate) (models.Sensor, error) {
	return s.db.CreateSensor(ctx, sensorCreate)
}

// UpdateSensor updates an existing sensor
func (s *SensorService) UpdateSensor(ctx context.Context, id int, sensorUpdate models.SensorUpdate) (models.Sensor, error) {
	return s.db.UpdateSensor(ctx, id, sensorUpdate)
}

// DeleteSensor deletes a sensor
func (s *SensorService) DeleteSensor(ctx context.Context, id int) error {
	return s.db.DeleteSensor(ctx, id)
}

// UpdateSensorValue updates a sensor's value and status
func (s *SensorService) UpdateSensorValue(ctx context.Context, id int, value float64, status string) error {
	return s.db.UpdateSensorValue(ctx, id, value, status)
}

// GetTemperatureByLocation gets temperature data for a specific location
func (s *SensorService) GetTemperatureByLocation(ctx context.Context, location string) (*models.Sensor, error) {
	tempData, err := s.temperatureService.GetTemperature(location)
	if err != nil {
		return nil, err
	}

	return &models.Sensor{
		Type:        models.Temperature,
		Location:    location,
		Value:       tempData.Value,
		Status:      tempData.Status,
		LastUpdated: tempData.Timestamp,
		Unit:        tempData.Unit,
	}, nil
}
