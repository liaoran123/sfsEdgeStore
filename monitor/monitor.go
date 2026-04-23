package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sfsEdgeStore/alert"
	sfsAlert "sfsEdgeStore/alert"
	"sfsEdgeStore/common"
	"sfsEdgeStore/config"
	"sfsEdgeStore/device"

	"github.com/liaoran123/sfsDb/management/monitor"
)

// Monitor 监控管理器
type Monitor struct {
	monitorManager  *monitor.MonitorManager // 监控管理器
	metrics         InternalMetrics         // 内部监控指标（使用atomic）
	startTime       time.Time               // 启动时间
	alertThresholds AlertThresholds         // 告警阈值
	alerts          []common.Alert          // 告警列表
	lastMetrics     InternalMetrics         // 上次收集的指标
	lastCollectTime time.Time               // 上次收集时间
	mutex           sync.Mutex              // 保护alerts切片的互斥锁
	notifier        *alert.Notifier         // 告警通知器
	deviceStatus    map[string]DeviceStatus // 设备状态映射
	deviceMutex     sync.Mutex              // 保护deviceStatus的互斥锁
	config          *config.Config          // 配置
	profileManager  *device.ProfileManager  // 设备配置管理器
	// 告警去重机制
	alertDedupe map[string]time.Time // 告警去重映射 (key: type+deviceName)
	dedupeMutex sync.Mutex           // 保护去重映射
}

// 告警分组信息
type AlertGroup struct {
	Type        string    // 告警类型
	Message     string    // 告警消息
	DeviceCount int       // 受影响设备数
	Devices     []string  // 受影响设备列表
	LastTime    time.Time // 最新告警时间
	Severity    string    // 严重级别
	Count       int       // 告警总数
}

