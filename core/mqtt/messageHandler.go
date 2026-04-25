package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"sfsEdgeStore/common"
	"sfsEdgeStore/core/database"
	"sfsEdgeStore/core/edgex"
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

		//// 将消息发送到 Worker 队列，由固定数量的 Worker 处理
		select {
		case c.messageQueue <- msg:
			// 消息成功入队
		default:
			// 队列已满，丢弃消息或记录警告
			if c.monitor != nil {
				c.monitor.RecordError("queue_full", "Message queue is full, dropping message")
			}
		}
	}
}

// messageWorker 工作协程，从消息队列中接收并处理消息
func (c *Client) messageWorker(workerID int) {
	for msg := range c.messageQueue {
		c.processMessageAsync(msg)
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
			return false
		}
	}
	return true
}

// handleStorageError 处理数据存储错误
func (c *Client) handleStorageError(err error, records []*map[string]any) {
	// 记录错误到监控系统
	if c.monitor != nil {
		c.monitor.RecordError("storage_error", err.Error())
	}

	// 入队重试
	if c.dataQueue != nil {
		c.dataQueue.Enqueue(records)
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
	if c.analyzer == nil || !c.analyzer.IsEnabled() {
		return
	}

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

		// 处理告警
		if len(alerts) > 0 {
			for _, alert := range alerts {
				if c.monitor != nil {
					c.monitor.RecordError(alert.AlertType, alert.Message)
				}
			}

			// 推送告警数据到 WebSocket（Web 端已展示，无需日志）
			c.broadcastData("alerts", map[string]interface{}{
				"deviceName": deviceName,
				"alerts":     alerts,
			})
		}

		// 及时释放分析结果
		_ = results
	}
}
