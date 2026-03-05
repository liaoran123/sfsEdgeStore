package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liaoran123/sfsDb/management/monitor"
)

// Monitor 监控管理器
type Monitor struct {
	monitorManager  *monitor.MonitorManager // 监控管理器
	metrics         Metrics                 // 监控指标
	startTime       time.Time               // 启动时间
	alertThresholds AlertThresholds         // 告警阈值
	alerts          []Alert                 // 告警列表
	lastMetrics     Metrics                 // 上次收集的指标
	lastCollectTime time.Time               // 上次收集时间
	mutex           sync.Mutex              // 保护alerts切片的互斥锁
}

// Metrics 监控指标
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

// ApplicationMetrics 应用指标
type ApplicationMetrics struct {
	MQTTMessagesReceived  atomic.Int64 `json:"mqtt_messages_received"`  // MQTT消息接收计数
	MQTTMessagesProcessed atomic.Int64 `json:"mqtt_messages_processed"` // MQTT消息处理计数
	HTTPRequests          atomic.Int64 `json:"http_requests"`           // HTTP请求计数
	DatabaseOperations    atomic.Int64 `json:"database_operations"`     // 数据库操作计数
	Errors                atomic.Int64 `json:"errors"`                  // 错误计数
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	HTTPRequestsPerMinute       int64 `json:"http_requests_per_minute"`       // 每分钟HTTP请求阈值
	ErrorsPerMinute             int64 `json:"errors_per_minute"`              // 每分钟错误数阈值
	DatabaseOperationsPerMinute int64 `json:"database_operations_per_minute"` // 每分钟数据库操作阈值
}

// Alert 告警信息
type Alert struct {
	Type      string    `json:"type"`      // 告警类型
	Message   string    `json:"message"`   // 告警消息
	Severity  string    `json:"severity"`  // 告警级别
	Timestamp time.Time `json:"timestamp"` // 告警时间
	Resolved  bool      `json:"resolved"`  // 是否已解决
}

// NewMonitor 创建监控管理器
func NewMonitor() *Monitor {
	return &Monitor{
		monitorManager: monitor.NewMonitorManager(),
		metrics: Metrics{
			System: SystemMetrics{
				Goroutines: runtime.NumGoroutine(),
			},
			Application: ApplicationMetrics{},
		},
		startTime: time.Now(),
		alertThresholds: AlertThresholds{
			HTTPRequestsPerMinute:       1000, // 默认每分钟1000个请求
			ErrorsPerMinute:             10,   // 默认每分钟10个错误
			DatabaseOperationsPerMinute: 5000, // 默认每分钟5000个数据库操作
		},
		alerts:          []Alert{},
		lastCollectTime: time.Now(),
	}
}

// CollectMetrics 收集监控指标
func (m *Monitor) CollectMetrics() Metrics {
	// 收集系统指标
	m.collectSystemMetrics()

	// 收集数据库指标
	m.collectDatabaseMetrics()

	return m.metrics
}

// collectSystemMetrics 收集系统指标
func (m *Monitor) collectSystemMetrics() {
	m.metrics.System.Goroutines = runtime.NumGoroutine()
	m.metrics.System.Uptime = int64(time.Since(m.startTime).Seconds())

	// 简化的CPU和内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.metrics.System.MemoryUsage = float64(memStats.Alloc) / 1024 / 1024 // MB
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

// RecordError 记录错误并触发告警
func (m *Monitor) RecordError(errorType, message string) {
	m.IncrementErrors()

	// 创建新告警
	alert := Alert{
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
}

// CheckAlerts 检查告警
func (m *Monitor) CheckAlerts() []Alert {
	var newAlerts []Alert

	// 计算时间差（分钟）
	timeDiff := time.Since(m.lastCollectTime).Minutes()
	if timeDiff == 0 {
		timeDiff = 1 // 避免除以零
	}

	// 计算每分钟的指标
	httpRequestsPerMinute := (m.metrics.Application.HTTPRequests.Load() - m.lastMetrics.Application.HTTPRequests.Load()) / int64(timeDiff)
	errorsPerMinute := (m.metrics.Application.Errors.Load() - m.lastMetrics.Application.Errors.Load()) / int64(timeDiff)
	dbOperationsPerMinute := (m.metrics.Application.DatabaseOperations.Load() - m.lastMetrics.Application.DatabaseOperations.Load()) / int64(timeDiff)

	// 检查HTTP请求告警
	if httpRequestsPerMinute > m.alertThresholds.HTTPRequestsPerMinute {
		newAlerts = append(newAlerts, Alert{
			Type:      "http_requests",
			Message:   fmt.Sprintf("HTTP requests rate too high: %d per minute", httpRequestsPerMinute),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// 检查错误告警
	if errorsPerMinute > m.alertThresholds.ErrorsPerMinute {
		newAlerts = append(newAlerts, Alert{
			Type:      "errors",
			Message:   fmt.Sprintf("Error rate too high: %d per minute", errorsPerMinute),
			Severity:  "critical",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// 检查数据库操作告警
	if dbOperationsPerMinute > m.alertThresholds.DatabaseOperationsPerMinute {
		newAlerts = append(newAlerts, Alert{
			Type:      "database_operations",
			Message:   fmt.Sprintf("Database operations rate too high: %d per minute", dbOperationsPerMinute),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// 添加新告警（加锁保护）
	m.mutex.Lock()
	m.alerts = append(m.alerts, newAlerts...)
	// 更新上次收集的指标和时间
	m.lastMetrics = m.metrics
	m.lastCollectTime = time.Now()
	m.mutex.Unlock()

	return newAlerts
}

// GetAlerts 获取所有告警
func (m *Monitor) GetAlerts() []Alert {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.alerts
}

// RegisterHandlers 注册HTTP处理函数
func (m *Monitor) RegisterHandlers() {
	http.HandleFunc("/metrics", m.handleMetrics)
	http.HandleFunc("/health", m.handleHealth)
	http.HandleFunc("/alerts", m.handleAlerts)
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
