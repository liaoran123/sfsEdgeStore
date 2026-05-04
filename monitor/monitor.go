package monitor

import (
	"encoding/json"
	"net/http"
)

// Monitor 监控管理器
type Monitor struct {
	*MetricsCollector
}

// NewMonitor 创建监控器
func NewMonitor() *Monitor {
	return &Monitor{MetricsCollector: NewMetricsCollector()}
}

// handleMetrics 返回监控指标 JSON
func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.CollectMetrics())
}

// handleHealth 返回健康状态 JSON
func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := m.CollectMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":          "healthy",
		"uptime_seconds":  metrics.System.Uptime,
		"goroutines":      metrics.System.Goroutines,
		"memory_usage_mb": metrics.System.MemoryUsage,
	})
}
