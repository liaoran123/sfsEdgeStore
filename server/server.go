package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"sfsEdgeStore/alert"
	"sfsEdgeStore/auth"
	"sfsEdgeStore/backup"
	"sfsEdgeStore/common"
	"sfsEdgeStore/config"
	"sfsEdgeStore/configwizard"
	"sfsEdgeStore/core/database"
	"sfsEdgeStore/core/edgex"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/resource"
	"sfsEdgeStore/retention"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// Server 结构
type Server struct {
	Table           *engine.Table
	Config          *config.Config
	Monitor         *monitor.Monitor
	RetentionMgr    *retention.RetentionManager
	AlertNotifier   *alert.Notifier
	ResourceMonitor *resource.ResourceMonitor
}

// NewServer 创建一个新的服务器实例
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor, retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier, resourceMonitor *resource.ResourceMonitor) *Server {
	return &Server{
		Table:           table,
		Config:          cfg,
		Monitor:         monitor,
		RetentionMgr:    retentionMgr,
		AlertNotifier:   alertNotifier,
		ResourceMonitor: resourceMonitor,
	}
}

// HTTP 用于提供外部接口和管理功能
// Start 启动HTTP服务器
func (s *Server) Start() error {
	// 注册路由
	s.registerRoutes()

	// 在后台启动HTTP服务器
	go func() {
		port := s.Config.HTTPPort
		if port == "" {
			port = "8081" // 默认端口
		}

		if s.Config.HTTPUseTLS && s.Config.HTTPCert != "" && s.Config.HTTPKey != "" {
			// 使用 HTTPS
			log.Printf("Starting HTTPS server for health checks on port %s", port)
			if err := http.ListenAndServeTLS(":"+port, s.Config.HTTPCert, s.Config.HTTPKey, nil); err != nil {
				log.Printf("HTTPS server error: %v", err)
			}
		} else {
			// 使用 HTTP
			log.Printf("Starting HTTP server for health checks on port %s", port)
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}
	}()

	return nil
}

// DeviceNameMiddleware 处理HTTP请求中的deviceName参数格式化
func DeviceNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取并格式化deviceName参数
		deviceName := r.URL.Query().Get("deviceName")
		if deviceName != "" {
			formattedDeviceName := common.FormatDeviceName(deviceName)
			// 重写URL参数
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
	// 健康检查端点 - 不需要认证，用于负载均衡器等
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/healthz", s.handleHealth)
	http.HandleFunc("/ready", s.handleReady)

	// Web界面路由 - 开源版基础功能
	http.HandleFunc("/", s.handleWebIndex)
	http.HandleFunc("/dashboard", s.handleWebIndex)
	http.HandleFunc("/mqtt-config", s.handleWebIndex)
	http.HandleFunc("/auth-config", s.handleWebIndex)
	http.HandleFunc("/subscription-status", s.handleWebIndex)
	http.HandleFunc("/subscription-topics", s.handleWebIndex)
	http.HandleFunc("/static/", s.handleStaticFiles)

	// 许可证信息API - 不需要认证
	http.HandleFunc("/api/license", s.handleLicenseInfo)

	// 监控指标API - 不需要认证，用于Web界面展示
	http.HandleFunc("/metrics", s.handleMetrics)

	// 数据查询API - 使用中间件处理deviceName格式化和认证
	http.HandleFunc("/api/readings", auth.AuthMiddleware(DeviceNameMiddleware(s.handleQueryReadings)))

	// 数据备份API - 需要认证和备份权限
	http.HandleFunc("/api/backup", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionBackup, s.handleBackup)))

	// 数据恢复API - 需要认证和恢复权限
	http.HandleFunc("/api/restore", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionRestore, s.handleRestore)))

	// 测试端点，用于模拟EdgeX消息 - 需要认证和写权限
	http.HandleFunc("/api/test-edgex", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionWrite, s.handleTestEdgeX)))

	// 认证管理API - 需要认证和管理员权限（除了创建API Key）
	http.HandleFunc("/api/auth/create-key", s.handleCreateAPIKey)
	http.HandleFunc("/api/auth/list-keys", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleListAPIKeys)))
	http.HandleFunc("/api/auth/revoke-key", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleRevokeAPIKey)))

	// 加密管理API - 需要认证和管理员权限
	http.HandleFunc("/api/encryption/rotate-key", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleRotateEncryptionKey)))
	http.HandleFunc("/api/encryption/status", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleGetEncryptionStatus)))

	// 表导入导出API - 需要认证和备份权限
	http.HandleFunc("/api/export/csv", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionBackup, s.handleExportCSV)))
	http.HandleFunc("/api/export/json", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionBackup, s.handleExportJSON)))
	http.HandleFunc("/api/export/sql", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionBackup, s.handleExportSQL)))
	http.HandleFunc("/api/import/csv", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionRestore, s.handleImportCSV)))
	http.HandleFunc("/api/import/json", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionRestore, s.handleImportJSON)))
	// 数据格式参数化API - 需要认证和备份权限
	http.HandleFunc("/api/data/export", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionBackup, s.handleDataExport)))

	// 数据保留策略API - 需要认证和管理员权限
	http.HandleFunc("/api/retention/status", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleRetentionStatus)))
	http.HandleFunc("/api/retention/cleanup", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleManualCleanup)))

	// 告警通知API - 需要认证和管理员权限
	http.HandleFunc("/api/alerts/notifier/status", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleAlertNotifierStatus)))
	http.HandleFunc("/api/alerts/test", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleTestAlert)))

	// 配置热更新API - 需要认证和管理员权限
	// 配置API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/config/get", s.handleGetConfig)
	http.HandleFunc("/api/config/update", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleUpdateConfig)))
	http.HandleFunc("/api/config/reload", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleReloadConfig)))
	// 一键配置API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/config/oneclick", s.handleOneClickConfig)

	// 资源监控API - 需要认证和管理员权限
	http.HandleFunc("/api/resources/status", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleResourceStatus)))

	// 订阅状态API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/subscription/status", s.handleSubscriptionStatus)
	http.HandleFunc("/api/subscription/test", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleSubscriptionTest)))
	// 订阅主题管理API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/subscription/themes", s.handleSubscriptionThemes)

}

