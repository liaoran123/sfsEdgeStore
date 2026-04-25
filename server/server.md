# 服务器包 (Server Package) 技术文档

## 1. 概述

服务器包是sfsEdgeStore系统的核心组件，负责提供HTTP API接口和WebSocket实时通信功能。它集成了数据存储、设备监控、告警管理、数据保留等多项功能，为前端Web界面和第三方系统提供统一的API服务。

### 1.1 主要功能

- **HTTP API服务**：提供丰富的RESTful API接口，支持数据查询、配置管理、设备状态监控等
- **WebSocket通信**：支持实时数据推送和告警通知
- **静态文件服务**：提供Web界面所需的HTML、CSS、JavaScript等静态资源
- **多协议支持**：支持HTTP/HTTPS、WebSocket等通信协议
- **中间件支持**：提供设备名称格式化、请求限流等中间件功能

## 2. 架构设计

### 2.1 核心组件

```
Server Package
├── Server                  # HTTP服务器主结构
│   ├── Table              # 数据库表引用
│   ├── Config             # 配置对象
│   ├── Monitor            # 监控管理器
│   ├── RetentionMgr       # 数据保留管理器
│   ├── AlertNotifier      # 告警通知器
│   ├── ResourceMonitor    # 资源监控器
│   ├── MQTTClient         # MQTT客户端
│   ├── wsManager          # WebSocket管理器
│   ├── TemplateManager    # 模板管理器
│   └── BaselineManager    # 基线管理器
│
├── WSManager              # WebSocket连接管理器
│   ├── clients            # 客户端连接映射
│   ├── broadcast          # 广播消息通道
│   ├── register           # 注册通道
│   └── unregister         # 注销通道
│
└── RateLimiter           # 请求限流器
    ├── requests           # 客户端请求记录
    ├── limit              # 请求限制数
    └── window             # 时间窗口
```

### 2.2 工作流程

```
HTTP请求进入
    ↓
限流中间件检查
    ↓
设备名称格式化中间件
    ↓
路由匹配
    ↓
执行对应处理器
    ↓
返回JSON响应
```

## 3. 数据结构

### 3.1 Server 结构体

```go
type Server struct {
    Table           *engine.Table           // 数据库表引用
    Config          *config.Config         // 配置对象
    Monitor         *monitor.Monitor       // 监控管理器
    RetentionMgr    *retention.RetentionManager  // 数据保留管理器
    AlertNotifier   *alert.Notifier        // 告警通知器
    ResourceMonitor *resource.ResourceMonitor  // 资源监控器
    MQTTClient      mqtt.Client            // MQTT客户端
    wsManager       *ws.WSManager          // WebSocket管理器
    TemplateManager *template.Manager      // 模板管理器
    BaselineManager *baseline.Manager      // 基线管理器
}
```

### 3.2 WSManager 结构体

```go
type WSManager struct {
    clients    map[*websocket.Conn]bool  // 客户端连接映射
    broadcast  chan []byte               // 广播消息通道
    register   chan *websocket.Conn      // 注册通道
    unregister chan *websocket.Conn      // 注销通道
}
```

### 3.3 RateLimiter 结构体

```go
type RateLimiter struct {
    requests map[string]*clientRequest  // 客户端请求记录
    mu       sync.Mutex                // 互斥锁
    limit    int                        // 请求限制数
    window   time.Duration              // 时间窗口
}

type clientRequest struct {
    count     int          // 当前请求数
    resetTime time.Time    // 重置时间
}
```

## 4. 函数说明

### 4.1 服务器生命周期管理

#### NewServer
```go
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor,
    retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier,
    resourceMonitor *resource.ResourceMonitor) *Server
```

**功能**：创建并初始化服务器实例

**处理逻辑**：
1. 创建WebSocket管理器并启动
2. 初始化模板管理器并加载模板
3. 初始化基线管理器
4. 返回配置好的服务器实例

#### Start
```go
func (s *Server) Start() error
```

**功能**：启动HTTP服务器

**处理逻辑**：
1. 注册所有HTTP路由
2. 启动后台goroutine运行HTTP服务器
3. 根据配置选择HTTP或HTTPS模式

---

### 4.2 中间件

