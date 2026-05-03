package monitor

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"sfsEdgeStore/alert"
	"sfsEdgeStore/common"
)

type AlertManager struct {
	alerts         []common.Alert
	mutex          sync.Mutex
	notifier       *alert.Notifier
	alertDedupe    map[string]time.Time
	dedupeMutex    sync.Mutex
	thresholds     AlertThresholds
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:      make([]common.Alert, 0, 100),
		alertDedupe: make(map[string]time.Time),
		thresholds: AlertThresholds{
			HTTPRequestsPerMinute:       1000,
			ErrorsPerMinute:             10,
			DatabaseOperationsPerMinute: 5000,
		},
	}
}

func (m *AlertManager) SetNotifier(notifier *alert.Notifier) {
	m.notifier = notifier
}

func (m *AlertManager) SetThresholds(thresholds AlertThresholds) {
	m.thresholds = thresholds
}

func (m *AlertManager) AddAlert(alert common.Alert) {
	m.mutex.Lock()
	m.alerts = append(m.alerts, alert)
	if len(m.alerts) > 1000 {
		m.alerts = m.alerts[len(m.alerts)-1000:]
	}
	m.mutex.Unlock()

	if m.notifier != nil {
		m.notifier.SendAlert(alert)
	}
}

func (m *AlertManager) RecordError(errorType, message string) {
	if errorType == "invalid_value_discarded" {
		return
	}

	alert := common.Alert{
		Type:      errorType,
		Message:   message,
		Severity:  "critical",
		Timestamp: time.Now(),
		Resolved:  false,
	}

	m.AddAlert(alert)
	log.Printf("Critical error recorded: %s - %s", errorType, message)
}

func (m *AlertManager) GetAlerts() []common.Alert {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.alerts
}

func (m *AlertManager) TrimAlerts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.alerts) > 1000 {
		m.alerts = m.alerts[len(m.alerts)-1000:]
	}
}

func (m *AlertManager) GetAlertGroups() []AlertGroup {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	groupMap := make(map[string]*AlertGroup)

	for i := len(m.alerts) - 1; i >= 0; i-- {
		alert := m.alerts[i]

		if IsSystemResourceAlert(alert.Type) {
			continue
		}

		key := alert.Type

		if group, exists := groupMap[key]; exists {
			deviceName := extractDeviceNameFromMessage(alert.Message)
			if deviceName != "" && deviceName != "未知设备" {
				found := false
				for _, d := range group.Devices {
					if d == deviceName {
						found = true
						break
					}
				}
				if !found && len(group.Devices) < 10 {
					group.Devices = append(group.Devices, deviceName)
					group.DeviceCount = len(group.Devices)
				}
			}
			group.Count++
			if alert.Timestamp.After(group.LastTime) {
				group.LastTime = alert.Timestamp
			}
		} else {
			deviceName := extractDeviceNameFromMessage(alert.Message)
			deviceType := extractDeviceTypeFromMessage(alert.Message)

			devices := []string{}
			if deviceName != "" {
				devices = []string{deviceName}
			}

			groupMap[key] = &AlertGroup{
				Type:        alert.Type,
				Message:     generateGroupMessage(alert.Type, deviceType),
				DeviceCount: len(devices),
				Devices:     devices,
				LastTime:    alert.Timestamp,
				Severity:    alert.Severity,
				Count:       1,
			}
		}
	}

	groups := make([]AlertGroup, 0, len(groupMap))
	for _, g := range groupMap {
		groups = append(groups, *g)
	}

	for i := 0; i < len(groups)-1; i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[j].LastTime.After(groups[i].LastTime) {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}

	return groups
}

func (m *AlertManager) ShouldSendAlert(alertType, deviceName string) bool {
	m.dedupeMutex.Lock()
	defer m.dedupeMutex.Unlock()

	key := alertType + ":" + deviceName
	lastTime, exists := m.alertDedupe[key]
	if exists && time.Since(lastTime) < 10*time.Minute {
		return false
	}

	m.alertDedupe[key] = time.Now()
	return true
}

func IsSystemResourceAlert(alertType string) bool {
	resourceAlerts := []string{
		"high_cpu_usage",
		"high_memory_usage",
		"high_goroutine_count",
	}
	for _, a := range resourceAlerts {
		if a == alertType {
			return true
		}
	}
	return false
}

func extractDeviceNameFromMessage(message string) string {
	if len(message) >= 8 && message[:8] == "Device " {
		remainder := message[8:]
		for i, c := range remainder {
			if c == ' ' {
				if i > 0 {
					return remainder[:i]
				}
				break
			}
		}
	}

	if len(message) >= 15 && message[:15] == "Device temperature" {
		colonIndex := -1
		for i, c := range message {
			if c == ':' {
				colonIndex = i
				break
			}
		}

		if colonIndex != -1 && colonIndex+2 < len(message) {
			remainder := message[colonIndex+2:]
			dashIndex := -1
			for i := 0; i < len(remainder)-2; i++ {
				if remainder[i:i+3] == " - " {
					dashIndex = i
					break
				}
			}

			if dashIndex != -1 {
				return remainder[:dashIndex]
			}
		}
	}

	return ""
}

func extractDeviceTypeFromMessage(message string) string {
	deviceName := extractDeviceNameFromMessage(message)
	if deviceName == "" {
		return ""
	}
	for i := len(deviceName) - 1; i >= 0; i-- {
		if deviceName[i] == '-' {
			return deviceName[:i]
		}
	}
	return deviceName
}

var alertMessageTemplates = map[string]string{
	"device_temperature_over_limit": "%s Temperature Over Limit",
	"device_humidity_over_limit":    "%s Humidity Over Limit",
	"device_data_anomaly":           "%s Data Anomaly",
	"http_requests":                 "HTTP Request Rate Too High",
	"errors":                        "Error Rate Too High",
	"database_operations":           "Database Operations Rate Too High",
	"device_offline":                "Device Offline",
	"device_online":                 "Device Online",
	"device_data_trend":             "%s Data Trend Anomaly",
}

func generateGroupMessage(alertType, deviceType string) string {
	if template, ok := alertMessageTemplates[alertType]; ok {
		deviceTypeName := getDeviceTypeName(deviceType)
		if strings.Contains(template, "%s") {
			return fmt.Sprintf(template, deviceTypeName)
		} else {
			return template
		}
	}
	return fmt.Sprintf("%s Alert", getDeviceTypeName(deviceType))
}

func getDeviceTypeName(deviceType string) string {
	switch deviceType {
	case "temperature-sensor":
		return "Temperature Sensor"
	case "pressure-sensor":
		return "Pressure Sensor"
	case "vibration-sensor":
		return "Vibration Sensor"
	case "power-meter":
		return "Power Meter"
	case "flow-meter":
		return "Flow Meter"
	default:
		return deviceType
	}
}
