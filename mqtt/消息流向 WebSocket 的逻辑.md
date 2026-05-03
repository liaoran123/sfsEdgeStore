

## 重构后的完整流程

```
MQTT 消息到达
    ↓
mqtt.MQTTClient.onMessage (mqtt.go)
    ↓
MessageProcessor.Handler (messageProcessor.go)
    ↓
解析消息 → batchWriter.Enqueue() (batchWriter.go:81-84)
    ↓
批量写入 LevelDB (batchWriter.go:97)
    ↓
[关键] broadcastData() (batchWriter.go:112-117)
    ↓
broadcaster.Broadcast() (batchWriter.go:153-171)
    ↓
ws.WSManager.Broadcast()
    ↓
ws.WSManager.broadcast channel
    ↓
遍历 clients → client.WriteMessage()
    ↓
WebSocket 前端收到！
```

## 代码对应

| 位置 | 代码 |
|------|------|
| 1 | `mqtt/client.go:29` → 创建 `BatchWriter`，传入 `broadcaster` |
| 2 | `mqtt/batchWriter.go:81` → 提交到 `writePool` |
| 3 | `mqtt/batchWriter.go:97` → 写入 LevelDB |
| 4 | `mqtt/batchWriter.go:112` → `if w.broadcaster != nil` → 广播数据 |
| 5 | `mqtt/batchWriter.go:153` → `broadcastData()` 函数 |

## 和原 `core/mqtt/client.go` 对比

| 部分 | 原代码 | 新代码 |
|------|--------|--------|
| **广播逻辑** | `core/mqtt/messageHandler.go:174` | `mqtt/batchWriter.go:112` |
| **广播函数** | `c.broadcastData()` | `w.broadcastData()` |
| **封装** | 都在一个大 Client 里 | 拆分为三个组件 |
| **功能** | ✅ 一样 | ✅ 一样 |

## 结论

重构后**消息流向 WebSocket 的功能完全保留**，只是代码更清晰、职责更单一！