[中文](./README.md) | [English](./README_en.md)

# sfsEdgeStore

Lightweight Industrial IoT Edge Data Storage Adapter for EdgeX Foundry.

> This project is built by the **official sfsDb team**, serving as the official edge computing adapter for sfsDb database.

## Data Sovereignty & Compliance

**sfsEdgeStore** follows a **local-first architecture** that ensures **full data sovereignty**. All data is stored locally on the edge device - no cloud dependency, no data transfer to third parties.

- **GDPR Compliant**: Data never leaves your premises. No cross-border data transfer.
- **EU Cyber Resilience Act Ready**: No dependency on external services.
- **Encryption at Rest**: AES-256 database encryption ensures data is unreadable without your keys.
- **Zero Vendor Lock-in**: Pure Go binary, embedded database, no external services required.

## ⚡ Performance

### Production Environment (100 devices, 1-60s interval)

| Metric | Value | Description |
|--------|-------|-------------|
| **Memory** | ~30 MB | Stable under normal operation |
| **CPU** | 1.7% | Minimal overhead, ultra-low power |
| **Message Rate** | ~30 msg/sec | Realistic industrial sensor data |
| **Startup** | <0.2s | Fast startup, ready in milliseconds |

### Stress Test (500 devices, 0.05-0.2s interval)

| Metric | Value | Description |
|--------|-------|-------------|
| **Memory** | ~44 MB | Under heavy load (4000 msg/sec) |
| **CPU** | 6.8% | Handles extreme load gracefully |
| **Message Rate** | ~4000 msg/sec | 133x normal production load |
| **Zero Data Loss** | 100% | All messages stored successfully |

### Resource Comparison

| Scenario | Devices | Interval | Rate | CPU | Memory |
|----------|---------|----------|------|-----|--------|
| Production | 100 | 1-60s | ~30 msg/sec | 1.7% | ~30 MB |
| Stress Test | 500 | 0.05-0.2s | ~4000 msg/sec | 6.8% | ~44 MB |

![Dashboard](./img/sfsEdgeStoreEn.png)

## 🎯 Core Problem Solved

### Edge Computing Challenges

| Challenge | Solution |
|-----------|----------|
| Limited edge device resources | Memory <30MB, CPU <7%, ultra-lightweight |
| Data loss during network outages | Local storage, offline operation |
| Complex heavy database deployment | 5-minute deployment, zero configuration |
| EdgeX Foundry storage gap | Native EdgeX integration, seamless setup |
| Slow data query response | Local LevelDB, millisecond query response |
| Cloud dependency | Independent operation, no center system required |

## 📋 Product Overview

**sfsEdgeStore** is a lightweight data storage adapter designed for industrial IoT edge scenarios. It serves as a bridge between EdgeX Foundry and sfsDb, providing efficient local data read/write and caching capabilities.

### Why EdgeX Needs sfsEdgeStore

> EdgeX is the best connectivity framework, but it doesn't persist data by default. Don't use Redis (data loss on power failure), don't use InfluxDB (too heavy for edge). sfsEdgeStore is EdgeX's native storage plugin, designed for resource-constrained edge gateways.

## ✨ Key Features

- 📡 **MQTT Data Ingestion**: Subscribe to EdgeX Foundry event topics
- 💾 **Local Data Storage**: sfsDb/LevelDB for efficient edge data storage
- 🔒 **Data Encryption**: AES-256 encryption for data at rest
- 🔄 **Reliable Queue**: Power-failure recovery and data retry mechanism
- 📊 **Real-time Monitoring**: Built-in system and business metrics
- ⚠️ **Smart Alerts**: Threshold alerts and anomaly detection
- 🗑️ **Data Retention**: Automatic cleanup of expired data
- 🔐 **Authentication**: API Key and RBAC access control
- 🌐 **HTTP API**: RESTful interfaces for external queries
- 📦 **Backup & Restore**: Automated backup and recovery

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (for building from source)
- EdgeX Foundry (optional, for data source)
- MQTT Broker (e.g., Mosquitto)

### Method 1: Binary Deployment (Recommended)

```bash
# Download from GitHub Releases
# https://github.com/liaoran123/sfsEdgeStore/releases

# Run directly (for testing)
./sfsedgestore

# For production, use systemd (Linux) or Windows Service
```

### Method 2: Docker Compose (Recommended)

```bash
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore
docker-compose up -d
```

Starts sfsEdgeStore + MQTT Broker together. Access dashboard at `http://localhost:8081`.

### Method 3: Docker

```bash
docker run -d \
  --name sfsedgestore \
  -p 8081:8081 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  liaoran123/sfsedgestore:latest
```

### Method 4: From Source

```bash
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore
go build -ldflags="-s -w" -o sfsedgestore .
./sfsedgestore
```

### Verify Installation

```bash
# Health check
curl http://localhost:8081/health

# Open Dashboard in browser
# http://localhost:8081
```

### Zero Configuration

sfsEdgeStore uses intelligent defaults. Start without any configuration:

| Setting | Default |
|---------|---------|
| MQTT Broker | `tcp://localhost:1883` |
| MQTT Topic | `edgex/events/#` |
| HTTP Port | `8081` |
| Database Path | `data` |

