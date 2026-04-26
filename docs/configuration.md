# Configuration

Complete configuration reference for sfsEdgeStore.

## Configuration Methods (Priority Order)

1. **Environment Variables** (highest priority)
2. **Configuration File** (`config.json`)
3. **Default Values**

## Configuration File

Create `config.json` in the application root directory.

## Common Configuration (Start Here)

Most customers only need to customize a few key settings. Here are the most frequently modified configurations, ordered by priority.

### Must Customize (Almost Every Deployment)

| Key | Default | Typical Value | Why Change |
|-----|---------|---------------|------------|
| `mqtt_broker` | `tcp://localhost:1883` | Your MQTT server address | Each customer has their own MQTT broker |
| `db_path` | `data/sfs.db` | `/var/lib/sfsedgestore/data/sfs.db` | Production usually requires fixed storage paths |
| `http_port` | `8081` | Available port | May conflict with existing services |

> **Note:** MQTT Topic is managed through the **Topic Subscription** page (`/mqtt-subscription`) in the web dashboard.

### Often Customize (Resource-Constrained or Compliance)

| Key | Default | Why Change |
|-----|---------|------------|
| `enable_retention_policy` + `retention_days` | `false` / `30` | Limited edge device disk, need to control data volume |
| `db_scenario` | `edge` | Match hardware performance (embedded/IoT/edge/game) |
| `enable_resource_monitoring` + `max_memory_mb` | `false` / `45` | Resource-constrained devices need tuning |
| `db_use_encryption` + `db_encryption_key` | `false` | GDPR compliance, healthcare, and other regulated industries |

### Advanced (Enterprise / Special Scenarios)

| Key | Default | Why Change |
|-----|---------|------------|
| `mqtt_use_tls` + certificate paths | `false` | Secure transport required |
| `http_use_tls` + certificate paths | `false` | HTTPS required for external access |
| `enable_data_sync` | `false` | Enterprise cloud sync feature |
| `enable_alert_notifications` | `false` | Need alerting via MQTT/Webhook |
| `enable_prometheus` | `false` | Integrate with Prometheus monitoring |
| `custom_topics` | `[]` | Subscribe to multiple custom topics |

### Rarely Customize (Defaults are Fine)

| Key | Why Not Change |
|-----|----------------|
| `client_id` | Auto-generated with timestamp |
| `analyzer_*` parameters | Default values already optimized |
| `simulator_*` parameters | Only for development/testing |
| `cleanup_batch_size` | 1000 is sufficient for most cases |
| `connection_timeout` / `keep_alive` | Default values work universally |

### Quick Start Example

Minimal configuration for a production deployment:

```json
{
  "mqtt_broker": "tcp://192.168.1.100:1883",
  "mqtt_topic": "edgex/events/#",
  "db_path": "/var/lib/sfsedgestore/data/sfs.db",
  "http_port": "8081",
  "enable_retention_policy": true,
  "retention_days": 30
}
```

## Core Configuration

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `db_path` | string | `EDGEX_DB_PATH` | `data/sfs.db` | Database storage path |
| `mqtt_broker` | string | `EDGEX_MQTT_BROKER` | `tcp://localhost:1883` | MQTT broker URL |
| `mqtt_topic` | string | `EDGEX_MQTT_TOPIC` | `edgex/events/#` | MQTT subscription topic |
| `client_id` | string | `EDGEX_CLIENT_ID` | `sfsdb-edgex-adapter` | MQTT client ID |
| `http_port` | string | `EDGEX_HTTP_PORT` | `8081` | HTTP server port |
| `auto_subscribe` | bool | `EDGEX_AUTO_SUBSCRIBE` | `true` | Auto-subscribe to MQTT topic |
| `dev_config_path` | string | `EDGEX_DEV_CONFIG_PATH` | - | Device configuration path |

## Database Configuration

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `db_use_encryption` | bool | `EDGEX_DB_USE_ENCRYPTION` | `false` | Enable database encryption |
| `db_encryption_key` | string | `EDGEX_DB_ENCRYPTION_KEY` | - | Encryption key |
| `db_encryption_algorithm` | string | `EDGEX_DB_ENCRYPTION_ALGORITHM` | `AES-256-GCM` | Encryption algorithm |
| `db_scenario` | string | `EDGEX_DB_SCENARIO` | `edge` | Database scenario: `embedded`, `iot`, `edge`, `game`, `default` |

### Database Scenarios

| Scenario | Write Buffer | Block Cache | Use Case |
|----------|-------------|-------------|----------|
| `embedded` | 2MB | 4MB | Embedded devices |
| `iot` | 4MB | 8MB | IoT devices |
| `edge` (default) | 16MB | 32MB | Edge computing |
| `game` | 64MB | 128MB | Gaming scenarios |
| `default` | 8MB | 16MB | General purpose |

## MQTT TLS Configuration

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `mqtt_use_tls` | bool | `EDGEX_MQTT_USE_TLS` | `false` | Enable MQTT TLS |
| `mqtt_ca_cert` | string | `EDGEX_MQTT_CA_CERT` | - | CA certificate path |
| `mqtt_client_cert` | string | `EDGEX_MQTT_CLIENT_CERT` | - | Client certificate path |
| `mqtt_client_key` | string | `EDGEX_MQTT_CLIENT_KEY` | - | Client key path |
| `connection_timeout` | int | `EDGEX_MQTT_CONNECTION_TIMEOUT` | `30` | Connection timeout (seconds) |
| `keep_alive` | int | `EDGEX_MQTT_KEEP_ALIVE` | `60` | Keep-alive interval (seconds) |
| `mqtt_username` | string | `EDGEX_MQTT_USERNAME` | - | MQTT username |
| `mqtt_password` | string | `EDGEX_MQTT_PASSWORD` | - | MQTT password |

