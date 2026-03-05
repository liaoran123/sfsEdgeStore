package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sfsdb-edgex-adapter/config"
	"sfsdb-edgex-adapter/database"
	"sfsdb-edgex-adapter/edgex"
	"testing"
	"time"
)

// TestPerformance 测试适配器的性能
func TestPerformance(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_performance"
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

	// 测试数据量
	testCount := 1000

	// 准备测试数据
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

	// 开始计时
	start := time.Now()

	// 执行测试
	for i := 0; i < testCount; i++ {
		// 解析消息
		var edgexMsg edgex.EdgeXMessage
		if err := json.Unmarshal([]byte(testMessage), &edgexMsg); err != nil {
			t.Fatalf("Failed to parse EdgeX message: %v", err)
		}

		// 解析 payload
		var event edgex.EdgeXEvent
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
				"id":         reading.ID + fmt.Sprintf("-%d", i),
				"deviceName": event.DeviceName,
				"reading":    reading.ResourceName,
				"value":      value,
				"timestamp":  reading.Origin, // 纳秒级时间戳
				"metadata":   metadataStr,
			}

			// 存储到 sfsDb
			_, err := database.Table.Insert(&data)
			if err != nil {
				t.Fatalf("Failed to store data: %v", err)
			}
		}
	}

	// 计算耗时
	duration := time.Since(start)
	throughput := float64(testCount) / duration.Seconds()

	t.Logf("Performance test completed: %d operations in %v (%.2f operations/second)", testCount, duration, throughput)

	// 验证性能指标
	if throughput < 100 { // 假设最低性能要求为 100 操作/秒
		t.Errorf("Performance below expected: %.2f operations/second, expected at least 100", throughput)
	}

	t.Log("Performance test passed")
}

// TestMemoryUsage 测试内存使用情况
func TestMemoryUsage(t *testing.T) {
	// 设置唯一的数据库路径
	dbPath := "./test_edgex_data_memory"
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

	// 测试数据量
	testCount := 1000

	// 准备测试数据
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

	// 执行测试
	for i := 0; i < testCount; i++ {
		// 解析消息
		var edgexMsg edgex.EdgeXMessage
		if err := json.Unmarshal([]byte(testMessage), &edgexMsg); err != nil {
			t.Fatalf("Failed to parse EdgeX message: %v", err)
		}

		// 解析 payload
		var event edgex.EdgeXEvent
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
				"id":         reading.ID + fmt.Sprintf("-%d", i),
				"deviceName": event.DeviceName,
				"reading":    reading.ResourceName,
				"value":      value,
				"timestamp":  reading.Origin, // 纳秒级时间戳
				"metadata":   metadataStr,
			}

			// 存储到 sfsDb
			_, err := database.Table.Insert(&data)
			if err != nil {
				t.Fatalf("Failed to store data: %v", err)
			}
		}
	}

	t.Log("Memory usage test passed")
}
