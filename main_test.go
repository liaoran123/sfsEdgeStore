package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"
)

// TestHealthCheck 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	// 启动 HTTP 服务器
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})
		http.ListenAndServe(":8081", nil)
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 发送健康检查请求
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Fatalf("Failed to send health check request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// 检查响应内容
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", result["status"])
	}

	t.Log("Health check test passed")
}

// TestEdgeXMessageParsing 测试 EdgeX 消息解析
func TestEdgeXMessageParsing(t *testing.T) {
	// 模拟 EdgeX 消息
	message := `{
		"correlationId": "test-correlation-id",
		"messageType": "Event",
		"origin": 1640995200000000000,
		"payload": {
			"id": "test-event-id",
			"deviceName": "test-device",
			"readings": [
				{
					"id": "test-reading-id",
					"resourceName": "temperature",
					"value": "25.5",
					"origin": 1640995200000000000
				}
			],
			"origin": 1640995200000000000
		}
	}`

	// 解析消息
	var edgexMsg EdgeXMessage
	if err := json.Unmarshal([]byte(message), &edgexMsg); err != nil {
		t.Fatalf("Failed to parse EdgeX message: %v", err)
	}

	// 解析 payload
	var event EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}

	// 验证消息内容
	if event.DeviceName != "test-device" {
		t.Errorf("Expected deviceName 'test-device', got '%s'", event.DeviceName)
	}

	if len(event.Readings) != 1 {
		t.Errorf("Expected 1 reading, got %d", len(event.Readings))
	}

	if event.Readings[0].ResourceName != "temperature" {
		t.Errorf("Expected resourceName 'temperature', got '%s'", event.Readings[0].ResourceName)
	}

	t.Log("EdgeX message parsing test passed")
}

// TestDatabaseInitialization 测试数据库初始化
func TestDatabaseInitialization(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_init"
	os.Setenv("EDGEX_DB_PATH", dbPath)

	// 加载配置
	var err error
	appConfig, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Init(appConfig.DBPath); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	t.Log("Database initialization test passed")
}