## HTTP TLS Configuration

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `http_use_tls` | bool | `EDGEX_HTTP_USE_TLS` | `false` | Enable HTTPS |
| `http_cert` | string | `EDGEX_HTTP_CERT` | - | Server certificate path |
| `http_key` | string | `EDGEX_HTTP_KEY` | - | Server key path |

## Data Retention Policy

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_retention_policy` | bool | `EDGEX_ENABLE_RETENTION_POLICY` | `false` | Enable retention policy |
| `retention_days` | int | `EDGEX_RETENTION_DAYS` | `30` | Data retention days |
| `cleanup_interval_hours` | int | `EDGEX_CLEANUP_INTERVAL_HOURS` | `24` | Cleanup interval |
| `cleanup_batch_size` | int | `EDGEX_CLEANUP_BATCH_SIZE` | `1000` | Batch size per cleanup |

## Resource Monitoring

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_resource_monitoring` | bool | `EDGEX_ENABLE_RESOURCE_MONITORING` | `false` | Enable resource monitoring |
| `max_memory_mb` | float | `EDGEX_MAX_MEMORY_MB` | `45` | Max memory (MB) |
| `max_cpu_percent` | float | `EDGEX_MAX_CPU_PERCENT` | `5` | Max CPU (%) |
| `resource_monitor_interval_seconds` | int | `EDGEX_RESOURCE_MONITOR_INTERVAL_SECONDS` | `30` | Monitor interval |

## Alert Configuration

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_alert_notifications` | bool | `EDGEX_ENABLE_ALERT_NOTIFICATIONS` | `false` | Enable alerts |
| `alert_notification_channels` | []string | `EDGEX_ALERT_NOTIFICATION_CHANNELS` | - | Notification channels |
| `alert_mqtt_topic` | string | `EDGEX_ALERT_MQTT_TOPIC` | - | Alert MQTT topic |
| `alert_webhook_url` | string | `EDGEX_ALERT_WEBHOOK_URL` | - | Webhook URL |
| `alert_min_severity` | string | `EDGEX_ALERT_MIN_SEVERITY` | `warning` | Minimum severity level |

## Data Sync (Enterprise)

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_data_sync` | bool | `EDGEX_ENABLE_DATA_SYNC` | `false` | Enable cloud sync |
| `data_sync_mqtt_topic` | string | `EDGEX_DATA_SYNC_MQTT_TOPIC` | - | Sync MQTT topic |
| `data_sync_queue_dir` | string | `EDGEX_DATA_SYNC_QUEUE_DIR` | `./data_sync_queue` | Queue directory |
| `data_sync_batch_size` | int | `EDGEX_DATA_SYNC_BATCH_SIZE` | `100` | Batch size |
| `data_sync_interval_seconds` | int | `EDGEX_DATA_SYNC_INTERVAL_SECONDS` | `60` | Sync interval |
| `data_sync_max_retry_count` | int | `EDGEX_DATA_SYNC_MAX_RETRY_COUNT` | `3` | Max retry count |

## Device Monitoring

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `device_offline_threshold_seconds` | int | `EDGEX_DEVICE_OFFLINE_THRESHOLD_SECONDS` | `300` | Offline threshold |
| `data_anomaly_threshold_percent` | int | `EDGEX_DATA_ANOMALY_THRESHOLD_PERCENT` | `50` | Anomaly threshold (%) |
| `data_trend_min_points` | int | `EDGEX_DATA_TREND_MIN_POINTS` | `5` | Min points for trend |

## Analyzer Configuration

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_analyzer` | bool | `EDGEX_ENABLE_ANALYZER` | `false` | Enable analyzer |
| `analyzer_max_memory` | int | `EDGEX_ANALYZER_MAX_MEMORY` | `10` | Max memory (MB) |
| `analyzer_max_time_per_run` | int | `EDGEX_ANALYZER_MAX_TIME_PER_RUN` | `500` | Max time per run (ms) |

## Prometheus Metrics

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_prometheus` | bool | `EDGEX_ENABLE_PROMETHEUS` | `false` | Enable Prometheus |
| `prometheus_path` | string | `EDGEX_PROMETHEUS_PATH` | `/metrics` | Metrics endpoint path |

## Simulator

| Key | Type | Environment Variable | Default | Description |
|-----|------|---------------------|---------|-------------|
| `enable_simulator` | bool | `EDGEX_ENABLE_SIMULATOR` | `false` | Enable simulator |
| `simulator_interval_min` | int | `EDGEX_SIMULATOR_INTERVAL_MIN` | `100` | Min interval (ms) |
| `simulator_interval_max` | int | `EDGEX_SIMULATOR_INTERVAL_MAX` | `1000` | Max interval (ms) |

## Example Configuration

```json
{
  "db_path": "data/sfs.db",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "client_id": "sfsdb-edgex-adapter",
  "http_port": "8081",
  "auto_subscribe": true,
  "db_scenario": "edge",
  "enable_resource_monitoring": true,
  "max_memory_mb": 256,
  "max_cpu_percent": 80,
  "enable_retention_policy": true,
  "retention_days": 30,
  "enable_analyzer": false,
  "enable_prometheus": false
}
```
