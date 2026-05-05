package monitor

import (
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

// MetricsCollector 收集系统和应用指标
type MetricsCollector struct {
	startTime     time.Time    // 启动时间
	mqttReceived  atomic.Int64 // MQTT 接收消息数
	mqttProcessed atomic.Int64 // MQTT 处理消息数
	mqttFiltered  atomic.Int64 // MQTT 过滤消息数
	totalRecords  atomic.Int64 // 总存储记录数
	mqttConnected atomic.Bool  // MQTT 连接状态
	httpRequests  atomic.Int64 // HTTP 请求数（仅计数，不返回前端）
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{startTime: time.Now()}
}

// 指标递增方法
func (c *MetricsCollector) IncrementMQTTMessagesReceived()      { c.mqttReceived.Add(1) }
func (c *MetricsCollector) IncrementMQTTMessagesProcessed(n int64) { c.mqttProcessed.Add(n) }
func (c *MetricsCollector) IncrementMQTTMessagesFiltered()      { c.mqttFiltered.Add(1) }
func (c *MetricsCollector) IncrementTotalRecordsStored(n int64) { c.totalRecords.Add(n) }

// MQTT 连接状态
func (c *MetricsCollector) SetMQTTConnectionStatus(ok bool) { c.mqttConnected.Store(ok) }
func (c *MetricsCollector) IsMQTTConnected() bool           { return c.mqttConnected.Load() }

// 错误和日志记录（告警需要）
func (c *MetricsCollector) RecordError(errorType, message string) {
	log.Printf("[ERROR] %s: %s", errorType, message)
}
func (c *MetricsCollector) RecordInfo(infoType, message string) {
	log.Printf("[INFO] %s: %s", infoType, message)
}

// HTTP 请求计数（内部使用，不返回前端）
func (c *MetricsCollector) IncrementHTTPRequests() { c.httpRequests.Add(1) }

// CollectMetrics 采集并返回监控指标（供前端使用）
func (c *MetricsCollector) CollectMetrics() Metrics {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return Metrics{
		System: SystemMetrics{
			MemoryUsage: float64(ms.Alloc) / 1024 / 1024,          // 进程内存使用（MB）
			Goroutines:  runtime.NumGoroutine(),                   // Goroutine 数量
			Uptime:      int64(time.Since(c.startTime).Seconds()), // 运行时间（秒）
		},
		Application: ApplicationMetrics{
			MQTTMessagesReceived:  c.mqttReceived.Load(),  // MQTT 接收消息数
			MQTTMessagesProcessed: c.mqttProcessed.Load(), // MQTT 处理消息数
			MQTTMessagesFiltered:  c.mqttFiltered.Load(),  // MQTT 过滤消息数
			TotalRecordsStored:    c.totalRecords.Load(),  // 总存储记录数
		},
	}
}
