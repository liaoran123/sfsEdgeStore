          
## 数据分发流程

### ✅ 是的，`Add` 是数据分发入口

```
MQTT消息到达
    ↓
MessageProcessor.ProcessEvent() 解析数据
    ↓
batchWriter.Add(records)  ← 所有数据经过这里
    ↓
    ├── 积累到阈值 (batchSize) 或超时 (batchTime)
    ↓
    └── flush() → writePool.Submit(doWrite)
         ↓
         doWrite() 分发到三个出口：
         ├── 1. database.BatchInsertWithRetry()  → 写入数据库
         ├── 2. broadcastData()                  → 推送到 WebSocket
         └── 3. analyzeData()                    → 数据分析
```

### 📊 分发逻辑

| 出口 | 触发条件 | 说明 |
|------|----------|------|
| **数据库** | 批量写入成功后 | LevelDB 持久化 |
| **WebSocket** | 数据库写入成功后 | 推送到前端 |
| **Analyzer** | 数据库写入成功后，且 records ≤ 50 | 数据分析/告警 |

