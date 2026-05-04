package mqtt

import (
	"sfsEdgeStore/common"
	"sfsEdgeStore/edgex"
	"sfsEdgeStore/filter"
)

// MessageProcessor 消息处理层 - 只负责消息解析、过滤、设备状态更新
type MessageProcessor struct {
	filterManager *filter.FilterManager
}

func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		filterManager: filter.NewFilterManager(),
	}
}

// ProcessEvent 处理单个 EdgeX 事件，返回解析后的记录
func (p *MessageProcessor) ProcessEvent(event *edgex.EdgeXEvent) []*map[string]any {
	if event == nil {
		return nil
	}

	records := p.processReadings(event)
	if len(records) > 0 {
		return records
	}
	return nil
}

func (p *MessageProcessor) processReadings(event *edgex.EdgeXEvent) []*map[string]any {
	records := make([]*map[string]any, 0, len(event.Readings))

	for _, reading := range event.Readings {
		value := common.ParseValue(reading.Value)

		if value == "" {
			continue
		}

		if !p.shouldStoreData(event.DeviceName, reading.ResourceName, value) {
			continue
		}

		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		data := map[string]any{
			"id":          reading.ID,           // 读数唯一标识
			"deviceName":  event.DeviceName,     // 设备名称（已格式化固定长度）
			"reading":     reading.ResourceName, // 资源/读数名称
			"value":       value,                // 读数实际值（自动解析为 bool/float64/[]byte/string）
			"valueType":   reading.ValueType,    // 值的数据类型
			"profileName": event.ProfileName,    // 设备配置文件名称
			"baseType":    reading.BaseType,     // 值的基础类型
			"timestamp":   reading.Origin,       // 纳秒级时间戳
			"metadata":    metadataStr,          // 额外元数据（JSON字符串）
		}

		records = append(records, &data)
	}

	return records
}

// shouldStoreData 判断是否需要存储数据
func (p *MessageProcessor) shouldStoreData(deviceName, resourceName string, value any) bool {
	if p.filterManager != nil {
		if !p.filterManager.ShouldStore(deviceName, resourceName, value) {
			return false
		}
	}
	return true
}
