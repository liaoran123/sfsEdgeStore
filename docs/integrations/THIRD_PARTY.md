# 第三方集成说明

## 概述

本文档介绍 sfsEdgeStore 与常见第三方系统的集成方法。

## 消息队列集成

### Kafka

#### 配置 Kafka 消费者

```java
import org.apache.kafka.clients.consumer.*;
import org.apache.kafka.common.serialization.StringDeserializer;
import java.util.Properties;
import java.time.Duration;

public class SfsEdgeStoreKafkaConsumer {
    public static void main(String[] args) {
        Properties props = new Properties();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, "localhost:9092");
        props.put(ConsumerConfig.GROUP_ID_CONFIG, "sfsedgestore-group");
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class.getName());
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class.getName());
        
        KafkaConsumer<String, String> consumer = new KafkaConsumer<>(props);
        consumer.subscribe(Arrays.asList("iot-data"));
        
        while (true) {
            ConsumerRecords<String, String> records = consumer.poll(Duration.ofMillis(100));
            for (ConsumerRecord<String, String> record : records) {
                // 发送到 sfsEdgeStore
                sendToSfsEdgeStore(record.value());
            }
        }
    }
    
    private static void sendToSfsEdgeStore(String data) {
        // 使用 HTTP API 发送到 sfsEdgeStore
    }
}
```

### RabbitMQ

#### 配置 RabbitMQ 消费者

```python
import pika
import json
import requests

def callback(ch, method, properties, body):
    data = json.loads(body)
    
    # 发送到 sfsEdgeStore
    response = requests.post(
        'http://localhost:8080/api/readings',
        json=data
    )
    
    print(f"Sent to sfsEdgeStore: {response.status_code}")
    ch.basic_ack(delivery_tag=method.delivery_tag)

connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
channel = connection.channel()

channel.queue_declare(queue='iot-data')
channel.basic_consume(queue='iot-data', on_message_callback=callback)

print('Waiting for messages...')
channel.start_consuming()
```

## 时序数据库集成

### InfluxDB

#### 从 sfsEdgeStore 同步到 InfluxDB

```python
import requests
from influxdb_client import InfluxDBClient, Point
from influxdb_client.client.write_api import SYNCHRONOUS

# sfsEdgeStore 配置
SFS_API = 'http://localhost:8080/api'

# InfluxDB 配置
INFLUX_URL = 'http://localhost:8086'
INFLUX_TOKEN = 'your-token'
INFLUX_ORG = 'your-org'
INFLUX_BUCKET = 'sfsedgestore'

client = InfluxDBClient(url=INFLUX_URL, token=INFLUX_TOKEN, org=INFLUX_ORG)
write_api = client.write_api(write_options=SYNCHRONOUS)

def sync_to_influx(device_name, start, end):
    # 从 sfsEdgeStore 查询数据
    response = requests.get(f'{SFS_API}/readings', params={
        'device': device_name,
        'start': start,
        'end': end,
        'limit': 10000
    })
    
    readings = response.json()
    
    # 写入 InfluxDB
    points = []
    for reading in readings:
        point = Point("readings") \
            .tag("device", reading['deviceName']) \
            .tag("resource", reading['resourceName']) \
            .field("value", float(reading['value'])) \
            .time(reading['timestamp'])
        points.append(point)
    
    write_api.write(bucket=INFLUX_BUCKET, org=INFLUX_ORG, record=points)
    print(f"Synced {len(points)} points to InfluxDB")
```

### Prometheus

#### 导出指标到 Prometheus

sfsEdgeStore 原生支持 Prometheus 格式：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'sfsedgestore'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

### TimescaleDB

#### 从 sfsEdgeStore 同步到 TimescaleDB

```sql
-- 创建超表
CREATE TABLE readings (
    time TIMESTAMPTZ NOT NULL,
    device_name TEXT NOT NULL,
    resource_name TEXT NOT NULL,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL
);

SELECT create_hypertable('readings', 'time');

-- 创建索引
CREATE INDEX ON readings (device_name, resource_name, time DESC);
```

```python
import psycopg2
import requests

def sync_to_timescale(device_name, start, end):
    # 从 sfsEdgeStore 查询
    readings = requests.get('http://localhost:8080/api/readings', params={
        'device': device_name,
        'start': start,
        'end': end
    }).json()
    
    # 写入 TimescaleDB
    conn = psycopg2.connect("dbname=sfsedgestore user=postgres")
    cur = conn.cursor()
    
    for reading in readings:
        cur.execute("""
            INSERT INTO readings (time, device_name, resource_name, value, value_type)
            VALUES (%s, %s, %s, %s, %s)
        """, (
            reading['timestamp'],
            reading['deviceName'],
            reading['resourceName'],
            reading['value'],
            reading['valueType']
        ))
    
    conn.commit()
    cur.close()
    conn.close()
```

## 可视化集成

### Grafana

#### 配置 sfsEdgeStore 数据源

1. 安装 JSON API 数据源插件
2. 配置数据源：
   - URL: `http://localhost:8080/api`
   - Access: Server

#### 创建仪表盘面板

