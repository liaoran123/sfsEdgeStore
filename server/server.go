package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sfsdb-edgex-adapter/backup"
	"sfsdb-edgex-adapter/common"
	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// EdgeX 消息结构（符合 MessageEnvelope 格式）
type EdgeXMessage struct {
	CorrelationID string          `json:"correlationId,omitempty"`
	MessageType   string          `json:"messageType,omitempty"`
	Origin        int64           `json:"origin,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// EdgeX 事件结构
type EdgeXEvent struct {
	ID          string         `json:"id"`
	DeviceName  string         `json:"deviceName"`
	Readings    []EdgeXReading `json:"readings"`
	Origin      int64          `json:"origin"`
	ProfileName string         `json:"profileName,omitempty"`
	SourceName  string         `json:"sourceName,omitempty"`
}

// EdgeX 读数结构
type EdgeXReading struct {
	ID           string          `json:"id"`
	ResourceName string          `json:"resourceName"`
	Value        string          `json:"value"`
	ValueType    string          `json:"valueType,omitempty"`
	Origin       int64           `json:"origin"`
	ProfileName  string          `json:"profileName,omitempty"`
	DeviceName   string          `json:"deviceName,omitempty"`
	BaseType     string          `json:"baseType,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// Server 结构

type Server struct {
	Table  *engine.Table
	Config *config.Config
}

// NewServer 创建一个新的服务器实例
func NewServer(table *engine.Table, cfg *config.Config) *Server {
	return &Server{
		Table:  table,
		Config: cfg,
	}
}

// HTTP 用于提供外部接口和管理功能
// Start 启动HTTP服务器
func (s *Server) Start() error {
	// 注册路由
	s.registerRoutes()

	// 在后台启动HTTP服务器
	go func() {
		log.Println("Starting HTTP server for health checks on port 8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// registerRoutes 注册HTTP路由
func (s *Server) registerRoutes() {
	// 健康检查接口
	http.HandleFunc("/health", s.handleHealthCheck)

	// 数据查询API
	http.HandleFunc("/api/readings", s.handleQueryReadings)

	// 数据备份API
	http.HandleFunc("/api/backup", s.handleBackup)

	// 数据恢复API
	http.HandleFunc("/api/restore", s.handleRestore)

	// 测试端点，用于模拟EdgeX消息
	http.HandleFunc("/api/test-edgex", s.handleTestEdgeX)
}

// handleHealthCheck 处理健康检查请求
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleQueryReadings 处理数据查询请求
func (s *Server) handleQueryReadings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取查询参数
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
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// 模拟EdgeX消息
	edgexMsg := EdgeXMessage{
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

	// 解析payload中的事件
	var event EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
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
			"deviceName": event.DeviceName,
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
