# sfsDb EdgeX 适配器

[English Version](README.md)

## 概述

sfsDb EdgeX 适配器提供了 EdgeX Foundry 与 sfsDb 嵌入式数据库之间的无缝集成，实现了边缘设备数据的高效存储和检索。该适配器遵循 EdgeX Foundry 的最佳实践和边缘计算解决方案标准。

## 特性

- **EdgeX Foundry 兼容**：实现了 EdgeX MessageEnvelope 格式和 MQTT 消息总线集成
- **高效存储**：使用 sfsDb 嵌入式数据库进行轻量级、高性能的数据存储
- **实时处理**：实时处理和存储 EdgeX 事件
- **自动数据库管理**：自动初始化数据库并创建优化的索引
- **可配置**：支持通过环境变量和配置文件进行配置
- **健康监控**：提供 HTTP 健康检查端点
- **数据备份**：包含数据持久化的备份和恢复功能

## 前提条件

- **EdgeX Foundry**：v3.0.0 或更高版本
- **MQTT 代理**：Mosquitto 2.0+ 或兼容的 MQTT 代理
- **Go**：1.25 或更高版本
- **sfsDb**：v1.0.0 或更高版本
- **操作系统**：Linux、macOS 或 Windows

## 快速开始

### 从源代码构建

```bash
git clone https://github.com/your-org/sfsdb-edgex-adapter.git
cd sfsdb-edgex-adapter
go build
```

### 运行适配器

```bash
./sfsdb-edgex-adapter
```

### 默认配置

默认情况下，适配器将：
- 连接到 `tcp://localhost:1883` 的 MQTT 代理
- 订阅 `edgex/events/core/#` 主题
- 将数据存储在 `./edgex_data` 目录中
- 创建带有优化索引的 `edgex_readings` 表
- 在端口 8080 上启动 HTTP 健康检查服务器

## 配置

### EdgeX 配置标准

适配器遵循 EdgeX Foundry 配置标准，支持多种配置源：

1. **环境变量**（最高优先级）
2. **配置文件**（`config.json`）
3. **默认值**（最低优先级）

### 环境变量

- `EDGEX_DB_PATH` - 数据库存储路径
- `EDGEX_MQTT_BROKER` - MQTT 代理地址
- `EDGEX_MQTT_TOPIC` - 要订阅的 MQTT 主题
- `EDGEX_CLIENT_ID` - MQTT 客户端 ID

### 配置文件示例

```json
{
  "db_path": "./edgex_data",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsdb-edgex-adapter"
}
```

## EdgeX 集成

### 消息格式

适配器处理 EdgeX Foundry 定义的标准 **MessageEnvelope** 格式的消息：

```json
{
  "correlationId": "5e8a3c9d-1b2c-4d5e-8f9a-1b2c3d4e5f6g",
  "messageType": "event",
  "origin": 1677721600000000000,
  "payload": {
    "id": "event-123",
    "deviceName": "Thermostat-001",
    "readings": [
      {
        "id": "reading-123",
        "resourceName": "temperature",
        "value": "25.5",
        "origin": 1677721600000000000,
        "profileName": "ThermostatProfile",
        "deviceName": "Thermostat-001",
        "metadata": {"unit": "Celsius"}
      }
    ],
    "origin": 1677721600000000000,
    "profileName": "ThermostatProfile",
    "sourceName": "ThermostatSource"
  }
}
```

### 数据存储

数据存储在 `edgex_readings` 表中，具有以下架构，针对 EdgeX 数据模式进行了优化：

| 字段 | 类型 | 描述 |
|------|------|------|
| `id` | 字符串 | 唯一读数 ID |
| `deviceName` | 字符串 | 设备名称（复合主键的一部分） |
| `reading` | 字符串 | 资源名称（例如，温度、湿度） |
| `value` | 浮点数 | 读数值 |
| `timestamp` | 整数 | 时间戳（秒）（复合主键的一部分） |
| `metadata` | 字符串 | JSON 元数据 |

### 索引

- **复合主键**：`(deviceName, timestamp)` 用于高效的时间范围查询
- **时间索引**：用于基于时间的过滤

### 查询优化

适配器通过 `queryReadings` 函数实现高效的查询功能，该函数：
- 支持按设备名称和时间范围过滤
- 利用复合主键进行优化的范围查询
- 解析 RFC3339 格式的时间戳用于基于时间的过滤
- 以一致的格式返回结构化结果

**查询参数**：
- `deviceName`：可选的设备名称过滤器
- `startTime`：可选的开始时间（RFC3339 格式）
- `endTime`：可选的结束时间（RFC3339 格式）

**查询示例**：

```bash
# 查询特定设备的所有读数
GET /api/readings?deviceName=Thermostat-001

# 查询设备在特定时间范围内的读数
GET /api/readings?deviceName=Thermostat-001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-02T00:00:00Z
```

## API

### 健康检查端点

- **URL**：`/health`
- **方法**：GET
- **响应**：适配器的 JSON 状态

示例响应：

```json
{
  "status": "healthy",
  "components": {
    "database": "connected",
    "mqtt": "connected",
    "adapter": "running"
  }
}
```

### 读数查询端点

- **URL**：`/api/readings`
- **方法**：GET
- **参数**：
  - `deviceName`（可选）：按设备名称过滤
  - `startTime`（可选）：RFC3339 格式的开始时间
  - `endTime`（可选）：RFC3339 格式的结束时间
- **响应**：读数的 JSON 数组

