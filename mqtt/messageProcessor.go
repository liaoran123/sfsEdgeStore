package mqtt

import (
	"sfsEdgeStore/common"
	"sfsEdgeStore/edgex"
	"sfsEdgeStore/filter"
	"sfsEdgeStore/monitor"
)

// MessageProcessor 消息处理层 - 只负责消息解析、过滤、设备状态更新
type MessageProcessor struct {
	monitor       *monitor.Monitor
	filterManager *filter.FilterManager
}

func NewMessageProcessor(monitor *monitor.Monitor) *MessageProcessor {
	return &MessageProcessor{
		monitor:       monitor,
		filterManager: filter.NewFilterManager(),
	}
}

// ProcessEvent 处理单个 EdgeX 事件，返回解析后的记录
func (p *MessageProcessor) ProcessEvent(event *edgex.EdgeXEvent) []*map[string]any {
	//
	p.recordMessageReceived(0)

	if event == nil {
		if p.monitor != nil {
			p.monitor.IncrementMQTTMessagesFiltered()
		}
		return nil
	}

	records := p.processReadings(event)
	if len(records) > 0 {
		return records
	}
	return nil
}

// recordMessageReceived 记录收到的消息，包括消息数量和数据量
func (p *MessageProcessor) recordMessageReceived(payloadSize int) {
	if p.monitor != nil {
		p.monitor.IncrementMQTTMessagesReceived()
		if payloadSize > 0 {
			p.monitor.IncrementDataReceivedBytes(int64(payloadSize))
		}
	}
}

func (p *MessageProcessor) processReadings(event *edgex.EdgeXEvent) []*map[string]any {
	records := make([]*map[string]any, 0, len(event.Readings))

	for _, reading := range event.Readings {
		value := common.ParseValue(reading.Value)

		if value == "" {
			continue
		}

		p.updateDeviceStatus(event.DeviceName, reading.ResourceName, value)

		if !p.shouldStoreData(event.DeviceName, reading.ResourceName, value) {
			continue
		}

		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		data := map[string]any{
			"id":          reading.ID,           //string 读数的唯一标识符，由 EdgeX 自动生成
			"deviceName":  event.DeviceName,     //string 设备名称，标识数据来源设备（已格式化固定长度）
			"reading":     reading.ResourceName, //string 资源/读数名称，如 temperature 、 humidity
			"value":       value,                //arseValue(reading.Value) any 读数实际值，自动解析为 bool/float64/[]byte/string
			"valueType":   reading.ValueType,    //string 值的数据类型，如 Float64 、 Int32 、 Bool
			"profileName": event.ProfileName,    //string 设备配置文件名称，标识设备类型/模型
			"baseType":    reading.BaseType,     //string 值的基础类型，如 simple 、 array
			"timestamp":   reading.Origin,       //reading.Origin int64 纳秒级时间戳，表示读数产生时间
			"metadata":    metadataStr,          //string(reading.Metadata) string 额外元数据（JSON 字符串），用于扩展信息
		}

		records = append(records, &data)
	}

	return records
}

// updateDeviceStatus 更新设备状态
// 支持 float64, int, int64 类型
func (p *MessageProcessor) updateDeviceStatus(deviceName, resourceName string, value any) {
	if p.monitor != nil {
		floatValue := 0.0
		switch v := value.(type) {
		case float64:
			floatValue = v
		case int:
			floatValue = float64(v)
		case int64:
			floatValue = float64(v)
		}
		p.monitor.UpdateDeviceStatus(deviceName, resourceName, floatValue)
	}
}

// shouldStoreData 判断是否需要存储数据
// 支持 any 类型
func (p *MessageProcessor) shouldStoreData(deviceName, resourceName string, value any) bool {
	if p.filterManager != nil {
		if !p.filterManager.ShouldStore(deviceName, resourceName, value) {
			return false
		}
	}
	return true
}
