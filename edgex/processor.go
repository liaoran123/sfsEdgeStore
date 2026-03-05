package edgex

import (
	"encoding/json"
	"log"

	"sfsdb-edgex-adapter/common"
)

// ProcessMessage 处理EdgeX消息
func ProcessMessage(payload []byte) (*EdgeXEvent, error) {
	var edgexMsg EdgeXMessage
	if err := json.Unmarshal(payload, &edgexMsg); err != nil {
		return nil, err
	}

	// 检查 MessageType 是否为 "event"
	if edgexMsg.MessageType != "event" && edgexMsg.MessageType != "Event" {
		log.Printf("Ignoring message with type: %s", edgexMsg.MessageType)
		return nil, nil
	}

	// 解析 payload 中的事件
	var event EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
		return nil, err
	}

	// 从源头格式化设备名称，确保长度为64字符
	event.DeviceName = common.FormatDeviceName(event.DeviceName)

	return &event, nil
}
