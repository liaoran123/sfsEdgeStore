# 监控指南

## 概述

本文档详细介绍如何监控 sfsEdgeStore 的运行状态、性能指标和系统健康度。

## 监控架构

```
┌─────────────────┐
│  sfsEdgeStore   │
│  (Metrics API)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐    ┌─────────────────┐
│  Prometheus     │───▶│  Grafana        │
│  (采集/存储)    │    │  (可视化)       │
└─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐
│  Alertmanager   │
│  (告警)         │
└─────────────────┘
```

## 内置监控端点

### 健康检查端点

#### 基本健康检查

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": 1620000000000,
  "version": "1.0.0"
}
```

#### 详细健康检查

```bash
curl http://localhost:8080/health/detailed
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": 1620000000000,
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "latency": 5
    },
    "mqtt": {
      "status": "healthy",
      "connected": true
    },
    "queue": {
      "status": "healthy",
      "length": 0
    }
  }
}
```

### 指标端点

#### Prometheus 格式

```bash
curl http://localhost:8080/metrics
```

#### JSON 格式

```bash
curl http://localhost:8080/metrics/json
```

## 关键监控指标

### 系统指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `system_cpu_usage` | Gauge | CPU 使用率 (%) |
| `system_memory_usage` | Gauge | 内存使用量 (bytes) |
| `system_disk_usage` | Gauge | 磁盘使用量 (bytes) |
| `system_disk_available` | Gauge | 磁盘可用空间 (bytes) |
| `system_uptime` | Counter | 系统运行时间 (seconds) |

### HTTP 指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `http_requests_total` | Counter | HTTP 请求总数 |
| `http_request_duration_seconds` | Histogram | HTTP 请求耗时分布 |
| `http_requests_in_flight` | Gauge | 进行中的 HTTP 请求数 |
| `http_errors_total` | Counter | HTTP 错误总数 |

### 数据库指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `db_operations_total` | Counter | 数据库操作总数 |
| `db_operation_duration_seconds` | Histogram | 数据库操作耗时 |
| `db_size_bytes` | Gauge | 数据库大小 |
| `db_compactions_total` | Counter | 数据库压缩次数 |

### 队列指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `queue_length` | Gauge | 队列当前长度 |
| `queue_messages_total` | Counter | 队列消息总数 |
| `queue_processing_duration_seconds` | Histogram | 消息处理耗时 |
| `queue_dropped_total` | Counter | 丢弃的消息数 |

### MQTT 指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `mqtt_connected` | Gauge | MQTT 连接状态 (1=连接, 0=断开) |
| `mqtt_messages_received_total` | Counter | 接收的 MQTT 消息总数 |
| `mqtt_messages_published_total` | Counter | 发布的 MQTT 消息总数 |
| `mqtt_connection_errors_total` | Counter | MQTT 连接错误数 |

### 业务指标

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `readings_written_total` | Counter | 写入的读数总数 |
| `readings_query_total` | Counter | 查询请求总数 |
| `devices_active` | Gauge | 活跃设备数 |
| `alerts_active` | Gauge | 活跃告警数 |

## Prometheus 集成

### 配置 Prometheus

在 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'sfsedgestore'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

### 验证配置

```bash
# 访问 Prometheus UI
# http://localhost:9090

# 查询指标
up{job="sfsedgestore"}
```

## Grafana 可视化

### 导入仪表盘

1. 打开 Grafana UI
2. 导航到 Dashboards → Import
3. 导入 sfsEdgeStore 仪表盘 JSON
4. 配置 Prometheus 数据源

### 推荐仪表盘

1. **系统概览仪表盘**
   - 系统资源使用
   - 请求量和响应时间
   - 错误率

2. **数据库仪表盘**
   - 数据库操作统计
   - 存储使用趋势
   - 慢查询分析

3. **MQTT/队列仪表盘**
   - 消息吞吐量
   - 队列长度
   - 处理延迟

## 告警配置

### 告警规则示例

创建 `alerts.yml`：

```yaml
groups:
  - name: sfsedgestore_alerts
    interval: 30s
    rules:
      - alert: ServiceDown
        expr: up{job="sfsedgestore"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "sfsEdgeStore 服务不可用"
          description: "sfsEdgeStore 服务已停止超过 1 分钟"

      - alert: HighCPUUsage
        expr: system_cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU 使用率过高"
          description: "CPU 使用率已超过 80% 达 5 分钟"

      - alert: HighMemoryUsage
        expr: system_memory_usage > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "内存使用率过高"
          description: "内存使用率已超过 85%"

      - alert: HighQueueLength
        expr: queue_length > 1000
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "队列长度过长"
          description: "消息队列积压超过 1000 条"

      - alert: DiskSpaceLow
        expr: system_disk_available < 10737418240
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "磁盘空间不足"
          description: "可用磁盘空间少于 10GB"

      - alert: HighErrorRate
        expr: rate(http_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "错误率过高"
          description: "HTTP 错误率超过 10%"
```

### Alertmanager 配置

```yaml
route:
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical'

receivers:
  - name: 'default'
    email_configs:
      - to: 'admin@example.com'
        send_resolved: true

  - name: 'critical'
    email_configs:
      - to: 'oncall@example.com'
    webhook_configs:
      - url: 'http://alert-webhook.example.com'
```

## 日志监控

### 关键日志模式

监控以下日志模式：

```
ERROR - 错误日志
WARN - 警告日志
connection lost - 连接丢失
queue full - 队列满
disk full - 磁盘满
```

### 使用 ELK Stack

#### Filebeat 配置

```yaml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /path/to/sfsedgestore/logs/*.log
    fields:
      service: sfsedgestore

output.elasticsearch:
  hosts: ["localhost:9200"]
```

## 自定义监控脚本

### 健康检查脚本

```bash
#!/bin/bash

SERVICE_URL="http://localhost:8080/health"

response=$(curl -s -o /dev/null -w "%{http_code}" $SERVICE_URL)

if [ $response -eq 200 ]; then
    echo "sfsEdgeStore is healthy"
    exit 0
else
    echo "sfsEdgeStore is unhealthy (HTTP $response)"
    exit 1
fi
```

### 性能监控脚本

```bash
#!/bin/bash

echo "=== sfsEdgeStore Performance Metrics ==="
echo "Timestamp: $(date)"
echo ""

curl -s http://localhost:8080/metrics/json | jq '.'
```

## 最佳实践

1. **分层监控**
   - 基础设施层（CPU、内存、磁盘）
   - 应用层（请求、错误、延迟）
   - 业务层（读数、设备、告警）

2. **设置合理的阈值**
   - 基于历史数据
   - 考虑业务影响
   - 避免告警风暴

3. **定期审查**
   - 每周审查告警配置
   - 每月审查监控指标
   - 及时调整阈值

4. **文档化**
   - 记录告警处理流程
   - 维护监控仪表盘
   - 更新监控文档

## 更多资源

- [运维操作手册](./OPERATIONS_GUIDE.md)
- [管理员指南](./ADMIN_GUIDE.md)
- [架构文档](../architecture/ARCHITECTURE.md)
- [故障排除](../user-guide/TROUBLESHOOTING.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
