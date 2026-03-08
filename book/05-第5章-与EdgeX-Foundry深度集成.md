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




