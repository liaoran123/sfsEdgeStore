# sfsEdgeStore

轻量级工业物联网边缘数据存储适配器，专为 EdgeX Foundry 设计。

> 本项目由 **sfsDb 官方团队** 精心打造，作为 sfsDb 数据库的官方边缘计算适配器。

### 性能表现

#### 生产环境（100设备）
- **内存**：~30MB（稳定运行）
- **CPU**：1.7%（超低功耗）
- **消息速率**：~30条/秒（真实工业数据）
- **启动时间**：<0.2秒（毫秒级启动）

#### 高压测试（500设备）
- **内存**：~44MB（4000条/秒负载）
- **CPU**：6.8%（应对极端负载）
- **消息速率**：~4000条/秒（133倍生产负载）
- **数据可靠性**：100%无丢失

![监控仪表盘](./img/sfsEdgeStoreCn.png)

## 🎯 解决的核心问题

### 边缘计算挑战

| 挑战                 | 解决方案                   |
| ------------------ | ---------------------- |
| 边缘设备资源有限           | 内存 <15MB，CPU <5%，超轻量设计 |
| 网络中断时数据丢失          | 本地存储，断网正常运行            |
| 重型数据库部署复杂          | 5 分钟部署，零配置             |
| EdgeX Foundry 存储缺失 | 原生 EdgeX 集成，开箱即用       |
| 数据查询响应慢            | 本地 LevelDB，毫秒级查询       |
| 依赖云端               | 独立运行，无需中心系统            |

## 📋 产品简介

**sfsEdgeStore** 是专为工业物联网边缘场景设计的轻量级数据存储适配器，作为 EdgeX Foundry 和 sfsDb 之间的桥梁，提供高效的本地数据读写和缓存能力。

### 为什么 EdgeX 需要 sfsEdgeStore

> EdgeX 是最好的连接框架，但它默认不存数据。别用 Redis（断电丢数据），别用 InfluxDB（边缘太卡）。sfsEdgeStore 是 EdgeX 的原生存储插件，专为资源受限的边缘网关设计。

## ✨ 核心功能

- 📡 **MQTT 数据接入**：订阅 EdgeX Foundry 事件主题
- 💾 **本地数据存储**：sfsDb/LevelDB 高效边缘存储
- 🔒 **数据加密**：AES-256 静态数据加密
- 🔄 **可靠队列**：断电恢复与数据重试机制
- 📊 **实时监控**：内置系统与业务指标监控
- ⚠️ **智能告警**：阈值告警与异常检测
- 🗑️ **数据保留**：自动清理过期数据
- 🔐 **认证授权**：API Key 与 RBAC 访问控制
- 🌐 **HTTP API**：RESTful 接口供外部查询
- 📦 **备份恢复**：自动备份与恢复功能

## 🚀 快速开始

### 前置条件

- Go 1.21+（源码编译需要）
- EdgeX Foundry（可选，数据源）
- MQTT Broker（如 Mosquitto）

### 方式一：二进制部署（推荐）

```bash
# 从 GitHub Releases 下载
# https://github.com/liaoran123/sfsEdgeStore/releases

# 直接运行（测试用）
./sfsedgestore

# 生产环境使用 systemd（Linux）或 Windows 服务
```

### 方式二：Docker Compose（推荐）

```bash
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore
docker-compose up -d
```

同时启动 sfsEdgeStore + MQTT Broker。访问仪表板 `http://localhost:8081`。

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

# 在浏览器中打开仪表板
# http://localhost:8081
```

### 零配置启动

sfsEdgeStore 使用智能默认值，无需配置即可启动：

| 配置项         | 默认值                    |
| ----------- | ---------------------- |
| MQTT Broker | `tcp://localhost:1883` |
| MQTT 主题     | `edgex/events/#`       |
| HTTP 端口     | `8081`                 |
| 数据库路径       | `data`                 |

### 配置示例

创建 `config.json` 自定义配置：

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
| `enable_retention_policy` | 自动清理过期数据                                      |
| `retention_days`          | 数据保留天数                                        |

> MQTT 主题通过 Web 仪表板的 **主题订阅** 页面（`http://localhost:8081/mqtt-subscription`）管理。

