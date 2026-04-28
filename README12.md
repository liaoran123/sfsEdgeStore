[中文](./README.md) | [English](./README_en.md)

# sfsEdgeStore

轻量级工业物联网边缘数据存储适配器，专为 EdgeX Foundry 设计。

> 本项目由 **sfsDb 官方团队** 精心打造，作为 sfsDb 数据库的官方边缘计算适配器。

## ⚡ 性能表现

### 生产环境（100 设备，1-60 秒间隔）

| 指标       | 数值           | 说明         |
| -------- | ------------ | ---------- |
| **内存**   | \~30 MB      | 正常运行时稳定占用  |
| **CPU**  | 1.7%         | 极低开销，超低功耗  |
| **消息速率** | \~30 msg/sec | 真实工业传感器数据  |
| **启动**   | <0.2 秒       | 毫秒级启动，即刻可用 |

### 高压测试（500 设备，0.05-0.2 秒间隔）

| 指标        | 数值             | 说明                 |
| --------- | -------------- | ------------------ |
| **内存**    | \~44 MB        | 重负载下（4000 msg/sec） |
| **CPU**   | 6.8%           | 从容应对极端负载           |
| **消息速率**  | \~4000 msg/sec | 133 倍于正常生产负载       |
| **零数据丢失** | 100%           | 所有消息全部存储成功         |

### 资源对比

| 场景   | 设备数 | 间隔         | 速率             | CPU  | 内存      |
| ---- | --- | ---------- | -------------- | ---- | ------- |
| 生产环境 | 100 | 1-60 秒     | \~30 msg/sec   | 1.7% | \~30 MB |
| 高压测试 | 500 | 0.05-0.2 秒 | \~4000 msg/sec | 6.8% | \~44 MB |

![监控仪表盘](./img/sfsEdgeStoreCn.png)

## 🎯 解决的核心问题

### 边缘计算挑战

| 挑战                 | 解决方案                   |
| ------------------ | ---------------------- |
| 边缘设备资源有限           | 内存 <30MB，CPU <7%，超轻量设计 |
| 网络中断时数据丢失          | 本地存储，断网正常运行            |
| 重型数据库部署复杂          | 5 分钟部署，零配置             |
| EdgeX Foundry 存储缺失 | 原生 EdgeX 集成，开箱即用       |
| 数据查询响应慢            | 本地 LevelDB，毫秒级查询       |
| 依赖云端               | 独立运行，无需中心系统            |

## 📋 产品概述

**sfsEdgeStore** 是一款专为工业物联网边缘场景设计的轻量级数据存储适配器。它作为 EdgeX Foundry 与 sfsDb 之间的桥梁，提供高效的本地数据读写和缓存能力。

### 为什么 EdgeX 需要 sfsEdgeStore

> EdgeX 是最好的连接框架，但它默认不持久化数据。不要用 Redis（断电丢数据），不要用 InfluxDB（边缘设备扛不住）。sfsEdgeStore 是 EdgeX 的原生存储插件，专为资源受限的边缘网关设计。

## ✨ 核心功能

- 📡 **MQTT 数据采集**：订阅 EdgeX Foundry 事件主题
- 💾 **本地数据存储**：sfsDb/LevelDB 高效边缘数据存储
- 🔒 **数据加密**：AES-256 静态数据加密
- 🔄 **可靠队列**：断电恢复、数据重试机制
- 📊 **实时监控**：内置系统和业务指标监控
- ⚠️ **智能告警**：阈值告警、异常检测
- 🗑️ **数据保留**：自动清理过期数据
- 🔐 **身份认证**：API Key 和 RBAC 访问控制
- 🌐 **HTTP API**：RESTful 接口支持外部查询
- 📦 **备份恢复**：自动化备份与恢复

## 🚀 快速开始

### 前置要求

- Go 1.21+（源码编译）
- EdgeX Foundry（可选，作为数据源）
- MQTT Broker（如 Mosquitto）

### 方式一：二进制部署（推荐）

```bash
# 从 GitHub Releases 下载
# https://github.com/liaoran123/sfsEdgeStore/releases

# 直接运行（测试用）
./sfsedgestore

# 生产环境请使用 systemd（Linux）或 Windows Service
```

### 方式二：Docker Compose（推荐）

```bash
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore
docker-compose up -d
```

同时启动 sfsEdgeStore + MQTT Broker。访问仪表盘：`http://localhost:8081`。

### 方式三：Docker

```bash
docker run -d \
  --name sfsedgestore \
  -p 8081:8081 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  liaoran123/sfsedgestore:latest
```

### 方式四：源码编译

```bash
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore
go build -ldflags="-s -w" -o sfsedgestore .
./sfsedgestore
```

### 验证安装

```bash
# 健康检查
curl http://localhost:8081/health

# 浏览器打开仪表盘
# http://localhost:8081
```

### 零配置

sfsEdgeStore 使用智能默认值，无需任何配置即可启动：

| 设置          | 默认值                    |
| ----------- | ---------------------- |
| MQTT Broker | `tcp://localhost:1883` |
| MQTT Topic  | `edgex/events/#`       |
| HTTP 端口     | `8081`                 |
| 数据库路径       | `data`                 |

### 配置示例

