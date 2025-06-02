package models

import "time"

type DeviceType string

const (
	Thermostat DeviceType = "thermostat"
	Unknown    DeviceType = "unknown"
)

type DeviceStatus string

const (
	StatusInactive DeviceStatus = "inactive"
	StatusActive   DeviceStatus = "active"
	StatusUnkown   DeviceStatus = "unknown"
)

type DeviceState struct {
	Status      DeviceStatus `json:"status,omitempty"`
	LastUpdated time.Time    `json:"last_updated"`
	Temperature *float64     `json:"temperature_data,omitempty"`
}

type DeviceRegistration struct {
	Name          string              `json:"name" validate:"required"`
	HubID         uint64              `json:"hub_id" db:"hub_id"`
	Type          DeviceType          `json:"type" validate:"required"`
	SerialNumber  string              `json:"serial_number" db:"serial_number"`
	Functionality []DeviceCommandDesc `json:"functionality"`
	Metadata      map[string]string   `json:"metadata"`
}

type Device struct {
	ID            uint64              `json:"id" db:"id"`
	Name          string              `json:"name" db:"name"`
	Type          DeviceType          `json:"type" db:"type"`
	SerialNumber  string              `json:"serial_number" db:"serial_number"`
	Functionality []DeviceCommandDesc `json:"functionality" db:"functionality"`
	Metadata      map[string]string   `json:"metadata" db:"metadata"`
	LastSeen      time.Time           `json:"last_seen" db:"last_seen"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" db:"updated_at"`
}