// handleQueryReadings 处理数据查询请求
func (s *Server) handleQueryReadings(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	// 获取查询参数（deviceName已由中间件格式化）
	deviceName := r.URL.Query().Get("deviceName")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")

	// 查询数据
	readings, err := database.QueryRecords(database.Table, deviceName, startTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer readings.Release()

	// 转换为map切片以进行JSON编码
	readingsMap := make([]map[string]any, len(readings))
	for i, reading := range readings {
		readingsMap[i] = reading
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":    len(readings),
		"readings": readingsMap,
	})
}

// handleBackup 处理数据备份请求
func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取备份路径参数
	backupPath := r.URL.Query().Get("path")
	if backupPath == "" {
		backupPath = "./backups"
	}

	// 创建备份管理器
	backupManager := backup.NewBackupManager(storage.KVDb)

	// 执行备份
	backupFile, err := backupManager.Backup(backupPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"backupFile": backupFile,
	})
}

// handleRestore 处理数据恢复请求
func (s *Server) handleRestore(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取备份文件路径
	backupFile := r.URL.Query().Get("file")
	if backupFile == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Backup file path is required"})
		return
	}

	// 创建备份管理器
	backupManager := backup.NewBackupManager(storage.KVDb)

	// 执行恢复
	if err := backupManager.Restore(backupFile); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Database restored successfully",
	})
}

// handleTestEdgeX 处理测试EdgeX消息请求
func (s *Server) handleTestEdgeX(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 模拟EdgeX消息
	edgexMsg := edgex.EdgeXMessage{
		CorrelationID: "test-correlation-id",
		MessageType:   "event",
		Origin:        time.Now().UnixNano(),
		Payload: json.RawMessage(`{
			"id": "test-event-id",
			"deviceName": "TestDevice-001",
			"readings": [
				{
					"id": "reading-1",
				"resourceName": "temperature",
				"value": "25.5",
				"valueType": "Float32",
				"baseType": "Float",
				"origin": 1677721600000000000,
				"deviceName": "TestDevice-001"
				},
				{
					"id": "reading-2",
				"resourceName": "humidity",
				"value": "45",
				"valueType": "Int32",
				"baseType": "Int",
				"origin": 1677721600000000000,
				"deviceName": "TestDevice-001"
				},
				{
					"id": "reading-3",
				"resourceName": "pressure",
				"value": "1013.25",
				"valueType": "Float64",
				"baseType": "Float",
				"origin": 1677721600000000000,
				"deviceName": "TestDevice-001"
				}
			],
			"origin": 1677721600000000000
		}`),
	}

	// 转换为字节数组并使用edgex包处理
	msgBytes, err := json.Marshal(edgexMsg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	event, err := edgex.ProcessMessage(msgBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 收集所有读数，准备批量插入
	var records []*map[string]any

	// 处理每个读数
	for _, reading := range event.Readings {
		// 准备数据
		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		// 解析值的类型
		value := common.ParseValue(reading.Value)

		data := map[string]any{
			"id":         reading.ID,
			"deviceName": event.DeviceName, // 设备名称已经在ProcessMessage中格式化
			"reading":    reading.ResourceName,
			"value":      value,
			"valueType":  reading.ValueType,
			"baseType":   reading.BaseType,
			"timestamp":  reading.Origin, // 纳秒级时间戳，类型为 int64
			"metadata":   metadataStr,
		}

		records = append(records, &data)
	}

	// 批量存储到sfsDb
	if len(records) > 0 {
		_, err := s.Table.BatchInsertNoInc(records)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		} else {
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": fmt.Sprintf("Batch stored %d readings from %s", len(records), event.DeviceName),
			})
		}
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "No readings to store",
		})
	}
}

