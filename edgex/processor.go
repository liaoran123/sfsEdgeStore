package edgex

import (
	"encoding/json"
	"strings"

	"sfsEdgeStore/common"
)

// ProcessMessage 处理EdgeX消息
/*
 将payload 解析为 EdgeXMessage结构体，再根据消息类型判断是否为事件消息
 如果是事件消息，解析Payload字段到EdgeXEvent结构体
 如果不是事件消息，返回nil，或错误
 * */
func ProcessMessage(payload []byte) (*EdgeXEvent, error) {
	var edgexMsg EdgeXMessage
	// 解析EdgeX消息到EdgeXMessage结构体
	if err := json.Unmarshal(payload, &edgexMsg); err != nil {
		return nil, err
	}

	// 非event类型消息静默过滤，避免日志泛滥
	if strings.ToLower(edgexMsg.MessageType) != "event" {
		return nil, nil
	}

	var event EdgeXEvent
	// 解析EdgeXMessage结构体中的Payload字段到EdgeXEvent结构体
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
		return nil, err
	}
	/*
	 * 这个设计是sfsDb的特性决定的，由于设备名称是字符串，属于不定长数据类型。
	 * 表设置是设备名称+时间戳作为组合主键，组合主键必须是定长类型。
	 * 所以设备名称必须格式化为定长的字符串。
	 */
	event.DeviceName = common.FormatDeviceName(event.DeviceName)

	return &event, nil
}
