package models

import (
	"time"
)

type SensorTelemetry struct {
	SensorID   uint64    `json:"sensor_id"`
	Temerature *float64  `json:"temperature"`
	Timestamp  time.Time `json:"timestamp"`
}
