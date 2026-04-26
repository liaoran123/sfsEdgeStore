# API 参考

sfsEdgeStore 完整 REST API 文档。

## 基础 URL

```
http://localhost:8081
```

## 健康与状态

### 健康检查

```http
GET /health
GET /healthz
```

响应：
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "1h23m45s"
}
```

### 就绪检查

```http
GET /ready
```

### 系统状态

```http
GET /api/status
```

### 资源状态

```http
GET /api/resources/status
```

### 指标（Prometheus 格式）

```http
GET /metrics
```

## 数据 API

### 查询读数

```http
GET /api/readings
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceName | string | 否 | 按设备名称过滤 |
| startTime | string | 否 | ISO 8601 时间戳 |
| endTime | string | 否 | ISO 8601 时间戳 |
| limit | int | 否 | 限制结果数量 |

示例：
```bash
curl "http://localhost:8081/api/readings?deviceName=Device001&limit=10"
```

响应：
```json
{
  "count": 10,
  "readings": [...]
}
```

### 设备状态

```http
GET /api/device-status
```

## 配置 API

### 获取配置

```http
GET /api/config/get
```

### 更新配置

```http
POST /api/config/update
```

### 重载配置

```http
POST /api/config/reload
```

### 一键配置

```http
POST /api/config/oneclick
```

## 备份与恢复

### 创建备份

```http
POST /api/backup
```

| 参数 | 类型 | 说明 |
|------|------|------|
| path | string | 备份目录（默认：`./backups`） |

### 从备份恢复

```http
POST /api/restore
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | string | 是 | 备份文件路径 |

## 数据导入/导出

### 导出 CSV

```http
GET /api/export/csv
```

### 导出 JSON

```http
GET /api/export/json
```

### 导出 SQL

```http
GET /api/export/sql
```

### 导入 CSV

```http
POST /api/import/csv
```

### 导入 JSON

```http
POST /api/import/json
```

## 监控与告警

### 订阅状态

```http
GET /api/subscription/status
```

### 测试订阅

```http
POST /api/subscription/test
```

### 订阅主题列表

```http
GET /api/subscription/themes
```

### 告警列表

```http
GET /api/alerts
```

### 告警分组

```http
GET /api/alert-groups
```

### 告警通知器状态

```http
GET /api/alerts/notifier/status
```

### 测试告警

```http
POST /api/alerts/test
```

## 数据保留

### 保留状态

```http
GET /api/retention/status
```

### 手动清理

```http
POST /api/retention/cleanup
```

## 模板

### 模板列表

```http
GET /api/templates
```

### 应用模板

```http
POST /api/templates/apply
```

请求体：
```json
{
  "industry": "motor"
}
```

## 基线

### 基线列表

```http
GET /api/baselines
```

### 计算基线

```http
POST /api/baselines/calculate
```

请求体：
```json
{
  "deviceName": "temperature-sensor-001",
  "readingName": "temperature"
}
```

## 认证

### 创建 API Key

```http
POST /api/auth/create-key
```

### 列出 API Keys

```http
GET /api/auth/list-keys
```

### 撤销 API Key

```http
POST /api/auth/revoke-key
```

## 加密

### 加密状态

```http
GET /api/encryption/status
```

### 轮换加密密钥

```http
POST /api/encryption/rotate-key
```

## 许可证

### 许可证信息

```http
GET /api/license
```

## MQTT 配置

### 更新 MQTT 配置

```http
POST /api/config/mqtt
```

## WebSocket

### 实时数据流

```
WS /ws
```

通过 WebSocket 连接获取实时数据流。
