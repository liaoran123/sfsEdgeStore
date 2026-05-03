# EdgeX 核心处理模块 (EdgeX Core) 技术文档

## 1. 概述

EdgeX核心处理模块是sfsEdgeStore系统中负责解析和处理EdgeX Foundry消息的核心组件。它负责接收来自EdgeX的MQTT消息，解析符合MessageEnvelope格式的JSON消息，提取设备事件和读数数据，并进行设备名称格式化等预处理工作。

### 1.1 主要功能

- **消息解析**：解析符合EdgeX MessageEnvelope格式的JSON消息
- **事件提取**：从消息Payload中提取设备事件数据
- **消息类型过滤**：只处理类型为"event"的消息，忽略其他类型
- **设备名称格式化**：将设备名称格式化为64字符长度
- **读数提取**：提取设备读数数据，包括资源名称、值、值类型等

### 1.2 消息流程

```
MQTT消息接收
    ↓
JSON反序列化为 EdgeXMessage
    ↓
检查 MessageType 是否为 "event"
    ↓
解析 Payload 为 EdgeXEvent
    ↓
格式化设备名称 (FormatDeviceName)
    ↓
返回 EdgeXEvent 给上层处理
```

## 2. 数据结构

### 2.1 EdgeXMessage 结构体

```go
type EdgeXMessage struct {
    CorrelationID string          `json:"correlationId,omitempty"`  // 关联ID（可选）
    MessageType   string          `json:"messageType,omitempty"`    // 消息类型
    Origin        int64           `json:"origin,omitempty"`        // 起源时间戳
    Payload       json.RawMessage `json:"payload"`                 // 消息载荷
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| CorrelationID | string | 否 | 用于追踪消息的关联ID |
| MessageType | string | 是 | 消息类型，"event"表示设备事件 |
| Origin | int64 | 否 | 消息产生的时间戳（纳秒） |
| Payload | json.RawMessage | 是 | 消息的实际载荷（JSON格式） |

### 2.2 EdgeXEvent 结构体

```go
type EdgeXEvent struct {
    ID          string         `json:"id"`                    // 事件ID
    DeviceName  string         `json:"deviceName"`            // 设备名称
    Readings    []EdgeXReading `json:"readings"`              // 读数列表
    Origin      int64          `json:"origin"`               // 起源时间戳
    ProfileName string         `json:"profileName,omitempty"` // 设备Profile名称
    SourceName  string         `json:"sourceName,omitempty"`  // 事件源名称
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ID | string | 是 | 事件的唯一标识符 |
| DeviceName | string | 是 | 产生事件的设备名称 |
| Readings | []EdgeXReading | 是 | 设备产生的读数列表 |
| Origin | int64 | 否 | 事件产生的时间戳（纳秒） |
| ProfileName | string | 否 | 设备关联的Profile名称 |
| SourceName | string | 否 | 事件源（如资源名称） |

### 2.3 EdgeXReading 结构体

```go
type EdgeXReading struct {
    ID           string          `json:"id"`                    // 读数ID
    ResourceName string          `json:"resourceName"`          // 资源名称
    Value        string          `json:"value"`                 // 读数值（字符串形式）
    ValueType    string          `json:"valueType,omitempty"`  // 值类型
    Origin       int64           `json:"origin"`               // 起源时间戳
    ProfileName  string          `json:"profileName,omitempty"` // Profile名称
    DeviceName   string          `json:"deviceName,omitempty"`  // 设备名称
    BaseType     string          `json:"baseType,omitempty"`    // 基本类型
    Metadata     json.RawMessage `json:"metadata,omitempty"`    // 元数据
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ID | string | 是 | 读数的唯一标识符 |
| ResourceName | string | 是 | 资源名称（如"GetValue"） |
| Value | string | 是 | 读数值（始终为字符串格式） |
| ValueType | string | 否 | 值的原始类型（如"Float64"） |
| Origin | int64 | 否 | 读数产生的时间戳 |
| ProfileName | string | 否 | 关联的Profile名称 |
| DeviceName | string | 否 | 关联的设备名称 |
| BaseType | string | 否 | 基本类型 |
| Metadata | json.RawMessage | 否 | 额外的元数据（JSON格式） |

## 3. 函数说明

### ProcessMessage

```go
func ProcessMessage(payload []byte) (*EdgeXEvent, error)
```

**功能**：处理EdgeX MQTT消息

**参数**：
- `payload`: MQTT消息的原始载荷（字节数组）

**处理逻辑**：
1. 将payload反序列化为EdgeXMessage结构
2. 检查MessageType是否为"event"或"Event"，不是则忽略并返回nil
3. 将Payload字段反序列化为EdgeXEvent结构
4. 调用common.FormatDeviceName格式化设备名称
5. 返回解析后的EdgeXEvent

**返回值**：
- `*EdgeXEvent`: 解析后的EdgeX事件对象（如果消息被忽略则返回nil）
- `error`: 解析过程中的错误（成功时为nil）

**示例**：

```go
// 假设这是从MQTT接收到的消息载荷
payload := []byte(`{
    "correlationId": "abc123",
    "messageType": "event",
    "origin": 1234567890,
    "payload": {
        "id": "event-001",
        "deviceName": "Temperature-Sensor-01",
        "readings": [{
            "id": "reading-001",
            "resourceName": "GetValue",
            "value": "25.5",
            "valueType": "Float64",
            "origin": 1234567890
        }],
        "origin": 1234567890
    }
}`)

event, err := edgex.ProcessMessage(payload)
if err != nil {
    log.Printf("Failed to process message: %v", err)
    return
}

if event != nil {
    fmt.Printf("Device: %s, Readings: %d\n", event.DeviceName, len(event.Readings))
}
```

## 4. 消息格式

### 4.1 EdgeX MessageEnvelope 格式

EdgeX消息遵循MessageEnvelope规范，外层为消息包装器：

```json
{
    "correlationId": "string",      // 可选，消息关联ID
    "messageType": "event",        // 必填，消息类型
    "origin": 1234567890,          // 可选，时间戳
    "payload": "{...}"             // 必填，JSON字符串
}
```

### 4.2 Event 事件格式

Payload中的事件数据格式：

```json
{
    "id": "event-001",             // 事件ID
    "deviceName": "sensor-01",     // 设备名称
    "readings": [...],             // 读数数组
    "origin": 1234567890,          // 时间戳
    "profileName": "temp-profile", // 可选
    "sourceName": "GetValue"       // 可选
}
```

### 4.3 Reading 读数格式

```json
{
    "id": "reading-001",
    "resourceName": "GetValue",
    "value": "25.5",
    "valueType": "Float64",
    "origin": 1234567890,
    "profileName": "temp-profile",
    "deviceName": "sensor-01",
    "baseType": "simple",
    "metadata": {}
}
```

## 5. 使用示例

### 5.1 基本使用

```go
package main

import (
    "encoding/json"
    "log"

    "sfsEdgeStore/core/edgex"
)

func main() {
    // 模拟MQTT消息载荷
    rawMessage := `{
        "correlationId": "msg-12345",
        "messageType": "event",
        "origin": 1640000000000000000,
        "payload": {
            "id": "evt-001",
            "deviceName": "Temperature-Sensor-01",
            "readings": [
                {
                    "id": "rd-001",
                    "resourceName": "GetValue",
                    "value": "25.5",
                    "valueType": "Float64",
                    "origin": 1640000000000000000
                }
            ],
            "origin": 1640000000000000000,
            "profileName": "TemperatureSensor",
            "sourceName": "GetValue"
        }
    }`

    // 处理消息
    event, err := edgex.ProcessMessage([]byte(rawMessage))
    if err != nil {
        log.Fatalf("Failed to process message: %v", err)
    }

    if event != nil {
        log.Printf("Processed event from device: %s", event.DeviceName)
        log.Printf("Event ID: %s", event.ID)
        log.Printf("Number of readings: %d", len(event.Readings))

        for i, reading := range event.Readings {
            log.Printf("  Reading %d: Resource=%s, Value=%s, Type=%s",
                i+1, reading.ResourceName, reading.Value, reading.ValueType)
        }
    }
}
```

### 5.2 在MQTT消息处理器中使用

```go
package mqtt

import (
    "log"

    "sfsEdgeStore/core/edgex"
)

func messageHandler(payload []byte) {
    event, err := edgex.ProcessMessage(payload)
    if err != nil {
        log.Printf("Error processing EdgeX message: %v", err)
        return
    }

    if event == nil {
        // 消息被忽略（非event类型）
        return
    }

    // 处理事件...
    for _, reading := range event.Readings {
        log.Printf("Device %s: %s = %s (%s)",
            event.DeviceName,
            reading.ResourceName,
            reading.Value,
            reading.ValueType)
    }
}
```

### 5.3 错误处理

```go
func handleMessage(payload []byte) {
    event, err := edgex.ProcessMessage(payload)
    if err != nil {
        // JSON解析错误
        log.Printf("Invalid JSON format: %v", err)
        return
    }

    if event == nil {
        // 消息类型不是event，被正常忽略
        return
    }

    // 正常处理事件
    processEvent(event)
}

func processEvent(event *edgex.EdgeXEvent) {
    // 业务逻辑处理
}
```

## 6. 注意事项

### 6.1 消息类型过滤

系统只处理MessageType为"event"或"Event"的消息，其他类型的消息会被忽略并返回nil：

```go
if edgexMsg.MessageType != "event" && edgexMsg.MessageType != "Event" {
    log.Printf("Ignoring message with type: %s", edgexMsg.MessageType)
    return nil, nil
}
```

### 6.2 设备名称格式化

设备名称会被格式化为64字符长度，以确保数据库存储的一致性：

```go
event.DeviceName = common.FormatDeviceName(event.DeviceName)
```

### 6.3 Payload格式

Payload字段本身是一个JSON字符串，需要进行二次解析：

```go
// 第一次解析：提取MessageEnvelope
var edgexMsg EdgeXMessage
json.Unmarshal(payload, &edgexMsg)

// 第二次解析：从Payload字段提取Event
var event EdgeXEvent
json.Unmarshal(edgexMsg.Payload, &event)
```

### 6.4 值类型

EdgeXReading.Value字段始终是字符串格式，需要根据ValueType进行类型转换：

```go
value := reading.Value
switch reading.ValueType {
case "Float64":
    f, _ := strconv.ParseFloat(value, 64)
    // 使用浮点值
case "Int64":
    i, _ := strconv.ParseInt(value, 10, 64)
    // 使用整数值
case "Bool":
    b, _ := strconv.ParseBool(value)
    // 使用布尔值
default:
    // 作为字符串使用
}
```

## 7. 性能优化

### 7.1 JSON解析优化

- 使用json.RawMessage延迟解析Payload，减少不必要的内存分配
- Payload字段在确认消息类型后再解析，避免处理无用数据

### 7.2 消息过滤

- 在消息入口处进行类型检查，快速跳过不需要处理的消息
- 非event类型的消息不会进入后续处理流程

### 7.3 内存管理

- 使用json.RawMessage避免Payload的二次拷贝
- 读数切片按需分配，减少预分配内存

## 8. 错误处理

### 8.1 常见错误

| 错误场景 | 原因 | 处理方式 |
|---------|------|---------|
| JSON解析失败 | payload不是有效的JSON | 返回错误 |
| MessageType不匹配 | 消息类型不是event | 返回nil, nil |
| Payload解析失败 | Payload不是有效的JSON | 返回错误 |

### 8.2 错误返回

```go
// 解析Envelope失败
if err := json.Unmarshal(payload, &edgexMsg); err != nil {
    return nil, err
}

// MessageType不匹配，返回nil表示忽略
if edgexMsg.MessageType != "event" && edgexMsg.MessageType != "Event" {
    return nil, nil
}

// 解析Payload失败
if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
    return nil, err
}
```

## 9. 与其他模块的集成

### 9.1 上游模块（MQTT客户端）

MQTT客户端接收消息后调用ProcessMessage：

```go
// mqtt/client.go
func onMessage(topic string, payload []byte) {
    event, err := edgex.ProcessMessage(payload)
    if err != nil {
        log.Printf("Failed to process message: %v", err)
        return
    }
    if event != nil {
        // 传递给下游处理
        processEvent(event)
    }
}
```

### 9.2 下游处理

处理返回的EdgeXEvent：

```go
func processEvent(event *edgex.EdgeXEvent) {
    // 存储到数据库
    storeReadings(event)

    // 更新设备状态
    updateDeviceStatus(event)

    // 检查告警条件
    checkAlerts(event)

    // 广播到WebSocket
    broadcastEvent(event)
}
```

## 10. 总结

EdgeX核心处理模块是sfsEdgeStore与EdgeX Foundry通信的桥梁，通过标准化的消息解析和设备名称格式化，确保了系统与EdgeX的无缝集成。其设计遵循以下原则：

- **标准化**：完全遵循EdgeX MessageEnvelope规范
- **高效性**：延迟解析和消息过滤提高处理效率
- **可靠性**：完善的错误处理机制
- **一致性**：统一的设备名称格式化
- **可扩展性**：清晰的数据结构便于扩展新功能