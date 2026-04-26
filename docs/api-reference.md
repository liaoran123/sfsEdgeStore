# API Reference

Complete REST API documentation for sfsEdgeStore.

## Base URL

```
http://localhost:8081
```

## Health & Status

### Health Check

```http
GET /health
GET /healthz
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "1h23m45s"
}
```

### Readiness Check

```http
GET /ready
```

### System Status

```http
GET /api/status
```

### Resource Status

```http
GET /api/resources/status
```

### Metrics (Prometheus format)

```http
GET /metrics
```

## Data API

### Query Readings

```http
GET /api/readings
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| deviceName | string | No | Filter by device name |
| startTime | string | No | ISO 8601 timestamp |
| endTime | string | No | ISO 8601 timestamp |
| limit | int | No | Limit result count |

Example:
```bash
curl "http://localhost:8081/api/readings?deviceName=Device001&limit=10"
```

Response:
```json
{
  "count": 10,
  "readings": [...]
}
```

### Device Status

```http
GET /api/device-status
```

## Configuration API

### Get Configuration

```http
GET /api/config/get
```

### Update Configuration

```http
POST /api/config/update
```

### Reload Configuration

```http
POST /api/config/reload
```

### One-Click Configuration

```http
POST /api/config/oneclick
```

## Backup & Restore

### Create Backup

```http
POST /api/backup
```

| Parameter | Type | Description |
|-----------|------|-------------|
| path | string | Backup directory (default: `./backups`) |

### Restore from Backup

```http
POST /api/restore
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| file | string | Yes | Path to backup file |

## Data Export/Import

### Export CSV

```http
GET /api/export/csv
```

### Export JSON

```http
GET /api/export/json
```

### Export SQL

```http
GET /api/export/sql
```

### Import CSV

```http
POST /api/import/csv
```

### Import JSON

```http
POST /api/import/json
```

## Monitoring & Alerts

### Subscription Status

```http
GET /api/subscription/status
```

### Test Subscription

```http
POST /api/subscription/test
```

### Subscription Themes

```http
GET /api/subscription/themes
```

### Alerts

```http
GET /api/alerts
```

### Alert Groups

```http
GET /api/alert-groups
```

### Alert Notifier Status

```http
GET /api/alerts/notifier/status
```

### Test Alert

```http
POST /api/alerts/test
```

## Data Retention

### Retention Status

```http
GET /api/retention/status
```

### Manual Cleanup

```http
POST /api/retention/cleanup
```

## Templates

### List Templates

```http
GET /api/templates
```

### Apply Template

```http
POST /api/templates/apply
```

Body:
```json
{
  "industry": "motor"
}
```

## Baselines

### List Baselines

```http
GET /api/baselines
```

### Calculate Baseline

```http
POST /api/baselines/calculate
```

Body:
```json
{
  "deviceName": "temperature-sensor-001",
  "readingName": "temperature"
}
```

## Authentication

### Create API Key

```http
POST /api/auth/create-key
```

### List API Keys

```http
GET /api/auth/list-keys
```

### Revoke API Key

```http
POST /api/auth/revoke-key
```

## Encryption

### Encryption Status

```http
GET /api/encryption/status
```

### Rotate Encryption Key

```http
POST /api/encryption/rotate-key
```

## License

### License Information

```http
GET /api/license
```

## MQTT Configuration

### Update MQTT Configuration

```http
POST /api/config/mqtt
```

## WebSocket

### Real-time Data Stream

```
WS /ws
```

Connect via WebSocket for real-time data streaming.
