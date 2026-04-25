---
sidebar_position: 1
---

# Reference

Detailed technical reference documentation.

## Available References

- [API Reference](api/overview.md) - REST API endpoints and authentication
- [Configuration](configuration.md) - All configuration options
- [CLI Commands](cli-commands.md) - Command-line interface
- [Metrics](metrics.md) - Prometheus metrics reference
- [Architecture](architecture/overview.md) - System architecture details

## API Quick Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/status` | GET | System status |
| `/api/readings` | GET | Query sensor data |
| `/api/config` | GET | Get configuration |
| `/api/resources/status` | GET | Resource usage |

## Configuration Quick Reference

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `db_path` | string | `./data` | Database storage path |
| `mqtt_broker` | string | `tcp://localhost:1883` | MQTT broker URL |
| `http_port` | string | `8081` | HTTP server port |
| `license_type` | string | `community` | License edition |
