package analyzer

import (
	"log"
	"time"

	"sfsEdgeStore/config"
)

type Analyzer struct {
	config        *config.Config
	isEnabled     bool
	maxTimePerRun time.Duration
}

func NewAnalyzer(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config:        cfg,
		isEnabled:     cfg.EnableAnalyzer,
		maxTimePerRun: time.Duration(cfg.AnalyzerMaxTimePerRun) * time.Millisecond,
	}
}

func (a *Analyzer) Analyze(data []map[string]interface{}, deviceName, reading string) []Alert {
	if !a.isEnabled {
		return nil
	}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if elapsed > a.maxTimePerRun {
			log.Printf("Analyzer execution time exceeded limit: %v", elapsed)
		}
	}()

	var alerts []Alert

	alerts = append(alerts, a.detectThresholdAnomalies(data, deviceName, reading)...)
	alerts = append(alerts, a.detectTrendAnomalies(data, deviceName, reading)...)

	return alerts
}

func (a *Analyzer) detectThresholdAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert {
	var alerts []Alert

	thresholds := a.getThreshold(deviceName, reading)
	if thresholds.Min == 0 && thresholds.Max == 0 {
		return alerts
	}

	for _, d := range data {
		value, ok := d["value"].(float64)
		if !ok {
			continue
		}

		timestamp, ok := d["timestamp"].(time.Time)
		if !ok {
			timestamp = time.Now()
		}

		if thresholds.Max != 0 && value > thresholds.Max {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "threshold_exceeded",
				Message:    "Value exceeded maximum threshold",
				Value:      value,
				Threshold:  thresholds.Max,
				Timestamp:  timestamp,
				Severity:   "high",
			})
		}

		if thresholds.Min != 0 && value < thresholds.Min {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "threshold_below",
				Message:    "Value below minimum threshold",
				Value:      value,
				Threshold:  thresholds.Min,
				Timestamp:  timestamp,
				Severity:   "medium",
			})
		}
	}

	return alerts
}

func (a *Analyzer) detectTrendAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert {
	if len(data) < 3 {
		return nil
	}

	var alerts []Alert

	values := make([]float64, 0, len(data))
	timestamps := make([]time.Time, 0, len(data))

	for _, d := range data {
		value, ok := d["value"].(float64)
		if !ok {
			continue
		}

		timestamp, ok := d["timestamp"].(time.Time)
		if !ok {
			timestamp = time.Now()
		}

		values = append(values, value)
		timestamps = append(timestamps, timestamp)
	}

	if len(values) < 3 {
		return alerts
	}

	for i := 2; i < len(values); i++ {
		change1 := (values[i-1] - values[i-2]) / values[i-2]
		change2 := (values[i] - values[i-1]) / values[i-1]

		if change1 > 0.1 && change2 > 0.1 {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "trend_anomaly",
				Message:    "Rapid continuous increase detected",
				Value:      values[i],
				Threshold:  0,
				Timestamp:  timestamps[i],
				Severity:   "medium",
			})
		}

		if change1 < -0.1 && change2 < -0.1 {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "trend_anomaly",
				Message:    "Rapid continuous decrease detected",
				Value:      values[i],
				Threshold:  0,
				Timestamp:  timestamps[i],
				Severity:   "medium",
			})
		}
	}

	return alerts
}

func (a *Analyzer) Enable() {
	a.isEnabled = true
}

func (a *Analyzer) Disable() {
	a.isEnabled = false
}

func (a *Analyzer) IsEnabled() bool {
	return a.isEnabled
}

func (a *Analyzer) getThreshold(deviceName, reading string) config.ThresholdConfig {
	key := deviceName + ":" + reading
	if t, ok := a.config.AnalyzerThresholds[key]; ok {
		return t
	}

	if t, ok := a.config.AnalyzerThresholds[reading]; ok {
		return t
	}

	return config.ThresholdConfig{}
}
