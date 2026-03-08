---
title: EdgeX Foundry 与 sfsDb 结合
subtitle: 工业物联网边缘计算数据存储实战
author: sfsEdgeStore 团队
date: 2026-03-08
version: 1.0.0
---

# EdgeX Foundry 与 sfsDb 结合

## 工业物联网边缘计算数据存储实战

---

## 关于本书

本书将带领您从零开始，学习如何使用 sfsEdgeStore 构建轻量级的工业物联网边缘计算数据存储解决方案。通过结合 EdgeX Foundry 和 sfsDb 数据库，您将掌握解决边缘场景数据存储痛点的核心技术。

### 本书特色

- ✅ **实战导向**：大量代码示例和实战案例
- ✅ **由浅入深**：从基础概念到企业级部署
- ✅ **完整覆盖**：包括架构设计、部署、监控、安全等全流程
- ✅ **真实案例**：基于实际生产环境的成功经验

### 读者对象

- 工业物联网开发者
- 边缘计算架构师
- DevOps 工程师
- 系统管理员
- 对边缘存储感兴趣的技术人员

---

## 目录

### 第一部分：基础篇

1. [工业物联网边缘计算概述](#第1章工业物联网边缘计算概述)
2. [EdgeX Foundry 基础](#第2章edgex-foundry-基础)
3. [sfsDb 数据库入门](#第3章sfsdb-数据库入门)

### 第二部分：实践篇

4. [sfsEdgeStore 快速开始](#第4章sfsedgestore-快速开始)
5. [与 EdgeX Foundry 深度集成](#第5章与-edgex-foundry-深度集成)
6. [数据存储与查询](#第6章数据存储与查询)

### 第三部分：进阶篇

7. [监控与告警](#第7章监控与告警)
8. [认证与安全](#第8章认证与安全)
9. [备份与恢复](#第9章备份与恢复)

### 第四部分：企业篇

10. [生产部署最佳实践](#第10章生产部署最佳实践)
11. [商业服务与支持](#第11章商业服务与支持)
12. [成功案例](#第12章成功案例)

---

## 第一部分：基础篇

---

## 第1章：工业物联网边缘计算概述

### 1.1 边缘计算的发展背景

在工业物联网（IIoT）和边缘计算快速发展的今天，越来越多的企业开始在边缘节点部署数据采集和处理系统。然而，在实际落地过程中，我们发现了一个普遍的痛点：**边缘数据存储太难了！**

### 1.2 工业物联网的六大困境

让我们看看大多数企业在边缘计算场景中遇到的六大困境：

| 困境 | 痛点描述 | 影响 |
|------|---------|------|
| **设备资源有限** | 边缘设备 CPU 弱、内存小，跑不起重型数据库 | 性能差、经常卡顿 |
| **网络不稳定** | 工厂车间、户外场景网络时常中断，数据丢失 | 业务中断、数据损失 |
| **部署太复杂** | 需要配置数据库、调优参数、处理依赖 | 上线周期长、人力成本高 |
| **查询响应慢** | 云端查询延迟高，本地查询也不给力 | 实时性差、决策滞后 |
| **过度依赖云端** | 一旦云端连接断开，整个系统瘫痪 | 可用性低、风险高 |
| **与 EdgeX 集成难** | EdgeX Foundry 缺少简单高效的数据存储方案 | 集成成本高、扩展性差 |

这些困境不仅增加了项目成本，还大大延长了上线周期。

### 1.3 sfsEdgeStore 的定位和价值

**sfsEdgeStore** 是专为工业物联网边缘场景设计的轻量级数据存储适配器，作为 EdgeX Foundry 和 sfsDb 数据库之间的桥梁，提供高效的本地数据读写和缓存能力。

#### 核心价值

- ✅ **极轻量**：资源占用极低，可在任何边缘设备上运行
- ✅ **高可靠**：本地存储，网络中断不影响数据采集
- ✅ **易集成**：与 EdgeX Foundry 原生集成，开箱即用
- ✅ **高性能**：LevelDB 底层，本地查询毫秒级响应
- ✅ **开源免费**：完整功能，无限制使用

#### 性能有多强？用数据说话

我们不玩虚的，直接上实测数据：

| 性能指标 | sfsEdgeStore 实测值 | 业界同类产品对比 | 测试文件 |
|---------|-------------------|----------------|---------|
| **内存占用** | 20.85 MB | 通常 200MB+（是我们的 10 倍） | [PERFORMANCE_REPORT.md](https://github.com/liaoran123/sfsEdgeStore/blob/main/PERFORMANCE_REPORT.md) |
| **CPU 使用率** | 2.9% | 通常 15-30% | [PERFORMANCE_REPORT.md](https://github.com/liaoran123/sfsEdgeStore/blob/main/PERFORMANCE_REPORT.md) |
| **启动时间** | 0.187 秒 | 通常 5-30 秒 | [test_startup_time.ps1](https://github.com/liaoran123/sfsEdgeStore/blob/main/test_startup_time.ps1) |
| **数据库大小** | 18,681 条记录仅占 0.25 MB | 通常需要数 MB 甚至更多 | [PERFORMANCE_REPORT.md](https://github.com/liaoran123/sfsEdgeStore/blob/main/PERFORMANCE_REPORT.md) |

这意味着什么？
- ✅ **可以在任何边缘设备上运行**：即使是资源极其受限的设备也没问题
- ✅ **后台运行几乎不占用资源**：不影响其他业务系统运行
- ✅ **毫秒级响应**：0.187 秒启动，查询也是毫秒级
- ✅ **存储效率极高**：18,000+ 条记录仅占 0.25 MB

### 1.4 本书学习路线

在本书中，我们将按照以下路线逐步学习：

1. **基础篇**：理解边缘计算、EdgeX Foundry 和 sfsDb 的基础概念
2. **实践篇**：动手部署 sfsEdgeStore 并与 EdgeX Foundry 集成
3. **进阶篇**：掌握监控、安全、备份等高级功能
4. **企业篇**：学习生产环境部署和商业服务支持

准备好了吗？让我们开始学习之旅！

---

## 第2章：EdgeX Foundry 基础

### 2.1 EdgeX Foundry 简介

EdgeX Foundry 是一个高度灵活、可扩展的开源边缘计算框架，专为工业物联网设计。它提供了标准化的设备连接、数据采集和边缘处理能力。

### 2.2 EdgeX Foundry 架构介绍

EdgeX Foundry 采用分层架构，主要包括以下层次：

```
┌─────────────────────────────────────────────────────────────┐
│                    应用服务层 (Export Services)              │
│  - 数据导出、转换、转发到云端                                │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                    核心服务层 (Core Services)                │
│  - Core Data：数据存储和管理                                  │
│  - Core Metadata：设备元数据管理                              │
│  - Core Command：设备命令控制                                │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                  设备服务层 (Device Services)                │
│  - 连接各类物理设备（传感器、执行器等）                        │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 核心组件和数据流

#### 主要组件

| 组件 | 职责 |
|------|------|
| **Core Data** | 存储和管理设备数据 |
| **Core Metadata** | 管理设备配置文件和元数据 |
| **Core Command** | 向设备发送命令 |
| **Device Services** | 与实际设备通信 |
| **Application Services** | 处理和导出数据 |

#### 典型数据流

1. 设备采集数据
2. Device Service 接收数据
3. Core Data 存储数据
4. Application Service 处理数据
5. 数据导出到云端或其他系统

### 2.4 部署和配置

#### 部署方式

EdgeX Foundry 支持多种部署方式：

1. **Docker Compose**：快速测试和开发
2. **Kubernetes**：生产环境高可用部署
3. **二进制部署**：资源受限设备

#### 基本配置

主要配置项包括：
- MQTT Broker 配置
- 数据存储配置
- 设备服务配置
- 安全配置

### 2.5 与 sfsEdgeStore 的关系

sfsEdgeStore 作为 EdgeX Foundry 的数据存储适配器，通过 MQTT 订阅 EdgeX Foundry 的事件主题，将数据高效存储到本地 sfsDb 数据库中。

---

## 第3章：sfsDb 数据库入门

### 3.1 sfsDb 设计理念

sfsDb 是一个轻量级的嵌入式数据库，基于 LevelDB 封装而成，专为边缘计算和 IoT 场景优化。

#### 设计原则

- **小而美**：只做一件事，做到极致
- **边缘优先**：所有功能优先考虑边缘场景
- **零依赖**：除了 LevelDB，不依赖其他重型组件
- **高可用**：断电恢复、数据重试、本地存储

### 3.2 与 LevelDB 的关系

sfsDb 是对 Google LevelDB 的高级封装，提供了更友好的 API 和更多企业级功能：

| 特性 | LevelDB | sfsDb |
|------|---------|-------|
| 基础存储 | ✅ | ✅ |
| Go 语言绑定 | ✅ | ✅ |
| 数据加密 | ❌ | ✅ (AES-256-GCM) |
| 批量操作优化 | 基础 | 增强 |
| 场景化配置 | ❌ | ✅ (embedded/iot/edge/game) |
| 事务支持 | 基础 | 增强 |

### 3.3 核心特性和优势

#### 核心特性

- ✅ **嵌入式存储**：无需单独的数据库进程
- ✅ **高性能读写**：LevelDB 底层，优化的 LSM 树
- ✅ **数据压缩**：自动压缩，节省存储空间
- ✅ **AES-256 加密**：保护敏感数据
- ✅ **多种场景配置**：针对不同场景优化

#### 数据库场景配置

| 场景 | 写缓冲 | 块缓存 | 适用场景 |
|------|--------|----------|----------|
| embedded | 2MB | 4MB | 嵌入式设备 |
| iot | 4MB | 8MB | IoT 设备 |
| **edge** (默认) | 16MB | 32MB | 边缘计算 |
| game | 64MB | 128MB | 游戏场景 |

### 3.4 sfsEdgeStore 中的 sfsDb

在 sfsEdgeStore 中，sfsDb 作为核心存储引擎，负责：
- 存储 EdgeX Foundry 发送的传感器数据
- 提供高效的时间范围查询
- 支持数据加密和压缩
- 管理数据保留策略

---

## 第二部分：实践篇

---

## 第4章：sfsEdgeStore 快速开始

### 4.1 前置条件

在开始之前，请确保您的环境满足以下要求：

- Go 1.25+（如需从源码编译）
- EdgeX Foundry（可选，用于数据源）
- MQTT Broker（如 Mosquitto）

### 4.2 5 分钟快速部署

我们提供三种部署方式，您可以根据需求选择。

#### 方式 1：二进制部署（推荐，高性能）

追求极致性能？使用二进制文件配合 systemd（Linux）或 Windows 服务，零虚拟化开销！

**Linux（systemd）：**

```bash
# 从 GitHub Releases 下载对应平台的二进制文件
# https://github.com/liaoran123/sfsEdgeStore/releases

# 直接运行（测试用）
./sfsedgestore

# 生产环境推荐使用 systemd 守护（开机自启、崩溃重启）
```

**Windows：**

```bash
# 使用 Windows 服务或 NSSM 等工具配置为系统服务
```

**优势：**
- ⚡ 最高性能
- 🚀 零虚拟化开销
- 🎯 生产环境首选

**适用场景：**
- 生产环境
- 性能要求极高的场景
- 资源受限的设备

#### 方式 2：Docker 部署（方便，快速体验）

Docker 适合快速测试和部署，**自带守护功能（开机自启、崩溃重启），但会有轻微的性能开销（约 5-10%）。**

```bash
# 拉取镜像
docker pull sfsedgestore/sfsedgestore:latest

# 运行
docker run -d \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  sfsedgestore/sfsedgestore:latest
```

**优势：**
- 🐳 部署方便
- 🔄 自带守护功能
- 📦 环境一致

**适用场景：**
- 快速测试
- 开发环境
- 需要快速体验的场景

#### 方式 3：从源码编译

```bash
# 克隆仓库
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore

# 安装依赖
go mod download

# 编译
go build -o sfsedgestore

# 运行
./sfsedgestore
```

### 4.3 验证安装

安装完成后，让我们验证一下：

```bash
# 健康检查
curl http://localhost:8080/health

# 查看指标
curl http://localhost:8080/metrics
```

健康检查响应示例：

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "1h23m45s"
}
```

### 4.4 配置文件说明

sfsEdgeStore 使用 `config.json` 进行配置，主要配置项包括：

```json
{
  "mqtt": {
    "broker": "tcp://localhost:1883",
    "topic": "edgex/events/#",
    "clientId": "sfsedgestore",
    "qos": 1
  },
  "db": {
    "path": "./edgex_data",
    "scenario": "edge"
  },
  "http": {
    "port": "8080"
  },
  "simulator": {
    "enabled": true,
    "intervalMin": 1,
    "intervalMax": 2
  }
}
```

---

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

---

## 第6章：数据存储与查询

### 6.1 数据存储策略

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

### 6.2 基本使用

#### 启动和停止

##### 前台运行

```bash
./sfsedgestore
```

##### 后台运行（Linux/macOS）

```bash
nohup ./sfsedgestore > sfsedgestore.log 2>&1 &
```

##### 停止服务

```bash
# 查找进程
ps aux | grep sfsedgestore

# 优雅停止
kill -TERM <pid>

# 强制停止
kill -9 <pid>
```

### 6.3 数据查询

#### 查询所有数据

```bash
curl "http://localhost:8080/api/readings"
```

#### 按设备查询

```bash
curl "http://localhost:8080/api/readings?deviceName=Device001"
```

#### 按时间范围查询

```bash
curl "http://localhost:8080/api/readings?startTime=2026-03-01T00:00:00Z&endTime=2026-03-07T23:59:59Z"
```

#### 分页查询

```bash
curl "http://localhost:8080/api/readings?limit=100&offset=0"
```

#### 组合查询

```bash
curl "http://localhost:8080/api/readings?deviceName=Device001&startTime=2026-03-01T00:00:00Z&limit=50"
```

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

### 6.4 数据导出

#### 导出 JSON 格式

```bash
curl "http://localhost:8080/api/readings?format=json" > data.json
```

#### 导出 CSV 格式

```bash
curl "http://localhost:8080/api/readings?format=csv" > data.csv
```

### 6.5 健康检查和监控

#### 健康检查

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": "2026-03-07T10:00:00Z",
  "version": "1.0.0",
  "uptime": "3600s"
}
```

#### 获取指标

```bash
curl http://localhost:8080/metrics
```

响应示例：

```json
{
  "system": {
    "cpu_usage": 2.5,
    "memory_usage": 20.8,
    "disk_usage": 45.2,
    "goroutines": 15
  },
  "business": {
    "total_readings": 18681,
    "mqtt_messages_received": 25000,
    "mqtt_messages_processed": 24980,
    "queue_length": 0
  },
  "database": {
    "size_mb": 0.25,
    "read_count": 150000,
    "write_count": 25000
  }
}
```

#### 查看告警

```bash
curl http://localhost:8080/alerts
```

---

## 第三部分：进阶篇

---

## 第7章：监控与告警

### 7.1 监控概述

sfsEdgeStore 提供了完整的监控体系，包括系统指标、业务指标、健康检查和告警管理。

### 7.2 监控架构

```
┌─────────────────┐
│  sfsEdgeStore   │
│  (Metrics API)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐    ┌─────────────────┐
│  Prometheus     │───▶│  Grafana        │
│  (采集/存储)    │    │  (可视化)       │
└─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐
│  Alertmanager   │
│  (告警)         │
└─────────────────┘
```

### 7.3 内置监控端点

#### 健康检查端点

##### 基本健康检查

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": 1620000000000,
  "version": "1.0.0"
}
```

##### 详细健康检查

```bash
curl http://localhost:8080/health/detailed
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": 1620000000000,
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "latency": 5
    },
    "mqtt": {
      "status": "healthy",
      "connected": true
    },
    "queue": {
      "status": "healthy",
      "length": 0
    }
  }
}
```

#### 指标端点

##### Prometheus 格式

```bash
curl http://localhost:8080/metrics
```

##### JSON 格式

```bash
curl http://localhost:8080/metrics/json
```

### 7.4 关键监控指标

#### 系统指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `system_cpu_usage` | Gauge | CPU 使用率 (%) |
| `system_memory_usage` | Gauge | 内存使用量 (bytes) |
| `system_disk_usage` | Gauge | 磁盘使用量 (bytes) |
| `system_disk_available` | Gauge | 磁盘可用空间 (bytes) |
| `system_uptime` | Counter | 系统运行时间 (seconds) |

#### HTTP 指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `http_requests_total` | Counter | HTTP 请求总数 |
| `http_request_duration_seconds` | Histogram | HTTP 请求耗时分布 |
| `http_requests_in_flight` | Gauge | 进行中的 HTTP 请求数 |
| `http_errors_total` | Counter | HTTP 错误总数 |

#### 数据库指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `db_operations_total` | Counter | 数据库操作总数 |
| `db_operation_duration_seconds` | Histogram | 数据库操作耗时 |
| `db_size_bytes` | Gauge | 数据库大小 |
| `db_compactions_total` | Counter | 数据库压缩次数 |

#### 队列指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `queue_length` | Gauge | 队列当前长度 |
| `queue_messages_total` | Counter | 队列消息总数 |
| `queue_processing_duration_seconds` | Histogram | 消息处理耗时 |
| `queue_dropped_total` | Counter | 丢弃的消息数 |

#### MQTT 指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `mqtt_connected` | Gauge | MQTT 连接状态 (1=连接, 0=断开) |
| `mqtt_messages_received_total` | Counter | 接收的 MQTT 消息总数 |
| `mqtt_messages_published_total` | Counter | 发布的 MQTT 消息总数 |
| `mqtt_connection_errors_total` | Counter | MQTT 连接错误数 |

#### 业务指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `readings_written_total` | Counter | 写入的读数总数 |
| `readings_query_total` | Counter | 查询请求总数 |
| `devices_active` | Gauge | 活跃设备数 |
| `alerts_active` | Gauge | 活跃告警数 |

### 7.5 Prometheus 集成

#### 配置 Prometheus

在 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'sfsedgestore'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

#### 验证配置

```bash
# 访问 Prometheus UI
# http://localhost:9090

# 查询指标
up{job="sfsedgestore"}
```

### 7.6 Grafana 可视化

#### 导入仪表盘

1. 打开 Grafana UI
2. 导航到 Dashboards → Import
3. 导入 sfsEdgeStore 仪表盘 JSON
4. 配置 Prometheus 数据源

#### 推荐仪表盘

1. **系统概览仪表盘**
   - 系统资源使用
   - 请求量和响应时间
   - 错误率

2. **数据库仪表盘**
   - 数据库操作统计
   - 存储使用趋势
   - 慢查询分析

3. **MQTT/队列仪表盘**
   - 消息吞吐量
   - 队列长度
   - 处理延迟

### 7.7 告警配置

#### 告警阈值配置

```json
{
  "EnableAlert": true,
  "AlertThresholds": {
    "cpu": 80,
    "memory": 85,
    "disk": 90,
    "mqtt_lag": 1000,
    "queue_length": 5000
  },
  "AlertNotifiers": [
    {
      "type": "webhook",
      "url": "https://api.example.com/alerts",
      "secret": "your-webhook-secret",
      "timeout": 10
    },
    {
      "type": "email",
      "smtp_server": "smtp.example.com",
      "smtp_port": 587,
      "smtp_username": "alerts@example.com",
      "smtp_password": "your-smtp-password",
      "from": "alerts@example.com",
      "to": ["admin@example.com"],
      "subject_prefix": "[sfsEdgeStore Alert]"
    }
  ]
}
```

#### 告警级别

| 级别 | 说明 | 响应时间 |
|------|------|----------|
| Critical | 系统不可用 | 立即响应 |
| Warning | 性能降级 | 1 小时内响应 |
| Info | 信息通知 | 24 小时内响应 |

#### 告警示例

```json
{
  "alert_id": "alert-001",
  "level": "critical",
  "title": "Disk Usage Critical",
  "message": "Disk usage has reached 95% on node sfsedgestore-01",
  "timestamp": "2026-03-07T10:00:00Z",
  "node": "sfsedgestore-01",
  "metric": "disk_usage_percent",
  "value": 95.0,
  "threshold": 90.0,
  "status": "firing",
  "duration": "5m"
}
```

#### 查看告警

```bash
# 获取所有活动告警
curl -H "X-API-Key: <api-key>" http://localhost:8080/alerts

# 获取告警历史
curl -H "X-API-Key: <api-key>" "http://localhost:8080/alerts/history?limit=100"

# 确认告警
curl -X POST -H "X-API-Key: <api-key>" \
  http://localhost:8080/alerts/alert-001/acknowledge

# 解决告警
curl -X POST -H "X-API-Key: <api-key>" \
  http://localhost:8080/alerts/alert-001/resolve
```

### 7.8 告警规则示例（Prometheus）

创建 `alerts.yml`：

```yaml
groups:
  - name: sfsedgestore_alerts
    interval: 30s
    rules:
      - alert: ServiceDown
        expr: up{job="sfsedgestore"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "sfsEdgeStore 服务不可用"
          description: "sfsEdgeStore 服务已停止超过 1 分钟"

      - alert: HighCPUUsage
        expr: system_cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU 使用率过高"
          description: "CPU 使用率已超过 80% 达 5 分钟"

      - alert: HighMemoryUsage
        expr: system_memory_usage > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "内存使用率过高"
          description: "内存使用率已超过 85%"

      - alert: HighQueueLength
        expr: queue_length > 1000
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "队列长度过长"
          description: "消息队列积压超过 1000 条"

      - alert: DiskSpaceLow
        expr: system_disk_available < 10737418240
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "磁盘空间不足"
          description: "可用磁盘空间少于 10GB"

      - alert: HighErrorRate
        expr: rate(http_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "错误率过高"
          description: "HTTP 错误率超过 10%"
```

---

## 第8章：认证与安全

### 8.1 认证授权

#### API Key 管理

##### 创建 API Key

首次启动时，需要创建管理员 API Key：

```bash
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 8760,
    "description": "System Administrator API Key"
  }'
```

响应：

```json
{
  "status": "success",
  "api_key": "sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c",
  "user_id": "admin",
  "role": "admin",
  "expires_at": "2027-03-07T10:00:00Z",
  "created_at": "2026-03-07T10:00:00Z"
}
```

##### 创建普通用户 API Key

```bash
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "X-API-Key: <admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "app1",
    "role": "user",
    "expires_in": 720,
    "description": "Application 1 API Key"
  }'
```

##### 创建只读用户 API Key

```bash
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "X-API-Key: <admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "monitoring",
    "role": "readonly",
    "expires_in": 0,
    "description": "Monitoring System API Key"
  }'
```

> 注意：`expires_in: 0` 表示永不过期（不推荐用于生产环境）

##### 列出所有 API Key

```bash
curl -H "X-API-Key: <admin-api-key>" \
  http://localhost:8080/api/auth/list-keys
```

响应示例：

```json
{
  "status": "success",
  "keys": [
    {
      "user_id": "admin",
      "role": "admin",
      "expires_at": "2027-03-07T10:00:00Z",
      "created_at": "2026-03-07T10:00:00Z",
      "last_used_at": "2026-03-07T12:00:00Z",
      "description": "System Administrator API Key"
    },
    {
      "user_id": "app1",
      "role": "user",
      "expires_at": "2026-04-07T10:00:00Z",
      "created_at": "2026-03-07T10:00:00Z",
      "last_used_at": null,
      "description": "Application 1 API Key"
    }
  ]
}
```

##### 撤销 API Key

```bash
curl -X POST http://localhost:8080/api/auth/revoke-key \
  -H "X-API-Key: <admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "sk_1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d"
  }'
```

##### 轮换 API Key

```bash
# 1. 创建新的 API Key
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "X-API-Key: <admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 8760,
    "description": "System Administrator API Key (New)"
  }'

# 2. 更新应用使用新 Key
# ... 更新配置 ...

# 3. 验证新 Key 工作正常
curl -H "X-API-Key: <new-api-key>" http://localhost:8080/health

# 4. 撤销旧 Key
curl -X POST http://localhost:8080/api/auth/revoke-key \
  -H "X-API-Key: <new-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "<old-api-key>"
  }'
```

#### 使用 API Key

```bash
curl -H "X-API-Key: sk_abc123def456..." \
  "http://localhost:8080/api/readings"
```

#### 角色和权限

##### 权限矩阵

| 权限 | admin | user | readonly |
|------|-------|------|----------|
| 读取数据 | ✅ | ✅ | ✅ |
| 写入数据 | ✅ | ✅ | ❌ |
| 删除数据 | ✅ | ✅ | ❌ |
| 查询数据 | ✅ | ✅ | ✅ |
| 创建备份 | ✅ | ✅ | ❌ |
| 恢复数据 | ✅ | ❌ | ❌ |
| 管理 API Key | ✅ | ❌ | ❌ |
| 查看监控 | ✅ | ✅ | ✅ |
| 管理告警 | ✅ | ❌ | ❌ |
| 修改配置 | ✅ | ❌ | ❌ |

### 8.2 安全基线

#### CIS 基准检查清单

- [ ] 使用专用用户运行
- [ ] 配置文件权限正确（600）
- [ ] 启用认证
- [ ] 使用 HTTPS
- [ ] 使用 MQTT TLS
- [ ] 数据库加密启用
- [ ] 定期轮换 API Key
- [ ] 日志集中化
- [ ] 定期备份
- [ ] 系统更新及时

### 8.3 网络安全

#### TLS 配置

##### 生成证书

```bash
# 生成 CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=sfsEdgeStore CA"

# 生成服务器证书
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=sfsedgestore.example.com"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

# 生成客户端证书
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=sfsedgestore-client"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365
```

##### 配置 HTTPS

```json
{
  "HTTPUseTLS": true,
  "HTTPCert": "/etc/sfsedgestore/certs/server.crt",
  "HTTPKey": "/etc/sfsedgestore/certs/server.key",
  "HTTPClientCA": "/etc/sfsedgestore/certs/ca.crt",
  "HTTPRequireClientCert": true
}
```

##### 配置 MQTT TLS

```json
{
  "MQTTUseTLS": true,
  "MQTTCACert": "ca.pem",
  "MQTTClientCert": "client.pem",
  "MQTTClientKey": "client.key"
}
```

### 8.4 API Key 安全最佳实践

1. **使用强密钥**
   - 长度至少 32 字符
   - 包含大小写字母、数字、符号
   - 不要使用可预测的模式

2. **密钥管理**
   ```bash
   # 使用密码管理器存储
   # 定期轮换（建议 90 天）
   # 立即撤销泄露的密钥
   ```

3. **最小权限原则**
   - 为每个应用使用独立的 Key
   - 只授予必要的权限
   - 使用只读 Key 用于监控

#### 强制认证配置

```json
{
  "EnableAuth": true,
  "AuthAPIKeyRequired": true,
  "AuthRateLimit": 1000,
  "AuthRateLimitWindow": 60,
  "AuthIPWhitelist": [
    "192.168.1.0/24",
    "10.0.0.0/8"
  ]
}
```

---

## 第9章：备份与恢复

### 9.1 备份策略

#### 备份类型

| 类型 | 说明 | 频率 | 保留期 |
|------|------|------|--------|
| 完整备份 | 完整数据库备份 | 每日 | 30 天 |
| 增量备份 | 变更数据备份 | 每小时 | 7 天 |
| 快照备份 | 时间点快照 | 每周 | 90 天 |

#### 3-2-1 备份原则

- **3 份数据副本**
- **2 种不同介质**
- **1 份异地备份**

### 9.2 手动备份

#### 使用 API 备份

```bash
# 创建备份
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=/var/lib/sfsedgestore/backups"

# 指定备份文件名
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=/var/lib/sfsedgestore/backups&name=backup-$(date +%Y%m%d_%H%M%S)"

# 压缩备份
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=/var/lib/sfsedgestore/backups&compress=true"
```

#### 文件系统备份

```bash
# 停止服务
sudo systemctl stop sfsedgestore

# 创建备份
sudo tar -czf \
  /var/lib/sfsedgestore/backups/sfsedgestore-$(date +%Y%m%d_%H%M%S).tar.gz \
  -C /var/lib/sfsedgestore data \
  -C /etc/sfsedgestore config.json

# 启动服务
sudo systemctl start sfsedgestore

# 验证备份
tar -tzf /var/lib/sfsedgestore/backups/sfsedgestore-*.tar.gz
```

### 9.3 自动备份

#### Cron 定时备份

```bash
# 编辑 crontab
sudo crontab -e -u sfsedgestore

# 添加备份任务
# 每小时增量备份
0 * * * * /opt/sfsedgestore/scripts/backup.sh incremental >> /var/log/sfsedgestore/backup.log 2>&1

# 每日完整备份
0 2 * * * /opt/sfsedgestore/scripts/backup.sh full >> /var/log/sfsedgestore/backup.log 2>&1

# 每周快照备份
0 3 * * 0 /opt/sfsedgestore/scripts/backup.sh snapshot >> /var/log/sfsedgestore/backup.log 2>&1

# 清理旧备份
0 4 * * * /opt/sfsedgestore/scripts/cleanup-backups.sh >> /var/log/sfsedgestore/backup.log 2>&1
```

#### 备份脚本示例

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/backup.sh

set -e

BACKUP_TYPE=${1:-full}
BACKUP_DIR="/var/lib/sfsedgestore/backups"
DATA_DIR="/var/lib/sfsedgestore/data"
CONFIG_DIR="/etc/sfsedgestore"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

backup_full() {
    log "Starting full backup..."
    
    BACKUP_FILE="$BACKUP_DIR/sfsedgestore-full-$DATE.tar.gz"
    
    # 使用 API 热备份
    curl -X POST \
      -H "X-API-Key: ${BACKUP_API_KEY}" \
      "http://localhost:8080/api/backup?path=$BACKUP_DIR&name=sfsedgestore-full-$DATE&compress=true"
    
    log "Full backup completed: $BACKUP_FILE"
}

backup_incremental() {
    log "Starting incremental backup..."
    
    # 查找上次完整备份
    LAST_FULL=$(find "$BACKUP_DIR" -name "sfsedgestore-full-*.tar.gz" -type f -mtime -1 | sort -r | head -n 1)
    
    if [ -z "$LAST_FULL" ]; then
        log "No recent full backup found, doing full backup instead"
        backup_full
        return
    fi
    
    # 增量备份逻辑
    # ...
    
    log "Incremental backup completed"
}

cleanup_old() {
    log "Cleaning up old backups..."
    
    find "$BACKUP_DIR" -name "sfsedgestore-full-*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete
    find "$BACKUP_DIR" -name "sfsedgestore-incr-*.tar.gz" -type f -mtime +7 -delete
    
    log "Old backups cleaned up"
}

case "$BACKUP_TYPE" in
    full)
        backup_full
        ;;
    incremental)
        backup_incremental
        ;;
    snapshot)
        # LVM 快照或文件系统快照
        ;;
    cleanup)
        cleanup_old
        ;;
    *)
        echo "Usage: $0 {full|incremental|snapshot|cleanup}"
        exit 1
        ;;
esac

log "Backup operation completed successfully"
```

### 9.4 数据恢复

#### 恢复前检查清单

- [ ] 确认备份文件完整性
- [ ] 停止相关服务
- [ ] 备份当前状态
- [ ] 通知相关人员
- [ ] 准备回滚方案
- [ ] 验证恢复环境

#### 使用 API 恢复

```bash
# 列出可用备份
curl -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup/list?path=/var/lib/sfsedgestore/backups"

# 恢复备份
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/restore?file=/var/lib/sfsedgestore/backups/backup-20260307_020000.db"
```

#### 文件系统恢复

```bash
# 1. 停止服务
sudo systemctl stop sfsedgestore

# 2. 备份当前数据（以防万一）
sudo mv /var/lib/sfsedgestore/data /var/lib/sfsedgestore/data.bak.$(date +%Y%m%d_%H%M%S)

# 3. 解压备份
sudo tar -xzf \
  /var/lib/sfsedgestore/backups/sfsedgestore-full-20260307_020000.tar.gz \
  -C /var/lib/sfsedgestore

# 4. 验证权限
sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore/data

# 5. 启动服务
sudo systemctl start sfsedgestore

# 6. 验证恢复
sudo systemctl status sfsedgestore
curl http://localhost:8080/health
```

### 9.5 备份验证

#### 自动验证脚本

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/verify-backup.sh

BACKUP_FILE="$1"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file>"
    exit 1
fi

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# 1. 检查文件存在
if [ ! -f "$BACKUP_FILE" ]; then
    log "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

# 2. 检查文件大小
FILE_SIZE=$(stat -f%z "$BACKUP_FILE" 2>/dev/null || stat -c%s "$BACKUP_FILE")
if [ "$FILE_SIZE" -lt 1024 ]; then
    log "ERROR: Backup file too small: $FILE_SIZE bytes"
    exit 1
fi

# 3. 验证文件完整性（如果有 checksum）
if [ -f "$BACKUP_FILE.sha256" ]; then
    cd "$(dirname "$BACKUP_FILE")"
    if ! sha256sum -c "$(basename "$BACKUP_FILE").sha256"; then
        log "ERROR: Checksum verification failed"
        exit 1
    fi
    log "Checksum verification passed"
fi

# 4. 测试恢复（使用临时目录）
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

log "Testing backup restoration in $TEST_DIR..."

# 解压到临时目录
if tar -tzf "$BACKUP_FILE" >/dev/null 2>&1; then
    tar -xzf "$BACKUP_FILE" -C "$TEST_DIR"
else
    log "ERROR: Invalid backup file format"
    exit 1
fi

# 检查关键文件
if [ ! -d "$TEST_DIR/data" ]; then
    log "ERROR: Data directory not found in backup"
    exit 1
fi

log "Backup verification passed: $BACKUP_FILE"
```

---

## 第四部分：企业篇

---

## 第10章：生产部署最佳实践

### 10.1 系统要求

#### 硬件要求

##### 最低配置

| 资源 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 1 核 | 2 核或以上 |
| 内存 | 64 MB | 128 MB 或以上 |
| 存储 | 500 MB | 1 GB 或以上 |
| 网络 | 10 Mbps | 100 Mbps 或以上 |

##### 生产环境推荐配置

| 资源 | 小型部署 | 中型部署 | 大型部署 |
|------|----------|----------|----------|
| CPU | 2 核 | 4 核 | 8 核 |
| 内存 | 128 MB | 256 MB | 512 MB |
| 存储 | 1 GB | 10 GB | 100 GB |
| 网络 | 100 Mbps | 1 Gbps | 1 Gbps |
| 并发连接 | 100 | 1000 | 10000 |

#### 软件要求

##### 操作系统

- **Linux** (推荐): Ubuntu 20.04+, Debian 11+, CentOS 8+, RHEL 8+
- **Windows**: Windows 10+, Windows Server 2019+
- **macOS**: macOS 11+ (仅用于开发和测试)

##### 依赖软件

- **MQTT Broker**: Eclipse Mosquitto 2.0+, EMQX 5.0+, 或其他兼容 MQTT 3.1.1/5.0 的 Broker
- **Go**: 1.25+ (仅编译时需要)
- **OpenSSL**: 1.1.1+ (如需 TLS 支持)

#### 网络要求

##### 端口使用

| 端口 | 协议 | 用途 | 必填 |
|------|------|------|------|
| 8080 | TCP | HTTP API 服务 | 是 |
| 1883 | TCP | MQTT (非 TLS) | 是 |
| 8883 | TCP | MQTT (TLS) | 否 |
| 8443 | TCP | HTTPS API 服务 | 否 |

### 10.2 部署管理

#### 部署前检查清单

- [ ] 硬件资源满足要求
- [ ] 操作系统已更新到最新版本
- [ ] 防火墙规则已配置
- [ ] MQTT Broker 已部署并运行
- [ ] 网络连接测试通过
- [ ] DNS 解析正常（如需要）
- [ ] NTP 时间同步已配置
- [ ] 备份策略已制定
- [ ] 监控方案已准备

#### 二进制部署（Linux 详细）

##### 1. 准备系统

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 创建专用用户
sudo useradd -r -s /bin/false -d /var/lib/sfsedgestore sfsedgestore

# 创建目录结构
sudo mkdir -p /opt/sfsedgestore
sudo mkdir -p /var/lib/sfsedgestore/data
sudo mkdir -p /var/lib/sfsedgestore/backups
sudo mkdir -p /var/log/sfsedgestore
sudo mkdir -p /etc/sfsedgestore

# 设置权限
sudo chown -R sfsedgestore:sfsedgestore /opt/sfsedgestore
sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore
sudo chown -R sfsedgestore:sfsedgestore /var/log/sfsedgestore
sudo chown -R sfsedgestore:sfsedgestore /etc/sfsedgestore
```

##### 2. 安装程序

```bash
# 下载或上传二进制文件
cd /tmp
wget https://github.com/your-org/sfsEdgeStore/releases/download/v1.0.0/sfsedgestore-linux-amd64
chmod +x sfsedgestore-linux-amd64
sudo mv sfsedgestore-linux-amd64 /opt/sfsedgestore/sfsedgestore

# 复制配置文件
sudo cp config.example.json /etc/sfsedgestore/config.json
sudo chown sfsedgestore:sfsedgestore /etc/sfsedgestore/config.json

# 编辑配置
sudo -u sfsedgestore vim /etc/sfsedgestore/config.json
```

##### 3. 创建 Systemd 服务

```ini
# /etc/systemd/system/sfsedgestore.service
[Unit]
Description=sfsEdgeStore Edge Data Storage Service
Documentation=https://docs.example.com/sfsedgestore
After=network.target mosquitto.service
Wants=network-online.target

[Service]
Type=simple
User=sfsedgestore
Group=sfsedgestore
WorkingDirectory=/var/lib/sfsedgestore
ExecStart=/opt/sfsedgestore/sfsedgestore -config /etc/sfsedgestore/config.json
Restart=always
RestartSec=10
StartLimitInterval=0

# 资源限制
MemoryMax=256M
CPUQuota=50%
TasksMax=100

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sfsedgestore /var/log/sfsedgestore
PrivateDevices=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
LockPersonality=true
MemoryDenyWriteExecute=true

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sfsedgestore

[Install]
WantedBy=multi-user.target
```

##### 4. 启用并启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable sfsedgestore

# 启动服务
sudo systemctl start sfsedgestore

# 检查状态
sudo systemctl status sfsedgestore

# 查看日志
sudo journalctl -u sfsedgestore -f -n 100
```

### 10.3 容器化部署

#### Docker Compose 高可用部署

```yaml
version: '3.8'

services:
  sfsedgestore:
    image: your-org/sfsedgestore:latest
    hostname: sfsedgestore-01
    networks:
      - sfsedgestore-net
    ports:
      - "8080:8080"
    volumes:
      - sfsedgestore-data:/app/data
      - ./config.json:/app/config.json:ro
      - ./logs:/app/logs
    environment:
      - EDGEX_LOG_LEVEL=info
      - EDGEX_ENABLE_MONITOR=true
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
      restart_policy:
        condition: on-failure
        delay: 10s
        max_attempts: 5
        window: 120s
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    depends_on:
      - mqtt-broker
      - redis

  mqtt-broker:
    image: eclipse-mosquitto:2.0
    networks:
      - sfsedgestore-net
    ports:
      - "1883:1883"
    volumes:
      - mosquitto-config:/mosquitto/config
      - mosquitto-data:/mosquitto/data
      - mosquitto-logs:/mosquitto/logs
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.25'
          memory: 128M

  redis:
    image: redis:7-alpine
    networks:
      - sfsedgestore-net
    volumes:
      - redis-data:/data
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '0.25'
          memory: 64M

  haproxy:
    image: haproxy:2.8-alpine
    networks:
      - sfsedgestore-net
    ports:
      - "80:8080"
      - "8404:8404"
    volumes:
      - ./haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.25'
          memory: 64M
    depends_on:
      - sfsedgestore

networks:
  sfsedgestore-net:
    driver: overlay

volumes:
  sfsedgestore-data:
  mosquitto-config:
  mosquitto-data:
  mosquitto-logs:
  redis-data:
```

### 10.4 升级和维护

#### 版本管理

##### 版本号规则

sfsEdgeStore 使用语义化版本（Semantic Versioning）：

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向下兼容的功能新增
- **PATCH**: 向下兼容的问题修复

##### 查看版本

```bash
# 查看程序版本
./sfsedgestore -version

# 通过 API 查看
curl http://localhost:8080/health | jq '.version'
```

#### 升级流程

##### 标准升级流程

```bash
# 1. 下载新版本
cd /tmp
wget https://github.com/your-org/sfsEdgeStore/releases/download/v1.1.0/sfsedgestore-linux-amd64
chmod +x sfsedgestore-linux-amd64

# 2. 验证新版本（可选）
./sfsedgestore-linux-amd64 -version
./sfsedgestore-linux-amd64 -config /etc/sfsedgestore/config.json -validate

# 3. 停止服务
sudo systemctl stop sfsedgestore

# 4. 备份旧版本
sudo cp /opt/sfsedgestore/sfsedgestore /opt/sfsedgestore/sfsedgestore.old

# 5. 安装新版本
sudo mv sfsedgestore-linux-amd64 /opt/sfsedgestore/sfsedgestore
sudo chown sfsedgestore:sfsedgestore /opt/sfsedgestore/sfsedgestore
sudo chmod +x /opt/sfsedgestore/sfsedgestore

# 6. 检查配置兼容性
# 如有配置变更，更新配置文件

# 7. 启动服务
sudo systemctl start sfsedgestore

# 8. 验证升级
sudo systemctl status sfsedgestore
sudo journalctl -u sfsedgestore -n 100 --no-pager
curl http://localhost:8080/health

# 9. 清理旧版本（确认一切正常后）
sudo rm /opt/sfsedgestore/sfsedgestore.old
```

### 10.5 性能优化

#### 数据库优化

```json
{
  "DBCompression": true,
  "DBMaxOpenFiles": 2000
}
```

#### MQTT 优化

```json
{
  "MQTTQoS": 1,
  "QueueSize": 50000
}
```

#### HTTP 优化

```json
{
  "HTTPMaxConnections": 200,
  "HTTPReadTimeout": 60
}
```

### 10.6 安全最佳实践

1. **启用认证**

```json
{
  "EnableAuth": true,
  "AuthAPIKeyRequired": true
}
```

2. **使用 TLS**

```json
{
  "HTTPUseTLS": true,
  "MQTTUseTLS": true
}
```

3. **API Key 管理**
   - 定期轮换 API Key
   - 使用不同的 Key 给不同的应用
   - 设置合理的过期时间
   - 及时撤销不再使用的 Key

### 10.7 监控和维护

1. **关键指标监控**
   - CPU 使用率
   - 内存使用率
   - 磁盘使用率
   - MQTT 消息处理速率
   - 数据库大小

2. **日志管理**

```json
{
  "LogLevel": "info",
  "LogFormat": "json",
  "LogOutput": "/var/log/sfsedgestore/app.log"
}
```

3. **定期维护**
   - 每日检查日志
   - 每周检查备份
   - 每月检查磁盘空间
   - 每季度进行性能评估

---

## 第11章：商业服务与支持

### 11.1 免费试用计划

为了让您在投入前充分体验我们的专业服务，我们提供以下免费试用方案。通过全方位的体验，让您深度了解我们的专业服务能力。

#### sfsEdgeStore 企业级支持体验包

| 服务 | 试用时长 | 说明 |
|------|---------|------|
| **标准技术支持** | 30 天 | 享受与付费客户相同的工单提交渠道，8 小时内响应，获得配置优化建议和故障排查指导 |
| **基础托管服务** | 14 天（可选） | 为您部署一个 sfsEdgeStore 实例，提供基础监控和告警设置，一次远程健康检查报告 |
| **快速启动咨询** | 1 次（1 小时） | 免费线上咨询，包括初步需求分析、架构建议或疑难解答 |

#### 各服务试用详情

##### 技术支持服务免费试用

这是最核心、最应该提供免费试用的服务。

**试用时长**：30 天标准支持服务免费试用

**试用内容**：
- 享受与付费客户相同的工单提交渠道
- 体验 8 小时内响应的专业服务
- 获得基本的配置优化建议和故障排查指导

##### 托管运维服务免费试用

这个服务的试用可以作为"技术支持"的升级版。

**试用时长**：14 天基础托管服务试用

**试用内容**：
- 为您部署一个 sfsEdgeStore 实例（您需提供服务器）
- 提供基础的监控和告警设置
- 提供一次远程健康检查报告

##### 实施咨询服务免费试用

这类服务通常价值高、成本高，不适合长时间免费。

**试用内容**：一次免费的、时长 1 小时的线上咨询（或"快速启动咨询"）

#### 试用期管理

**明确条款**：在提供试用前，需要您签署一份简单的试用协议，明确试用范围、期限和保密条款。

**自动提醒**：在试用期结束前 3-5 天，自动发送邮件提醒您试用即将到期，并引导购买。

**数据追踪**：追踪试用客户的使用情况和问题，这些数据能帮助我们优化产品和服务，也能为销售提供有价值的信息。

### 11.2 商业服务说明

虽然 sfsEdgeStore 软件本身 100% 开源免费，但我们提供专业的商业服务来帮助您成功落地和高效使用产品。

#### 核心价值

- ✅ **降低风险**：专业团队支持，减少实施风险
- ✅ **节省时间**：快速上手，避免踩坑
- ✅ **获得保障**：Bug 优先修复，SLA 保障
- ✅ **持续优化**：性能调优，架构优化

#### 技术支持服务

| 服务等级 | 响应时间 | 服务时间 | 价格 | 适用场景 |
|---------|---------|---------|------|---------|
| **基础支持** | 24 小时 | 工作日 9:00-18:00 | ¥5,000/年 | 小型项目、测试环境 |
| **标准支持** | 8 小时 | 7x12 小时 | ¥20,000/年 | 中型项目、生产环境 |
| **企业支持** | 2 小时（严重） | 7x24 小时 | ¥150,000/年 | 大型项目、关键业务 |

#### 服务套餐

为了满足不同客户的需求，我们提供以下组合套餐：

**成长套餐**
- **价格**：¥20,000/年
- **包含内容**：基础技术支持、快速启动咨询、基础培训等
- **适用客户**：小团队、初创企业

**商业套餐**
- **价格**：¥60,000/年
- **包含内容**：标准技术支持、标准实施咨询、进阶培训等
- **适用客户**：中小企业、中等规模项目

**企业套餐**
- **价格**：¥200,000/年
- **包含内容**：企业技术支持、企业实施咨询、认证培训、专属顾问等
- **适用客户**：中大型企业、关键业务项目

### 11.3 如何购买

#### 购买方式

1. **在线购买**：访问官网，选择所需服务，在线支付
2. **联系我们**：发送邮件至 sfsweb@qq.com，我们的销售团队会联系您
3. **云市场**：在各大云市场（阿里云、腾讯云、AWS 等）购买

#### 支付方式

- ✅ 支付宝
- ✅ 微信支付
- ✅ 银行转账
- ✅ 信用卡（Visa、MasterCard）

---

## 第12章：成功案例

### 12.1 汽车零部件制造厂

#### 客户背景

某汽车零部件制造厂，拥有 5 条产线，100+ 台设备，之前使用的是传统的云端数据存储方案。

#### 遇到的问题

1. **网络中断时数据丢失**：工厂网络不稳定，每月平均有 3-5 次网络中断，每次中断都会丢失数据
2. **查询响应慢**：查询产线数据需要 3-5 秒，无法满足实时监控需求
3. **部署复杂**：每次新增产线都需要 1-2 周的部署时间
4. **成本高**：重型数据库 license 费用昂贵

#### 解决方案

部署 sfsEdgeStore：
- 每条产线部署 1 个 sfsEdgeStore 节点
- 本地存储数据，网络中断不影响采集
- 5 分钟完成每条产线的部署
- 与现有的 EdgeX Foundry 系统无缝集成

#### 取得的成果

| 指标 | 之前 | sfsEdgeStore | 改善 |
|------|------|-------------|------|
| **数据丢失率** | 每月 3-5 次 | 0 次 | 100% 消除 |
| **查询响应时间** | 3-5 秒 | 50-100 毫秒 | 30-50 倍提升 |
| **部署时间** | 1-2 周/产线 | 5 分钟/产线 | 200-400 倍提升 |
| **年度成本** | ¥500,000+ | ¥0（软件免费） | 成本降低 100% |

### 12.2 物流园区监控系统

（内容待补充）

### 12.3 能源监控系统

（内容待补充）

---

## 附录

### 附录 A：API 参考

（内容待补充）

### 附录 B：配置文件完整示例

（内容待补充）

### 附录 C：故障排除指南

（内容待补充）

---

## 联系我们

- **官网**: https://www.sfsedgestore.com
- **GitHub**: [https://github.com/liaoran123/sfsEdgeStore](https://github.com/liaoran123/sfsEdgeStore)
- **邮箱**: sfsweb@qq.com

---

**本书版本**：1.0.0  
**最后更新**：2026-03-08  
**sfsEdgeStore** - 让边缘数据存储更简单！🚀
