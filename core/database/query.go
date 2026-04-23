package database

import (
	"fmt"
	"log"
	"sfsEdgeStore/common"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/record"
)

// QueryRecords 查询记录数据
func QueryRecords(tbl *engine.Table, deviceName, startTime, endTime string, esc bool, limit ...int) (record.Records, error) {
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
	records := iter.GetRecords(esc, limit...)
	return records, nil
}