// handleCreateAPIKey 处理创建API Key请求
func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 解析请求参数
	var req struct {
		UserID    string `json:"user_id"`
		Role      string `json:"role"`
		ExpiresIn int    `json:"expires_in"` // 过期时间（小时）
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 验证参数
	if req.UserID == "" || req.Role == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID and role are required"})
		return
	}

	// 创建认证管理器
	authManager := auth.NewAuthManager()

	// 创建API Key
	apiKey, err := authManager.CreateAPIKey(req.UserID, req.Role, time.Duration(req.ExpiresIn)*time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"api_key":    apiKey.Key,
		"user_id":    apiKey.UserID,
		"role":       apiKey.Role,
		"expires_at": apiKey.ExpiresAt,
	})
}

// handleListAPIKeys 处理列出API Key请求
func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 创建认证管理器
	authManager := auth.NewAuthManager()

	// 获取API Keys
	apiKeys, err := authManager.ListAPIKeys()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 转换为响应格式
	var response []map[string]interface{}
	for _, key := range apiKeys {
		response = append(response, map[string]interface{}{
			"id":         key.ID,
			"user_id":    key.UserID,
			"role":       key.Role,
			"created_at": key.CreatedAt,
			"expires_at": key.ExpiresAt,
			"active":     key.Active,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"api_keys": response,
	})
}

// handleRevokeAPIKey 处理撤销API Key请求
func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 解析请求参数
	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 验证参数
	if req.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "API key is required"})
		return
	}

	// 创建认证管理器
	authManager := auth.NewAuthManager()

	// 撤销API Key
	if err := authManager.RevokeAPIKey(req.APIKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "API key revoked successfully",
	})
}