创建 `config.json` 自定义设置：

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "http_port": "8081",
  "db_path": "data",
  "db_scenario": "edge",
  "enable_resource_monitoring": true,
  "max_memory_mb": 256,
  "enable_retention_policy": true,
  "retention_days": 30
}
```

| 配置项                       | 说明                                            |
| ------------------------- | --------------------------------------------- |
| `mqtt_broker`             | MQTT 服务器地址（如 Mosquitto）                       |
| `mqtt_topic`              | EdgeX 事件主题模式                                  |
| `db_path`                 | 本地数据库存储路径                                     |
| `db_scenario`             | 性能配置：`embedded`、`iot`、`edge`、`game`、`default` |
| `enable_retention_policy` | 自动清理旧数据                                       |
| `retention_days`          | 数据保留天数                                        |

> MQTT 主题通过 Web 仪表盘的**主题订阅**页面（`http://localhost:8081/mqtt-subscription`）管理。

## 🏗️ 架构

### 轻量级边缘架构

```
┌─────────────────────────────────────────────────────────────┐
│                      边缘节点（资源受限）                      │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              EdgeX Foundry                              │ │
│  │  （数据采集、设备管理）                                  │ │
│  └────────────────────┬──────────────────────────────────┘ │
│                       │ MQTT                              │
│                       ▼                                   │
│  ┌───────────────────────────────────────────────────────┐ │
│  │         sfsEdgeStore（轻量级适配器）                     │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  │ │
│  │  │  MQTT 客户端  │→ │  数据队列    │→ │  sfsDb   │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────┘  │ │
│  │  ┌──────────────┐  ┌──────────────┐                  │ │
│  │  │  HTTP 服务    │  │  监控模块    │                  │ │
│  │  └──────────────┘  └──────────────┘                  │ │
│  └───────────────────────────────────────────────────────┘ │
│                       │ HTTP API                         │
└───────────────────────┼─────────────────────────────────────┘
                        ▼
                  外部查询/监控
```

### 设计原则

- **小而美**：专注一件事，做到极致
- **数据主权**：数据留在本地，不依赖云端
- **边缘优先**：所有功能优先考虑边缘场景
- **零依赖**：仅依赖 sfsDb，无重型组件
- **高可用**：断电恢复、数据重试、本地存储

## 📚 文档

完整文档位于 [docs](./docs/) 目录：

| 文档                                | 说明          |
| --------------------------------- | ----------- |
| [快速开始](./docs/quick-start.md)     | 5 分钟入门指南    |
| [安装部署](./docs/installation.md)    | 安装和部署指南     |
| [API 参考](./docs/api-reference.md) | REST API 文档 |
| [配置说明](./docs/configuration.md)   | 所有配置项说明     |
| [架构设计](./docs/architecture.md)    | 系统架构概述      |
| [安全指南](./docs/security.md)        | 认证、TLS、加密   |
| [故障排查](./docs/troubleshooting.md) | 常见问题和解决方案   |

📖 [查看文档索引](./docs/README.md)

## 💼 合作与投资

> **我们卖的是解决方案，不是软件。**

**sfsEdgeStore** 定位为工业物联网边缘数据解决方案——提供完整的硬件+软件+部署服务包，而非单一产品。

### 🤝 寻求合作 / 融资

我们正在积极寻找战略合作伙伴和投资人，共同加速在工业物联网市场的布局。

**投资人关心的核心指标**

| 投资人关心的点  | 传统工业软件/方案      | sfsEdgeStore（我们的方案） | 我们的融资故事                  |
| :------- | :------------- | :------------------ | :----------------------- |
| **硬件成本** | 需昂贵工控机（500 元+） | 普通 ARM 网关（200 元）    | *"我们帮客户省下了 60% 的硬件预算。"*  |
| **资源占用** | 臃肿，吃内存         | **27.6MB** 极致轻量     | *"老旧设备也能跑，市场存量改造空间巨大。"*  |
| **部署难度** | 需专业实施团队        | **5 分钟**开箱即用        | *"可以像 SaaS 一样快速规模化复制。"*  |
| **技术壁垒** | 依赖外部库，重        | **纯 Go**，零依赖        | *"极高的工程效率，一个人顶一个团队。"*    |
| **可扩展性** | 受限于基础设施        | 无状态，水平扩展            | *"为下一个 10,000 个边缘节点而生。"* |

### 💡 我们能提供什么

| 服务       | 说明               |
| -------- | ---------------- |
| **方案咨询** | 评估边缘计算需求，设计最优架构  |
| **定制开发** | 定制协议、集成、行业专属功能   |
| **技术支持** | 优先邮件支持、部署协助、故障排查 |
| **培训赋能** | 现场或远程技术培训        |

**联系：** <liao010203kk@gmail.com>

***

<details>
<summary><strong>English Version</strong></summary>

## About This Project

sfsEdgeStore is a lightweight Industrial IoT Edge Data Storage Adapter designed for EdgeX Foundry. Built by the official sfsDb team, it provides efficient local data storage with minimal resource footprint.

### Key Highlights

| Metric    | Value                               |
| --------- | ----------------------------------- |
| Memory    | \~30 MB (production)                |
| CPU       | 1.7% (normal load)                  |
| Startup   | <0.2 seconds                        |
| Data Loss | 0% (even under 4000 msg/sec stress) |

### Contact

For partnership inquiries: <liao010203kk@gmail.com>

</details>

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解贡献指南。

## 🔒 安全

请参阅 [SECURITY.md](./SECURITY.md) 了解安全政策和漏洞报告。

## 📄 许可证

Apache License 2.0

## 🙏 致谢

- [EdgeX Foundry](https://www.edgexfoundry.org/)
- [sfsDb](https://github.com/liaoran123/sfsDb)
- [Eclipse Paho MQTT](https://www.eclipse.org/paho/)