示例响应：

```json
[
  {
    "id": "reading-123",
    "deviceName": "Thermostat-001",
    "reading": "temperature",
    "value": 25.5,
    "timestamp": 1677721600,
    "metadata": "{\"unit\": \"Celsius\"}"
  }
]
```

### 备份 API 端点

- **URL**：`/api/backup`
- **方法**：POST
- **参数**：
  - `path`（可选）：备份存储路径（默认：`./backups`）
- **响应**：带有备份状态和文件路径的 JSON 对象

示例请求：

```bash
# 使用默认路径创建备份
POST /api/backup

# 使用自定义路径创建备份
POST /api/backup?path=/path/to/backups
```

示例响应：

```json
{
  "status": "success",
  "backupFile": "./backups/backup_20240101_120000"
}
```

### 恢复 API 端点

- **URL**：`/api/restore`
- **方法**：POST
- **参数**：
  - `file`（必需）：备份文件路径
- **响应**：带有恢复状态的 JSON 对象

示例请求：

```bash
# 从备份文件恢复
POST /api/restore?file=./backups/backup_20240101_120000
```

示例响应：

```json
{
  "status": "success",
  "message": "Database restored successfully"
}
```

## 监控

### 日志

适配器遵循 EdgeX Foundry 日志标准，提供结构化日志用于：
- 配置加载
- 数据库初始化和操作
- MQTT 连接状态
- 消息处理
- 错误条件和警告

### 指标

适配器公开以下指标用于监控：
- 消息处理速率
- 数据库操作延迟
- MQTT 连接状态
- 存储利用率

## 备份和恢复

### 备份

创建数据库备份：

```bash
# 备份功能已集成到适配器中
# 有关详细信息，请参阅 backup 包
```

### 恢复

从备份恢复：

```bash
# 恢复功能已集成到适配器中
# 有关详细信息，请参阅 backup 包
```

## 故障排除

### 常见问题

1. **MQTT 连接失败**
   - 确保 MQTT 代理正在运行
   - 验证代理地址和端口配置
   - 检查网络连接

2. **数据库初始化失败**
   - 确保数据库目录可写
   - 检查文件系统权限
   - 验证有足够的磁盘空间

3. **消息处理错误**
   - 验证消息格式是否符合 EdgeX MessageEnvelope 标准
   - 检查日志输出以获取详细的错误消息
   - 确保 EdgeX Foundry 版本兼容性

4. **性能问题**
   - 考虑增加 MQTT 消息队列大小
   - 验证数据库目录是否位于快速存储上
   - 监控系统资源（CPU、内存、磁盘 I/O）

## 版本兼容性

| 组件 | 版本 |
|------|------|
| EdgeX Foundry | v3.0.0+ |
| Go | 1.25+ |
| sfsDb | v1.0.0+ |
| MQTT 代理 | Mosquitto 2.0+ |

## 安全

### 最佳实践

- 使用安全的 MQTT 连接（TLS）
- 为数据库实现适当的访问控制
- 定期更新依赖项
- 遵循 EdgeX Foundry 安全指南

## 部署

### 容器化

适配器可以使用 Docker 进行容器化：

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o sfsdb-edgex-adapter

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sfsdb-edgex-adapter .
COPY config.json .
EXPOSE 8080
CMD ["./sfsdb-edgex-adapter"]
```

### EdgeX Foundry 部署

适配器可以作为 EdgeX Foundry 实例的一部分使用 Docker Compose 进行部署：

```yaml
version: '3.8'
services:
  sfsdb-edgex-adapter:
    image: sfsdb-edgex-adapter:latest
    depends_on:
      - mqtt-broker
    environment:
      - EDGEX_MQTT_BROKER=tcp://mqtt-broker:1883
    volumes:
      - ./edgex_data:/app/edgex_data
    ports:
      - "8080:8080"
```

## 测试

### 单元测试

```bash
go test -v ./...
```

### 集成测试

```bash
go test -v -run TestIntegration
```

### EdgeX 兼容性测试

适配器已通过 EdgeX Foundry v3.0.0+ 测试，以确保与 EdgeX 消息格式和集成模式的兼容性。

## 贡献

### EdgeX 贡献指南

本项目遵循 EdgeX Foundry 贡献指南。有关如何贡献的详细信息，请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)。

### 代码风格

- 遵循 Go 代码风格标准
- 使用 EdgeX Foundry 推荐的模式
- 包含全面的测试覆盖
- 记录所有公共 API

## 许可证

本项目采用 [Apache 2.0 许可证](LICENSE)，与 EdgeX Foundry 许可要求一致。

## 支持

### 社区支持

- EdgeX Foundry 社区论坛
- GitHub issues 用于错误报告和功能请求

### 商业支持

我们为 sfsdb-edgex-adapter 提供商业支持和企业解决方案：

- **企业版**：高级功能，包括增强的数据压缩、安全集成和监控能力
- **咨询服务**：EdgeX 集成咨询和架构设计
- **定制开发**：为特定行业量身定制的解决方案
- **技术培训**：EdgeX 和 sfsDb 相关培训
- **优先支持**：专用技术支持，保证响应时间

有关商业产品的更多信息，请通过 [sfsweb@qq.com](mailto:sfsweb@qq.com) 联系我们。

## 路线图

### 未来增强

- 支持 EdgeX Foundry v4.0.0
- 高级数据压缩和保留策略
- 与 EdgeX 安全服务集成
- 增强的监控和分析功能
- 支持其他数据库后端