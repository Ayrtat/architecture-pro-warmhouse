package services

import "fmt"

// DeviceError represents a device service error
type DeviceError struct {
	Type    DeviceErrorType
	Message string
	Err     error
}

type DeviceErrorType int

const (
	ErrTypeNotFound DeviceErrorType = iota
	ErrTypeInvalidInput
	ErrTypeInternal
	ErrTypeUnimplemented
)

func (e *DeviceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *DeviceError) Unwrap() error {
	return e.Err
}

// NewDeviceError creates a new device error
func NewDeviceError(errType DeviceErrorType, message string, err error) error {
	return &DeviceError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}
