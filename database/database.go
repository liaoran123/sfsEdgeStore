package database

import (
	"fmt"
	"log"
	"os"
	"sfsdb-edgex-adapter/common"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// Table 全局表实例
var Table *engine.Table

// Init 初始化数据库
func Init(dbPath string) error {
	// 确保数据库目录存在
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// 打开数据库
	_, err := storage.GetDBManager().OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	/*
		   FormatDeviceName 格式化设备名称，确保长度为64字符
		   可以通过util的
		    RegisterTypeSize("string", func(value any) int {
			// 固定string类型大小为64
			return 64
		})
		即可添加其他二级索引，如果需要的话。（组合主键必须是固定长度的字段，字符串是不定长类型，需要通过RegisterTypeSize注册固定大小）
	*/
	// 创建或获取表
	tableName := "edgex_readings"
	var createErr error
	Table, createErr = engine.TableNew(tableName)
	if createErr != nil {
		return fmt.Errorf("failed to create table: %v", createErr)
	}

	// 设置表字段
	fields := map[string]any{
		"id":         "",
		"deviceName": "",
		"reading":    "",
		"value":      0.0,
		"valueType":  "",
		"baseType":   "",
		"timestamp":  int64(0),
		"metadata":   "",
	}
	if err := Table.SetFields(fields); err != nil {
		return fmt.Errorf("failed to set table fields: %v", err)
	}

	// 创建组合主键索引 (deviceName + timestamp)
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		return fmt.Errorf("failed to create primary key: %v", err)
	}
	primaryKey.AddFields("deviceName", "timestamp") // 创建deviceName和timestamp的组合主键
	if err := Table.CreateIndex(primaryKey); err != nil {
		// 忽略索引已存在的错误
		if err.Error() != "index already exists" {
			return fmt.Errorf("failed to create primary key index: %v", err)
		}
	}
	/*
		//由于组合主键中包括deviceName不定长类型，所以不能单独创建其他索引
		//如果业务确实需要使用单独索引，如按时间查询，则需要将组合主键更改为组合索引。时间戳为主键。

		//现在的实现需要按设备和时间戳进行查询的场景性能最高，资源最少。权衡之下，是最优解。

				// 创建设备名称索引
				deviceIndex, err := engine.DefaultNormalIndexNew("device_index")
				if err != nil {
					return fmt.Errorf("failed to create device index: %v", err)
				}
				deviceIndex.AddFields("deviceName")
				if err := table.CreateIndex(deviceIndex); err != nil {
					// 忽略索引已存在的错误
					if err.Error() != "index already exists" {
						return fmt.Errorf("failed to create device index: %v", err)
					}
				}

			// 创建时间戳索引
			timeIndex, err := engine.DefaultNormalIndexNew("time_index")
			if err != nil {
				return fmt.Errorf("failed to create time index: %v", err)
			}
			timeIndex.AddFields("timestamp")
			if err := table.CreateIndex(timeIndex); err != nil {
				// 忽略索引已存在的错误
				if err.Error() != "index already exists" {
					return fmt.Errorf("failed to create time index: %v", err)
				}
			}
	*/
	log.Println("Database initialized successfully")
	return nil
}

// BatchInsertWithRetry 批量插入数据并带有重试机制，以防数据库暂时不可用。
func BatchInsertWithRetry(tbl *engine.Table, records []*map[string]any, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		_, err := tbl.BatchInsertNoInc(records)
		if err == nil {
			return nil
		}

		log.Printf("Failed to batch insert data (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("failed to batch insert data after %d attempts", maxRetries)
}

// QueryRecords 查询记录数据
func QueryRecords(tbl *engine.Table, deviceName, startTime, endTime string) (record.Records, error) {
	// 格式化设备名称，确保长度为64字符
	formattedDeviceName := common.FormatDeviceName(deviceName)

	log.Println("Querying readings with filters:")
	log.Printf("  deviceName: %s", deviceName)
	log.Printf("  formattedDeviceName: %s", formattedDeviceName)
	log.Printf("  startTime: %s", startTime)
	log.Printf("  endTime: %s", endTime)

	// 构建时间范围查询
	var startTimestamp, endTimestamp *int64

	// 解析开始时间
	if startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			ts := start.UnixNano()
			startTimestamp = &ts
		}
	}

	// 解析结束时间
	if endTime != "" {
		end, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			ts := end.UnixNano()
			endTimestamp = &ts
		}
	}

	// 构建查询范围
	startRange := make(map[string]any)
	endRange := make(map[string]any)

	// 利用组合主键 (deviceName + timestamp) 进行更高效的查询
	// 设置设备名称
	startRange["deviceName"] = formattedDeviceName
	endRange["deviceName"] = formattedDeviceName

	// 设置时间范围
	if startTimestamp != nil {
		startRange["timestamp"] = *startTimestamp
	} else {
		startRange["timestamp"] = nil // 从最小值开始
	}

	if endTimestamp != nil {
		endRange["timestamp"] = *endTimestamp
	} else {
		endRange["timestamp"] = nil // 到最大值结束
	}

	// 执行范围查询
	iter, err := tbl.SearchRange(nil, &startRange, &endRange)
	if err != nil {
		return nil, fmt.Errorf("failed to search readings: %v", err)
	}
	defer iter.Release()

	// 获取记录
	records := iter.GetRecords(true)
	return records, nil
}
