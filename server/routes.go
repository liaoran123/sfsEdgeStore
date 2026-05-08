package server

import (
	"net/http"

	"sfsEdgeStore/common"
)

// DeviceNameMiddleware 处理HTTP请求中的deviceName参数格式化
func DeviceNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceName := r.URL.Query().Get("deviceName")
		if deviceName != "" {
			formattedDeviceName := common.FormatDeviceName(deviceName)
			url := *r.URL
			q := url.Query()
			q.Set("deviceName", formattedDeviceName)
			url.RawQuery = q.Encode()
			*r.URL = url
		}
		next(w, r)
	}
}

// registerRoutes 注册HTTP路由
func (s *Server) registerRoutes() {
	// 健康检查
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/healthz", s.handleHealth)
	http.HandleFunc("/ready", s.handleReady)

	// Web界面
	http.HandleFunc("/", s.handleWebIndex)
	http.HandleFunc("/dashboard", s.handleWebIndex)
	http.HandleFunc("/subscription-topics", s.handleWebIndex)
	http.HandleFunc("/static/", s.handleStaticFiles)

	// 监控指标
	http.HandleFunc("/metrics", s.handleMetrics)

	// 数据查询
	http.HandleFunc("/api/readings", DeviceNameMiddleware(s.handleQueryReadings))

	// 数据备份/恢复
	http.HandleFunc("/api/backup", s.handleBackup)
	http.HandleFunc("/api/restore", s.handleRestore)

	// 测试端点
	http.HandleFunc("/api/test-edgex", s.handleTestEdgeX)

	// 认证管理
	http.HandleFunc("/api/auth/create-key", s.handleCreateAPIKey)
	http.HandleFunc("/api/auth/list-keys", s.handleListAPIKeys)
	http.HandleFunc("/api/auth/revoke-key", s.handleRevokeAPIKey)

	// 加密管理
	http.HandleFunc("/api/encryption/rotate-key", s.handleRotateEncryptionKey)
	http.HandleFunc("/api/encryption/status", s.handleGetEncryptionStatus)

	// 表导入导出
	http.HandleFunc("/api/export/csv", s.handleExportCSV)
	http.HandleFunc("/api/export/json", s.handleExportJSON)
	http.HandleFunc("/api/export/sql", s.handleExportSQL)
	http.HandleFunc("/api/import/csv", s.handleImportCSV)
	http.HandleFunc("/api/import/json", s.handleImportJSON)
	http.HandleFunc("/api/data/export", s.handleDataExport)

	// 数据保留策略
	http.HandleFunc("/api/retention/status", s.handleRetentionStatus)
	http.HandleFunc("/api/retention/cleanup", s.handleManualCleanup)

	// 告警通知
	http.HandleFunc("/api/alerts/notifier/status", s.handleAlertNotifierStatus)
	http.HandleFunc("/api/alerts/test", s.handleTestAlert)
	http.HandleFunc("/api/alert-groups", s.handleAlertGroups)

	// 配置管理
	http.HandleFunc("/api/config/mqtt", s.handleMQTTConfigUpdate)
	http.HandleFunc("/api/config/get", s.handleGetConfig)
	http.HandleFunc("/api/config/update", s.handleUpdateConfig)
	http.HandleFunc("/api/config/reload", s.handleReloadConfig)
	http.HandleFunc("/api/config/oneclick", s.handleOneClickConfig)

	// 资源监控
	http.HandleFunc("/api/resources/status", s.handleResourceStatus)

	// 设备状态
	http.HandleFunc("/api/device-status", s.handleDeviceStatus)

	// 订阅管理
	http.HandleFunc("/api/subscription/status", s.handleSubscriptionStatus)
	http.HandleFunc("/api/subscription/test", s.handleSubscriptionTest)
	http.HandleFunc("/api/subscription/themes", s.handleSubscriptionThemes)

	// WebSocket
	http.HandleFunc("/ws", s.handleWebSocket)
}
