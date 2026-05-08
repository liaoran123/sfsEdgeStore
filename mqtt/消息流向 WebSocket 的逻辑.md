## 消息流向 WebSocket 的逻辑

## 完整流程

```
MQTT 消息到达
    ↓
mqtt.MQTTClient.onMessage (mqtt.go)
    ↓
Client.handleMessage (client.go)
    ↓
MessageProcessor.Handler (messageProcessor.go) → 解析消息
    ↓
BatchWriter.Add() (batchWriter.go) → 放入缓冲区
    ↓
flushLoop() 定时触发 doWrite()
    ↓
doWrite() 分发到三个出口：
    ├── 写入 LevelDB (database.Insert)
    ├── NewBroadcastMessage("device_data") → broadcastChan
    └── analyzeData() → 分析告警 → NewBroadcastMessage("alerts") → broadcastChan
    ↓
Server.SetBroadcastChan() 监听 channel (server.go)
    ↓
Server.broadcastLoop() → msg.MarshalJSON()
    ↓
Server.Broadcast() → wsManager.Broadcast()
    ↓
wsManager.Run() 遍历 clients → conn.WriteMessage()
    ↓
WebSocket 前端收到！
```

## 广播消息类型

| 类型 | 触发点 | 数据内容 |
|------|--------|----------|
| `device_data` | 写入数据库后 | 实时设备记录 |
| `alerts` | analyzeData() 分析后 | 告警列表 |

## 代码对应

| 位置 | 代码 |
|------|------|
| `mqtt/client.go:55` | `BroadcastChan()` 返回广播通道 |
| `mqtt/batchWriter.go:113` | `NewBroadcastMessage()` 广播实时数据 |
| `mqtt/analyze.go:38` | `NewBroadcastMessage()` 广播告警数据 |
| `mqtt/broadcast.go:55` | `NewBroadcastMessage()` 辅助函数 |
| `mqtt/broadcast.go:45` | `PutTo()` 推入通道，满时静默丢弃 |
| `server/server.go:61` | `SetBroadcastChan()` 启动监听 |
| `server/server.go:66` | `broadcastLoop()` 读取通道并广播 |
| `server/ws_manager.go:30` | `Run()` 单协程事件循环 |

## 对象池管理

| 阶段 | 释放方式 |
|------|----------|
| 通道满时 | `PutTo()` 内部 `PutBroadcastMessage()` |
| 广播成功 | `broadcastLoop()` → `PutBroadcastMessage()` |
| Marshal 失败 | `broadcastLoop()` → `PutBroadcastMessage()` |

## 结论

消息流向 WebSocket 的功能完整保留，职责清晰：
- **BatchWriter**：写库 + 产生广播消息放入 channel
- **mqtt.Client**：暴露 channel，不处理广播逻辑
- **Server**：直接监听 channel → 推送到 WebSocket
