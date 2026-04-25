---
sidebar_position: 4
---

# First Deployment

Deploy sfsEdgeStore in a real production environment.

## Pre-Deployment Checklist

- [ ] MQTT Broker installed and running
- [ ] EdgeX Foundry configured to publish events via MQTT
- [ ] Sufficient disk space (at least 50MB free)
- [ ] Firewall rules configured (port 8081)
- [ ] SSL/TLS certificates ready (if using HTTPS)

## Configure for Production

Update `config.json`:

```json
{
  "db_path": "/var/lib/sfsedgestore/data",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "http_port": "8081",
  "license_type": "enterprise",
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "/etc/ssl/certs/ca.pem",
  "http_use_tls": true,
  "http_cert": "/etc/ssl/certs/server.pem",
  "http_key": "/etc/ssl/private/server.key",
  "enable_resource_monitoring": true,
  "max_memory_mb": 256,
  "max_cpu_percent": 80
}
```

## Start the Service

```bash
sudo systemctl start sfsedgestore
sudo systemctl status sfsedgestore
```

## Verify Data Flow

1. Check MQTT connection:
   ```bash
   curl http://localhost:8081/api/status
   ```

2. Verify data is being stored:
   ```bash
   curl http://localhost:8081/api/readings?limit=10
   ```

3. Check resource usage:
   ```bash
   curl http://localhost:8081/api/resources/status
   ```

## Monitor Logs

```bash
# View real-time logs
sudo journalctl -u sfsedgestore -f

# Check for errors
sudo journalctl -u sfsedgestore --priority=err --since="1 hour ago"
```

## Next Steps

- [Configure Monitoring](../how-to/configure-monitoring.md)
- [Setup Alerts](../how-to/setup-alerts.md)
- [Backup Configuration](../how-to/backup-restore.md)
