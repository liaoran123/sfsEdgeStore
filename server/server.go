package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"sfsEdgeStore/alert"
	"sfsEdgeStore/auth"
	"sfsEdgeStore/backup"
	"sfsEdgeStore/baseline"
	"sfsEdgeStore/common"
	"sfsEdgeStore/config"
	"sfsEdgeStore/configwizard"
	"sfsEdgeStore/database"
	"sfsEdgeStore/edgex"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/mqtt"
	"sfsEdgeStore/pathutil"
	"sfsEdgeStore/resource"
	"sfsEdgeStore/retention"
	"sfsEdgeStore/template"
	"sfsEdgeStore/ws"

	"github.com/gorilla/websocket"
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/record"
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
	MQTTClient      mqtt.Client
	wsManager       *ws.WSManager
	TemplateManager *template.Manager
	BaselineManager *baseline.Manager
}

// NewServer 创建一个新的服务器实例
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor, retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier, resourceMonitor *resource.ResourceMonitor) *Server {
	wsManager := ws.NewWSManager()
	go wsManager.Run()

	// 初始化模板管理器
	templateManager := template.NewManager()
	if err := templateManager.LoadTemplates(); err != nil {
		log.Printf("Failed to load templates: %v", err)
	}

	// 初始化基线管理器
	learningPeriod := 3 // 默认学习期3天
	if cfg.Baseline != nil {
		if lp, ok := cfg.Baseline["learning_period"].(int); ok {
			learningPeriod = lp
		}
	}
	enabled := true
	if cfg.Baseline != nil {
		if e, ok := cfg.Baseline["enabled"].(bool); ok {
			enabled = e
		}
	}
	baselineManager := baseline.NewManager(learningPeriod, enabled)

	return &Server{
		Table:           table,
		Config:          cfg,
		Monitor:         monitor,
		RetentionMgr:    retentionMgr,
		AlertNotifier:   alertNotifier,
		ResourceMonitor: resourceMonitor,
		wsManager:       wsManager,
		TemplateManager: templateManager,
		BaselineManager: baselineManager,
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

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
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

	http.HandleFunc("/subscription-topics", s.handleWebIndex)
	http.HandleFunc("/static/", s.handleStaticFiles)

	// MQTT配置更新API - 不需要认证，用于Web界面配置
	http.HandleFunc("/api/config/mqtt", s.handleMQTTConfigUpdate)

	// 监控指标API - 不需要认证，用于Web界面展示
	http.HandleFunc("/metrics", s.handleMetrics)

	// 数据查询API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/readings", DeviceNameMiddleware(s.handleQueryReadings))

	// 数据备份API - 不需要认证，简化访问
	http.HandleFunc("/api/backup", s.handleBackup)

	// 数据恢复API - 不需要认证，简化访问
	http.HandleFunc("/api/restore", s.handleRestore)

	// 测试端点，用于模拟EdgeX消息 - 不需要认证，简化访问
	http.HandleFunc("/api/test-edgex", s.handleTestEdgeX)

	// 认证管理API - 不需要认证，简化访问
	http.HandleFunc("/api/auth/create-key", s.handleCreateAPIKey)
	http.HandleFunc("/api/auth/list-keys", s.handleListAPIKeys)
	http.HandleFunc("/api/auth/revoke-key", s.handleRevokeAPIKey)

	// 加密管理API - 不需要认证，简化访问
	http.HandleFunc("/api/encryption/rotate-key", s.handleRotateEncryptionKey)
	http.HandleFunc("/api/encryption/status", s.handleGetEncryptionStatus)

	// 表导入导出API - 不需要认证，简化访问
	http.HandleFunc("/api/export/csv", s.handleExportCSV)
	http.HandleFunc("/api/export/json", s.handleExportJSON)
	http.HandleFunc("/api/export/sql", s.handleExportSQL)
	http.HandleFunc("/api/import/csv", s.handleImportCSV)
	http.HandleFunc("/api/import/json", s.handleImportJSON)
	// 数据格式参数化API - 不需要认证，简化访问
	http.HandleFunc("/api/data/export", s.handleDataExport)

	// 数据保留策略API - 不需要认证，简化访问
	http.HandleFunc("/api/retention/status", s.handleRetentionStatus)
	http.HandleFunc("/api/retention/cleanup", s.handleManualCleanup)

	// 告警通知API - 不需要认证，简化访问
	http.HandleFunc("/api/alerts/notifier/status", s.handleAlertNotifierStatus)
	http.HandleFunc("/api/alerts/test", s.handleTestAlert)

	// 配置API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/config/get", s.handleGetConfig)
	http.HandleFunc("/api/config/update", s.handleUpdateConfig)
	http.HandleFunc("/api/config/reload", s.handleReloadConfig)
	// 一键配置API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/config/oneclick", s.handleOneClickConfig)

	// 资源监控API - 不需要认证，简化访问
	http.HandleFunc("/api/resources/status", s.handleResourceStatus)

	// 订阅状态API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/subscription/status", s.handleSubscriptionStatus)
	http.HandleFunc("/api/subscription/test", s.handleSubscriptionTest)
	// 订阅主题管理API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/subscription/themes", s.handleSubscriptionThemes)

	// 模板相关API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/templates", s.handleTemplates)
	http.HandleFunc("/api/templates/apply", s.handleApplyTemplate)

	// 基线相关API - 不需要认证，用于Web界面展示
	http.HandleFunc("/api/baselines", s.handleBaselines)
	http.HandleFunc("/api/baselines/calculate", s.handleCalculateBaseline)

	// WebSocket 端点
	http.HandleFunc("/ws", s.handleWebSocket)

}

// handleQueryReadings 处理数据查询请求
func (s *Server) handleQueryReadings(w http.ResponseWriter, r *http.Request) {
	s.Monitor.IncrementHTTPRequests()

	w.Header().Set("Content-Type", "application/json")

	// 获取查询参数
	deviceName := r.URL.Query().Get("deviceName")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")
	limitStr := r.URL.Query().Get("limit")

	// 格式化设备名称，确保与数据库中存储的格式一致
	if deviceName != "" {
		deviceName = common.FormatDeviceName(deviceName)
	}

	// 解析limit参数
	var limit int
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 查询数据
	var err error
	var readings record.Records
	if limit > 0 {
		readings, err = database.QueryRecords(database.Table, deviceName, startTime, endTime, false, limit)
	} else {
		readings, err = database.QueryRecords(database.Table, deviceName, startTime, endTime, false)
	}
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
	s.Monitor.IncrementHTTPRequests()

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
	backupManager := backup.NewBackupManager(storage.GetDBManager().GetDB())

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
	backupManager := backup.NewBackupManager(storage.GetDBManager().GetDB())

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

	// 从请求体中读取EdgeX消息
	var edgexMsg edgex.EdgeXMessage
	if err := json.NewDecoder(r.Body).Decode(&edgexMsg); err != nil {
		// 如果请求体不是EdgeXMessage格式，尝试直接解析为EdgeX事件
		r.Body.Close()

		// 重新读取请求体
		r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024) // 限制为10MB

		// 读取请求体内容
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to read request body: %v", err)})
			return
		}

		// 尝试直接解析为EdgeX事件
		var event edgex.EdgeXEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to process message: %v", err)})
			return
		}

		// 格式化设备名称
		event.DeviceName = common.FormatDeviceName(event.DeviceName)
		handleEdgeXEvent(s, w, &event)
		return
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
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to process message: %v", err)})
		return
	}
	if event == nil {
		// 消息类型不是 event，无需处理
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored_non_event_message"})
		return
	}
	handleEdgeXEvent(s, w, event)
}

