package analyzer

import (
	"time"
)

// AnalysisResult 分析结果
type AnalysisResult struct {
	DeviceName   string    `json:"device_name"`
	Reading      string    `json:"reading"`
	AnalysisType string    `json:"analysis_type"`
	Value        float64   `json:"value"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Timestamp    time.Time `json:"timestamp"`
}

// Alert 告警信息
type Alert struct {
	DeviceName  string    `json:"device_name"`
	Reading     string    `json:"reading"`
	AlertType   string    `json:"alert_type"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
