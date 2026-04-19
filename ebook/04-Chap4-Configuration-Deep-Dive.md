# Chapter 4: Configuration Deep Dive

## 4.1 Database Scenario Configurations

sfsDb provides three pre-configured scenarios optimized for different use cases. Understanding when to use each scenario is crucial for optimal performance.

### Scenario Overview

| Scenario | Best For | Write Performance | Read Performance | Memory Usage |
|----------|----------|------------------|------------------|--------------|
| `edge` | IoT gateways, edge devices | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| `standard` | General workloads | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| `analytics` | Data analysis, dashboards | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

### Edge Scenario (Recommended for sfsEdgeStore)

The `edge` scenario is optimized for the constrained resources and write-heavy workloads typical of IoT gateways:

```json
{
  "DBScenario": "edge",
  "DBPath": "/var/lib/sfsedgestore/data",
  "DBCompression": true
}
```

**Configuration Details:**

```go
// From sfsDb scenario configuration
case ScenarioEdge:
    return ScenarioOptions{
        WriteBuffer:       4 * 1024 * 1024,   // 4MB MemTable
        OpenFilesCapacity: 100,               // Keep 100 files open
        BlockSize:         4096,              // 4KB blocks
        Compression:       true,               // Snappy compression
    }
```

**Why This Configuration:**

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| `WriteBuffer` | 4MB | Balance between memory and flush frequency |
| `OpenFiles` | 100 | Enough for active data without exhausting file handles |
| `BlockSize` | 4KB | Optimal for small IoT readings (typically < 1KB) |
| `Compression` | true | 60-70% compression ratio, saves storage and I/O |

### Standard Scenario

Use for balanced workloads where you need decent performance for both reads and writes:

```json
{
  "DBScenario": "standard",
  "DBPath": "/var/lib/sfsedgestore/data"
}
```

**Configuration Details:**

```go
case ScenarioStandard:
    return ScenarioOptions{
        WriteBuffer:       4 * 1024 * 1024,   // 4MB
        OpenFilesCapacity: 1000,              // More files for larger datasets
        BlockSize:         4096,              // 4KB
        Compression:       true,              // Still compress
    }
```

### Analytics Scenario

Optimized for read-heavy workloads like dashboards and historical analysis:

```json
{
  "DBScenario": "analytics",
  "DBPath": "/var/lib/sfsedgestore/analytics",
  "DBCompression": false
}
```

**Configuration Details:**

```go
case ScenarioAnalytics:
    return ScenarioOptions{
        WriteBuffer:       8 * 1024 * 1024,   // 8MB - larger buffer
        OpenFilesCapacity: 5000,              // Much more files
        BlockSize:         8192,              // 8KB blocks - larger for better scan performance
        Compression:       false,             // No compression - faster reads
    }
```

### Custom Scenario

For fine-grained control, you can override individual parameters:

```json
{
  "DBScenario": "edge",
  "DBCustomOptions": {
    "WriteBufferSize": 8388608,
    "MaxOpenFiles": 200,
    "BlockSize": 8192,
    "Compression": true
  }
}
```

## 4.2 MQTT Connection Settings

MQTT is the lifeline between EdgeX and sfsEdgeStore. Proper configuration ensures reliable data ingestion.

### Basic MQTT Configuration

```json
{
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-001",
  "MQTTQoS": 1,
  "MQTTKeepAlive": 30
}
```

**Parameter Explanation:**

| Parameter | Description | Recommended Value | Notes |
|-----------|-------------|-------------------|-------|
| `MQTTBroker` | Broker URL | - | Must match EdgeX export config |
| `MQTTTopic` | Subscription topic | `edgex/events` | Must match EdgeX publish topic |
| `ClientID` | Unique client ID | Auto-generated | Should be unique per instance |
| `MQTTQoS` | Quality of Service | 1 | 0=at most once, 1=at least once, 2=exactly once |
| `MQTTKeepAlive` | Keep-alive interval | 30-60 seconds | Longer for unstable networks |

### QoS Levels Explained

**QoS 0 - At Most Once:**
```yaml
# Fire and forget
# Suitable for: High-frequency data where loss is acceptable
# Example: Sensor readings, metrics

client.Publish(topic, 0, false, payload)
```

**QoS 1 - At Least Once (Recommended):**
```yaml
# Acknowledged delivery
# Suitable for: Most IoT data
# Guarantees: Message will be delivered, may be duplicated

client.Publish(topic, 1, false, payload)
```

**QoS 2 - Exactly Once:**
```yaml
# Handshake-based delivery
# Suitable for: Critical data, commands
# Cost: Higher latency, more overhead

client.Publish(topic, 2, false, payload)
```

### TLS/SSL Configuration

For production environments, enable TLS encryption:

