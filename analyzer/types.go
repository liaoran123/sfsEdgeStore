package analyzer

import (
	"time"
)

type Alert struct {
	DeviceName string    `json:"device_name"`
	Reading    string    `json:"reading"`
	AlertType  string    `json:"alert_type"`
	Message    string    `json:"message"`
	Value      float64   `json:"value"`
	Threshold  float64   `json:"threshold"`
	Timestamp  time.Time `json:"timestamp"`
	Severity   string    `json:"severity"`
}
