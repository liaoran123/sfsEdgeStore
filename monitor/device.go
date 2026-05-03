package monitor

import (
	"fmt"
	"math"
	"sync"
	"time"

	"sfsEdgeStore/common"
	"sfsEdgeStore/config"
	"sfsEdgeStore/device"
)

type DeviceTracker struct {
	deviceStatus   map[string]DeviceStatus
	mutex          sync.Mutex
	config         *config.Config
	profileManager *device.ProfileManager
	alertManager   *AlertManager
}

func NewDeviceTracker(cfg *config.Config, alertManager *AlertManager) *DeviceTracker {
	profileManager := device.GetProfileManager()
	if err := profileManager.LoadProfiles("device_profiles"); err != nil {
		// Profile loading failure is not fatal
	}

	return &DeviceTracker{
		deviceStatus:   make(map[string]DeviceStatus),
		config:         cfg,
		profileManager: profileManager,
		alertManager:   alertManager,
	}
}

func (t *DeviceTracker) UpdateStatus(deviceName, reading string, value float64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	status, exists := t.deviceStatus[deviceName]
	if !exists {
		status = DeviceStatus{
			LastActive:  now,
			IsOnline:    true,
			DataHistory: make([]float64, 0, 10),
		}
	}

	var dataChange float64
	if status.DataCount > 0 && status.LastDataValue != 0 {
		dataChange = math.Abs(value - status.LastDataValue)
		status.LastDataChange = dataChange
	}

	status.DataHistory = append(status.DataHistory, value)
	if len(status.DataHistory) > 10 {
		status.DataHistory = status.DataHistory[len(status.DataHistory)-10:]
	}

	t.detectAnomaly(deviceName, value, dataChange, &status, now)
	t.detectTrend(deviceName, &status, now)
	t.checkThresholds(deviceName, reading, value, now)

	status.LastActive = now
	status.IsOnline = true
	status.DataCount++
	status.LastDataValue = value
	status.LastDataTime = now

	t.deviceStatus[deviceName] = status
}

func (t *DeviceTracker) GetStatus() map[string]DeviceStatus {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	statusCopy := make(map[string]DeviceStatus)
	for k, v := range t.deviceStatus {
		statusCopy[k] = v
	}
	return statusCopy
}

func (t *DeviceTracker) CheckStatus() []common.Alert {
	var newAlerts []common.Alert
	now := time.Now()

	offlineThreshold := 300
	if t.config != nil && t.config.DeviceOfflineThreshold > 0 {
		offlineThreshold = t.config.DeviceOfflineThreshold
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	for deviceName, status := range t.deviceStatus {
		if now.Sub(status.LastActive) > time.Duration(offlineThreshold)*time.Second && status.IsOnline {
			if t.alertManager.ShouldSendAlert("device_offline", deviceName) {
				alert := common.Alert{
					Type:      "device_offline",
					Message:   fmt.Sprintf("Device %s is offline", deviceName),
					Severity:  "warning",
					Timestamp: now,
					Resolved:  false,
				}
				newAlerts = append(newAlerts, alert)

				status.IsOnline = false
				status.LastAlertTime = now
				t.deviceStatus[deviceName] = status
			}
		}

		if now.Sub(status.LastActive) < time.Duration(offlineThreshold)*time.Second && !status.IsOnline {
			alert := common.Alert{
				Type:      "device_online",
				Message:   fmt.Sprintf("Device %s is back online", deviceName),
				Severity:  "info",
				Timestamp: now,
				Resolved:  false,
			}
			newAlerts = append(newAlerts, alert)

			status.IsOnline = true
			t.deviceStatus[deviceName] = status
		}
	}

	return newAlerts
}

func (t *DeviceTracker) CleanupInactive() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	for deviceName, status := range t.deviceStatus {
		if now.Sub(status.LastActive) > 30*time.Minute {
			delete(t.deviceStatus, deviceName)
		}
	}
}

func (t *DeviceTracker) detectAnomaly(deviceName string, value float64, dataChange float64, status *DeviceStatus, now time.Time) {
	if status.DataCount <= 2 || status.LastDataValue == 0 {
		return
	}

	threshold := 50.0
	if t.config != nil && t.config.DataAnomalyThreshold > 0 {
		threshold = float64(t.config.DataAnomalyThreshold)
	}

	if dataChange/status.LastDataValue <= threshold/100 {
		return
	}

	t.alertManager.AddAlert(common.Alert{
		Type:      "device_data_anomaly",
		Message:   fmt.Sprintf("Device %s data anomaly: value changed from %.2f to %.2f", deviceName, status.LastDataValue, value),
		Severity:  "warning",
		Timestamp: now,
		Resolved:  false,
	})
}

func (t *DeviceTracker) detectTrend(deviceName string, status *DeviceStatus, now time.Time) {
	minPoints := 5
	if t.config != nil && t.config.DataTrendMinPoints > 0 {
		minPoints = t.config.DataTrendMinPoints
	}

	if len(status.DataHistory) < minPoints {
		return
	}

	isIncreasing, isDecreasing := true, true
	for i := 1; i < len(status.DataHistory); i++ {
		if status.DataHistory[i] <= status.DataHistory[i-1] {
			isIncreasing = false
		}
		if status.DataHistory[i] >= status.DataHistory[i-1] {
			isDecreasing = false
		}
	}

	if !isIncreasing && !isDecreasing {
		return
	}

	trend := "increasing"
	if isDecreasing {
		trend = "decreasing"
	}

	t.alertManager.AddAlert(common.Alert{
		Type:      "device_data_trend",
		Message:   fmt.Sprintf("Device %s data trend: continuous %s for %d consecutive readings", deviceName, trend, minPoints),
		Severity:  "info",
		Timestamp: now,
		Resolved:  false,
	})
}

func (t *DeviceTracker) checkThresholds(deviceName string, reading string, value float64, now time.Time) {
	if t.profileManager == nil {
		return
	}

	profileName, found := t.profileManager.FindProfileByDeviceName(deviceName)
	if !found {
		return
	}

	checks := []struct {
		resourceName string
		alertType    string
		severity     string
	}{
		{"Temperature", "device_temperature_over_limit", "critical"},
		{"Humidity", "device_humidity_over_limit", "warning"},
	}

	for _, check := range checks {
		if threshold, exists := t.profileManager.GetResourceThreshold(profileName, check.resourceName); exists {
			if value > threshold {
				t.alertManager.AddAlert(common.Alert{
					Type:      check.alertType,
					Message:   fmt.Sprintf("Device %s %s over limit: %.2f > %.2f", deviceName, check.resourceName, value, threshold),
					Severity:  check.severity,
					Timestamp: now,
					Resolved:  false,
				})
			}
		}
	}
}
