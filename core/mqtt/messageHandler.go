package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"sfsEdgeStore/common"
	"sfsEdgeStore/core/database"
	"sfsEdgeStore/core/edgex"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var flushCount atomic.Int64

func (c *Client) messageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		c.recordMessageReceived(len(msg.Payload()))

		event, err := edgex.ProcessMessage(msg.Payload())
		if err != nil {
			log.Printf("Failed to process message: %v", err)
			return
		}

		if event == nil {
			if c.monitor != nil {
				c.monitor.IncrementMQTTMessagesFiltered()
			}
			return
		}

		if !c.isDeviceAllowed(event.DeviceName) {
			log.Printf("Device limit reached, rejecting data from device: %s", event.DeviceName)
			if c.monitor != nil {
				c.monitor.RecordError("device_limit_reached", fmt.Sprintf("Device %s rejected due to limit", event.DeviceName))
				c.monitor.IncrementMQTTMessagesFiltered()
			}
			return
		}

		records := c.processReadings(event)
		if len(records) > 0 {
			c.enqueueRecords(records)
		}
	}
}

func (c *Client) recordMessageReceived(payloadSize int) {
	if c.monitor != nil {
		c.monitor.IncrementMQTTMessagesReceived()
		c.monitor.IncrementDataReceivedBytes(int64(payloadSize))
	}
}

func (c *Client) processReadings(event *edgex.EdgeXEvent) []*map[string]any {
	records := make([]*map[string]any, 0, len(event.Readings))

	for _, reading := range event.Readings {
		value := common.ParseValue(reading.Value)

		if isInvalidValue(value) {
			if c.monitor != nil {
				c.monitor.RecordError("invalid_value_discarded", fmt.Sprintf("Device: %s, Reading: %s, Value: %v", event.DeviceName, reading.ResourceName, reading.Value))
				c.monitor.IncrementMQTTMessagesFiltered()
			}
			continue
		}

		c.updateDeviceStatus(event.DeviceName, reading.ResourceName, value)

		if !c.shouldStoreData(event.DeviceName, reading.ResourceName, value) {
			continue
		}

		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		data := map[string]any{
			"id":         reading.ID,
			"deviceName": event.DeviceName,
			"reading":    reading.ResourceName,
			"value":      value,
			"valueType":  reading.ValueType,
			"baseType":   reading.BaseType,
			"timestamp":  reading.Origin,
			"metadata":   metadataStr,
		}

		records = append(records, &data)
	}

	return records
}

func (c *Client) enqueueRecords(newRecords []*map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pendingRecords = append(c.pendingRecords, newRecords...)

	// 检查缓冲区大小：如果积累的记录数超过 batchSize 或者
	// 等待时间超过 batchTime，则立即触发写入
	// 这样可以避免 LevelDB 的 BlockCache 积累过多数据
	shouldFlush := len(c.pendingRecords) >= batchSize ||
		time.Since(c.lastBatchTime) >= batchTime*time.Millisecond

	if shouldFlush {
		records := c.pendingRecords
		c.pendingRecords = make([]*map[string]any, 0, batchSize)
		c.lastBatchTime = time.Now()
		if c.batchTimer != nil {
			c.batchTimer.Stop()
			c.batchTimer = nil
		}
		c.writePool.Submit(func() {
			c.flushRecords(records)
		})
		return
	}

	if c.batchTimer == nil {
		c.batchTimer = time.AfterFunc(batchTime*time.Millisecond, func() {
			c.mu.Lock()
			if len(c.pendingRecords) > 0 {
				records := c.pendingRecords
				c.pendingRecords = make([]*map[string]any, 0, batchSize)
				c.lastBatchTime = time.Now()
				c.batchTimer = nil
				c.mu.Unlock()
				c.writePool.Submit(func() {
					c.flushRecords(records)
				})
			} else {
				c.batchTimer = nil
				c.mu.Unlock()
			}
		})
	}
}

func (c *Client) flushRecords(records []*map[string]any) {
	if len(records) == 0 {
		return
	}

	if c.monitor != nil {
		c.monitor.IncrementDatabaseOperations()
		c.monitor.IncrementTotalRecordsStored(int64(len(records)))
		c.monitor.IncrementDataStoredBytes(int64(len(records) * 100))
	}

	err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
	if err != nil {
		c.handleStorageError(err, records)
		return
	}

	if c.monitor != nil {
		c.monitor.IncrementMQTTMessagesProcessed()
	}

	// 每 50 次写入后手动触发 GC 并释放内存给 OS
	count := flushCount.Add(1)
	if count%50 == 0 {
		debug.FreeOSMemory()
	}

	if c.broadcaster != nil && len(records) > 0 {
		c.broadcastData("device_data", map[string]any{
			"deviceName": (*records[0])["deviceName"],
			"records":    records,
		})
	}

	if c.analyzer != nil && c.analyzer.IsEnabled() && len(records) <= 50 {
		c.analyzeData(records, "")
	}
}

func (c *Client) updateDeviceStatus(deviceName, resourceName string, value interface{}) {
	if c.monitor != nil {
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

func (c *Client) shouldStoreData(deviceName, resourceName string, value interface{}) bool {
	if c.filterManager != nil {
		if !c.filterManager.ShouldStore(deviceName, resourceName, value) {
			return false
		}
	}
	return true
}

func isInvalidValue(value any) bool {
	switch v := value.(type) {
	case string:
		return v == "" || v == "invalid_value"
	}
	return false
}

func (c *Client) handleStorageError(err error, records []*map[string]any) {
	if c.monitor != nil {
		c.monitor.RecordError("storage_error", err.Error())
	}

	if isInvalidData(records) {
		log.Printf("Discarding invalid data (retry won't help): %v", err)
		return
	}

	if c.dataQueue != nil {
		c.dataQueue.Enqueue(records)
	}
}

func isInvalidData(records []*map[string]any) bool {
	for _, record := range records {
		val, ok := (*record)["value"]
		if !ok {
			return true
		}
		switch v := val.(type) {
		case string:
			if v == "" || v == "invalid_value" {
				return true
			}
		}
	}
	return false
}

func (c *Client) broadcastData(dataType string, data any) {
	if c.broadcaster == nil {
		return
	}

	broadcastData := map[string]any{
		"type":      dataType,
		"data":      data,
		"timestamp": time.Now().UnixNano(),
	}

	jsonData, err := json.Marshal(broadcastData)
	if err != nil {
		log.Printf("Failed to marshal broadcast data: %v", err)
		return
	}

	c.broadcaster.Broadcast(jsonData)
}

func (c *Client) analyzeData(records []*map[string]any, deviceName string) {
	if c.analyzer == nil || !c.analyzer.IsEnabled() {
		return
	}

	readingDataMap := make(map[string][]map[string]any, len(records))
	for _, record := range records {
		readingName, ok := (*record)["reading"].(string)
		if !ok {
			continue
		}
		readingDataMap[readingName] = append(readingDataMap[readingName], *record)
	}

	for readingName, analysisData := range readingDataMap {
		results, alerts := c.analyzer.Analyze(analysisData, deviceName, readingName)

		if len(alerts) > 0 {
			for _, alert := range alerts {
				if c.monitor != nil {
					c.monitor.RecordError(alert.AlertType, alert.Message)
				}
			}

			c.broadcastData("alerts", map[string]any{
				"deviceName": deviceName,
				"alerts":     alerts,
			})
		}

		_ = results
	}
}