#### DeviceNameMiddleware
```go
func DeviceNameMiddleware(next http.HandlerFunc) http.HandlerFunc
```

**功能**：格式化HTTP请求中的deviceName参数

**处理逻辑**：
1. 从URL查询参数中获取deviceName
2. 调用common.FormatDeviceName格式化设备名称
3. 重写URL参数，使用格式化后的设备名称
4. 传递给下一个处理器

---

### 4.3 API端点

#### 健康检查API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查，返回服务状态 |
| `/healthz` | GET | Kubernetes就绪探测 |
| `/ready` | GET | 服务就绪检查 |

#### 数据管理API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/readings` | GET | 查询设备读数数据 |
| `/api/backup` | POST | 备份数据库 |
| `/api/restore` | POST | 恢复数据库 |
| `/api/export/csv` | GET | 导出数据为CSV |
| `/api/export/json` | GET | 导出数据为JSON |
| `/api/export/sql` | GET | 导出数据为SQL |
| `/api/data/export` | GET | 参数化数据导出 |

#### 配置管理API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/config/get` | GET | 获取当前配置 |
| `/api/config/update` | POST | 更新配置 |
| `/api/config/reload` | POST | 重新加载配置 |
| `/api/config/oneclick` | POST | 一键智能配置 |
| `/api/config/mqtt` | POST | 更新MQTT配置 |

#### 监控指标API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/metrics` | GET | 获取Prometheus格式指标 |
| `/api/device-status` | GET | 获取设备状态列表 |
| `/api/resources/status` | GET | 获取资源使用状态 |

#### MQTT订阅管理API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/subscription/status` | GET | 获取订阅状态 |
| `/api/subscription/test` | POST | 测试订阅主题 |
| `/api/subscription/themes` | GET/POST/DELETE | 管理订阅主题 |

#### 告警管理API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/alerts` | GET | 获取告警列表 |
| `/api/alert-groups` | GET | 获取分组告警 |
| `/api/alerts/notifier/status` | GET | 获取告警通知器状态 |
| `/api/alerts/test` | POST | 发送测试告警 |

#### 数据保留API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/retention/status` | GET | 获取保留策略状态 |
| `/api/retention/cleanup` | POST | 手动触发数据清理 |

#### 认证管理API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/auth/create-key` | POST | 创建API密钥 |
| `/api/auth/list-keys` | GET | 列出API密钥 |
| `/api/auth/revoke-key` | POST | 撤销API密钥 |

#### 加密管理API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/encryption/rotate-key` | POST | 轮换加密密钥 |
| `/api/encryption/status` | GET | 获取加密状态 |

#### 模板和基线API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/templates` | GET | 获取行业模板 |
| `/api/templates/apply` | POST | 应用模板配置 |
| `/api/baselines` | GET | 获取基线列表 |
| `/api/baselines/calculate` | POST | 计算基线 |

#### WebSocket API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/ws` | GET | 建立WebSocket连接 |

---

### 4.4 WebSocket管理器

#### NewWSManager
```go
func NewWSManager() *WSManager
```

**功能**：创建WebSocket管理器实例

#### Run
```go
func (manager *WSManager) Run()
```

**功能**：启动WebSocket管理器事件循环

**处理逻辑**：
1. 监听register、unregister和broadcast通道
2. 处理客户端连接注册
3. 处理客户端连接注销
4. 广播消息给所有已连接客户端

#### Broadcast
```go
func (manager *WSManager) Broadcast(message []byte)
```

**功能**：广播消息给所有已连接客户端

---

### 4.5 限流器

#### NewRateLimiter
```go
func NewRateLimiter(limit int, window time.Duration) *RateLimiter
```

**功能**：创建限流器实例

#### Allow
```go
func (rl *RateLimiter) Allow(clientID string) bool
```

**功能**：检查是否允许请求

**返回值**：true表示允许，false表示超出限制

#### Middleware
```go
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc
```

**功能**：创建限流中间件

---

## 5. 配置说明

### 5.1 HTTP服务器配置

```go
type Config struct {
    HTTPPort      string  // HTTP端口，默认"8081"
    HTTPUseTLS    bool    // 是否使用TLS
    HTTPCert      string  // TLS证书路径
    HTTPKey       string  // TLS密钥路径
}
```