```json
{
  "MQTTBroker": "ssl://mqtt.example.com:8883",
  "MQTTUseTLS": true,
  "MQTTCACert": "/etc/sfsedgestore/certs/ca.crt",
  "MQTTClientCert": "/etc/sfsedgestore/certs/client.crt",
  "MQTTClientKey": "/etc/sfsedgestore/certs/client.key"
}
```

**Certificate Setup:**

```bash
# Generate self-signed certificates for testing
mkdir -p /etc/sfsedgestore/certs

# CA Certificate (for production, use a real CA)
openssl req -x509 -newkey rsa:2048 -keyout ca.key -out ca.crt -days 365 -nodes

# Client Certificate
openssl req -newkey rsa:2048 -keyout client.key -out client.csr -nodes
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365

# Set permissions
chmod 600 /etc/sfsedgestore/certs/*
```

### Connection Reliability Features

```json
{
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "MQTTQoS": 1,
  "MQTTKeepAlive": 60,
  "MQTTUseTLS": false,

  "MQTTConnectionOptions": {
    "CleanSession": false,
    "AutoReconnect": true,
    "MaxReconnectInterval": 300,
    "ConnectTimeout": 10
  }
}
```

**Key Reliability Settings:**

```go
// From mqtt/client.go - NewClient function
opts.SetCleanSession(false)              // Persist subscription across reconnects
opts.SetAutoReconnect(true)               // Automatic reconnection
opts.SetMaxReconnectInterval(5 * time.Minute)  // Backoff on repeated failures
```

## 4.3 Security and Encryption Options

Data security is critical for production deployments. sfsEdgeStore supports encryption at rest.

### Database Encryption

```json
{
  "DBUseEncryption": true,
  "DBEncryptionKey": "your-32-character-encryption-key!!",
  "DBEncryptionAlgorithm": "AES-256"
}
```

**Important Security Notes:**

1. **Key Length**: AES-256 requires exactly 32 characters
2. **Key Storage**: Never commit keys to version control
3. **Key Management**: Use environment variables or secrets management

```bash
# Using environment variable (recommended)
export DB_ENCRYPTION_KEY="your-32-character-encryption-key!!"

# In config.json, reference the environment variable
# sfsEdgeStore will read from DB_ENCRYPTION_KEY env var
```

### Encryption Configuration Example

```json
{
  "DBPath": "/var/lib/sfsedgestore/data",
  "DBUseEncryption": true,
  "DBEncryptionKey": "${DB_ENCRYPTION_KEY}",
  "DBEncryptionAlgorithm": "AES-256",
  "DBCompression": true
}
```

### HTTP Server Security

For production, enable HTTPS:

```json
{
  "HTTPPort": "8443",
  "HTTPUseTLS": true,
  "HTTPCert": "/etc/sfsedgestore/certs/server.crt",
  "HTTPKey": "/etc/sfsedgestore/certs/server.key"
}
```

### API Key Authentication

Protect API endpoints with API key authentication:

```json
{
  "AuthEnabled": true,
  "RequireAPIKey": true
}
```

**API Key Management:**

```bash
# Create API key via API
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 8760
  }'

# Use API key in requests
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/readings?deviceName=Device001"
```

## 4.4 Retention and Data Management

Proper data retention policies prevent storage exhaustion and help manage costs.

### Retention Configuration

```json
{
  "EnableRetention": true,
  "RetentionDays": 30,
  "RetentionCheckInterval": "24h"
}
```

**How Retention Works:**

```go
// From retention/retention.go - RetentionManager
func (rm *RetentionManager) Start() error {
    ticker := time.NewTicker(rm.checkInterval)
    go func() {
        for {
            <-ticker.C
            deleted, err := rm.cleanupExpiredData()
            if err != nil {
                log.Printf("Retention cleanup error: %v", err)
            }
            log.Printf("Retention cleanup: deleted %d expired records", deleted)
        }
    }()
    return nil
}
```

### Data Export Before Deletion

Always export data before the retention period expires:

```bash
# Export before retention deletes it
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/export/json?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z" \
  -o backup_2024_01.json
```

### Backup Configuration

```json
{
  "EnableBackup": true,
  "BackupPath": "/var/backups/sfsedgestore",
  "BackupSchedule": "0 2 * * *",
  "BackupRetentionDays": 90
}
```

## 4.5 Alert Configuration

Configure alerts to monitor system health and respond to issues proactively.

### Alert Channels

```json
{
  "EnableAlertNotifications": true,
  "AlertNotificationChannels": ["mqtt", "webhook", "log"],
  "AlertMinSeverity": "warning"
}
```

**Alert Severity Levels:**

| Level | Description | Example |
|-------|-------------|---------|
| `info` | Informational | Normal operations |
| `warning` | Attention needed | High memory usage |
| `critical` | Immediate action | Database failure |

### MQTT Alerts

```json
{
  "AlertMQTTTopic": "edgex/alerts"
}
```

### Webhook Alerts

```json
{
  "AlertWebhookURL": "https://your-monitoring-system.com/webhook"
}
```

**Webhook Payload Format:**

