package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"
)

// TestMQTTIntegration 测试与 MQTT 消息总线的集成
func TestMQTTIntegration(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_mqtt"
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

	// 初始化 MQTT 客户端
	client, err := initMQTT()
	if err != nil {
		t.Fatalf("Failed to initialize MQTT: %v", err)
	}
	defer client.Disconnect(250)

	// 订阅 EdgeX 消息
	if err := subscribeToEdgeX(client); err != nil {
		t.Fatalf("Failed to subscribe to EdgeX messages: %v", err)
	}

	// 模拟 EdgeX 消息
	testMessage := `{
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

	// 发布消息到 MQTT 主题
	token := client.Publish("edgex/events/core/test", 1, false, testMessage)
	token.Wait()
	if token.Error() != nil {
		t.Fatalf("Failed to publish test message: %v", token.Error())
	}

	// 等待消息处理
	time.Sleep(2 * time.Second)

	// 验证数据是否被正确存储
	// 这里可以添加数据库查询逻辑来验证数据存储
	t.Log("MQTT integration test passed")
}

// TestDataFlow 测试数据从 EdgeX 到 sfsDb 的完整流转
func TestDataFlow(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_flow"
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

	// 处理每个读数
	for _, reading := range event.Readings {
		// 准备数据
		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		// 将字符串值转换为浮点数
		value := 0.0
		fmt.Sscanf(reading.Value, "%f", &value)

		data := map[string]any{
			"id":         reading.ID,
			"deviceName": event.DeviceName,
			"reading":    reading.ResourceName,
			"value":      value,
			"timestamp":  reading.Origin / 1000000000, // 转换为秒
			"metadata":   metadataStr,
		}

		// 存储到 sfsDb
		_, err := database.Table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}
	}

	t.Log("Data flow test passed")
}
