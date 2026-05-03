package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sfsEdgeStore/alert"
	"sfsEdgeStore/common"
	"sfsEdgeStore/config"
)

type Monitor struct {
	*MetricsCollector
	*AlertManager
	*DeviceTracker
	config *config.Config
}

func NewMonitor(cfg *config.Config) *Monitor {
	if cfg == nil {
		cfg = config.GetConfigManager().GetConfig()
	}
	mc := NewMetricsCollector()
	am := NewAlertManager()
	dt := NewDeviceTracker(cfg, am)
	return &Monitor{MetricsCollector: mc, AlertManager: am, DeviceTracker: dt, config: cfg}
}

func (m *Monitor) SetNotifier(notifier *alert.Notifier) {
	m.AlertManager.SetNotifier(notifier)
}

func (m *Monitor) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			m.DeviceTracker.CleanupInactive()
			m.AlertManager.TrimAlerts()
		}
	}()
}

func (m *Monitor) RecordInfo(infoType, message string) {
	log.Printf("Info recorded: %s - %s", infoType, message)
}

// 向后兼容别名
func (m *Monitor) UpdateDeviceStatus(d, r string, v float64) { m.UpdateStatus(d, r, v) }
func (m *Monitor) GetDeviceStatus() map[string]DeviceStatus  { return m.GetStatus() }
func (m *Monitor) CheckDeviceStatus() []common.Alert         { return m.CheckStatus() }

func (m *Monitor) CheckAlerts() []common.Alert {
	var newAlerts []common.Alert
	timeDiff := time.Since(m.GetLastCollectTime()).Minutes()
	if timeDiff < 1 {
		timeDiff = 1
	}

	last := m.GetLastMetrics()
	cur := &m.metrics.Application
	httpPerMin := (cur.HTTPRequests.Load() - last.Application.HTTPRequests.Load()) / int64(timeDiff)
	errPerMin := (cur.Errors.Load() - last.Application.Errors.Load()) / int64(timeDiff)
	dbPerMin := (cur.DatabaseOperations.Load() - last.Application.DatabaseOperations.Load()) / int64(timeDiff)

	if httpPerMin > m.thresholds.HTTPRequestsPerMinute {
		newAlerts = append(newAlerts, common.Alert{Type: "http_requests",
			Message:  fmt.Sprintf("HTTP requests rate too high: %d per minute", httpPerMin),
			Severity: "warning", Timestamp: time.Now(), Resolved: false})
	}
	if errPerMin > m.thresholds.ErrorsPerMinute {
		newAlerts = append(newAlerts, common.Alert{Type: "errors",
			Message:  fmt.Sprintf("Error rate too high: %d per minute", errPerMin),
			Severity: "critical", Timestamp: time.Now(), Resolved: false})
	}
	if dbPerMin > m.thresholds.DatabaseOperationsPerMinute {
		newAlerts = append(newAlerts, common.Alert{Type: "database_operations",
			Message:  fmt.Sprintf("Database operations rate too high: %d per minute", dbPerMin),
			Severity: "warning", Timestamp: time.Now(), Resolved: false})
	}

	for _, a := range append(newAlerts, m.DeviceTracker.CheckStatus()...) {
		m.AddAlert(a)
	}
	m.UpdateLastMetrics()
	if m.notifier != nil {
		for _, a := range newAlerts {
			m.notifier.SendAlert(a)
		}
	}
	return newAlerts
}

func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.CollectMetrics())
}

func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := m.CollectMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy", "uptime_seconds": metrics.System.Uptime,
		"goroutines": metrics.System.Goroutines, "memory_usage_mb": metrics.System.MemoryUsage,
	})
}

func (m *Monitor) handleAlerts(w http.ResponseWriter, r *http.Request) {
	m.CheckAlerts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.GetAlerts())
}

func (m *Monitor) handleAlertGroups(w http.ResponseWriter, r *http.Request) {
	m.CheckAlerts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "groups": m.GetAlertGroups()})
}

func (m *Monitor) handleDeviceStatus(w http.ResponseWriter, r *http.Request) {
	m.CheckDeviceStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.GetDeviceStatus())
}