```json
{
  "datasource": "sfsEdgeStore",
  "targets": [
    {
      "url": "/readings?device=device001&resource=temperature&limit=1000",
      "path": "/",
      "fields": [
        { "name": "time", "jsonPath": "$.timestamp" },
        { "name": "value", "jsonPath": "$.value" }
      ]
    }
  ]
}
```

### Grafana (Prometheus)

如果使用 Prometheus 集成，可以直接使用 Prometheus 数据源。

## IoT 平台集成

### AWS IoT Core

#### 规则引擎配置

```sql
SELECT 
    deviceName,
    resourceName,
    value,
    valueType,
    timestamp() AS ts
FROM 'iot/topic'
```

#### Lambda 函数转发

```javascript
const https = require('https');

exports.handler = async (event) => {
    const data = {
        deviceName: event.deviceName,
        resourceName: event.resourceName,
        value: String(event.value),
        valueType: event.valueType,
        timestamp: event.ts
    };
    
    return new Promise((resolve, reject) => {
        const options = {
            hostname: 'your-sfsedgestore.com',
            port: 443,
            path: '/api/readings',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer your-api-key'
            }
        };
        
        const req = https.request(options, (res) => {
            resolve({ statusCode: res.statusCode });
        });
        
        req.write(JSON.stringify(data));
        req.end();
    });
};
```

### Azure IoT Hub

#### 路由配置

```json
{
  "routes": {
    "routeToSfsEdgeStore": {
      "source": "DeviceMessages",
      "condition": "true",
      "endpointNames": ["sfsEdgeStoreEndpoint"],
      "isEnabled": true
    }
  },
  "endpoints": {
    "serviceBusQueues": [
      {
        "connectionString": "your-connection-string",
        "name": "sfsEdgeStoreEndpoint"
      }
    ]
  }
}
```

### Google Cloud IoT Core

#### Pub/Sub 到 sfsEdgeStore

```python
import base64
import json
import requests
from google.cloud import pubsub_v1

def callback(message):
    data = json.loads(base64.b64decode(message.data))
    
    # 发送到 sfsEdgeStore
    requests.post(
        'http://localhost:8080/api/readings',
        json={
            'deviceName': data['deviceId'],
            'resourceName': data['resource'],
            'value': str(data['value']),
            'valueType': data['valueType'],
            'timestamp': data['timestamp']
        }
    )
    
    message.ack()

subscriber = pubsub_v1.SubscriberClient()
subscription_path = subscriber.subscription_path('your-project', 'your-subscription')

subscriber.subscribe(subscription_path, callback=callback)
```

## 监控和告警集成

### Alertmanager

#### 告警路由配置

```yaml
route:
  receiver: 'sfsedgestore-alerts'
  routes:
    - match:
        service: 'sfsedgestore'
      receiver: 'sms-and-email'

receivers:
  - name: 'sfsedgestore-alerts'
    webhook_configs:
      - url: 'http://your-alert-handler/webhook'
```

### PagerDuty

#### 集成配置

```yaml
receivers:
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'your-service-key'
        description: '{{ .CommonAnnotations.description }}'
```

### Slack

#### Slack 通知

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/xxx/yyy/zzz'
        channel: '#alerts'
        username: 'sfsEdgeStore Alert'
        text: |-
          *{{ .Status | toUpper }}*: {{ .CommonAnnotations.summary }}
          {{ range .Alerts }}
          - {{ .Annotations.description }}
          {{ end }}
```

## 数据管道集成

### Apache NiFi

#### 流程配置

1. **GetMQTT** - 从 MQTT broker 接收数据
2. **ConvertJSONToSQL** - 转换为 SQL（可选）
3. **InvokeHTTP** - 发送到 sfsEdgeStore API

### Apache Flink

#### Flink Job 示例

```java
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.connectors.mqtt.FlinkMqttConsumer;

public class SfsEdgeStoreJob {
    public static void main(String[] args) throws Exception {
        StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();
        
        FlinkMqttConsumer<String> consumer = new FlinkMqttConsumer<>(
            "tcp://localhost:1883",
            "edgex/events/#",
            new SimpleStringSchema()
        );
        
        env.addSource(consumer)
            .map(SfsEdgeStoreMapper::map)
            .addSink(new SfsEdgeStoreSink());
        
        env.execute("sfsEdgeStore Integration");
    }
}
```

## CI/CD 集成

### GitHub Actions

```yaml
name: sfsEdgeStore Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Start sfsEdgeStore
        run: |
          docker run -d -p 8080:8080 sfsedgestore/sfsedgestore:latest
          sleep 10
      
      - name: Run integration tests
        run: |
          ./scripts/integration-tests.sh
```

### Jenkins

```groovy
pipeline {
    agent any
    
    stages {
        stage('Deploy sfsEdgeStore') {
            steps {
                sh 'docker-compose up -d'
                sh 'sleep 30'
            }
        }
        
        stage('Run Tests') {
            steps {
                sh './scripts/test-integration.sh'
            }
        }
    }
}
```

## 更多资源

- [系统集成指南](./INTEGRATION_GUIDE.md)
- [EdgeX 集成指南](./EDGEX_INTEGRATION.md)
- [API 文档](../api.md)
- [架构文档](../architecture/ARCHITECTURE.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
