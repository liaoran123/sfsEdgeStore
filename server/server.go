package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sfsdb-edgex-adapter-enterprise/auth"
	"sfsdb-edgex-adapter-enterprise/backup"
	"sfsdb-edgex-adapter-enterprise/common"
	"sfsdb-edgex-adapter-enterprise/config"
	"sfsdb-edgex-adapter-enterprise/database"
	"sfsdb-edgex-adapter-enterprise/edgex"
	"sfsdb-edgex-adapter-enterprise/monitor"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// Server 结构
type Server struct {
	Table   *engine.Table
	Config  *config.Config
	Monitor *monitor.Monitor
}

// NewServer 创建一个新的服务器实例
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor) *Server {
	return &Server{
		Table:   table,
		Config:  cfg,
		Monitor: monitor,
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
	// 数据查询API - 使用中间件处理deviceName格式化和认证
	http.HandleFunc("/api/readings", auth.AuthMiddleware(DeviceNameMiddleware(s.handleQueryReadings)))

	// 数据备份API - 需要认证和备份权限
	http.HandleFunc("/api/backup", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionBackup, s.handleBackup)))

	// 数据恢复API - 需要认证和恢复权限
	http.HandleFunc("/api/restore", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionRestore, s.handleRestore)))

	// 测试端点，用于模拟EdgeX消息 - 需要认证和写权限
	http.HandleFunc("/api/test-edgex", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionWrite, s.handleTestEdgeX)))

	// 认证管理API - 需要认证和管理员权限
	http.HandleFunc("/api/auth/create-key", auth.AuthMiddleware(auth.PermissionMiddleware(auth.PermissionAdmin, s.handleCreateAPIKey)))
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