// handleRotateEncryptionKey 处理密钥轮换请求
func (s *Server) handleRotateEncryptionKey(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 解析请求参数
	var req struct {
		NewKey string `json:"new_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 验证参数
	if req.NewKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "New encryption key is required"})
		return
	}

	// 执行密钥轮换
	if err := database.RotateEncryptionKey(req.NewKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Encryption key rotated successfully",
	})
}

// handleGetEncryptionStatus 处理获取加密状态请求
func (s *Server) handleGetEncryptionStatus(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取加密状态
	enabled, algorithm, err := database.GetEncryptionStatus()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"enabled":   enabled,
		"algorithm": algorithm,
	})
}

// handleExportCSV 处理导出表数据为CSV格式请求
func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取文件路径参数
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		filePath = "./export.csv"
	}

	// 执行导出
	if err := database.ExportTableToCSV(database.Table, filePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"file":   filePath,
	})
}

// handleExportJSON 处理导出表数据为JSON格式请求
func (s *Server) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取文件路径参数
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		filePath = "./export.json"
	}

	// 执行导出
	if err := database.ExportTableToJSON(database.Table, filePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"file":   filePath,
	})
}

// handleExportSQL 处理导出表数据为SQL格式请求
func (s *Server) handleExportSQL(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取文件路径参数
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		filePath = "./export.sql"
	}

	// 执行导出
	if err := database.ExportTableToSQL(database.Table, filePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"file":   filePath,
	})
}

// handleImportCSV 处理从CSV文件导入数据到表请求
func (s *Server) handleImportCSV(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取文件路径参数
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File path is required"})
		return
	}

	// 获取批量大小参数
	batchSize := 100 // 默认值

	// 执行导入
	if err := database.ImportTableFromCSV(database.Table, filePath, batchSize); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data imported successfully from CSV",
	})
}

// handleImportJSON 处理从JSON文件导入数据到表请求
func (s *Server) handleImportJSON(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取文件路径参数
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File path is required"})
		return
	}

	// 获取批量大小参数
	batchSize := 100 // 默认值

	// 执行导入
	if err := database.ImportTableFromJSON(database.Table, filePath, batchSize); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data imported successfully from JSON",
	})
}

// handleDataExport 处理数据导出请求，支持格式参数化
func (s *Server) handleDataExport(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取查询参数
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // 默认格式
	}

	deviceName := r.URL.Query().Get("deviceName")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")

	// 查询数据
	readings, err := database.QueryRecords(database.Table, deviceName, startTime, endTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer readings.Release()

	// 转换为map切片
	readingsMap := make([]map[string]any, len(readings))
	for i, reading := range readings {
		readingsMap[i] = reading
	}

	// 根据格式返回数据
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(readings),
			"readings": readingsMap,
		})
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=data.csv")
		// 写入CSV表头
		if len(readingsMap) > 0 {
			// 获取所有字段名
			fields := make([]string, 0)
			for field := range readingsMap[0] {
				fields = append(fields, field)
			}
			// 创建CSV写入器
			writer := csv.NewWriter(w)
			defer writer.Flush()
			// 写入表头
			writer.Write(fields)
			// 写入数据
			for _, reading := range readingsMap {
				row := make([]string, len(fields))
				for i, field := range fields {
					value := reading[field]
					row[i] = fmt.Sprintf("%v", value)
				}
				writer.Write(row)
			}
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unsupported format"})
	}
}

// handleRetentionStatus 处理获取保留策略状态请求
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

// handleManualCleanup 处理手动数据清理请求
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

// handleAlertNotifierStatus 处理获取告警通知器状态请求
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

	s.Monitor.RecordError(req.Type, req.Message)

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

// handleGetConfig 处理获取配置请求
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   currentConfig,
	})
}

// handleUpdateConfig 处理更新配置请求
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var newConfig config.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	configManager := config.GetConfigManager()
	if err := configManager.UpdateConfig(&newConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Config updated successfully",
	})
}

// handleReloadConfig 处理重新加载配置请求
func (s *Server) handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	newConfig, err := config.ReloadFromFile()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	configManager := config.GetConfigManager()
	if err := configManager.UpdateConfig(newConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Config reloaded successfully",
	})
}

// handleOneClickConfig 处理一键配置请求
func (s *Server) handleOneClickConfig(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 创建配置向导
	wizard := configwizard.NewWizard(s.Config)

	// 运行配置向导进行智能检测
	if err := wizard.Run(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 获取配置管理器
	configManager := config.GetConfigManager()

	// 应用智能默认值
	smartConfig := &config.Config{
		DBPath:        "./data",
		MQTTBroker:    "tcp://localhost:1883",
		MQTTTopic:     "edgex/events/#",
		ClientID:      config.GenerateClientID(),
		HTTPPort:      "8081",
		DevConfigPath: "./devconfig",
		AutoSubscribe: true,
		// 智能默认值（简化初次使用）
		MQTTUseTLS:               false,
		HTTPUseTLS:               false,
		DBUseEncryption:          false,
		EnableAnalyzer:           false,
		EnableRetentionPolicy:    false,
		EnableAlertNotifications: false,
		EnableResourceMonitoring: true,
		DBScenario:               config.ScenarioEdge,
		LicenseType:              "community",
	}

	// 保存智能配置
	if err := configManager.UpdateConfig(smartConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "One-click configuration completed successfully",
		"config":  smartConfig,
	})
}

// handleResourceStatus 处理获取资源状态请求
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

// handleSubscriptionStatus 处理获取订阅状态请求
func (s *Server) handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	status := map[string]interface{}{
		"connected":      s.Monitor != nil && s.Monitor.IsMQTTConnected(),
		"broker":         currentConfig.MQTTBroker,
		"topic":          currentConfig.MQTTTopic,
		"client_id":      currentConfig.ClientID,
		"use_tls":        currentConfig.MQTTUseTLS,
		"username":       currentConfig.MQTTUsername,
		"ca_cert":        currentConfig.MQTTCACert != "",
		"client_cert":    currentConfig.MQTTClientCert != "",
		"auto_subscribe": true,
		"edgex_version":  "latest",
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

// handleSubscriptionTest 处理测试订阅请求
func (s *Server) handleSubscriptionTest(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		Broker string `json:"broker"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Broker == "" {
		req.Broker = "tcp://localhost:1883"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Connection test to %s would be performed here", req.Broker),
	})
}

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
		"mqtt":     true, // MQTT 连接状态后续可以增强
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

