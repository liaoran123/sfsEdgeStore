package alert

// AlertType 告警类型
type AlertType string

// 设备相关告警类型（仅这些告警会显示在告警中心）
const (
	AlertTypeDeviceTemperatureOverLimit AlertType = "device_temperature_over_limit"
	AlertTypeDeviceHumidityOverLimit    AlertType = "device_humidity_over_limit"
	AlertTypeDeviceDataAnomaly          AlertType = "device_data_anomaly"
	AlertTypeDeviceDataTrend            AlertType = "device_data_trend"
	AlertTypeDeviceOffline              AlertType = "device_offline"
	AlertTypeDeviceOnline               AlertType = "device_online"
)

// 非设备相关告警类型（这些告警将被过滤，不显示在告警中心）
const (
	AlertTypeHTTPRequests       AlertType = "http_requests"
	AlertTypeErrors             AlertType = "errors"
	AlertTypeDatabaseOperations AlertType = "database_operations"
	AlertTypeMemoryOverLimit    AlertType = "memory_over_limit"
	AlertTypeCPUOverLimit       AlertType = "cpu_over_limit"
)

// IsDeviceAlert 检查是否为设备相关告警
func IsDeviceAlert(alertType string) bool {
	deviceAlerts := map[string]bool{
		"device_temperature_over_limit": true,
		"device_humidity_over_limit":    true,
		"device_data_anomaly":           true,
		"device_data_trend":             true,
		"device_offline":                true,
		"device_online":                 true,
	}
	return deviceAlerts[alertType]
}

// IsSystemResourceAlert 检查是否为系统资源相关告警（已被过滤，不显示）
func IsSystemResourceAlert(alertType string) bool {
	return !IsDeviceAlert(alertType)
}
