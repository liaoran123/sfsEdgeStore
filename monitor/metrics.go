package monitor

import (
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

type MetricsCollector struct {
	metrics         metricsData
	startTime       time.Time
	lastMetrics     metricsData
	lastCollectTime time.Time
}

func NewMetricsCollector() *MetricsCollector {
	goroutines := runtime.NumGoroutine()
	now := time.Now()
	return &MetricsCollector{
		metrics: metricsData{
			System: SystemMetrics{
				Goroutines: goroutines,
			},
		},
		startTime: now,
		lastMetrics: metricsData{
			System: SystemMetrics{
				Goroutines: goroutines,
			},
		},
		lastCollectTime: now,
	}
}

func (c *MetricsCollector) IncrementMQTTMessagesReceived() {
	c.metrics.Application.MQTTMessagesReceived.Add(1)
}
func (c *MetricsCollector) IncrementMQTTMessagesProcessed() {
	c.metrics.Application.MQTTMessagesProcessed.Add(1)
}
func (c *MetricsCollector) IncrementMQTTMessagesFiltered() {
	c.metrics.Application.MQTTMessagesFiltered.Add(1)
}
func (c *MetricsCollector) IncrementTotalRecordsStored(n int64) {
	c.metrics.Application.TotalRecordsStored.Add(n)
}
func (c *MetricsCollector) IncrementHTTPRequests() { c.metrics.Application.HTTPRequests.Add(1) }
func (c *MetricsCollector) IncrementDatabaseOperations() {
	c.metrics.Application.DatabaseOperations.Add(1)
}
func (c *MetricsCollector) IncrementErrors() { c.metrics.Application.Errors.Add(1) }
func (c *MetricsCollector) SetMQTTConnectionStatus(ok bool) {
	c.metrics.Application.MQTTConnectionStatus.Store(ok)
}
func (c *MetricsCollector) IncrementDataReceivedBytes(n int64) {
	c.metrics.Application.DataReceivedBytes.Add(n)
}
func (c *MetricsCollector) IncrementDataStoredBytes(n int64) {
	c.metrics.Application.DataStoredBytes.Add(n)
}
func (c *MetricsCollector) IsMQTTConnected() bool {
	return c.metrics.Application.MQTTConnectionStatus.Load()
}

func (c *MetricsCollector) CollectMetrics() Metrics {
	c.collectSystemMetrics()
	return Metrics{
		System: c.metrics.System,
		Application: ApplicationMetrics{
			MQTTMessagesReceived:  c.metrics.Application.MQTTMessagesReceived.Load(),
			MQTTMessagesProcessed: c.metrics.Application.MQTTMessagesProcessed.Load(),
			MQTTMessagesFiltered:  c.metrics.Application.MQTTMessagesFiltered.Load(),
			TotalRecordsStored:    c.metrics.Application.TotalRecordsStored.Load(),
			HTTPRequests:          c.metrics.Application.HTTPRequests.Load(),
			DatabaseOperations:    c.metrics.Application.DatabaseOperations.Load(),
			Errors:                c.metrics.Application.Errors.Load(),
			MQTTConnected:         c.metrics.Application.MQTTConnectionStatus.Load(),
			DataReceivedBytes:     c.metrics.Application.DataReceivedBytes.Load(),
			DataStoredBytes:       c.metrics.Application.DataStoredBytes.Load(),
		},
	}
}

func (c *MetricsCollector) collectSystemMetrics() {
	c.metrics.System.Goroutines = runtime.NumGoroutine()
	c.metrics.System.Uptime = int64(time.Since(c.startTime).Seconds())
	c.metrics.System.MemoryUsage = c.getProcessMemoryMB()
	c.metrics.System.CPUUsage = 0
}

func (c *MetricsCollector) getProcessMemoryMB() float64 {
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		return float64(ms.Alloc) / 1024 / 1024
	}
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		return float64(ms.Alloc) / 1024 / 1024
	}
	return float64(memInfo.RSS) / 1024 / 1024
}

func (c *MetricsCollector) GetLastMetrics() *metricsData { return &c.lastMetrics }
func (c *MetricsCollector) GetLastCollectTime() time.Time {
	return c.lastCollectTime
}
func (c *MetricsCollector) UpdateLastMetrics() {
	m := &c.metrics.Application
	l := &c.lastMetrics.Application
	l.MQTTMessagesReceived.Store(m.MQTTMessagesReceived.Load())
	l.MQTTMessagesProcessed.Store(m.MQTTMessagesProcessed.Load())
	l.TotalRecordsStored.Store(m.TotalRecordsStored.Load())
	l.HTTPRequests.Store(m.HTTPRequests.Load())
	l.DatabaseOperations.Store(m.DatabaseOperations.Load())
	l.Errors.Store(m.Errors.Load())
	c.lastCollectTime = time.Now()
}
