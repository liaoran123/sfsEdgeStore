# Data Retention

Configure data retention policies for sfsEdgeStore.

## Overview

Data retention policies automatically clean up old data to prevent unbounded disk growth.

## Configuration

```json
{
  "enable_retention_policy": true,
  "retention_days": 30,
  "cleanup_interval_hours": 24,
  "cleanup_batch_size": 1000
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `retention_days` | 30 | Days to keep data |
| `cleanup_interval_hours` | 24 | How often to run cleanup |
| `cleanup_batch_size` | 1000 | Records deleted per batch |

## Check Retention Status

```bash
curl http://localhost:8081/api/retention/status
```

## Manual Cleanup

Trigger cleanup immediately:

```bash
curl -X POST http://localhost:8081/api/retention/cleanup
```

## Recommended Retention Periods

| Use Case | Retention Period |
|----------|-----------------|
| Real-time monitoring | 7 days |
| Historical analysis | 30 days |
| Compliance/audit | 90 days |
| Archival | Custom (export before cleanup) |

## Export Before Cleanup

Export data before it is deleted:

```bash
# Export as CSV
curl "http://localhost:8081/api/export/csv?startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z"

# Export as JSON
curl "http://localhost:8081/api/export/json?startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z"
```
