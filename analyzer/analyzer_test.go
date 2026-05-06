package analyzer

import (
	"testing"

	"sfsEdgeStore/config"
)

func TestGetThreshold_DevicePriority(t *testing.T) {
	cfg := &config.Config{
		AnalyzerThresholds: map[string]config.ThresholdConfig{
			"temperature":              {Min: -40, Max: 85},
			"Device_B:temperature":     {Min: 35, Max: 37},
			"humidity":                 {Min: 0, Max: 100},
			"Device_A:custom_reading":  {Min: 10, Max: 90},
		},
	}

	a := NewAnalyzer(cfg)
	a.Enable()

	tests := []struct {
		device   string
		reading  string
		wantMin  float64
		wantMax  float64
	}{
		{"Device_B", "temperature", 35, 37},
		{"Device_A", "temperature", -40, 85},
		{"Device_A", "humidity", 0, 100},
		{"Device_C", "temperature", -40, 85},
		{"Device_A", "custom_reading", 10, 90},
		{"Device_A", "unknown", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.device+"_"+tt.reading, func(t *testing.T) {
			tc := a.getThreshold(tt.device, tt.reading)
			if tc.Min != tt.wantMin {
				t.Errorf("getThreshold(%q, %q) Min = %v, want %v", tt.device, tt.reading, tc.Min, tt.wantMin)
			}
			if tc.Max != tt.wantMax {
				t.Errorf("getThreshold(%q, %q) Max = %v, want %v", tt.device, tt.reading, tc.Max, tt.wantMax)
			}
		})
	}
}

func TestAnalyze_ThresholdDetection(t *testing.T) {
	cfg := &config.Config{
		AnalyzerThresholds: map[string]config.ThresholdConfig{
			"temperature": {Min: -40, Max: 85},
		},
	}

	a := NewAnalyzer(cfg)
	a.Enable()

	data := []map[string]interface{}{
		{"value": 50.0},
		{"value": 90.0},
		{"value": 50.0},
		{"value": -50.0},
	}

	alerts := a.Analyze(data, "TestDevice", "temperature")

	var thresholdAlerts []Alert
	for _, alert := range alerts {
		if alert.AlertType == "threshold_exceeded" || alert.AlertType == "threshold_below" {
			thresholdAlerts = append(thresholdAlerts, alert)
		}
	}

	if len(thresholdAlerts) != 2 {
		t.Fatalf("Expected 2 threshold alerts, got %d", len(thresholdAlerts))
	}

	if thresholdAlerts[0].AlertType != "threshold_exceeded" {
		t.Errorf("Expected threshold_exceeded, got %s", thresholdAlerts[0].AlertType)
	}

	if thresholdAlerts[1].AlertType != "threshold_below" {
		t.Errorf("Expected threshold_below, got %s", thresholdAlerts[1].AlertType)
	}
}

func TestAnalyze_TrendDetection(t *testing.T) {
	cfg := &config.Config{
		AnalyzerThresholds: map[string]config.ThresholdConfig{},
	}

	a := NewAnalyzer(cfg)
	a.Enable()

	data := []map[string]interface{}{
		{"value": 10.0},
		{"value": 12.0},
		{"value": 15.0},
	}

	alerts := a.Analyze(data, "TestDevice", "temperature")

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].AlertType != "trend_anomaly" {
		t.Errorf("Expected trend_anomaly, got %s", alerts[0].AlertType)
	}
}

func TestAnalyze_Disabled(t *testing.T) {
	cfg := &config.Config{
		AnalyzerThresholds: map[string]config.ThresholdConfig{
			"temperature": {Min: -40, Max: 85},
		},
	}

	a := NewAnalyzer(cfg)
	a.Disable()

	data := []map[string]interface{}{
		{"value": 999.0},
	}

	alerts := a.Analyze(data, "TestDevice", "temperature")

	if len(alerts) != 0 {
		t.Fatalf("Expected 0 alerts when analyzer is disabled, got %d", len(alerts))
	}
}
