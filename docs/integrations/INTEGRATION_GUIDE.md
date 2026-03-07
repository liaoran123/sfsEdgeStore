# 系统集成指南

## 概述

本文档详细介绍如何将 sfsEdgeStore 集成到您的系统中，包括 API 集成、MQTT 集成、SDK 使用等内容。

## 集成方式

### 1. REST API 集成

#### 基础配置

```javascript
// 配置 API 客户端
const API_BASE = 'http://localhost:8080/api';

// 创建请求
async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${API_KEY}`,
      ...options.headers
    },
    ...options
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  return response.json();
}
```

#### 写入数据

```javascript
// 写入单个读数
async function writeReading(deviceName, resourceName, value, valueType) {
  return request('/readings', {
    method: 'POST',
    body: JSON.stringify({
      deviceName,
      resourceName,
      value: String(value),
      valueType,
      timestamp: Date.now()
    })
  });
}

// 批量写入
async function writeReadingsBatch(readings) {
  return request('/readings/batch', {
    method: 'POST',
    body: JSON.stringify(readings)
  });
}
```

#### 查询数据

```javascript
// 查询读数
async function queryReadings(deviceName, resourceName, start, end, limit = 100) {
  const params = new URLSearchParams({
    device: deviceName,
    resource: resourceName,
    start,
    end,
    limit
  });
  
  return request(`/readings?${params}`);
}

// 聚合查询
async function queryAggregate(deviceName, resourceName, start, end, granularity) {
  const params = new URLSearchParams({
    device: deviceName,
    resource: resourceName,
    start,
    end,
    granularity,
    function: 'avg'
  });
  
  return request(`/aggregate?${params}`);
}
```

### 2. MQTT 集成

#### Python 示例 (paho-mqtt)

```python
import paho.mqtt.client as mqtt
import json
import time

# 配置
MQTT_BROKER = 'localhost'
MQTT_PORT = 1883
MQTT_TOPIC = 'edgex/events'

def on_connect(client, userdata, flags, rc):
    print(f"Connected with result code {rc}")
    client.subscribe(MQTT_TOPIC)

def on_message(client, userdata, msg):
    print(f"Received message: {msg.topic}")
    try:
        data = json.loads(msg.payload.decode())
        print(f"Data: {data}")
    except Exception as e:
        print(f"Error parsing message: {e}")

# 创建客户端
client = mqtt.Client()
client.on_connect = on_connect
client.on_message = on_message

# 连接并订阅
client.connect(MQTT_BROKER, MQTT_PORT, 60)
client.loop_start()

# 发布数据
def publish_reading(device_name, resource_name, value):
    payload = {
        "correlationId": "sample-id",
        "messageType": "event",
        "origin": int(time.time() * 1000),
        "payload": {
            "id": "event-id",
            "deviceName": device_name,
            "origin": int(time.time() * 1000),
            "readings": [
                {
                    "id": "reading-id",
                    "resourceName": resource_name,
                    "value": str(value),
                    "valueType": "Float64",
                    "origin": int(time.time() * 1000)
                }
            ]
        }
    }
    
    client.publish(MQTT_TOPIC, json.dumps(payload))

# 示例：发布读数
publish_reading("device001", "temperature", 25.5)
```

#### JavaScript 示例 (MQTT.js)

```javascript
const mqtt = require('mqtt');

// 配置
const MQTT_BROKER = 'mqtt://localhost:1883';
const MQTT_TOPIC = 'edgex/events';

// 连接
const client = mqtt.connect(MQTT_BROKER);

client.on('connect', () => {
  console.log('Connected to MQTT broker');
  client.subscribe(MQTT_TOPIC);
});

client.on('message', (topic, message) => {
  console.log(`Received message on ${topic}:`, message.toString());
});

// 发布读数
function publishReading(deviceName, resourceName, value, valueType) {
  const payload = {
    correlationId: 'sample-id',
    messageType: 'event',
    origin: Date.now(),
    payload: {
      id: 'event-id',
      deviceName,
      origin: Date.now(),
      readings: [
        {
          id: 'reading-id',
          resourceName,
          value: String(value),
          valueType,
          origin: Date.now()
        }
      ]
    }
  };
  
  client.publish(MQTT_TOPIC, JSON.stringify(payload));
}

// 示例
publishReading('device001', 'temperature', 25.5, 'Float64');
```

### 3. Go SDK 集成

#### 安装 SDK

```bash
go get github.com/username/sfsedgestore/sdk
```

#### 使用示例

```go
package main

import (
    "context"
    "log"
    "time"
    
    sfs "github.com/username/sfsedgestore/sdk"
)

