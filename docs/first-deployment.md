# First Deployment

Guide for deploying sfsEdgeStore in production.

## Pre-Deployment Checklist

- [ ] MQTT Broker installed and running
- [ ] EdgeX Foundry configured to publish events via MQTT
- [ ] Sufficient disk space (at least 1GB free)
- [ ] Firewall rules configured (port 8081)
- [ ] SSL/TLS certificates ready (if using HTTPS)

## Production Configuration

Create `config.json`:

```json
{
  "db_path": "/var/lib/sfsedgestore/data",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "http_port": "8081",
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "/etc/ssl/certs/ca.pem",
  "http_use_tls": true,
  "http_cert": "/etc/ssl/certs/server.pem",
  "http_key": "/etc/ssl/private/server.key",
  "enable_resource_monitoring": true,
  "max_memory_mb": 256,
  "max_cpu_percent": 80,
  "enable_retention_policy": true,
  "retention_days": 30
}
```

## Start the Service

```bash
sudo systemctl start sfsedgestore
sudo systemctl status sfsedgestore
```

## Verify Data Flow

```bash
# Check system status
curl http://localhost:8081/api/status

# Verify data is being stored
curl "http://localhost:8081/api/readings?limit=10"

# Check resource usage
curl http://localhost:8081/api/resources/status
```

## Monitor Logs

```bash
# View real-time logs
sudo journalctl -u sfsedgestore -f

# Check for errors
sudo journalctl -u sfsedgestore --priority=err --since="1 hour ago"
```
