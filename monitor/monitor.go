package monitor

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// ErrorCounters 按类型分类的错误计数器（预定义，无锁设计）
type ErrorCounters struct {
	DBWriteFailed atomic.Int64 // 数据库写入失败
	PartialWrite  atomic.Int64 // 部分写入
	Memory        atomic.Int64 // 内存相关
	Database      atomic.Int64 // 数据库连接
	Restart       atomic.Int64 // 服务重启
	Rule          atomic.Int64 // 规则执行
	Script        atomic.Int64 // 脚本执行
	Storage       atomic.Int64 // 存储相关
	Other         atomic.Int64 // 其他/备用
}

// Monitor 监控管理器（合并 MetricsCollector 功能）
type Monitor struct {
	startTime     time.Time    // 启动时间
	mqttReceived  atomic.Int64 // MQTT 接收消息数
	mqttProcessed atomic.Int64 // MQTT 处理消息数
	mqttFiltered  atomic.Int64 // MQTT 过滤消息数
	totalRecords  atomic.Int64 // 总存储记录数
	mqttConnected atomic.Bool  // MQTT 连接状态
	httpRequests  atomic.Int64 // HTTP 请求数（仅计数，不返回前端）
	errorCount    atomic.Int64 // 错误总数
	ErrorCounters              // 内嵌错误计数器
}

// NewMonitor 创建监控器
func NewMonitor() *Monitor {
	return &Monitor{startTime: time.Now()}
}

// 指标递增方法
func (m *Monitor) IncrementMQTTMessagesReceived()         { m.mqttReceived.Add(1) }
func (m *Monitor) IncrementMQTTMessagesProcessed(n int64) { m.mqttProcessed.Add(n) }
func (m *Monitor) IncrementMQTTMessagesFiltered()         { m.mqttFiltered.Add(1) }
func (m *Monitor) IncrementTotalRecordsStored(n int64)    { m.totalRecords.Add(n) }
func (m *Monitor) SetMQTTConnectionStatus(ok bool)        { m.mqttConnected.Store(ok) }
func (m *Monitor) IsMQTTConnected() bool                  { return m.mqttConnected.Load() }
func (m *Monitor) IncrementHTTPRequests()                 { m.httpRequests.Add(1) }

// 集中记录各种错误类型
func (m *Monitor) RecordError(errorType, message string) {
	m.errorCount.Add(1)

	switch errorType {
	case "db_write_failed":
		m.DBWriteFailed.Add(1)
	case "partial_write":
		m.PartialWrite.Add(1)
	case "memory":
		m.Memory.Add(1)
	case "database":
		m.Database.Add(1)
	case "restart":
		m.Restart.Add(1)
	case "rule":
		m.Rule.Add(1)
	case "script":
		m.Script.Add(1)
	case "storage_error", "database_error", "resource_contention":
		m.Storage.Add(1)
	default:
		m.Other.Add(1)
	}

	log.Printf("[ERROR] %s: %s", errorType, message)
}
func (m *Monitor) RecordInfo(infoType, message string) {
	log.Printf("[INFO] %s: %s", infoType, message)
}

// handleMetrics 返回监控指标 JSON
func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.CollectMetrics(nil))
}

// CollectMetrics 采集并返回监控指标（供前端使用）
func (m *Monitor) CollectMetrics(resourceUsage *ResourceUsageData) Metrics {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	memoryMB := float64(ms.Alloc) / 1024 / 1024
	cpuPercent := 0.0
	goroutines := runtime.NumGoroutine()

	if resourceUsage != nil {
		memoryMB = resourceUsage.MemoryMB
		cpuPercent = resourceUsage.CPUPercent
		goroutines = resourceUsage.Goroutines
	}

	return Metrics{
		System: SystemMetrics{
			CPUPercent:  cpuPercent,                               // CPU 使用率（%）
			MemoryUsage: memoryMB,                                 // 进程内存使用（MB）
			Goroutines:  goroutines,                               // Goroutine 数量
			Uptime:      int64(time.Since(m.startTime).Seconds()), // 运行时间（秒）
		},
		Application: ApplicationMetrics{
			MQTTMessagesReceived:  m.mqttReceived.Load(),  // MQTT 接收消息数
			MQTTMessagesProcessed: m.mqttProcessed.Load(), // MQTT 处理消息数
			MQTTMessagesFiltered:  m.mqttFiltered.Load(),  // MQTT 过滤消息数
			TotalRecordsStored:    m.totalRecords.Load(),  // 总存储记录数
			ErrorCount:            m.errorCount.Load(),    // 错误总数
			ErrorByType: map[string]int64{
				"db_write_failed": m.DBWriteFailed.Load(),
				"partial_write":   m.PartialWrite.Load(),
				"memory":          m.Memory.Load(),
				"database":        m.Database.Load(),
				"restart":         m.Restart.Load(),
				"rule":            m.Rule.Load(),
				"script":          m.Script.Load(),
				"storage":         m.Storage.Load(),
				"other":           m.Other.Load(),
			},
		},
	}
}

// ResourceUsageData 资源使用数据（用于从外部传入）
type ResourceUsageData struct {
	MemoryMB   float64
	CPUPercent float64
	Goroutines int
}