## 🏗️ 架构设计

### 轻量级边缘架构

```
┌─────────────────────────────────────────────────────────────┐
│                      边缘节点（资源受限）                     │
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
│  │  │  HTTP 服务   │  │  监控告警    │                  │ │
│  │  └──────────────┘  └──────────────┘                  │ │
│  └───────────────────────────────────────────────────────┘ │
│                       │ HTTP API                         │
└───────────────────────┼─────────────────────────────────────┘
                        ▼
                  外部查询/监控
```

### 设计原则

- **小而美**：只做一件事，做到极致
- **数据主权**：数据留在本地，不依赖云端
- **边缘优先**：所有功能优先考虑边缘场景
- **零依赖**：仅依赖 sfsDb，无重型组件
- **高可用**：断电恢复、数据重试、本地存储

## 📚 文档

完整文档请查看 [docs](./docs/) 目录：

| 文档                                | 说明            |
| --------------------------------- | ------------- |
| [快速开始](./docs/quick-start.md)     | 5 分钟上手        |
| [安装指南](./docs/installation.md)    | 安装与部署         |
| [API 参考](./docs/api-reference.md) | REST API 文档   |
| [配置参考](./docscn/配置参考.md)          | 所有配置项（中文）     |
| [系统架构](./docs/architecture.md)    | 架构概览          |
| [安全指南](./docscn/安全指南.md)          | 认证、TLS、加密（中文） |
| [故障排查](./docscn/故障排查.md)          | 常见问题（中文）      |

#### 合作与投资

**sfsEdgeStore** 定位为工业物联网（IIoT）边缘数据解决方案提供商。我们致力于通过技术创新，解决工业现场数据采集与处理的成本与效率瓶颈。

#### 核心价值主张

传统工业软件方案普遍依赖昂贵的工控机硬件，导致项目部署成本高昂，难以在资源受限的边缘侧大规模推广。

**sfsEdgeStore 以 27.6MB 的极致轻量级架构，成功将完整的数据处理能力下沉至成本仅 200 元的 ARM 网关。** 这不仅是一款软件产品，更是对工业网关硬件标准的一次重新定义。

#### 方案对比与优势

下表清晰地展示了 sfsEdgeStore 相较于传统方案的颠覆性优势：

| 维度       | 传统工业软件/方案      | sfsEdgeStore    | 为客户创造的价值                      |
| :------- | :------------- | :-------------- | :---------------------------- |
| **硬件成本** | 依赖昂贵工控机（500元+） | 适配普通ARM网关（200元） | **节省60%硬件预算**，大幅降低项目初始投入。     |
| **资源占用** | 架构臃肿，内存占用高     | **27.6MB** 极致轻量 | 可在老旧设备上运行，激活巨大的**存量市场改造潜力**。  |
| **部署难度** | 需专业团队现场实施      | **5分钟**开箱即用     | 部署流程极简，具备**SaaS化的快速规模化复制能力**。 |
| **技术壁垒** | 依赖外部库，架构沉重     | **纯Go编写**，零外部依赖 | 极高的工程效率与稳定性，**一人团队即可支撑核心研发**。 |
| **可扩展性** | 受限于中心服务器性能     | 无状态设计，支持水平扩展    | 架构灵活，为支撑**未来数万个边缘节点**的部署而生。   |

### 💡 我们能提供什么

| 服务       | 说明               |
| -------- | ---------------- |
| **方案咨询** | 评估边缘计算需求，设计最优架构  |
| **定制开发** | 定制协议、集成、行业专属功能   |
| **技术支持** | 优先邮件支持、部署协助、故障排查 |
| **培训赋能** | 现场或远程技术培训        |

**联系：** <sfsweb@qq.com>

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解贡献指南。

## 🔒 安全

请查看 [SECURITY.md](./SECURITY.md) 了解安全策略和漏洞报告方式。

## 📄 许可证

Apache License 2.0

## 🙏 致谢

- [EdgeX Foundry](https://www.edgexfoundry.org/)
- [sfsDb](https://github.com/liaoran123/sfsDb)
- [Eclipse Paho MQTT](https://www.eclipse.org/paho/)

