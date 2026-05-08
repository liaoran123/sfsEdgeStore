# ws 包 - WebSocket 管理器技术文档

## 概述

`ws` 包提供 WebSocket 连接管理功能，支持多客户端连接、消息广播和连接生命周期管理。采用**单协程事件循环**架构，通过 channel 实现线程安全的连接操作。

## 核心组件

### WSManager 结构体

```go
type WSManager struct {
    clients    map[*websocket.Conn]bool  // 已连接的客户端集合
    broadcast  chan []byte               // 广播消息通道（缓冲 256）
    register   chan *websocket.Conn      // 连接注册通道
    unregister chan *websocket.Conn      // 连接注销通道
}
```

**字段说明：**

| 字段 | 类型 | 作用 |
|------|------|------|
| `clients` | `map[*websocket.Conn]bool` | 存储所有活跃的 WebSocket 连接 |
| `broadcast` | `chan []byte` | 接收需要广播的消息，缓冲区 256 |
| `register` | `chan *websocket.Conn` | 接收新连接的注册请求 |
| `unregister` | `chan *websocket.Conn` | 接收连接的注销请求 |

## API 参考

### 构造函数

#### NewWSManager()

创建新的 WebSocket 管理器实例。

```go
func NewWSManager() *WSManager
```

**默认配置：**
- `broadcast` channel 缓冲区：256 条消息
- 初始连接数：0

**示例：**
```go
manager := ws.NewWSManager()
go manager.Run()
```

---

### 方法

#### Run() - 事件循环

启动 WebSocket 管理器的核心事件循环，必须在独立 goroutine 中运行。

```go
func (manager *WSManager) Run()
```

**处理的事件类型：**

| 事件 | 通道 | 行为 |
|------|------|------|
| 注册连接 | `register` | 将连接添加到 `clients` map |
| 注销连接 | `unregister` | 从 `clients` map 移除并关闭连接 |
| 广播消息 | `broadcast` | 向所有客户端发送消息 |

**架构说明：**

该方法是**单协程串行处理**，无需互斥锁即可安全访问 `clients` map。

#### Broadcast(message []byte) - 广播消息

向所有已连接的客户端发送消息。

```go
func (manager *WSManager) Broadcast(message []byte)
```

**参数：**
- `message`：要广播的字节数据

**行为：**
- 消息被推入 `broadcast` channel
- `Run()` 方法中的事件循环会处理该消息并发送给所有客户端
- 如果 channel 缓冲区已满，调用会被阻塞

#### Register(conn *websocket.Conn) - 注册连接

将新的 WebSocket 连接注册到管理器。

```go
func (manager *WSManager) Register(conn *websocket.Conn)
```

**参数：**
- `conn`：WebSocket 连接实例

#### Unregister(conn *websocket.Conn) - 注销连接

从管理器中移除 WebSocket 连接。

```go
func (manager *WSManager) Unregister(conn *websocket.Conn)
```

**参数：**
- `conn`：要移除的 WebSocket 连接

## 使用示例

### 基本用法

```go
// 创建并启动管理器
manager := ws.NewWSManager()
go manager.Run()

// 注册新连接
manager.Register(conn)

// 广播消息
manager.Broadcast([]byte(`{"type": "update", "data": "..."}`))

// 注销连接
manager.Unregister(conn)
```

### 在 HTTP Handler 中使用

```go
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    manager.Register(conn)
    
    // 保持连接
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            manager.Unregister(conn)
            break
        }
    }
}
```

## 架构设计

### 线程安全机制

WSManager 采用 **Actor 模型**：所有对 `clients` map 的操作都在单一 goroutine 中执行，通过 channel 进行异步通信，避免了并发访问 map 的竞争条件。

```
[HTTP Handler] --register--> [Run() goroutine] --> clients map
[HTTP Handler] --unregister-> [Run() goroutine] --> clients map
[其他模块] ----broadcast----> [Run() goroutine] --> clients map
```

### 消息流

```
Broadcast([]byte)
    ↓
broadcast channel (buffer: 256)
    ↓
Run() select 接收
    ↓
遍历 clients map
    ↓
conn.WriteMessage()
```

## 注意事项

### 1. 广播阻塞

如果某个客户端的 `WriteMessage` 因网络慢而阻塞，会阻塞整个事件循环，导致其他客户端的广播也被延迟。

**建议优化：** 为每个连接的写入操作启动独立 goroutine，或使用独立的写通道。

### 2. Channel 缓冲区

`broadcast` channel 缓冲区为 256。如果广播频率过高，缓冲区满后 `Broadcast()` 调用会被阻塞。

### 3. 运行模式

`Run()` 必须在独立 goroutine 中启动，否则会阻塞主线程。

```go
manager := ws.NewWSManager()
go manager.Run()  // 必须使用 go 关键字
```

### 4. 连接关闭

调用 `Unregister()` 会自动关闭对应的 WebSocket 连接，不需要手动 `conn.Close()`。
