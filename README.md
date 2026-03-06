# sfsEdgeStore

轻量级边缘计算数据存储适配器 - EdgeX Foundry 与 sfsDb 之间的桥梁

## 📋 简介

**sfsEdgeStore** 是典型的"小前端"应用，专为边缘计算场景设计。它作为 EdgeX Foundry 和 sfsDb 数据库之间的桥梁，提供高效的本地数据读写和缓存能力。

### 🎯 核心特征

| 特性 | 说明 |
|------|------|
| **部署位置** | 直接部署在资源受限的边缘设备上，与 EdgeX Foundry 共同运行 |
| **资源占用** | 轻量级设计，内存 < 50MB，CPU < 5% |
| **核心功能** | EdgeX Foundry 和 sfsDb 之间的数据桥梁 |
| **独立运行** | 可独立于中心系统运行，网络中断不影响本地数据采集 |
| **本地化处理** | 数据存储在本地 sfsDb，实现边缘数据处理 |

## 🏗️ 架构

### "大后台 + 小前端" 模式

```
┌─────────────────────────────────────────────────────────────┐
│                        大后台                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ 中央管理平台  │  │ 数据分析中心  │  │ 告警运维中心  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ 全局管理 / 数据汇总
                              │
┌─────────────────────────────────────────────────────────────┐
│                        边缘节点                              │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  sfsEdgeStore (小前端)                                │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  │ │
│  │  │  MQTT 客户端  │→ │  数据队列    │→ │  sfsDb   │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────┘  │ │
│  │  ┌──────────────┐  ┌──────────────┐                  │ │
│  │  │  HTTP 服务   │  │  监控告警    │                  │ │
│  │  └──────────────┘  └──────────────┘                  │ │
│  └───────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              EdgeX Foundry                              │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## ✨ 功能特性

- 📡 **MQTT 数据接入**：订阅 EdgeX Foundry 事件主题
- 💾 **本地数据存储**：使用 sfsDb 高效存储边缘数据
- 📊 **实时监控**：内置系统指标和业务指标监控
- ⚠️ **智能告警**：阈值告警和异常检测
- 🔄 **数据队列**：断电恢复和数据重试机制
- 📈 **数据分析**：内置时间窗口聚合和预测
- 🔐 **认证授权**：API Key 和 RBAC 权限控制
- 🌐 **HTTP API**：RESTful 接口供外部查询
- 🔄 **数据同步**：可选的数据上云同步
- 🗑️ **数据保留**：自动清理过期数据

## 🚀 快速开始

### 前置条件

- Go 1.25+
- EdgeX Foundry (可选，用于数据源)
- MQTT Broker (如 Mosquitto)

### 安装

```bash
# 克隆仓库
git clone https://github.com/your-org/sfsEdgeStore.git
cd sfsEdgeStore

# 安装依赖
go mod download

# 编译
go build -o sfsedgestore
```

### 配置

复制配置示例文件：

```bash
cp config.example.json config.json
```

编辑 `config.json`：

```json
{
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTClientID": "sfsedgestore",
  "MQTTTopic": "edgex/events/core/#",
  "DBPath": "./data",
  "DBUseEncryption": false,
  "HTTPPort": 8080
}
```

### 运行

```bash
# 直接运行
./sfsedgestore

# 或使用 Go 运行
go run main.go
```

## 📡 API 接口

### 健康检查

```bash
curl http://localhost:8080/health
```

### 获取指标

```bash
curl http://localhost:8080/metrics
```

### 查询数据

```bash
# 查询设备数据
curl "http://localhost:8080/query?deviceName=Device001"

# 按时间范围查询
curl "http://localhost:8080/query?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-12-31T23:59:59Z"
```

### 查看告警

```bash
curl http://localhost:8080/alerts
```

## 📦 部署

### 二进制部署

1. 从 [GitHub Releases](https://github.com/your-org/sfsEdgeStore/releases) 下载对应平台的二进制文件
2. 配置 `config.json`
3. 运行二进制文件

### Docker 部署

```bash
docker pull your-org/sfsdb-edgex-adapter:latest
docker run -d \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  your-org/sfsdb-edgex-adapter:latest
```

### Systemd 服务

创建 `/etc/systemd/system/sfsedgestore.service`：

```ini
[Unit]
Description=sfsEdgeStore Edge Data Adapter
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/sfsedgestore
ExecStart=/opt/sfsedgestore/sfsedgestore
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable sfsedgestore
sudo systemctl start sfsedgestore
```

## 🔧 开发

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行带竞争检测
go test -v -race ./...

# 只运行数据库测试
go test -v ./database/...
```

### 项目结构

```
sfsEdgeStore/
├── main.go                 # 主程序入口
├── config/                 # 配置管理
├── database/               # 数据库操作
├── mqtt/                   # MQTT 客户端
├── server/                 # HTTP 服务器
├── monitor/                # 监控指标
├── alert/                  # 告警管理
├── queue/                  # 数据队列
├── auth/                   # 认证授权
├── agent/                  # 管理 Agent
├── analyzer/               # 数据分析
├── sync/                   # 数据同步
├── retention/              # 数据保留
├── resource/               # 资源监控
└── word/                   # 文档
```

## 📊 性能指标

| 指标 | 目标值 |
|------|--------|
| 内存占用 | < 50 MB |
| CPU 使用率 | < 5% |
| 启动时间 | < 2 秒 |
| 数据写入延迟 | < 10 ms |
| 支持并发 | 1000+ QPS |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

Apache License 2.0

## 🙏 致谢

- [EdgeX Foundry](https://www.edgexfoundry.org/)
- [sfsDb](https://github.com/liaoran123/sfsDb)
- [Eclipse Paho MQTT](https://www.eclipse.org/paho/)

---

**sfsEdgeStore** - 让边缘数据存储更简单！🚀