func main() {
    // 创建客户端
    client, err := sfs.NewClient(sfs.Config{
        BaseURL: "http://localhost:8080",
        APIKey:  "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 写入读数
    err = client.WriteReading(context.Background(), sfs.Reading{
        DeviceName:   "device001",
        ResourceName: "temperature",
        Value:        "25.5",
        ValueType:    "Float64",
        Timestamp:    time.Now().UnixMilli(),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 查询读数
    readings, err := client.QueryReadings(context.Background(), sfs.QueryOptions{
        DeviceName:   "device001",
        ResourceName: "temperature",
        Start:        time.Now().Add(-1 * time.Hour).UnixMilli(),
        End:          time.Now().UnixMilli(),
        Limit:        100,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d readings", len(readings))
}
```

### 4. Python SDK 集成

#### 安装 SDK

```bash
pip install sfsedgestore-sdk
```

#### 使用示例

```python
from sfsedgestore import SfsEdgeStoreClient
import time

# 创建客户端
client = SfsEdgeStoreClient(
    base_url="http://localhost:8080",
    api_key="your-api-key"
)

# 写入读数
client.write_reading(
    device_name="device001",
    resource_name="temperature",
    value="25.5",
    value_type="Float64",
    timestamp=int(time.time() * 1000)
)

# 查询读数
readings = client.query_readings(
    device_name="device001",
    resource_name="temperature",
    start=int(time.time() * 1000) - 3600000,
    end=int(time.time() * 1000),
    limit=100
)

print(f"Found {len(readings)} readings")
```

## WebSocket 集成

### 连接 WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/readings');

ws.onopen = () => {
  console.log('WebSocket connected');
  
  // 订阅设备
  ws.send(JSON.stringify({
    action: 'subscribe',
    devices: ['device001', 'device002']
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('New reading:', data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
  // 重连逻辑
};
```

## 认证集成

### API Key 认证

```bash
curl http://localhost:8080/api/readings \
  -H "Authorization: Bearer your-api-key"
```

### JWT 认证

```javascript
// 获取 Token
async function login(username, password) {
  const response = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  const data = await response.json();
  return data.token;
}

// 使用 Token
const token = await login('admin', 'password');
const response = await fetch('http://localhost:8080/api/readings', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

## 错误处理

### HTTP 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

### 错误响应格式

```json
{
  "code": 400,
  "message": "Invalid parameter",
  "details": {
    "parameter": "start",
    "reason": "must be a valid timestamp"
  },
  "requestId": "req-12345"
}
```

## 最佳实践

### 1. 批量写入

使用批量写入提高性能：

```javascript
const batch = [];
for (let i = 0; i < 100; i++) {
  batch.push({
    deviceName: 'device001',
    resourceName: 'temperature',
    value: String(20 + Math.random() * 10),
    valueType: 'Float64',
    timestamp: Date.now()
  });
}

await writeReadingsBatch(batch);
```

### 2. 分页查询

大量数据使用分页：

```javascript
async function queryAllReadings(deviceName, resourceName, start, end) {
  const allReadings = [];
  let offset = 0;
  const limit = 1000;
  
  while (true) {
    const readings = await queryReadings(
      deviceName, resourceName, start, end, limit, offset
    );
    
    allReadings.push(...readings);
    
    if (readings.length < limit) {
      break;
    }
    
    offset += limit;
  }
  
  return allReadings;
}
```

### 3. 重试机制

实现指数退避重试：

```javascript
async function requestWithRetry(path, options, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      return await request(path, options);
    } catch (error) {
      if (i === retries - 1) throw error;
      
      const delay = Math.pow(2, i) * 1000;
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
}
```

### 4. 连接池

使用 HTTP 连接池：

```javascript
const https = require('https');
const agent = new https.Agent({
  keepAlive: true,
  maxSockets: 10
});

// 在 fetch 中使用
const response = await fetch(url, { agent });
```

## 测试集成

### 健康检查

```bash
curl http://localhost:8080/health
```

### 写入测试数据

```bash
curl -X POST http://localhost:8080/api/readings \
  -H "Content-Type: application/json" \
  -d '{
    "deviceName": "test-device",
    "resourceName": "test-resource",
    "value": "123.45",
    "valueType": "Float64",
    "timestamp": '$(date +%s000)'
  }'
```

### 查询测试数据

```bash
curl "http://localhost:8080/api/readings?device=test-device&limit=10"
```

## 更多资源

- [API 文档](../api.md)
- [EdgeX 集成指南](./EDGEX_INTEGRATION.md)
- [第三方集成](./THIRD_PARTY.md)
- [用户指南](../user-guide/USER_MANUAL.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
