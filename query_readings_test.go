package main

import (
	"fmt"
	"os"
	"sfsdb-edgex-adapter/database"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// TestQueryReadings 测试查询读数功能
func TestQueryReadings(t *testing.T) {
	// 清理旧数据
	dbPath := "./test_query_readings_db"
	os.RemoveAll(dbPath)

	// 初始化数据库
	dbMgr := storage.GetDBManager()
	_, err := dbMgr.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer dbMgr.CloseDB()
	defer os.RemoveAll(dbPath)

	// 创建表
	tableName := "edgex_readings"
	table, err := engine.TableNew(tableName)
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":         "",
		"deviceName": "",
		"reading":    "",
		"value":      0.0,
		"timestamp":  0,
		"metadata":   "",
	}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置字段失败: %v", err)
	}

	// 创建组合主键索引 (deviceName + timestamp)
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	primaryKey.AddFields("deviceName", "timestamp")
	err = table.CreateIndex(primaryKey)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 生成测试数据
	now := time.Now()
	deviceName := "test-device"

	// 插入5条测试数据
	for i := 0; i < 5; i++ {
		timestamp := now.Add(time.Duration(i) * time.Minute).Unix()
		data := map[string]any{
			"id":         fmt.Sprintf("id-%d", i),
			"deviceName": deviceName,
			"reading":    "temperature",
			"value":      25.0 + float64(i),
			"timestamp":  timestamp,
			"metadata":   fmt.Sprintf("{\"location\": \"room-%d\"}", i),
		}

		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}
	}

	// 测试1: 查询特定设备的所有数据
	t.Run("QueryAllDeviceData", func(t *testing.T) {
		readings, err := database.QueryRecords(table, deviceName, "", "")
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		defer readings.Release()

		if len(readings) != 5 {
			t.Errorf("期望5条记录，实际得到%d条", len(readings))
		}
	})

	// 测试2: 查询特定设备的时间范围数据
	t.Run("QueryDeviceTimeRange", func(t *testing.T) {
		startTime := now.Add(1 * time.Minute).Format(time.RFC3339)
		endTime := now.Add(3 * time.Minute).Format(time.RFC3339)

		readings, err := database.QueryRecords(table, deviceName, startTime, endTime)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		defer readings.Release()

		if len(readings) != 3 {
			t.Errorf("期望3条记录，实际得到%d条", len(readings))
		}
	})
}