### 5.2 限流配置

```go
// 创建限流器，每分钟最多100个请求
limiter := NewRateLimiter(100, time.Minute)
```

## 6. 使用示例

### 6.1 创建和启动服务器

```go
// 创建数据库表
table := engine.OpenTable("data")

// 获取配置
cfg := config.GetConfig()

// 创建监控器
monitor := monitor.NewMonitor()

// 创建保留管理器
retentionMgr := retention.NewRetentionManager(table, cfg)

// 创建告警通知器
alertNotifier := alert.NewNotifier()

// 创建资源监控器
resourceMonitor := resource.NewMonitor()

// 创建服务器
server := server.NewServer(table, cfg, monitor, retentionMgr, alertNotifier, resourceMonitor)

// 启动服务器
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
```

### 6.2 使用限流中间件

```go
// 创建限流器
limiter := server.NewRateLimiter(100, time.Minute)

// 应用到路由
http.HandleFunc("/api/data/export", limiter.Middleware(server.handleDataExport))
```

### 6.3 WebSocket消息广播

```go
// 获取WebSocket管理器
wsManager := server.GetWSManager()

// 广播消息
message := []byte(`{"type":"alert","data":"High temperature detected"}`)
wsManager.Broadcast(message)
```

## 7. 性能优化

### 7.1 连接管理

- **WebSocket连接池**：使用map管理客户端连接，支持高效注册和注销
- **消息广播**：使用通道进行异步广播，避免阻塞主线程
- **连接超时**：设置合理的读写超时，防止资源泄漏

### 7.2 请求处理

- **异步处理**：长时间操作使用goroutine异步执行
- **批量操作**：数据库操作使用批量插入，提高吞吐量
- **中间件优化**：中间件轻量化设计，减少处理开销

### 7.3 限流策略

```go
// 不同端点使用不同限流策略
apiLimiter := NewRateLimiter(100, time.Minute)    // API: 100次/分钟
dataLimiter := NewRateLimiter(10, time.Minute)    // 数据导出: 10次/分钟
```

## 8. 错误处理

### 8.1 错误响应格式

```json
{
    "error": "错误描述信息"
}
```

### 8.2 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁（限流） |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

### 8.3 常见错误处理

```go
// 数据库操作错误
if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]string{
        "error": err.Error(),
    })
    return
}

// 参数验证错误
if req.Topic == "" {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Topic is required",
    })
    return
}
```

## 9. 安全性

### 9.1 CORS配置

WebSocket升级器允许所有来源的连接：

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境应限制为特定域名
    },
}
```

### 9.2 输入验证

所有API端点都对输入参数进行验证：

```go
// 验证必填参数
if req.UserID == "" || req.Role == "" {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "User ID and role are required",
    })
    return
}

// 验证参数范围
if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 10000 {
    limit = l
}
```

## 10. 监控指标

### 10.1 HTTP请求监控

每个API处理器都增加HTTP请求计数：

```go
func (s *Server) handleQueryReadings(w http.ResponseWriter, r *http.Request) {
    if s.Monitor != nil {
        s.Monitor.IncrementHTTPRequests()
    }
    // 处理请求...
}
```

### 10.2 WebSocket连接监控

```go
// 客户端连接时
log.Printf("WebSocket client connected, total: %d", len(manager.clients))

// 客户端断开时
log.Printf("WebSocket client disconnected, total: %d", len(manager.clients))
```

## 11. 注意事项

1. **线程安全**：限流器使用互斥锁保护共享数据
2. **资源清理**：WebSocket连接断开时正确释放资源
3. **配置更新**：部分配置需要重启服务才能生效
4. **错误日志**：关键操作应记录详细日志便于排查
5. **限流策略**：根据实际业务调整限流参数

## 12. 总结

服务器包通过模块化设计，提供了完整的HTTP API服务和WebSocket实时通信能力。其设计遵循以下原则：

- **高性能**：使用goroutine和通道实现异步处理
- **可扩展**：模块化设计便于功能扩展
- **易用性**：统一的API风格和错误处理
- **安全性**：输入验证和限流保护
- **可观测性**：完善的日志和监控指标