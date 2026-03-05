package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"
	"sfsdb-edgex-adapter/edgex"
)

// TestInvalidMessageFormat 测试无效的消息格式
func TestInvalidMessageFormat(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_invalid_message"
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

	// 测试无效的 JSON 格式
	invalidJSON := `invalid json format`

	// 解析消息
	var edgexMsg edgex.EdgeXMessage
	if err := json.Unmarshal([]byte(invalidJSON), &edgexMsg); err == nil {
		t.Errorf("Expected error when parsing invalid JSON, but got none")
	}

	t.Log("Invalid message format test passed")
}

// TestInvalidPayloadFormat 测试无效的 payload 格式
func TestInvalidPayloadFormat(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_invalid_payload"
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

	// 测试无效的 payload 格式
	message := `{
		"correlationId": "test-correlation-id",
		"messageType": "Event",
		"origin": 1640995200000000000,
		"payload": "invalid payload"
	}`

	// 解析消息
	var edgexMsg edgex.EdgeXMessage
	if err := json.Unmarshal([]byte(message), &edgexMsg); err != nil {
		t.Fatalf("Failed to parse EdgeX message: %v", err)
	}

	// 解析 payload
	var event edgex.EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err == nil {
		t.Errorf("Expected error when parsing invalid payload, but got none")
	}

	t.Log("Invalid payload format test passed")
}

// TestEmptyReadings 测试空的 readings 数组
func TestEmptyReadings(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_empty_readings"
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

	// 测试空的 readings 数组
	message := `{
		"correlationId": "test-correlation-id",
		"messageType": "Event",
		"origin": 1640995200000000000,
		"payload": {
			"id": "test-event-id",
			"deviceName": "test-device",
			"readings": [],
			"origin": 1640995200000000000
		}
	}`

	// 解析消息
	var edgexMsg edgex.EdgeXMessage
	if err := json.Unmarshal([]byte(message), &edgexMsg); err != nil {
		t.Fatalf("Failed to parse EdgeX message: %v", err)
	}

	// 解析 payload
	var event edgex.EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}

	// 验证 readings 数组为空
	if len(event.Readings) != 0 {
		t.Errorf("Expected empty readings array, got %d readings", len(event.Readings))
	}

	t.Log("Empty readings test passed")
}

// TestInvalidValueFormat 测试无效的 value 格式
func TestInvalidValueFormat(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_invalid_value"
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

	// 测试无效的 value 格式
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
					"value": "invalid-value",
					"origin": 1640995200000000000
				}
			],
			"origin": 1640995200000000000
		}
	}`

	// 解析消息
	var edgexMsg edgex.EdgeXMessage
	if err := json.Unmarshal([]byte(message), &edgexMsg); err != nil {
		t.Fatalf("Failed to parse EdgeX message: %v", err)
	}

	// 解析 payload
	var event edgex.EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}

	// 验证 value 解析
	for _, reading := range event.Readings {
		// 将字符串值转换为浮点数
		value := 0.0
		_, err := fmt.Sscanf(reading.Value, "%f", &value)
		if err == nil {
			t.Errorf("Expected error when parsing invalid value, but got none")
		}
	}

	t.Log("Invalid value format test passed")
}
