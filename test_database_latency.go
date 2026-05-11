package main

import (
	"fmt"
	"log"
	"sfsEdgeStore/config"
	"sfsEdgeStore/database"
	"time"
)

func main1() {
	fmt.Println("=== 数据库实时性检测 ===")
	fmt.Println()

	// 加载配置
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	fmt.Printf("正在初始化数据库 (路径: %s)...\n", appConfig.DBPath)
	if err := database.Init(appConfig.DBPath, appConfig.DBUseEncryption, appConfig.DBEncryptionKey, ""); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	fmt.Println("✓ 数据库初始化成功")
	fmt.Println()

	// 测试1: 单条记录写入延迟
	fmt.Println("=== 测试1: 单条记录写入延迟 ===")
	testSingleWriteLatency()
	fmt.Println()

	// 测试2: 批量记录写入延迟
	fmt.Println("=== 测试2: 批量记录写入延迟 ===")
	testBatchWriteLatency()
	fmt.Println()

	// 测试3: 查询最新数据延迟
	fmt.Println("=== 测试3: 查询最新数据延迟 ===")
	testQueryLatestData()
	fmt.Println()

	// 测试4: 写入后立即查询
	fmt.Println("=== 测试4: 写入后立即查询 ===")
	testWriteAndQuery()
	fmt.Println()

	fmt.Println("=== 检测完成 ===")
}

// 测试单条记录写入延迟
func testSingleWriteLatency() {
	testDevice := fmt.Sprintf("test-device-latency-%d", time.Now().UnixNano())

	startTime := time.Now()
	for i := 0; i < 10; i++ {
		record := map[string]any{
			"id":         fmt.Sprintf("test-id-%d", i),
			"deviceName": testDevice,
			"reading":    "temperature",
			"value":      25.0 + float64(i),
			"valueType":  "Float32",
			"baseType":   "Float",
			"timestamp":  time.Now().UnixNano(),
			"metadata":   "{\"test\": \"true\"}",
		}

		writeStart := time.Now()
		_, err := database.Insert(database.Table, []*map[string]any{&record})
		writeTime := time.Since(writeStart)

		if err != nil {
			log.Printf("写入失败 (i=%d): %v", i, err)
			continue
		}

		log.Printf("第 %2d 条记录写入耗时: %v", i+1, writeTime)
	}
	totalTime := time.Since(startTime)

	fmt.Printf("\n单条记录写入测试: 共写入 10 条记录, 总耗时: %v\n", totalTime)
}

// 测试批量记录写入延迟
func testBatchWriteLatency() {
	testDevice := fmt.Sprintf("test-device-batch-%d", time.Now().UnixNano())

	for batchSize := 10; batchSize <= 100; batchSize += 10 {
		records := make([]*map[string]any, 0, batchSize)
		for i := 0; i < batchSize; i++ {
			record := map[string]any{
				"id":         fmt.Sprintf("test-id-batch-%d-%d", batchSize, i),
				"deviceName": testDevice,
				"reading":    "humidity",
				"value":      60.0 + float64(i),
				"valueType":  "Float32",
				"baseType":   "Float",
				"timestamp":  time.Now().UnixNano() + int64(i),
				"metadata":   "{\"test\": \"true\"}",
			}
			records = append(records, &record)
		}

		writeStart := time.Now()
		_, err := database.Insert(database.Table, records)
		writeTime := time.Since(writeStart)

		if err != nil {
			log.Printf("批量写入失败 (size=%d): %v", batchSize, err)
			continue
		}

		log.Printf("批量写入 %3d 条记录耗时: %v (平均每条: %v)",
			batchSize, writeTime, writeTime/time.Duration(batchSize))
	}
}

// 测试查询最新数据
func testQueryLatestData() {
	testDevice := fmt.Sprintf("test-device-query-%d", time.Now().UnixNano())

	// 先写入一些测试数据
	now := time.Now()
	for i := 0; i < 5; i++ {
		record := map[string]any{
			"id":         fmt.Sprintf("test-id-query-%d", i),
			"deviceName": testDevice,
			"reading":    "pressure",
			"value":      1013.0 + float64(i),
			"valueType":  "Float32",
			"baseType":   "Float",
			"timestamp":  now.Add(time.Duration(i) * time.Second).UnixNano(),
			"metadata":   "{\"test\": \"true\"}",
		}
		_, err := database.Insert(database.Table, []*map[string]any{&record})
		if err != nil {
			log.Printf("写入失败: %v", err)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	// 查询最新数据
	queryStart := time.Now()
	records, err := database.QueryRecords(database.Table, testDevice, "", "", false)
	queryTime := time.Since(queryStart)

	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}
	defer records.Release()

	fmt.Printf("查询 %q 的数据耗时: %v\n", testDevice, queryTime)
	fmt.Printf("查询到 %d 条记录\n", len(records))

	if len(records) > 0 {
		latestRecord := records[len(records)-1]
		timestamp, ok := latestRecord["timestamp"].(int64)
		if ok {
			dataAge := time.Since(time.Unix(0, timestamp))
			fmt.Printf("最新记录年龄: %v\n", dataAge)
		}
	}
}

// 测试写入后立即查询
func testWriteAndQuery() {
	testDevice := fmt.Sprintf("test-device-write-query-%d", time.Now().UnixNano())

	for i := 0; i < 5; i++ {
		// 写入一条记录
		writeTimestamp := time.Now().UnixNano()
		record := map[string]any{
			"id":         fmt.Sprintf("test-id-write-query-%d", i),
			"deviceName": testDevice,
			"reading":    "voltage",
			"value":      12.0 + float64(i)*0.1,
			"valueType":  "Float32",
			"baseType":   "Float",
			"timestamp":  writeTimestamp,
			"metadata":   "{\"test\": \"true\"}",
		}

		_, err := database.Insert(database.Table, []*map[string]any{&record})
		if err != nil {
			log.Printf("写入失败: %v", err)
			continue
		}

		// 立即查询
		queryStart := time.Now()
		records, err := database.QueryRecords(database.Table, testDevice, "", "", false)
		queryTime := time.Since(queryStart)

		if err != nil {
			log.Printf("查询失败: %v", err)
			continue
		}

		found := false
		for _, r := range records {
			ts, ok := r["timestamp"].(int64)
			if ok && ts == writeTimestamp {
				found = true
				break
			}
		}
		records.Release()

		if found {
			log.Printf("第 %d 次: 写入后立即查询成功, 查询耗时: %v, ✓ 找到最新记录", i+1, queryTime)
		} else {
			log.Printf("第 %d 次: 写入后立即查询成功, 查询耗时: %v, ✗ 未找到最新记录", i+1, queryTime)
		}
	}
}
