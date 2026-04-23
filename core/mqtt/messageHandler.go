package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"sfsEdgeStore/common"
	"sfsEdgeStore/core/database"
	"sfsEdgeStore/core/edgex"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// messageHandler 适配器处理收到的EdgeX消息
/*
- messageHandler 函数是通过 MQTT 客户端订阅主题时设置的回调函数触发的，用于处理收到的EdgeX消息
- subscribeTopics 函数中，MQTT 客户端会订阅指定的主题
- 订阅时，会调用 mqttClient.Subscribe 方法，并将 c.messageHandler() 作为回调函数传入
*/
func (c *Client) messageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		// 增加MQTT消息接收计数
		c.recordMessageReceived(len(msg.Payload()))

		log.Printf("Received message on topic: %s", msg.Topic())

		// 使用goroutine异步处理消息，避免阻塞MQTT消息接收
		go c.processMessageAsync(msg)
	}
}

// recordMessageReceived 记录消息接收情况
func (c *Client) recordMessageReceived(payloadSize int) {
	if c.monitor != nil {
		c.monitor.IncrementMQTTMessagesReceived()
		c.monitor.IncrementDataReceivedBytes(int64(payloadSize))
	}
}

// processMessageAsync 异步处理MQTT消息
func (c *Client) processMessageAsync(msg mqtt.Message) {
	// 使用edgex包处理消息
	event, err := edgex.ProcessMessage(msg.Payload())
	if err != nil {
		log.Printf("Failed to process message: %v", err)
		return
	}

	// 如果消息类型不是event，event会为nil
	if event == nil {
		return
	}

	// 检查设备数量限制（仅对免费用户生效）
	if !c.isDeviceAllowed(event.DeviceName) {
		log.Printf("Device limit reached, rejecting data from device: %s", event.DeviceName)
		if c.monitor != nil {
			c.monitor.RecordError("device_limit_reached", fmt.Sprintf("Device %s rejected due to limit", event.DeviceName))
		}
		return
	}

	// 处理读数
	records := c.processReadings(event)

	// 批量存储到 sfsDb
	if len(records) > 0 {
		c.storeData(records, event.DeviceName)
	}
}

// processReadings 处理设备读数
func (c *Client) processReadings(event *edgex.EdgeXEvent) []*map[string]any {
	// 预分配切片容量，避免动态扩容
	records := make([]*map[string]any, 0, len(event.Readings))

	// 处理每个读数
	for _, reading := range event.Readings {
		// 解析值的类型
		value := common.ParseValue(reading.Value)

		// 更新设备状态
		c.updateDeviceStatus(event.DeviceName, reading.ResourceName, value)

		// 数据过滤
		if !c.shouldStoreData(event.DeviceName, reading.ResourceName, value) {
			continue
		}

		// 从对象池获取map，减少内存分配
		data := objPool.GetMap()

		// 准备数据
		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		data["id"] = reading.ID
		data["deviceName"] = event.DeviceName // 设备名称已经在ProcessMessage中格式化
		data["reading"] = reading.ResourceName
		data["value"] = value
		data["valueType"] = reading.ValueType
		data["baseType"] = reading.BaseType
		data["timestamp"] = reading.Origin // 纳秒级时间戳，类型为 int64
		data["metadata"] = metadataStr

		records = append(records, &data)
	}

	return records
}

// updateDeviceStatus 更新设备状态
func (c *Client) updateDeviceStatus(deviceName, resourceName string, value interface{}) {
	if c.monitor != nil {
		// 尝试将值转换为float64用于监控
		floatValue := 0.0
		switch v := value.(type) {
		case float64:
			floatValue = v
		case int:
			floatValue = float64(v)
		case int64:
			floatValue = float64(v)
		}
		c.monitor.UpdateDeviceStatus(deviceName, resourceName, floatValue)
	}
}

// shouldStoreData 检查是否应该存储数据
func (c *Client) shouldStoreData(deviceName, resourceName string, value interface{}) bool {
	if c.filterManager != nil {
		if !c.filterManager.ShouldStore(deviceName, resourceName, value) {
			log.Printf("Filtered out data from %s: %s", deviceName, resourceName)
			return false
		}
	}
	return true
}

