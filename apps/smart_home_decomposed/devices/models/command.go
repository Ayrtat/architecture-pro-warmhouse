package models

type DeviceCommand struct {
	Name       string        `json:"name"`
	Parameters []interface{} `json:"parameters,omitempty"`
}

type DeviceCommandDesc struct {
	Name           string   `json:"name"`
	ParametersType []string `json:"parameters,omitempty"`
	ReturnType     string   `json:"return_type,omitempty"`
	Description    string   `json:"description"`
}