```json
{
  "type": "high_error_rate",
  "message": "Error rate too high: 15 errors per minute",
  "severity": "critical",
  "timestamp": "2024-01-15T10:30:00Z",
  "source": "sfsedgestore-001"
}
```

## 4.6 Performance Tuning

Fine-tune for your specific workload.

### High-Frequency Data Scenarios

For 1000+ readings per second:

```json
{
  "DBScenario": "edge",
  "MQTTQoS": 1,
  "BatchSize": 1000,
  "BatchInterval": "100ms"
}
```

### Low-Power Scenarios

For battery-powered or thermal-constrained devices:

```json
{
  "DBScenario": "edge",
  "MQTTKeepAlive": 300,
  "EnableAnalytics": false,
  "LogLevel": "error"
}
```

## 4.7 Environment-Specific Configurations

### Raspberry Pi 3/4

```json
{
  "HTTPPort": "8081",
  "DBPath": "/var/lib/sfsedgestore/data",
  "DBScenario": "edge",
  "DBCompression": true,
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-pi",
  "EnableSimulator": false,
  "EnableAnalyzer": false,
  "EnableRetention": true,
  "RetentionDays": 14,
  "LogLevel": "info"
}
```

### Industrial Gateway

```json
{
  "HTTPPort": "8081",
  "HTTPUseTLS": true,
  "HTTPCert": "/etc/sfsedgestore/certs/server.crt",
  "HTTPKey": "/etc/sfsedgestore/certs/server.key",

  "DBPath": "/var/lib/sfsedgestore/data",
  "DBScenario": "standard",
  "DBUseEncryption": true,
  "DBEncryptionKey": "${DB_ENCRYPTION_KEY}",

  "MQTTBroker": "ssl://mqtt.factory.local:8883",
  "MQTTTopic": "factory/edgex/events",
  "ClientID": "sfsedgestore-gateway-001",
  "MQTTQoS": 2,
  "MQTTKeepAlive": 60,
  "MQTTUseTLS": true,

  "EnableAnalyzer": true,
  "EnableRetention": true,
  "RetentionDays": 90,

  "EnableAlertNotifications": true,
  "AlertNotificationChannels": ["mqtt", "webhook"],
  "AlertWebhookURL": "https://monitoring.factory.local/alerts",

  "EnableSync": true,
  "SyncEndpoint": "https://cloud.factory.local/api/sync",

  "LogLevel": "info"
}
```

### Docker Development

```json
{
  "HTTPPort": "8081",
  "DBPath": "/app/data",
  "DBScenario": "edge",
  "MQTTBroker": "tcp://mosquitto:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-dev",
  "EnableSimulator": true,
  "EnableAnalyzer": false,
  "EnableRetention": false,
  "LogLevel": "debug"
}
```

## 4.8 Configuration Best Practices

### 1. Use Environment Variables for Secrets

```bash
# Never put real keys in config files
export DB_ENCRYPTION_KEY="your-production-key-32chars!!"
export MQTT_PASSWORD="your-mqtt-password"
```

```json
{
  "DBEncryptionKey": "${DB_ENCRYPTION_KEY}",
  "MQTTPassword": "${MQTT_PASSWORD}"
}
```

### 2. Separate Development and Production

```bash
# Development
cp config.example.json config.dev.json
# Edit for local development

# Production
cp config.example.json config.prod.json
# Edit for production, never commit secrets
```

### 3. Validate Configuration on Startup

sfsEdgeStore validates configuration at startup:

```bash
./sfsedgestore --validate-config config.json
# Will print errors if configuration is invalid
```

### 4. Configuration Migration

When upgrading sfsEdgeStore:

```bash
# Backup current config
cp /etc/sfsedgestore/config.json /etc/sfsedgestore/config.json.bak

# Check migration guide
cat /path/to/sfsedgestore/MIGRATION.md

# Update config with new parameters
# Restart service
sudo systemctl restart sfsedgestore
```

## 4.9 Chapter Summary

This chapter covered all configuration options in detail.

**Key Takeaways:**

1. **Use the `edge` scenario** for most sfsEdgeStore deployments - it's optimized for IoT workloads.

2. **MQTT QoS 1** is recommended for most scenarios - balances reliability and performance.

3. **Enable encryption** in production - protect data at rest with AES-256.

4. **Set retention policies** - prevent storage exhaustion, export before deleting.

5. **Use environment variables** for secrets - never commit keys to version control.

**What's Next:**

➡️ Next: [Chapter 5: Common Issues and Solutions](./05-Chap5-Common-Issues-Solutions.md)

---

**Quick Configuration Checklist:**

```markdown
□ Database scenario set correctly
□ MQTT broker and topic match EdgeX config
□ ClientID is unique per instance
□ Encryption enabled for production
□ Retention policy configured
□ Alerts set up for critical metrics
□ Log level appropriate for environment
□ Environment variables used for secrets
□ Configuration backed up
```