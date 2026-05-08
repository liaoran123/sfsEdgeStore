# 服务器包 (Server Package) 技术文档

## 1. 概述

服务器包是 sfsEdgeStore 系统的核心组件，负责提供 HTTP API 接口和 WebSocket 实时通信功能，为前端 Web 界面和第三方系统提供统一的服务。

### 1.1 主要功能

- **HTTP API 服务**：提供 RESTful API，支持数据查询、配置管理、设备监控等
- **WebSocket 通信**：支持实时数据推送
- **静态文件服务**：提供 Web 界面所需的 HTML、CSS、JavaScript 等资源
- **广播机制**：将 MQTT 消息转发至 WebSocket 客户端

## 2. 文件结构

```
server/
├── server.go          # Server 结构体、构造函数、启动
├── routes.go          # 路由注册 + DeviceNameMiddleware
├── broadcast.go       # MQTT 广播监听
├── ws_manager.go      # WebSocket 连接管理器
├── handlers_data.go   # 数据：查询、设备状态、备份、恢复、EdgeX 测试
├── handlers_config.go # 配置：CRUD、MQTT、订阅、主题管理
├── handlers_export.go # 导入导出：CSV / JSON / SQL
├── handlers_auth.go   # 认证：API Key CRUD + 加密管理
├── handlers_ops.go    # 运维：健康、指标、保留、告警、资源
└── handlers_web.go    # Web 页面 + 静态文件 + WebSocket 端点
```

## 3. 架构设计

### 3.1 核心组件

```
Server
├── Table           *engine.Table          数据库表引用
├── Config          *config.Config         系统配置
├── Monitor         *monitor.Monitor       监控管理器
├── RetentionMgr    *retention.RetentionManager  数据保留管理器
├── AlertNotifier   *alert.Notifier        告警通知器
├── ResourceMonitor *resource.ResourceMonitor    资源监控器
└── wsManager       *WSManager             WebSocket 管理器（私有）
```

### 3.2 请求流程

```
HTTP 请求
    ↓
DeviceNameMiddleware（仅 /api/readings）
    ↓
路由匹配
    ↓
Handler 处理
    ↓
JSON 响应
```

### 3.3 广播流程

```
MQTT 消息
    ↓
BroadcastChan
    ↓
broadcastLoop (goroutine)
    ↓
JSON 序列化
    ↓
wsManager.Broadcast
    ↓
所有 WebSocket 客户端
```

## 4. 数据结构

### 4.1 Server 结构体

```go
type Server struct {
    Table           *engine.Table
    Config          *config.Config
    Monitor         *monitor.Monitor
    RetentionMgr    *retention.RetentionManager
    AlertNotifier   *alert.Notifier
    ResourceMonitor *resource.ResourceMonitor
    wsManager       *WSManager
}
```

### 4.2 WSManager 结构体

```go
type WSManager struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan []byte
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
}
```

## 5. 生命周期

### 5.1 创建实例

```go
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor,
    retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier,
    resourceMonitor *resource.ResourceMonitor) *Server
```

创建 WSManager → 启动 WSManager.Run() → 返回 Server 实例。

### 5.2 启动

```go
func (s *Server) Start() error
```

注册路由 → 后台启动 HTTP 服务器（默认端口 8081）。

### 5.3 广播启动

```go
func (s *Server) StartBroadcast(ch <-chan *mqtt.BroadcastMessage)
```

启动 goroutine 监听 MQTT 广播通道，将消息转发至 WebSocket 客户端。

## 6. API 端点

### 6.1 健康检查

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/healthz` | GET | 健康检查（兼容） |
| `/ready` | GET | 就绪检查 |

### 6.2 数据管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/readings` | GET | 查询设备读数 |
| `/api/device-status` | GET | 设备状态列表 |
| `/api/backup` | POST | 备份数据库 |
| `/api/restore` | POST | 恢复数据库 |
| `/api/test-edgex` | POST | 测试 EdgeX 消息 |

### 6.3 配置管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/config/get` | GET | 获取当前配置 |
| `/api/config/update` | POST | 更新配置 |
| `/api/config/reload` | POST | 从文件重新加载配置 |
| `/api/config/oneclick` | POST | 一键智能配置 |
| `/api/config/mqtt` | POST | 更新 MQTT 配置 |

### 6.4 订阅管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/subscription/status` | GET | 订阅状态 |
| `/api/subscription/test` | POST | 测试订阅主题 |
| `/api/subscription/themes` | GET/POST/DELETE | 主题管理 |

### 6.5 导入导出

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/export/csv` | GET | 导出 CSV |
| `/api/export/json` | GET | 导出 JSON |
| `/api/export/sql` | GET | 导出 SQL |
| `/api/import/csv` | POST | 导入 CSV |
| `/api/import/json` | POST | 导入 JSON |
| `/api/data/export` | GET | 参数化导出（JSON/CSV） |

### 6.6 认证与加密

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/auth/create-key` | POST | 创建 API Key |
| `/api/auth/list-keys` | GET | 列出 API Key |
| `/api/auth/revoke-key` | POST | 撤销 API Key |
| `/api/encryption/rotate-key` | POST | 轮换加密密钥 |
| `/api/encryption/status` | GET | 获取加密状态 |

### 6.7 运维监控

| 端点 | 方法 | 说明 |
|------|------|------|
| `/metrics` | GET | 监控指标 |
| `/api/resources/status` | GET | 资源使用状态 |
| `/api/retention/status` | GET | 保留策略状态 |
| `/api/retention/cleanup` | POST | 手动数据清理 |
| `/api/alerts/notifier/status` | GET | 告警通知器状态 |
| `/api/alerts/test` | POST | 发送测试告警 |
| `/api/alert-groups` | GET | 分组告警 |

### 6.8 WebSocket

| 端点 | 方法 | 说明 |
|------|------|------|
| `/ws` | GET | 建立 WebSocket 连接 |

## 7. 使用示例

### 7.1 创建和启动

```go
server := server.NewServer(table, cfg, monitor, retentionMgr, alertNotifier, resourceMonitor)
server.Start()
server.StartBroadcast(mqttClient.BroadcastChan())
```

### 7.2 WebSocket 广播

```go
wsManager := server.GetWSManager()
wsManager.Broadcast([]byte(`{"type":"alert","data":"High temperature"}`))
```

## 8. 错误处理

### 8.1 错误响应格式

```json
{ "error": "错误描述信息" }
```

### 8.2 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 405 | 方法不允许 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

## 9. 安全性

- **CORS**：WebSocket 升级器允许所有来源（生产环境应限制）
- **输入验证**：所有 API 端点均验证必填参数
- **请求体限制**：EdgeX 测试端点限制请求体为 10MB

## 10. 注意事项

1. **goroutine 安全**：WSManager 使用通道实现异步广播，无锁设计
2. **资源清理**：WebSocket 断开时自动注销连接
3. **配置更新**：部分配置需要重启服务才能生效
4. **Broadcast 调用**：`StartBroadcast` 必须在 MQTT 客户端初始化后调用
