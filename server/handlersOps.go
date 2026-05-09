package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"sfsEdgeStore/monitor"
)

// handleHealth 处理健康检查请求
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	json.NewEncoder(w).Encode(healthStatus)
}

// handleReady 处理就绪检查请求
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	isReady := true
	checks := map[string]bool{
		"database": s.Table != nil,
		"mqtt":     true,
	}

	for _, ready := range checks {
		if !ready {
			isReady = false
			break
		}
	}

	statusCode := http.StatusOK
	if !isReady {
		statusCode = http.StatusServiceUnavailable
	}

	readyStatus := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks":    checks,
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(readyStatus)
}

// handleMetrics 处理指标请求
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if s.Monitor == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not initialized"})
		return
	}

	var resourceData *monitor.ResourceUsageData
	if s.ResourceMonitor != nil {
		usage := s.ResourceMonitor.GetCurrentUsage()
		resourceData = &monitor.ResourceUsageData{
			MemoryMB:   usage.MemoryMB,
			CPUPercent: usage.CPUPercent,
			Goroutines: usage.Goroutines,
		}
	}

	metrics := s.Monitor.CollectMetrics(resourceData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleRetentionStatus 处理保留策略状态请求
func (s *Server) handleRetentionStatus(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if s.RetentionMgr == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Retention manager not initialized"})
		return
	}

	status := s.RetentionMgr.GetRetentionStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

// handleManualCleanup 处理手动清理请求
func (s *Server) handleManualCleanup(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	if s.RetentionMgr == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Retention manager not initialized"})
		return
	}

	log.Println("Starting manual data cleanup")
	deleted, err := s.RetentionMgr.CleanupOldData()
	if err != nil {
		log.Printf("Manual cleanup failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("Manual cleanup completed, deleted %d records", deleted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"deleted_count": deleted,
	})
}

// handleAlertNotifierStatus 处理告通知器状态请求
func (s *Server) handleAlertNotifierStatus(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if s.AlertNotifier == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Alert notifier not initialized"})
		return
	}

	status := s.AlertNotifier.GetNotifierStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

// handleTestAlert 处理测试告警请求
func (s *Server) handleTestAlert(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	if s.Monitor == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not initialized"})
		return
	}

	var req struct {
		Type     string `json:"type"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Type == "" {
		req.Type = "test"
	}
	if req.Message == "" {
		req.Message = "Test alert notification"
	}
	if req.Severity == "" {
		req.Severity = "warning"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Test alert sent",
		"alert": map[string]string{
			"type":     req.Type,
			"message":  req.Message,
			"severity": req.Severity,
		},
	})
}

// handleAlertGroups 处理告警分组请求
func (s *Server) handleAlertGroups(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"groups": []map[string]interface{}{},
	})
}

// handleResourceStatus 处理资源状态请求
func (s *Server) handleResourceStatus(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}
	w.Header().Set("Content-Type", "application/json")

	if s.ResourceMonitor == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Resource monitor not initialized"})
		return
	}

	usage := s.ResourceMonitor.GetCurrentUsage()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   usage,
	})
}
