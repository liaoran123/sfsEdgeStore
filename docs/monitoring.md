# Monitoring & Alerting

Monitor sfsEdgeStore health and configure alerts.

## Metrics

### Prometheus Metrics

Enable Prometheus metrics in configuration:

```json
{
  "enable_prometheus": true,
  "prometheus_path": "/metrics"
}
```

Access metrics at:
```bash
curl http://localhost:8081/metrics
```

### System Status

```bash
curl http://localhost:8081/api/status
```

### Resource Usage

```bash
curl http://localhost:8081/api/resources/status
```

Returns current memory and CPU usage.

## Resource Monitoring

sfsEdgeStore includes built-in resource monitoring that runs every 30 seconds.

### Configuration

```json
{
  "enable_resource_monitoring": true,
  "max_memory_mb": 512,
  "max_cpu_percent": 80,
  "resource_monitor_interval_seconds": 30
}
```

### Alerts

When thresholds are exceeded, alerts are triggered via configured channels.

## Alert Configuration

### Enable Alerts

```json
{
  "enable_alert_notifications": true,
  "alert_notification_channels": ["mqtt", "webhook"],
  "alert_mqtt_topic": "sfsEdgeStore/alerts",
  "alert_webhook_url": "https://hooks.example.com/alerts",
  "alert_min_severity": "warning"
}
```

### Notification Channels

| Channel | Configuration | Description |
|---------|--------------|-------------|
| MQTT | `alert_mqtt_topic` | Publish alerts to MQTT topic |
| Webhook | `alert_webhook_url` | Send HTTP POST to webhook URL |

### Severity Levels

- `info` - Informational messages
- `warning` - Warning conditions
- `error` - Error conditions
- `critical` - Critical failures

### Test Alerts

```bash
curl -X POST http://localhost:8081/api/alerts/test
```

### View Alerts

```bash
# All alerts
curl http://localhost:8081/api/alerts

# Grouped alerts
curl http://localhost:8081/api/alert-groups
```

### Alert Notifier Status

```bash
curl http://localhost:8081/api/alerts/notifier/status
```

## Device Monitoring

sfsEdgeStore automatically monitors device health.

### Offline Detection

Configured via `device_offline_threshold_seconds` (default: 300 seconds).

### Data Anomaly Detection

Detects abnormal data patterns based on `data_anomaly_threshold_percent`.

### Trend Analysis

Analyzes data trends when minimum points (`data_trend_min_points`) are reached.

## Subscription Monitoring

### Check Subscription Status

```bash
curl http://localhost:8081/api/subscription/status
```

### Test Subscription

```bash
curl -X POST http://localhost:8081/api/subscription/test
```

### View Subscribed Topics

```bash
curl http://localhost:8081/api/subscription/themes
```

## Dashboard

Access the web dashboard at:
```
http://localhost:8081/dashboard
```

Features:
- Real-time data visualization
- Device status overview
- Alert center
- Subscription topic management
- Template and baseline configuration
