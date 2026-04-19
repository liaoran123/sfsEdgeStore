# Appendix A: Complete API Reference

## A.1 HTTP API Endpoints

### Health Check

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/healthz` | GET | Health check |
| `/ready` | GET | Readiness check |

**Example:**
```bash
curl http://localhost:8081/health
```

**Response:**
```json
{
  "status": "healthy",
  "uptime": "1h23m45s",
  "version": "1.0.0"
}
```

### Data Query

| Endpoint | Method | Description | Parameters |
|----------|--------|-------------|------------|
| `/api/readings` | GET | Query readings | `deviceName`, `startTime`, `endTime`, `limit` |

**Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/readings?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-02T00:00:00Z&limit=100"
```

**Response:**
```json
{
  "records": [
    {
      "id": "reading-001",
      "deviceName": "Device001",
      "reading": "temperature",
      "value": 25.5,
      "valueType": "Float32",
      "timestamp": 1704067200000000000
    }
  ],
  "count": 1
}
```

### Metrics

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/metrics` | GET | Get monitoring metrics |

**Example:**
```bash
curl http://localhost:8081/metrics
```

**Response:**
```json
{
  "system": {
    "cpu_usage": 2.9,
    "memory_usage": 20.85,
    "goroutines": 15,
    "uptime_seconds": 5025
  },
  "application": {
    "mqtt_messages_received": 18681,
    "mqtt_messages_processed": 18681,
    "http_requests": 150,
    "database_operations": 187,
    "errors": 0
  }
}
```

### Alerts

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/alerts` | GET | Get current alerts |

**Example:**
```bash
curl http://localhost:8081/alerts
```

**Response:**
```json
{
  "alerts": [
    {
      "type": "high_memory",
      "message": "Memory usage exceeded threshold",
      "severity": "warning",
      "timestamp": "2024-01-15T10:30:00Z",
      "resolved": false
    }
  ]
}
```

### Backup and Restore

| Endpoint | Method | Description | Parameters |
|----------|--------|-------------|------------|
| `/api/backup` | POST | Backup database | `path` |
| `/api/restore` | POST | Restore database | `path` |

**Backup Example:**
```bash
curl -X POST -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/backup?path=/backups"
```

### Data Export

| Endpoint | Method | Description | Parameters |
|----------|--------|-------------|------------|
| `/api/export/csv` | GET | Export to CSV | `deviceName`, `startTime`, `endTime` |
| `/api/export/json` | GET | Export to JSON | `deviceName`, `startTime`, `endTime` |
| `/api/export/sql` | GET | Export to SQL | `deviceName`, `startTime`, `endTime` |

**Export Example:**
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/export/json?deviceName=Device001" \
  -o backup.json
```

### Authentication

| Endpoint | Method | Description | Body |
|----------|--------|-------------|------|
| `/api/auth/create-key` | POST | Create API key | `user_id`, `role`, `expires_in` |
| `/api/auth/list-keys` | GET | List API keys | - |
| `/api/auth/revoke-key` | POST | Revoke API key | `key` |

**Create Key Example:**
```bash
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "role": "user",
    "expires_in": 8760
  }'
```

**Response:**
```json
{
  "id": "key-001",
  "key": "a1b2c3d4e5f6...",
  "user_id": "user1",
  "role": "user",
  "expires_at": "2027-01-01T00:00:00Z",
  "active": true
}
```

### Encryption

| Endpoint | Method | Description | Body |
|----------|--------|-------------|------|
| `/api/encryption/rotate-key` | POST | Rotate encryption key | `new_key` |
| `/api/encryption/status` | GET | Get encryption status | - |

### Retention

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/retention/status` | GET | Get retention policy status |
| `/api/retention/cleanup` | POST | Trigger manual cleanup |

### Alerts Notification

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/alerts/notifier/status` | GET | Get notifier status |
| `/api/alerts/test` | POST | Send test alert |

### Sync

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/sync/status` | GET | Get sync status |
| `/api/sync/start` | POST | Start sync |
| `/api/sync/database` | POST | Sync from database |

### Configuration

| Endpoint | Method | Description | Body |
|----------|--------|-------------|------|
| `/api/config/get` | GET | Get configuration | - |
| `/api/config/update` | POST | Update configuration | `config` |
| `/api/config/reload` | POST | Reload configuration | - |

### Resources

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/resources/status` | GET | Get resource monitoring status |

## A.2 Configuration Reference

### Complete Configuration Schema

```json
{
  "HTTPPort": "string",
  "HTTPUseTLS": "boolean",
  "HTTPCert": "string",
  "HTTPKey": "string",

  "DBPath": "string",
  "DBScenario": "string (edge|standard|analytics)",
  "DBUseEncryption": "boolean",
  "DBEncryptionKey": "string",
  "DBEncryptionAlgorithm": "string",
  "DBCompression": "boolean",
  "DBCustomOptions": {
    "WriteBufferSize": "number",
    "MaxOpenFiles": "number",
    "BlockSize": "number",
    "Compression": "boolean"
  },

  "MQTTBroker": "string",
  "MQTTTopic": "string",
  "ClientID": "string",
  "MQTTQoS": "number (0|1|2)",
  "MQTTKeepAlive": "number",
  "MQTTUseTLS": "boolean",
  "MQTTCACert": "string",
  "MQTTClientCert": "string",
  "MQTTClientKey": "string",
  "MQTTConnectionOptions": {
    "CleanSession": "boolean",
    "AutoReconnect": "boolean",
    "MaxReconnectInterval": "number",
    "ConnectTimeout": "number"
  },

  "EnableSimulator": "boolean",
  "EnableAnalyzer": "boolean",
  "EnableRetention": "boolean",
  "RetentionDays": "number",
  "RetentionCheckInterval": "string",

  "EnableAlertNotifications": "boolean",
  "AlertNotificationChannels": ["mqtt", "webhook", "log"],
  "AlertMinSeverity": "string (info|warning|critical)",
  "AlertMQTTTopic": "string",
  "AlertWebhookURL": "string",

  "EnableSync": "boolean",
  "SyncEndpoint": "string",

  "LogLevel": "string (debug|info|warn|error)"
}
```

## A.3 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_ENCRYPTION_KEY` | Database encryption key | `your-32-char-key!!!` |
| `MQTT_PASSWORD` | MQTT password | `mqtt_password` |
| `HTTP_USERNAME` | HTTP basic auth username | `admin` |
| `HTTP_PASSWORD` | HTTP basic auth password | `password123` |
| `LICENSE_KEY` | sfsEdgeStore license key | `XXXX-XXXX-XXXX` |

## A.4 Exit Codes

| Code | Description |
|------|-------------|
| 0 | Normal exit |
| 1 | General error |
| 2 | Configuration error |
| 3 | Database error |
| 4 | MQTT connection error |
| 5 | HTTP server error |

---

➡️ Next: [Appendix B: Benchmark Test Results](./Appendix-B-Benchmark-Results.md)

➡️ Back to [Chapter 8](./08-Chap8-Product-Commercial-Services.md)