### Configuration Example

Create `config.json` to customize settings:

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "http_port": "8081",
  "db_path": "data",
  "db_scenario": "edge",
  "enable_resource_monitoring": true,
  "max_memory_mb": 256,
  "enable_retention_policy": true,
  "retention_days": 30
}
```

| Key | Description |
|-----|-------------|
| `mqtt_broker` | MQTT server address (e.g., Mosquitto) |
| `mqtt_topic` | EdgeX event topic pattern |
| `db_path` | Local database storage path |
| `db_scenario` | Performance profile: `embedded`, `iot`, `edge`, `game`, `default` |
| `enable_retention_policy` | Auto-cleanup old data |
| `retention_days` | Days to keep data before cleanup |

> MQTT Topic is managed through the **Topic Subscription** page (`http://localhost:8081/mqtt-subscription`) in the web dashboard.

## 🏗️ Architecture

### Lightweight Edge Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Edge Node (Resource-Constrained)       │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              EdgeX Foundry                              │ │
│  │  (Data Collection, Device Management)                  │ │
│  └────────────────────┬──────────────────────────────────┘ │
│                       │ MQTT                              │
│                       ▼                                   │
│  ┌───────────────────────────────────────────────────────┐ │
│  │         sfsEdgeStore (Lightweight Adapter)             │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  │ │
│  │  │  MQTT Client  │→ │  Data Queue  │→ │  sfsDb   │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────┘  │ │
│  │  ┌──────────────┐  ┌──────────────┐                  │ │
│  │  │  HTTP Server  │  │  Monitoring  │                  │ │
│  │  └──────────────┘  └──────────────┘                  │ │
│  └───────────────────────────────────────────────────────┘ │
│                       │ HTTP API                         │
└───────────────────────┼─────────────────────────────────────┘
                        ▼
                  External Query/Monitoring
```

### Design Principles

- **Small and Beautiful**: Does one thing, does it perfectly
- **Data Sovereignty**: Data stays local. No cloud dependency.
- **Edge First**: All features prioritize edge scenarios
- **Zero Dependencies**: Only depends on sfsDb, no heavy components
- **High Availability**: Power-failure recovery, data retry, local storage

## 📚 Documentation

Complete documentation available in the [docs](./docs/) directory:

| Document | Description |
|----------|-------------|
| [Quick Start](./docs/quick-start.md) | Get started in 5 minutes |
| [Installation](./docs/installation.md) | Installation and deployment guide |
| [API Reference](./docs/api-reference.md) | REST API documentation |
| [Configuration](./docs/configuration.md) | All configuration options |
| [Architecture](./docs/architecture.md) | System architecture overview |
| [Security](./docs/security.md) | Authentication, TLS, encryption |
| [Troubleshooting](./docs/troubleshooting.md) | Common issues and solutions |

📖 [View Documentation Index](./docs/README.md)

## 💼 Partnership & Investment

> **We sell solutions, not software.**

**sfsEdgeStore** is positioned as an industrial IoT edge data solution — providing complete hardware + software + deployment packages, not just a standalone product.

### 🤝 Seeking Partnership / Investment

We are actively seeking strategic partners and investors to accelerate our growth in the industrial IoT market.

**Why Invest in sfsEdgeStore?**

| What Investors Care About | Traditional Industrial Software | sfsEdgeStore (Our Solution) | Our Investment Story |
| :--- | :--- | :--- | :--- |
| **Hardware Cost** | Requires expensive industrial PCs ($500+) | Runs on ordinary ARM gateways ($200) | *"We helped clients save 60% on hardware budget."* |
| **Resource Footprint** | Bloated, memory-hungry | **27.6MB** ultra-lightweight | *"Runs on legacy equipment — massive retrofit market."* |
| **Deployment Difficulty** | Requires professional implementation teams | **5-minute** plug-and-play | *"Scales like SaaS — rapid deployment, zero friction."* |
| **Technical Barrier** | Heavy external dependencies | **Pure Go**, zero dependencies | *"Extreme engineering efficiency — one person equals a team."* |
| **Scalability** | Limited by infrastructure | Stateless, horizontally scalable | *"Built for the next 10,000 edge nodes."* |

### 💡 What We Offer

| Service | Description |
|---------|-------------|
| **Solution Consulting** | Assess your edge computing needs, design optimal architecture |
| **Custom Development** | Tailored protocols, integrations, and industry-specific features |
| **Technical Support** | Priority email support, deployment assistance, troubleshooting |
| **Training & Enablement** | On-site or remote training for your engineering team |

**Contact:** [liao010203kk@gmail.com](mailto:liao010203kk@gmail.com)

## 🤝 Contributing

Welcome to submit Issues and Pull Requests!

Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution guidelines.

## 🔒 Security

Please see [SECURITY.md](./SECURITY.md) for security policy and vulnerability reporting.

## 📄 License

Apache License 2.0

## 🙏 Acknowledgments

- [EdgeX Foundry](https://www.edgexfoundry.org/)
- [sfsDb](https://github.com/liaoran123/sfsDb)
- [Eclipse Paho MQTT](https://www.eclipse.org/paho/)
