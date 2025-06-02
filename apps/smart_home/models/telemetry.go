package models

import "time"

type TelemetryData struct {
	DeviceID  uint64      `json:"device_id"`
	Timestamp time.Time   `json:"timestamp"`
	State     DeviceState `json:"state"`
}
