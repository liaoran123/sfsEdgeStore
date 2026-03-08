## 第5章：与 EdgeX Foundry 深度集成

### 5.1 集成架构

sfsEdgeStore 提供了与 EdgeX Foundry 物联网边缘计算平台的深度集成能力。

```
┌─────────────┐     MQTT     ┌──────────────┐     HTTP/MQTT     ┌───────────────┐
│ EdgeX Foundry│ ───────────→ │  sfsEdgeStore│ ───────────────→ │  应用系统     │
│  (设备服务)  │              │  (数据存储)   │                   │  (查询/分析)   │
└─────────────┘              └──────────────┘                   └───────────────┘
```

### 5.2 数据模型

#### EdgeX Message 格式

sfsEdgeStore 支持 EdgeX Foundry 的 MessageEnvelope 格式：

```json
{
  "correlationId": "uuid-string",
  "messageType": "event",
  "origin": 1620000000000,
  "payload": {
    "id": "event-uuid",
    "deviceName": "sensor-001",
    "profileName": "temperature-sensor",
    "sourceName": "temperature",
    "origin": 1620000000000,
    "readings": [
      {
        "id": "reading-uuid",
        "resourceName": "temperature",
        "value": "25.5",
        "valueType": "Float64",
        "origin": 1620000000000,
        "deviceName": "sensor-001",
        "profileName": "temperature-sensor"
      }
    ]
  }
}
```

#### 数据模型说明

| 字段 | 类型 | 说明 |
|------|------|------|
| correlationId | string | 消息关联 ID，用于追踪 |
| messageType | string | 消息类型，必须为 "event" |
| origin | int64 | 消息时间戳（毫秒） |
| payload | object | EdgeX Event 数据 |
| deviceName | string | 设备名称（自动格式化至 64 字符） |
| readings | array | 传感器读数数组 |

### 5.3 配置步骤

#### 步骤 1：配置 MQTT 连接

编辑 `config.json` 文件：

```json
{
  "mqtt": {
    "broker": "tcp://localhost:1883",
    "topic": "edgex/events/#",
    "clientId": "sfsedgestore-edgex",
    "qos": 1,
    "username": "",
    "password": ""
  }
}
```

#### 步骤 2：配置 EdgeX Foundry

在 EdgeX Foundry 的配置文件中，设置消息发布到 sfsEdgeStore 的 MQTT broker：

```yaml
application:
  mqtt:
    url: tcp://sfsedgestore-host:1883
    topic: edgex/events
    client-id: edgex-application-service
```

#### 步骤 3：启动 sfsEdgeStore

```bash
# 启动服务
./sfsedgestore
```

### 5.4 MQTT 主题订阅

sfsEdgeStore 支持以下 MQTT 主题模式：

- `edgex/events/#` - 订阅所有 EdgeX 事件
- `edgex/events/device/+/+` - 按设备订阅
- `edgex/events/profile/+/+` - 按设备配置文件订阅

### 5.5 数据存储

#### 存储格式

EdgeX 数据转换为 sfsEdgeStore 的内部格式存储：

```go
type Reading struct {
    ID           string
    DeviceName   string
    ResourceName string
    Value        string
    ValueType    string
    Timestamp    int64
    Metadata     map[string]string
}
```

#### 索引策略

- **主键索引**: DeviceName + ResourceName + Timestamp
- **时间范围索引**: 支持高效的时间范围查询
- **设备索引**: 按设备名称快速检索

### 5.6 查询 API

#### 查询指定设备的数据

```bash
curl "http://localhost:8080/api/readings?device=sensor-001&start=1620000000000&end=1620003600000"
```

#### 查询指定资源的数据

```bash
curl "http://localhost:8080/api/readings?device=sensor-001&resource=temperature&limit=100"
```

#### 聚合查询

```bash
curl "http://localhost:8080/api/aggregate?device=sensor-001&resource=temperature&granularity=1h&function=avg"
```

### 5.7 完整集成示例

#### 使用 Docker Compose 部署

创建 `docker-compose.yml`：

```yaml
version: '3'

services:
  mosquitto:
    image: eclipse-mosquitto:2.0
    ports:
      - "1883:1883"

  sfsedgestore:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - mosquitto
    environment:
      - MQTT_BROKER=tcp://mosquitto:1883
      - MQTT_TOPIC=edgex/events/#

  edgex:
    image: edgexfoundry/docker-edgex-compose:2.3.0
    depends_on:
      - mosquitto
```

#### 发送测试数据

使用提供的示例脚本：

```bash
cd scripts
./test_mqtt_publish.ps1
```

或手动发送：

```bash
mosquitto_pub -h localhost -t edgex/events -f sample_edgex_payload.json
```