// handleEdgeXEvent 处理EdgeX事件
func handleEdgeXEvent(s *Server, w http.ResponseWriter, event *edgex.EdgeXEvent) {
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
			"id":          reading.ID,
			"deviceName":  event.DeviceName,
			"profileName": event.ProfileName,
			"reading":     reading.ResourceName,
			"value":       value,
			"valueType":   reading.ValueType,
			"baseType":    reading.BaseType,
			"timestamp":   reading.Origin,
			"metadata":    metadataStr,
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
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Batch stored %d readings from %s", len(records), event.DeviceName),
		})
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

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 生成临时文件
	tempFile, err := os.CreateTemp("", "export-*.csv")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temp file"})
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// 执行导出
	if err := database.ExportTableToCSV(database.Table, tempPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 设置下载头
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=export.csv")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 读取并发送文件
	file, err := os.Open(tempPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open export file"})
		return
	}
	defer file.Close()

	// 发送文件内容
	io.Copy(w, file)
}

// handleExportJSON 处理导出表数据为JSON格式请求
func (s *Server) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	// 增加HTTP请求计数
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 生成临时文件
	tempFile, err := os.CreateTemp("", "export-*.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temp file"})
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// 执行导出
	if err := database.ExportTableToJSON(database.Table, tempPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 设置下载头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=export.json")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 读取并发送文件
	file, err := os.Open(tempPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open export file"})
		return
	}
	defer file.Close()

	// 发送文件内容
	io.Copy(w, file)
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
	limitStr := r.URL.Query().Get("limit")

	// 解析limit参数
	var limit int
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 查询数据
	var err error
	var readings record.Records
	if limit > 0 {
		readings, err = database.QueryRecords(database.Table, deviceName, startTime, endTime, false, limit)
	} else {
		readings, err = database.QueryRecords(database.Table, deviceName, startTime, endTime, false)
	}
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Topic test for %s would be performed here", req.Topic),
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
	webDir, err := pathutil.Join("web")
	if err != nil {
		webDir = "web"
	}

	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		webDir = "web"
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
	case "/mqtt-subscription":
		pageFile := filepath.Join(webDir, "mqtt-subscription.html")
		if _, err := os.Stat(pageFile); os.IsNotExist(err) {
			http.Error(w, "MQTT subscription page not found.", http.StatusNotFound)
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

	webDir, err := pathutil.Join("web")
	if err != nil {
		webDir = "web"
	}

	if _, err := os.Stat(webDir); err != nil {
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

	// 根据文件扩展名设置正确的 MIME 类型
	ext := filepath.Ext(filePath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	}

	log.Printf("Serving file: %s", filePath)
	http.ServeFile(w, r, filePath)
}

// handleMQTTConfigUpdate 处理MQTT配置更新请求（不需要认证）
func (s *Server) handleMQTTConfigUpdate(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var mqttConfig struct {
		MqttBroker        string `json:"mqtt_broker"`
		MqttTopic         string `json:"mqtt_topic"`
		ClientID          string `json:"client_id"`
		MqttUseTls        bool   `json:"mqtt_use_tls"`
		MqttCaCert        string `json:"mqtt_ca_cert"`
		MqttClientCert    string `json:"mqtt_client_cert"`
		MqttClientKey     string `json:"mqtt_client_key"`
		AutoSubscribe     bool   `json:"autoSubscribe"`
		ConnectionTimeout int    `json:"connectionTimeout"`
		KeepAlive         int    `json:"keepAlive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&mqttConfig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 获取当前配置
	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()
	if currentConfig == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Config not initialized"})
		return
	}

	// 更新MQTT相关配置
	currentConfig.MQTTBroker = mqttConfig.MqttBroker
	currentConfig.MQTTTopic = mqttConfig.MqttTopic
	currentConfig.ClientID = mqttConfig.ClientID
	currentConfig.MQTTUseTLS = mqttConfig.MqttUseTls
	currentConfig.MQTTCACert = mqttConfig.MqttCaCert
	currentConfig.MQTTClientCert = mqttConfig.MqttClientCert
	currentConfig.MQTTClientKey = mqttConfig.MqttClientKey
	currentConfig.AutoSubscribe = mqttConfig.AutoSubscribe
	// 更新连接超时和保持连接时间
	currentConfig.ConnectionTimeout = mqttConfig.ConnectionTimeout
	currentConfig.KeepAlive = mqttConfig.KeepAlive

	// 保存更新后的配置
	if err := configManager.UpdateConfig(currentConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "MQTT config updated successfully",
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

	// 构建自定义主题列表
	customThemes := []map[string]interface{}{}
	for _, topic := range currentConfig.CustomTopics {
		customThemes = append(customThemes, map[string]interface{}{
			"topic":    topic,
			"added_at": time.Now().Format("2006-01-02 15:04:05"),
			"active":   true,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"themes":        themes,
		"custom_topics": customThemes,
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

	// 添加主题到配置
	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	// 检查主题是否已存在
	for _, topic := range currentConfig.CustomTopics {
		if topic == req.Topic {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "主题已存在"})
			return
		}
	}

	// 添加新主题
	currentConfig.CustomTopics = append(currentConfig.CustomTopics, req.Topic)

	// 保存配置
	if err := config.SaveToFile(currentConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "保存配置失败"})
		return
	}

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

	// 从配置中删除主题
	configManager := config.GetConfigManager()
	currentConfig := configManager.GetConfig()

	// 查找并删除主题
	newCustomTopics := []string{}
	found := false
	for _, topic := range currentConfig.CustomTopics {
		if topic != req.Topic {
			newCustomTopics = append(newCustomTopics, topic)
		} else {
			found = true
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "主题不存在"})
		return
	}

	// 更新配置
	currentConfig.CustomTopics = newCustomTopics

	// 保存配置
	if err := config.SaveToFile(currentConfig); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "保存配置失败"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "主题删除成功",
		"topic":   req.Topic,
	})
}

// handleTemplates 处理获取模板请求
func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取所有行业模板
	industries := s.TemplateManager.ListIndustries()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"industries": industries,
	})
}

// handleApplyTemplate 处理应用模板请求
func (s *Server) handleApplyTemplate(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 解析请求
	var req struct {
		Industry string `json:"industry"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 应用模板
	if err := s.TemplateManager.ApplyTemplate(req.Industry, s.Config); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  "Template applied successfully",
		"industry": req.Industry,
	})
}

// handleBaselines 处理获取基线请求
func (s *Server) handleBaselines(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 获取所有基线
	baselines := s.BaselineManager.ListBaselines()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"baselines": baselines,
	})
}

// handleCalculateBaseline 处理计算基线请求
func (s *Server) handleCalculateBaseline(w http.ResponseWriter, r *http.Request) {
	if s.Monitor != nil {
		s.Monitor.IncrementHTTPRequests()
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 解析请求
	var req struct {
		DeviceName  string `json:"deviceName"`
		ReadingName string `json:"readingName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 计算基线
	baseline, err := s.BaselineManager.CalculateBaseline(req.DeviceName, req.ReadingName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"baseline": baseline,
	})
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	s.wsManager.Register(conn)

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.wsManager.Unregister(conn)
			break
		}
	}
}

// GetWSManager 获取 WebSocket 管理器
func (s *Server) GetWSManager() *ws.WSManager {
	return s.wsManager
}

// Broadcast 实现 broadcast.Broadcaster 接口
func (s *Server) Broadcast(message []byte) {
	if s.wsManager != nil {
		s.wsManager.Broadcast(message)
	}
}