// handleWebIndex 处理Web界面首页
func (s *Server) handleWebIndex(w http.ResponseWriter, r *http.Request) {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		http.Error(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}
	webDir := filepath.Join(filepath.Dir(execPath), "web")
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		webDir = "./web"
	}

	// 根据路径返回不同的页面
	switch r.URL.Path {
	case "/", "/dashboard":
		indexFile := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			http.Error(w, "Web interface not found. Please ensure the 'web' directory exists.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, indexFile)
	case "/mqtt-config":
		pageFile := filepath.Join(webDir, "mqtt-config.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "MQTT config page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/auth-config":
		pageFile := filepath.Join(webDir, "auth-config.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "Auth config page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/subscription-status":
		pageFile := filepath.Join(webDir, "subscription-status.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "Subscription status page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/subscription-topics":
		pageFile := filepath.Join(webDir, "subscription-topics.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "Subscription topics page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/mqtt-subscription":
		pageFile := filepath.Join(webDir, "mqtt-subscription.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "MQTT subscription page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/api-keys":
		pageFile := filepath.Join(webDir, "api-keys.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "API keys page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/data-records":
		pageFile := filepath.Join(webDir, "data-records.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "Data records page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/system-health":
		pageFile := filepath.Join(webDir, "system-health.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "System health page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/settings":
		pageFile := filepath.Join(webDir, "settings.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "Settings page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	case "/tls-config":
		pageFile := filepath.Join(webDir, "tls-config.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "TLS config page not found.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, pageFile)
	default:
		// 对于其他路径，尝试作为静态文件处理
		s.handleStaticFiles(w, r)
	}
}

// handleStaticFiles 处理静态文件请求
func (s *Server) handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("Static file request: %s", r.URL.Path)

	// 尝试多个可能的web目录位置
	webDirs := []string{
		"./web",
		"web",
	}

	var webDir string
	for _, dir := range webDirs {
		if _, err := os.Stat(dir); err == nil {
			webDir = dir
			log.Printf("Found web directory: %s", webDir)
			break
		}
	}

	if webDir == "" {
		// 最后尝试基于可执行文件路径
		execPath, err := os.Executable()
		if err == nil {
			webDir = filepath.Join(filepath.Dir(execPath), "web")
			log.Printf("Trying web directory from executable path: %s", webDir)
		}
	}

	if webDir == "" {
		log.Println("Web directory not found")
		http.Error(w, "Web directory not found", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(webDir, "static", filepath.FromSlash(r.URL.Path[len("/static/"):]))
	log.Printf("Trying to serve file: %s", filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File not found: %s", filePath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	log.Printf("Serving file: %s", filePath)
	http.ServeFile(w, r, filePath)
}

// handleLicenseInfo 处理许可证信息请求
func (s *Server) handleLicenseInfo(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfigManager().GetConfig()
	if cfg == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"license_type": "opensource",
			"features": config.EnterpriseFeatures{
				MaxDevices: 50,
			},
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"license_type": cfg.LicenseType,
		"features":     cfg.EnterpriseFeatures,
	})
}

// handleMetrics 处理监控指标请求
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if s.Monitor == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Monitor not initialized"})
		return
	}

	metrics := s.Monitor.CollectMetrics()

	// 获取 ResourceMonitor 的真实 CPU 使用率
	if s.ResourceMonitor != nil {
		usage := s.ResourceMonitor.GetCurrentUsage()
		metrics.System.CPUUsage = usage.CPUPercent
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleSubscriptionThemes 处理订阅主题管理请求
func (s *Server) handleSubscriptionThemes(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.handleGetSubscriptionThemes(w, r)
	case http.MethodPost:
		s.handleAddSubscriptionTheme(w, r)
	case http.MethodDelete:
		s.handleDeleteSubscriptionTheme(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

// handleGetSubscriptionThemes 处理获取订阅主题列表请求
func (s *Server) handleGetSubscriptionThemes(w http.ResponseWriter, r *http.Request) {
	// 从配置中获取订阅主题
	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	// 构建主题列表
	themes := []string{}
	if currentConfig.MQTTTopic != "" {
		themes = append(themes, currentConfig.MQTTTopic)
	}

	// 这里可以从其他地方获取自定义主题
	// 暂时返回默认主题

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"themes": themes,
	})
}

// handleAddSubscriptionTheme 处理添加订阅主题请求
func (s *Server) handleAddSubscriptionTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Topic == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Topic is required"})
		return
	}

	// 这里可以添加主题到配置或其他存储
	// 暂时返回成功

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "主题添加成功",
		"topic":   req.Topic,
	})
}

// handleDeleteSubscriptionTheme 处理删除订阅主题请求
func (s *Server) handleDeleteSubscriptionTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Topic == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Topic is required"})
		return
	}

	// 这里可以从配置或其他存储中删除主题
	// 暂时返回成功

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "主题删除成功",
		"topic":   req.Topic,
	})
}
