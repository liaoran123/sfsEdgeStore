# EdgeX 消息处理

## 概述

sfsEdgeStore 通过 MQTT 协议接收 EdgeX Foundry 发送的数据，解析 EdgeX 的消息格式并转换为适合存储的格式。

## 数据结构

### EdgeXMessage 消息结构

```go
// edgex/models.go:6-11
type EdgeXMessage struct {
	CorrelationID string          `json:"correlationId,omitempty"`
	MessageType   string          `json:"messageType,omitempty"`
	Origin        int64           `json:"origin,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}
```

**字段说明：**
- `CorrelationID`：关联ID，用于追踪消息
- `MessageType`：消息类型，需为 "event" 或 "Event"
- `Origin`：消息的时间戳（纳秒）
- `Payload`：实际的事件数据（JSON格式）

### EdgeXEvent 事件结构

```go
// edgex/models.go:14-21
type EdgeXEvent struct {
	ID          string         `json:"id"`
	DeviceName  string         `json:"deviceName"`
	Readings    []EdgeXReading `json:"readings"`
	Origin      int64          `json:"origin"`
	ProfileName string         `json:"profileName,omitempty"`
	SourceName  string         `json:"sourceName,omitempty"`
}
```

**字段说明：**
- `ID`：事件的唯一标识
- `DeviceName`：设备名称（会被格式化为64字符）
- `Readings`：读数列表
- `Origin`：事件时间戳（纳秒）

### EdgeXReading 读数结构

```go
// edgex/models.go:24-33
type EdgeXReading struct {
	ID           string          `json:"id"`
	ResourceName string          `json:"resourceName"`
	Value        string          `json:"value"`
	ValueType    string          `json:"valueType,omitempty"`
	Origin       int64           `json:"origin"`
	ProfileName  string          `json:"profileName,omitempty"`
	DeviceName   string          `json:"deviceName,omitempty"`
	BaseType     string          `json:"baseType,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}
```

**字段说明：**
- `ResourceName`：资源名称（对应读数的名称）
- `Value`：读数的值（字符串格式）
- `ValueType`：值类型（如 Float32、Int64、String 等）
- `BaseType`：基础类型
- `Metadata`：元数据（JSON格式）

## 消息处理

### ProcessMessage 处理函数

```go
// edgex/processor.go:11-33
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
```

**处理流程：**
1. **解析外层消息**：将原始 MQTT 消息解析为 `EdgeXMessage` 结构
2. **验证消息类型**：只处理 MessageType 为 "event" 或 "Event" 的消息
3. **解析事件数据**：从 Payload 字段中解析出 `EdgeXEvent`
4. **格式化设备名**：确保设备名长度为64字符，便于索引查询

## 设备名称格式化

### 格式化逻辑

```go
// common/utils.go
func FormatDeviceName(deviceName string) string {
	const maxLength = 64
	if len(deviceName) >= maxLength {
		return deviceName[:maxLength]
	}
	
	// 用空格填充到64字符
	return deviceName + strings.Repeat(" ", maxLength-len(deviceName))
}
```

**设计原因：**
- 使用固定长度的设备名可以确保主键索引的一致性
- 便于按设备名范围查询
- 优化数据库查询性能

## 数据转换

### EdgeX 读数转存储记录

```go
func ConvertReadingsToRecords(event *EdgeXEvent) []*map[string]any {
	var records []*map[string]any

	for _, reading := range event.Readings {
		value, err := parseValue(reading.Value, reading.ValueType)
		if err != nil {
			log.Printf("Failed to parse value %s: %v", reading.Value, err)
			continue
		}

		record := map[string]any{
			"id":         reading.ID,
			"deviceName": event.DeviceName,
			"reading":    reading.ResourceName,
			"value":      value,
			"valueType":  reading.ValueType,
			"baseType":   reading.BaseType,
			"timestamp":  reading.Origin,
			"metadata":   string(reading.Metadata),
		}
		records = append(records, &record)
	}

	return records
}
```

## 示例

### 完整的 EdgeX 消息示例

```json
{
  "correlationId": "abc-123",
  "messageType": "event",
  "origin": 1704067200000000000,
  "payload": {
    "id": "event-001",
    "deviceName": "Device001",
    "readings": [
      {
        "id": "reading-001",
        "resourceName": "temperature",
        "value": "25.5",
        "valueType": "Float32",
        "origin": 1704067200000000000,
        "baseType": "Float"
      },
      {
        "id": "reading-002",
        "resourceName": "humidity",
        "value": "60",
        "valueType": "Int32",
        "origin": 1704067200000000000,
        "baseType": "Int"
      }
    ]
  }
}
```

### 转换后的存储记录

```go
[]*map[string]any{
  {
    "id":         "reading-001",
    "deviceName": "Device001                                                       ",
    "reading":    "temperature",
    "value":      25.5,
    "valueType":  "Float32",
    "baseType":   "Float",
    "timestamp":  1704067200000000000,
    "metadata":   "",
  },
  {
    "id":         "reading-002",
    "deviceName": "Device001                                                       ",
    "reading":    "humidity",
    "value":      60,
    "valueType":  "Int32",
    "baseType":   "Int",
    "timestamp":  1704067200000000000,
    "metadata":   "",
  },
}
```

## API 接口

### ProcessMessage 处理 EdgeX 消息

```go
func ProcessMessage(payload []byte) (*EdgeXEvent, error)
```

**参数：**
- `payload`：原始 MQTT 消息的字节数组

**返回值：**
- `*EdgeXEvent`：解析后的事件对象
- `error`：错误信息

**处理流程：**
1. 解析 JSON 消息
2. 验证消息类型
3. 解析事件数据
4. 格式化设备名称
5. 返回事件对象