// handleStorageError 处理存储错误
func (c *Client) handleStorageError(err error, records []*map[string]any) {
	log.Printf("Failed to batch store data after retries: %v", err)

	errorMsg := err.Error()
	var errorType string

	// 边缘设备常见故障类型判断
	switch {
	case strings.Contains(errorMsg, "no space left") ||
		strings.Contains(errorMsg, "disk full") ||
		strings.Contains(errorMsg, "file system") ||
		strings.Contains(errorMsg, "I/O error"):
		// 磁盘空间不足或文件系统错误
		errorType = "storage_error"
		log.Printf("Fatal storage error detected: %v", err)
	case strings.Contains(errorMsg, "lock") ||
		strings.Contains(errorMsg, "busy"):
		// 锁竞争或资源忙
		errorType = "resource_contention"
		log.Printf("Resource contention error detected: %v", err)
	default:
		// 其他错误
		errorType = "database_error"
		log.Printf("Other database error: %v", err)
	}

	// 触发监控告警
	if c.monitor != nil {
		c.monitor.RecordError(errorType, errorMsg)
	}

	// 将数据加入队列，以便后续处理
	if err := c.dataQueue.Enqueue(records); err != nil {
		log.Printf("Failed to enqueue data: %v", err)
	} else {
		log.Printf("Enqueued %d readings for later processing", len(records))
	}

	// 归还map对象到池中
	for _, data := range records {
		objPool.PutMap(*data)
	}
}

// broadcastData 广播数据到 WebSocket
func (c *Client) broadcastData(dataType string, data interface{}) {
	if c.broadcaster == nil {
		return
	}

	// 准备广播数据
	broadcastData := map[string]interface{}{
		"type":      dataType,
		"data":      data,
		"timestamp": time.Now().UnixNano(),
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(broadcastData)
	if err != nil {
		log.Printf("Failed to marshal broadcast data: %v", err)
		return
	}

	// 广播数据
	c.broadcaster.Broadcast(jsonData)
}

// storeData 存储数据到数据库
func (c *Client) storeData(records []*map[string]any, deviceName string) {
	// 增加数据库操作计数
	if c.monitor != nil {
		c.monitor.IncrementDatabaseOperations()
	}

	// 使用重试机制插入数据
	err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
	if err != nil {
		c.handleStorageError(err, records)
		return
	}

	log.Printf("Batch stored %d readings from %s", len(records), deviceName)

	// 增加MQTT消息处理计数
	if c.monitor != nil {
		c.monitor.IncrementMQTTMessagesProcessed()
		// 增加数据存储字节数
		c.monitor.IncrementDataStoredBytes(int64(len(records) * 100)) // 估算每条记录约100字节
	}

	// 推送设备数据
	if len(records) > 0 {
		c.broadcastData("device_data", map[string]interface{}{
			"deviceName": deviceName,
			"records":    records,
		})
	}

	// 分析数据
	c.analyzeData(records, deviceName)

	// 归还map对象到池中
	for _, data := range records {
		objPool.PutMap(*data)
	}
}

// analyzeData 分析数据
func (c *Client) analyzeData(records []*map[string]any, deviceName string) {
	if c.analyzer != nil && c.analyzer.IsEnabled() {
		// 按reading分组分析数据
		readingDataMap := make(map[string][]map[string]interface{}, len(records))
		for _, record := range records {
			// 从记录中获取reading信息
			readingName, ok := (*record)["reading"].(string)
			if !ok {
				continue
			}
			readingDataMap[readingName] = append(readingDataMap[readingName], *record)
		}

		// 对每个reading进行分析
		for readingName, analysisData := range readingDataMap {
			// 分析数据
			results, alerts := c.analyzer.Analyze(analysisData, deviceName, readingName)

			// 处理分析结果
			if len(results) > 0 {
				log.Printf("Analysis completed for %s: %d results", readingName, len(results))
				// 这里可以将分析结果存储或发送到其他系统
			}

			// 处理告警
			if len(alerts) > 0 {
				log.Printf("Detected %d alerts for %s", len(alerts), readingName)
				// 这里可以将告警发送到监控系统或其他通知渠道
				for _, alert := range alerts {
					log.Printf("Alert: %s - %s - %s", alert.Severity, alert.Message, alert.Reading)
					// 触发监控告警
					if c.monitor != nil {
						c.monitor.RecordError(alert.AlertType, alert.Message)
					}
				}

				// 推送告警数据
				c.broadcastData("alerts", map[string]interface{}{
					"deviceName": deviceName,
					"alerts":     alerts,
				})
			}
		}
	}
}