// GetAlertGroups 获取分组后的告警列表
func (m *Monitor) GetAlertGroups() []AlertGroup {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 按告警类型分组
	groupMap := make(map[string]*AlertGroup)

	for i := len(m.alerts) - 1; i >= 0; i-- {
		alert := m.alerts[i]

		// 跳过系统资源相关的告警
		if sfsAlert.IsSystemResourceAlert(alert.Type) {
			continue
		}

		// 直接使用 alert.Type 作为分组键
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

	// 转换为切片并按时间排序
	groups := make([]AlertGroup, 0, len(groupMap))
	for _, g := range groupMap {
		groups = append(groups, *g)
	}

	// 按最后时间排序（最新的在前）
	for i := 0; i < len(groups)-1; i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[j].LastTime.After(groups[i].LastTime) {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}

	return groups
}

// extractDeviceNameFromMessage 从消息中提取设备名称
func extractDeviceNameFromMessage(message string) string {
	// 处理多种消息格式：
	// 格式1: "Device temperature-sensor-001 temperature over limit: 55.00 > 50.00"
	// 格式2: "Device temperature over limit: temperature-sensor-001 - Device temperature-sensor-001 temperature over limit: 57.35 > 50.00"

	// 首先尝试格式1
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

	// 尝试格式2
	if len(message) >= 15 && message[:15] == "Device temperature" {
		// 查找 ": " 分隔符
		colonIndex := -1
		for i, c := range message {
			if c == ':' {
				colonIndex = i
				break
			}
		}

		if colonIndex != -1 && colonIndex+2 < len(message) {
			remainder := message[colonIndex+2:]
			// 查找 " - " 分隔符
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

	// 尝试其他可能的格式，如 "Device data anomaly: value changed from 1258.85 to 623.14"
	if len(message) >= 13 && message[:13] == "Device data anomaly" {
		// 这种格式可能不包含设备名称，返回空
		return ""
	}

	return ""
}

// extractDeviceTypeFromMessage 从消息中提取设备类型
func extractDeviceTypeFromMessage(message string) string {
	deviceName := extractDeviceNameFromMessage(message)
	if deviceName == "" {
		return ""
	}
	// 移除编号，如 temperature-sensor-001 -> temperature-sensor
	for i := len(deviceName) - 1; i >= 0; i-- {
		if deviceName[i] == '-' {
			return deviceName[:i]
		}
	}
	return deviceName
}

// 告警类型到消息模板的映射
var alertMessageTemplates = map[string]string{
	"device_temperature_over_limit": "%s 温度超限",
	"device_humidity_over_limit":    "%s 湿度超限",
	"device_data_anomaly":           "%s 数据异常",
	"http_requests":                 "HTTP请求频率过高",
	"errors":                        "错误率过高",
	"database_operations":           "数据库操作频率过高",
	"device_offline":                "设备离线",
	"device_online":                 "设备上线",
	"device_data_trend":             "%s 数据趋势异常",
}

// 生成分组消息
func generateGroupMessage(alertType, deviceType string) string {
	if template, ok := alertMessageTemplates[alertType]; ok {
		deviceTypeName := getDeviceTypeName(deviceType)
		// 检查模板是否包含 %s 占位符
		if strings.Contains(template, "%s") {
			return fmt.Sprintf(template, deviceTypeName)
		} else {
			return template
		}
	}
	return fmt.Sprintf("%s 告警", getDeviceTypeName(deviceType))
}

// 获取设备类型中文名
func getDeviceTypeName(deviceType string) string {
	switch deviceType {
	case "temperature-sensor":
		return "温度传感器"
	case "pressure-sensor":
		return "压力传感器"
	case "vibration-sensor":
		return "振动传感器"
	case "power-meter":
		return "功率计"
	case "flow-meter":
		return "流量计"
	default:
		return deviceType
	}
}

// DeviceStatus 设备状态
type DeviceStatus struct {
	LastActive     time.Time // 最后活跃时间
	IsOnline       bool      // 是否在线
	LastAlertTime  time.Time // 最后告警时间
	DataCount      int64     // 数据计数
	LastDataValue  float64   // 最后数据值
	LastDataTime   time.Time // 最后数据时间
	DataHistory    []float64 // 数据历史记录（用于异常检测）
	LastDataChange float64   // 最后数据变化量
}

// InternalMetrics 内部监控指标（使用atomic类型）
type InternalMetrics struct {
	System      SystemMetrics              `json:"system"`
	Database    DatabaseMetrics            `json:"database"`
	Application InternalApplicationMetrics `json:"application"`
}

// InternalApplicationMetrics 内部应用指标（使用atomic类型）
type InternalApplicationMetrics struct {
	MQTTMessagesReceived  atomic.Int64
	MQTTMessagesProcessed atomic.Int64
	HTTPRequests          atomic.Int64
	DatabaseOperations    atomic.Int64
	Errors                atomic.Int64
	MQTTConnectionStatus  atomic.Bool  // MQTT连接状态
	DataReceivedBytes     atomic.Int64 // 数据接收字节数
	DataStoredBytes       atomic.Int64 // 数据存储字节数
}

// Metrics 导出的监控指标（使用普通类型）
type Metrics struct {
	System      SystemMetrics      `json:"system"`
	Database    DatabaseMetrics    `json:"database"`
	Application ApplicationMetrics `json:"application"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`      // CPU使用率
	MemoryUsage float64 `json:"memory_usage"`   // 内存使用率
	Goroutines  int     `json:"goroutines"`     //  Goroutine数量
	Uptime      int64   `json:"uptime_seconds"` // 运行时间（秒）
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	KeyStats   map[int]interface{} `json:"key_stats"`   // 键值变化统计
	IndexStats map[int]interface{} `json:"index_stats"` // 索引统计
}

// ApplicationMetrics 应用指标（使用普通类型用于导出）
type ApplicationMetrics struct {
	MQTTMessagesReceived  int64 `json:"mqtt_messages_received"`  // MQTT消息接收计数
	MQTTMessagesProcessed int64 `json:"mqtt_messages_processed"` // MQTT消息处理计数
	HTTPRequests          int64 `json:"http_requests"`           // HTTP请求计数
	DatabaseOperations    int64 `json:"database_operations"`     // 数据库操作计数
	Errors                int64 `json:"errors"`                  // 错误计数
	MQTTConnected         bool  `json:"mqtt_connected"`          // MQTT连接状态
	DataReceivedBytes     int64 `json:"data_received_bytes"`     // 数据接收字节数
	DataStoredBytes       int64 `json:"data_stored_bytes"`       // 数据存储字节数
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	HTTPRequestsPerMinute       int64 `json:"http_requests_per_minute"`       // 每分钟HTTP请求阈值
	ErrorsPerMinute             int64 `json:"errors_per_minute"`              // 每分钟错误数阈值
	DatabaseOperationsPerMinute int64 `json:"database_operations_per_minute"` // 每分钟数据库操作阈值
}

// NewMonitor 创建监控管理器
func NewMonitor(cfg *config.Config) *Monitor {
	if cfg == nil {
		cfg = config.GetConfigManager().GetConfig()
	}

	// 初始化设备配置管理器
	profileManager := device.GetProfileManager()
	// 加载设备配置文件
	if err := profileManager.LoadProfiles("device_profiles"); err != nil {
		log.Printf("Failed to load device profiles: %v", err)
	}

	return &Monitor{
		monitorManager: monitor.NewMonitorManager(),
		metrics: InternalMetrics{
			System: SystemMetrics{
				Goroutines: runtime.NumGoroutine(),
			},
			Application: InternalApplicationMetrics{},
		},
		startTime: time.Now(),
		alertThresholds: AlertThresholds{
			HTTPRequestsPerMinute:       1000, // 默认每分钟1000个请求
			ErrorsPerMinute:             10,   // 默认每分钟10个错误
			DatabaseOperationsPerMinute: 5000, // 默认每分钟5000个数据库操作
		},
		alerts: []common.Alert{},
		lastMetrics: InternalMetrics{
			System: SystemMetrics{
				Goroutines: runtime.NumGoroutine(),
			},
			Application: InternalApplicationMetrics{},
		},
		lastCollectTime: time.Now(),
		deviceStatus:    make(map[string]DeviceStatus),
		config:          cfg,
		profileManager:  profileManager,
	}
}

// toExportedMetrics 将内部指标转换为导出指标
func (m *Monitor) toExportedMetrics() Metrics {
	return Metrics{
		System:   m.metrics.System,
		Database: m.metrics.Database,
		Application: ApplicationMetrics{
			MQTTMessagesReceived:  m.metrics.Application.MQTTMessagesReceived.Load(),
			MQTTMessagesProcessed: m.metrics.Application.MQTTMessagesProcessed.Load(),
			HTTPRequests:          m.metrics.Application.HTTPRequests.Load(),
			DatabaseOperations:    m.metrics.Application.DatabaseOperations.Load(),
			Errors:                m.metrics.Application.Errors.Load(),
			MQTTConnected:         m.metrics.Application.MQTTConnectionStatus.Load(),
			DataReceivedBytes:     m.metrics.Application.DataReceivedBytes.Load(),
			DataStoredBytes:       m.metrics.Application.DataStoredBytes.Load(),
		},
	}
}

// SetNotifier 设置告警通知器
func (m *Monitor) SetNotifier(notifier *alert.Notifier) {
	m.notifier = notifier
}

// CollectMetrics 收集监控指标
func (m *Monitor) CollectMetrics() Metrics {
	// 收集系统指标
	m.collectSystemMetrics()

	// 收集数据库指标
	m.collectDatabaseMetrics()

	return m.toExportedMetrics()
}

// collectSystemMetrics 收集系统指标
func (m *Monitor) collectSystemMetrics() {
	m.metrics.System.Goroutines = runtime.NumGoroutine()
	m.metrics.System.Uptime = int64(time.Since(m.startTime).Seconds())

	// 内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.metrics.System.MemoryUsage = float64(memStats.Alloc) / 1024 / 1024 // MB

	// CPU使用率：Go标准库不支持获取进程CPU使用率，保持为0
	// 可以通过goroutines数量间接判断负载
	m.metrics.System.CPUUsage = 0
}

// collectDatabaseMetrics 收集数据库指标
func (m *Monitor) collectDatabaseMetrics() {
	// 获取键值变化统计
	keyStats := m.monitorManager.GetKeyChangeStats()
	m.metrics.Database.KeyStats = make(map[int]interface{})
	for k, v := range keyStats {
		m.metrics.Database.KeyStats[k] = v
	}

	// 获取索引统计
	indexStats := m.monitorManager.GetIndexStats()
	m.metrics.Database.IndexStats = make(map[int]interface{})
	for k, v := range indexStats {
		m.metrics.Database.IndexStats[k] = v
	}
}

// IncrementMQTTMessagesReceived 增加MQTT消息接收计数
func (m *Monitor) IncrementMQTTMessagesReceived() {
	m.metrics.Application.MQTTMessagesReceived.Add(1)
}

// IncrementMQTTMessagesProcessed 增加MQTT消息处理计数
func (m *Monitor) IncrementMQTTMessagesProcessed() {
	m.metrics.Application.MQTTMessagesProcessed.Add(1)
}

// IncrementHTTPRequests 增加HTTP请求计数
func (m *Monitor) IncrementHTTPRequests() {
	m.metrics.Application.HTTPRequests.Add(1)
}

// IncrementDatabaseOperations 增加数据库操作计数
func (m *Monitor) IncrementDatabaseOperations() {
	m.metrics.Application.DatabaseOperations.Add(1)
}

// IncrementErrors 增加错误计数
func (m *Monitor) IncrementErrors() {
	m.metrics.Application.Errors.Add(1)
}

// SetMQTTConnectionStatus 设置MQTT连接状态
func (m *Monitor) SetMQTTConnectionStatus(connected bool) {
	m.metrics.Application.MQTTConnectionStatus.Store(connected)
	if !connected {
		m.RecordError("mqtt_disconnected", "MQTT connection lost")
	}
}

// IncrementDataReceivedBytes 增加数据接收字节数
func (m *Monitor) IncrementDataReceivedBytes(bytes int64) {
	m.metrics.Application.DataReceivedBytes.Add(bytes)
}

// IncrementDataStoredBytes 增加数据存储字节数
func (m *Monitor) IncrementDataStoredBytes(bytes int64) {
	m.metrics.Application.DataStoredBytes.Add(bytes)
}

// IsMQTTConnected 获取MQTT连接状态
func (m *Monitor) IsMQTTConnected() bool {
	return m.metrics.Application.MQTTConnectionStatus.Load()
}

// UpdateDeviceStatus 更新设备状态
func (m *Monitor) UpdateDeviceStatus(deviceName, reading string, value float64) {
	m.deviceMutex.Lock()
	defer m.deviceMutex.Unlock()

	now := time.Now()
	status, exists := m.deviceStatus[deviceName]

	if !exists {
		status = DeviceStatus{
			LastActive:    now,
			IsOnline:      true,
			LastAlertTime: time.Time{},
			DataCount:     0,
			DataHistory:   make([]float64, 0, 10),
		}
	}

	// 计算数据变化量
	var dataChange float64
	if status.DataCount > 0 && status.LastDataValue != 0 {
		dataChange = math.Abs(value - status.LastDataValue)
		status.LastDataChange = dataChange
	}

	// 更新数据历史记录（保持最近10个数据点）
	status.DataHistory = append(status.DataHistory, value)
	if len(status.DataHistory) > 10 {
		status.DataHistory = status.DataHistory[len(status.DataHistory)-10:]
	}

	// 简单的异常检测
	if status.DataCount > 2 {
		// 获取数据异常检测阈值
		dataAnomalyThreshold := 50.0 // 默认50%
		dataTrendMinPoints := 5      // 默认5个点

		if m.config != nil {
			if m.config.DataAnomalyThreshold > 0 {
				dataAnomalyThreshold = float64(m.config.DataAnomalyThreshold)
			}
			if m.config.DataTrendMinPoints > 0 {
				dataTrendMinPoints = m.config.DataTrendMinPoints
			}
		}

		// 检测数据突变
		if status.LastDataValue != 0 && dataChange/status.LastDataValue > dataAnomalyThreshold/100 {
			// 触发数据异常告警
			alert := common.Alert{
				Type:      "device_data_anomaly",
				Message:   fmt.Sprintf("Device %s data anomaly: value changed from %.2f to %.2f", deviceName, status.LastDataValue, value),
				Severity:  "warning",
				Timestamp: now,
				Resolved:  false,
			}

			m.mutex.Lock()
			m.alerts = append(m.alerts, alert)
			m.mutex.Unlock()

			log.Printf("Device data anomaly detected: %s - %s", deviceName, alert.Message)

			// 发送告警通知
			if m.notifier != nil {
				m.notifier.SendAlert(alert)
			}
		}

		// 检测数据趋势异常
		if len(status.DataHistory) >= dataTrendMinPoints {
			isIncreasing := true
			isDecreasing := true

			for i := 1; i < len(status.DataHistory); i++ {
				if status.DataHistory[i] <= status.DataHistory[i-1] {
					isIncreasing = false
				}
				if status.DataHistory[i] >= status.DataHistory[i-1] {
					isDecreasing = false
				}
			}

			if isIncreasing || isDecreasing {
				trend := "increasing"
				if isDecreasing {
					trend = "decreasing"
				}

				alert := common.Alert{
					Type:      "device_data_trend",
					Message:   fmt.Sprintf("Device %s data trend: continuous %s for %d consecutive readings", deviceName, trend, dataTrendMinPoints),
					Severity:  "info",
					Timestamp: now,
					Resolved:  false,
				}

				m.mutex.Lock()
				m.alerts = append(m.alerts, alert)
				m.mutex.Unlock()

				log.Printf("Device data trend detected: %s - %s", deviceName, alert.Message)

				// 发送告警通知
				if m.notifier != nil {
					m.notifier.SendAlert(alert)
				}
			}
		}
	}

	// 检查设备配置文件中的阈值
	if m.profileManager != nil {
		// 根据设备名称自动查找对应的设备配置
		profileName, found := m.profileManager.FindProfileByDeviceName(deviceName)
		if !found {
			// 如果没有找到对应的设备配置，跳过阈值检查
			// 但仍然更新设备状态
			status.LastActive = now
			status.IsOnline = true
			status.DataCount++
			status.LastDataValue = value
			status.LastDataTime = now

			m.deviceStatus[deviceName] = status
			return
		}

		// 检查温度阈值
		if threshold, exists := m.profileManager.GetResourceThreshold(profileName, "Temperature"); exists {
			if value > threshold {
				// 触发温度超限告警
				alert := common.Alert{
					Type:      "device_temperature_over_limit",
					Message:   fmt.Sprintf("Device %s temperature over limit: %.2f > %.2f", deviceName, value, threshold),
					Severity:  "critical",
					Timestamp: now,
					Resolved:  false,
				}

				m.mutex.Lock()
				m.alerts = append(m.alerts, alert)
				m.mutex.Unlock()

				log.Printf("Device temperature over limit: %s - %s", deviceName, alert.Message)

				// 发送告警通知
				if m.notifier != nil {
					m.notifier.SendAlert(alert)
				}
			}
		}

		// 检查湿度阈值
		if threshold, exists := m.profileManager.GetResourceThreshold(profileName, "Humidity"); exists {
			if value > threshold {
				// 触发湿度超限告警
				alert := common.Alert{
					Type:      "device_humidity_over_limit",
					Message:   fmt.Sprintf("Device %s humidity over limit: %.2f > %.2f", deviceName, value, threshold),
					Severity:  "warning",
					Timestamp: now,
					Resolved:  false,
				}

				m.mutex.Lock()
				m.alerts = append(m.alerts, alert)
				m.mutex.Unlock()

				log.Printf("Device humidity over limit: %s - %s", deviceName, alert.Message)

				// 发送告警通知
				if m.notifier != nil {
					m.notifier.SendAlert(alert)
				}
			}
		}
	}

	status.LastActive = now
	status.IsOnline = true
	status.DataCount++
	status.LastDataValue = value
	status.LastDataTime = now

	m.deviceStatus[deviceName] = status
}

// CheckDeviceStatus 检查设备状态
func (m *Monitor) CheckDeviceStatus() []common.Alert {
	var newAlerts []common.Alert
	now := time.Now()

	// 获取离线检测阈值
	offlineThreshold := 300 // 默认5分钟
	if m.config != nil && m.config.DeviceOfflineThreshold > 0 {
		offlineThreshold = m.config.DeviceOfflineThreshold
	}

	m.deviceMutex.Lock()
	defer m.deviceMutex.Unlock()

	for deviceName, status := range m.deviceStatus {
		// 检查设备是否离线
		if now.Sub(status.LastActive) > time.Duration(offlineThreshold)*time.Second && status.IsOnline {
			// 检查是否已经告警过（避免重复告警）
			if now.Sub(status.LastAlertTime) > 10*time.Minute {
				alert := common.Alert{
					Type:      "device_offline",
					Message:   fmt.Sprintf("Device %s is offline", deviceName),
					Severity:  "warning",
					Timestamp: now,
					Resolved:  false,
				}
				newAlerts = append(newAlerts, alert)

				// 更新设备状态和告警时间
				status.IsOnline = false
				status.LastAlertTime = now
				m.deviceStatus[deviceName] = status
			}
		}

		// 检查设备是否恢复在线
		if now.Sub(status.LastActive) < time.Duration(offlineThreshold)*time.Second && !status.IsOnline {
			alert := common.Alert{
				Type:      "device_online",
				Message:   fmt.Sprintf("Device %s is back online", deviceName),
				Severity:  "info",
				Timestamp: now,
				Resolved:  false,
			}
			newAlerts = append(newAlerts, alert)

			// 更新设备状态
			status.IsOnline = true
			m.deviceStatus[deviceName] = status
		}
	}

	return newAlerts
}

// GetDeviceStatus 获取设备状态
func (m *Monitor) GetDeviceStatus() map[string]DeviceStatus {
	m.deviceMutex.Lock()
	defer m.deviceMutex.Unlock()

	// 创建副本以避免并发访问问题
	statusCopy := make(map[string]DeviceStatus)
	for k, v := range m.deviceStatus {
		statusCopy[k] = v
	}

	return statusCopy
}

// RecordError 记录错误并触发告警
func (m *Monitor) RecordError(errorType, message string) {
	m.IncrementErrors()

	// 创建新告警
	alert := common.Alert{
		Type:      errorType,
		Message:   message,
		Severity:  "critical",
		Timestamp: time.Now(),
		Resolved:  false,
	}

	// 添加到告警列表（加锁保护）
	m.mutex.Lock()
	m.alerts = append(m.alerts, alert)
	m.mutex.Unlock()

	log.Printf("Critical error recorded: %s - %s", errorType, message)

	// 发送告警通知
	if m.notifier != nil {
		m.notifier.SendAlert(alert)
	}
}

// RecordInfo 记录信息
func (m *Monitor) RecordInfo(infoType, message string) {
	log.Printf("Info recorded: %s - %s", infoType, message)
}

// lastMetricValues 保存上次指标的数值快照
type lastMetricValues struct {
	httpRequests       int64
	errors             int64
	databaseOperations int64
}

// CheckAlerts 检查告警
func (m *Monitor) CheckAlerts() []common.Alert {
	var newAlerts []common.Alert

	// 计算时间差（分钟）
	timeDiff := time.Since(m.lastCollectTime).Minutes()
	if timeDiff < 1 {
		timeDiff = 1 // 避免除以零
	}

	// 获取当前指标值
	currentHTTP := m.metrics.Application.HTTPRequests.Load()
	currentErrors := m.metrics.Application.Errors.Load()
	currentDBOps := m.metrics.Application.DatabaseOperations.Load()

	// 获取上次指标值
	lastHTTP := m.lastMetrics.Application.HTTPRequests.Load()
	lastErrors := m.lastMetrics.Application.Errors.Load()
	lastDBOps := m.lastMetrics.Application.DatabaseOperations.Load()

	// 计算每分钟的指标
	httpRequestsPerMinute := (currentHTTP - lastHTTP) / int64(timeDiff)
	errorsPerMinute := (currentErrors - lastErrors) / int64(timeDiff)
	dbOperationsPerMinute := (currentDBOps - lastDBOps) / int64(timeDiff)

	// 检查HTTP请求告警
	if httpRequestsPerMinute > m.alertThresholds.HTTPRequestsPerMinute {
		newAlerts = append(newAlerts, common.Alert{
			Type:      "http_requests",
			Message:   fmt.Sprintf("HTTP requests rate too high: %d per minute", httpRequestsPerMinute),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// 检查错误告警
	if errorsPerMinute > m.alertThresholds.ErrorsPerMinute {
		newAlerts = append(newAlerts, common.Alert{
			Type:      "errors",
			Message:   fmt.Sprintf("Error rate too high: %d per minute", errorsPerMinute),
			Severity:  "critical",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// 检查数据库操作告警
	if dbOperationsPerMinute > m.alertThresholds.DatabaseOperationsPerMinute {
		newAlerts = append(newAlerts, common.Alert{
			Type:      "database_operations",
			Message:   fmt.Sprintf("Database operations rate too high: %d per minute", dbOperationsPerMinute),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// 检查设备状态告警
	deviceAlerts := m.CheckDeviceStatus()
	newAlerts = append(newAlerts, deviceAlerts...)

	// 添加新告警（加锁保护）
	m.mutex.Lock()
	m.alerts = append(m.alerts, newAlerts...)

	// 限制告警列表大小，防止内存泄漏（最多保留10000条）
	if len(m.alerts) > 10000 {
		m.alerts = m.alerts[len(m.alerts)-10000:]
	}

	// 更新上次收集的指标值（逐个存储，不复制整个结构体）
	m.lastMetrics.Application.MQTTMessagesReceived.Store(m.metrics.Application.MQTTMessagesReceived.Load())
	m.lastMetrics.Application.MQTTMessagesProcessed.Store(m.metrics.Application.MQTTMessagesProcessed.Load())
	m.lastMetrics.Application.HTTPRequests.Store(m.metrics.Application.HTTPRequests.Load())
	m.lastMetrics.Application.DatabaseOperations.Store(m.metrics.Application.DatabaseOperations.Load())
	m.lastMetrics.Application.Errors.Store(m.metrics.Application.Errors.Load())

	m.lastCollectTime = time.Now()
	m.mutex.Unlock()

	// 发送新告警通知
	if m.notifier != nil {
		for _, alert := range newAlerts {
			m.notifier.SendAlert(alert)
		}
	}

	return newAlerts
}

// GetAlerts 获取所有告警
func (m *Monitor) GetAlerts() []common.Alert {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.alerts
}

// RegisterHandlers 注册HTTP处理函数
func (m *Monitor) RegisterHandlers() {
	http.HandleFunc("/metrics", m.handleMetrics)
	http.HandleFunc("/health", m.handleHealth)
	http.HandleFunc("/alerts", m.handleAlerts)
	http.HandleFunc("/device-status", m.handleDeviceStatus)
}

// handleMetrics 处理指标请求
func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.CollectMetrics()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("Error encoding metrics: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleHealth 处理健康检查请求
func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := m.CollectMetrics()

	healthStatus := map[string]interface{}{
		"status":          "healthy",
		"uptime_seconds":  metrics.System.Uptime,
		"goroutines":      metrics.System.Goroutines,
		"memory_usage_mb": metrics.System.MemoryUsage,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
		log.Printf("Error encoding health status: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAlerts 处理告警请求
func (m *Monitor) handleAlerts(w http.ResponseWriter, r *http.Request) {
	// 检查新告警
	m.CheckAlerts()

	// 获取所有告警
	alerts := m.GetAlerts()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		log.Printf("Error encoding alerts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAlertGroups 处理分组告警请求
func (m *Monitor) handleAlertGroups(w http.ResponseWriter, r *http.Request) {
	// 检查新告警
	m.CheckAlerts()

	// 获取分组后的告警
	groups := m.GetAlertGroups()

	// 调试日志
	log.Printf("AlertGroups: %d groups", len(groups))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "groups": groups}); err != nil {
		log.Printf("Error encoding alert groups: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleDeviceStatus 处理设备状态请求
func (m *Monitor) handleDeviceStatus(w http.ResponseWriter, r *http.Request) {
	// 检查设备状态
	m.CheckDeviceStatus()

	// 获取设备状态
	deviceStatus := m.GetDeviceStatus()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deviceStatus); err != nil {
		log.Printf("Error encoding device status: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
