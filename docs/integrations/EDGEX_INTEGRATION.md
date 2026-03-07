# EdgeX Foundry 集成指南

## 概述

sfsEdgeStore 提供了与 EdgeX Foundry 物联网边缘计算平台的深度集成能力。本指南将详细介绍如何配置和使用 sfsEdgeStore 来存储和处理 EdgeX Foundry 产生的传感器数据。

## 集成架构

```
┌─────────────┐     MQTT     ┌──────────────┐     HTTP/MQTT     ┌───────────────┐
│ EdgeX Foundry│ ───────────→ │  sfsEdgeStore│ ───────────────→ │  应用系统     │
│  (设备服务)  │              │  (数据存储)   │                   │  (查询/分析)   │
└─────────────┘              └──────────────┘                   └───────────────┘
```

## 数据模型

### EdgeX Message 格式

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

### 数据模型说明

| 字段 | 类型 | 说明 |
|------|------|------|
| correlationId | string | 消息关联 ID，用于追踪 |
| messageType | string | 消息类型，必须为 "event" |
| origin | int64 | 消息时间戳（毫秒） |
| payload | object | EdgeX Event 数据 |
| deviceName | string | 设备名称（自动格式化至 64 字符） |
| readings | array | 传感器读数数组 |

## 配置步骤

### 1. 配置 MQTT 连接

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

### 2. 配置 EdgeX Foundry

在 EdgeX Foundry 的配置文件中，设置消息发布到 sfsEdgeStore 的 MQTT broker：

```yaml
application:
  mqtt:
    url: tcp://sfsedgestore-host:1883
    topic: edgex/events
    client-id: edgex-application-service
```

### 3. 启动 sfsEdgeStore

```bash
# 启动服务
./sfsedgestore
```

## MQTT 主题订阅

sfsEdgeStore 支持以下 MQTT 主题模式：

- `edgex/events/#` - 订阅所有 EdgeX 事件
- `edgex/events/device/+/+` - 按设备订阅
- `edgex/events/profile/+/+` - 按设备配置文件订阅

## 数据存储

### 存储格式

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

### 索引策略

- **主键索引**: DeviceName + ResourceName + Timestamp
- **时间范围索引**: 支持高效的时间范围查询
- **设备索引**: 按设备名称快速检索

## 查询 API

### 查询指定设备的数据

```bash
curl "http://localhost:8080/api/readings?device=sensor-001&start=1620000000000&end=1620003600000"
```

### 查询指定资源的数据

```bash
curl "http://localhost:8080/api/readings?device=sensor-001&resource=temperature&limit=100"
```

### 聚合查询

```bash
curl "http://localhost:8080/api/aggregate?device=sensor-001&resource=temperature&granularity=1h&function=avg"
```

## 完整集成示例

### 1. 使用 Docker Compose 部署

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

### 2. 发送测试数据

使用提供的示例脚本：

```bash
cd scripts
./test_mqtt_publish.ps1
```

或手动发送：

```bash
mosquitto_pub -h localhost -t edgex/events -f sample_edgex_payload.json
```

## 故障排除

### 常见问题

**1. 无法连接到 MQTT Broker**
- 检查 broker 地址和端口
- 验证网络连接
- 检查用户名和密码

**2. 消息未被处理**
- 确认 messageType 为 "event"
- 检查 JSON 格式是否正确
- 查看日志文件

**3. 数据未存储**
- 检查数据库路径权限
- 验证磁盘空间
- 查看错误日志

### 日志查看

```bash
# 查看实时日志
tail -f logs/sfsedgestore.log

# 过滤 EdgeX 相关日志
grep "EdgeX" logs/sfsedgestore.log
```

## 性能优化

### 批量处理

配置批量处理参数：

```json
{
  "queue": {
    "batchSize": 100,
    "flushInterval": 1000
  }
}
```

### 数据保留策略

配置数据保留时间：

```json
{
  "retention": {
    "enabled": true,
    "days": 90
  }
}
```

### 索引优化

根据查询模式优化索引配置，参考架构文档中的数据库优化建议。

## 安全建议

1. **启用 MQTT 认证**
   - 配置用户名和密码
   - 使用 TLS 加密连接

2. **网络隔离**
   - 将 sfsEdgeStore 部署在私有网络
   - 使用防火墙限制访问

3. **数据加密**
   - 启用数据库加密
   - 敏感数据进行脱敏处理

## 更多资源

- [API 文档](../api.md)
- [架构文档](../architecture/ARCHITECTURE.md)
- [部署指南](../../DEPLOYMENT.md)
- [故障排除](../user-guide/TROUBLESHOOTING.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
