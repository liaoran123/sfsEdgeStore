# sfsEdgeStore

Lightweight Edge Computing Data Storage Adapter - Solving Edge Data Storage Pain Points

---

## ⚡ Performance Highlights

| Metric | Measured Value | Description |
|--------|---------------|-------------|
| **Memory Usage** | 20.85 MB | Ultra-lightweight, suitable for resource-constrained devices |
| **CPU Usage** | 2.9% | Almost no resource consumption when running in background |
| **Startup Time** | 0.187 seconds | Lightning-fast startup, millisecond-level response |
| **Database Size** | 0.25 MB | 18,681 records take only 0.25 MB |

> 📊 **How to reproduce tests?** For detailed performance testing methods and benchmark code, please see [Performance Test Guide](./docs/BENCHMARK.md)

---

## 🎯 Core Problems Solved

### Edge Computing Pain Points

| Pain Point | sfsEdgeStore Solution |
|------------|----------------------|
| **Limited edge device resources** | Memory < 50MB, CPU < 5%, ultra-lightweight design |
| **Data loss during network outages** | Local storage, network disconnects don't affect data collection |
| **Complex heavy database deployment** | 5-minute deployment, ready to use out of the box |
| **EdgeX Foundry data storage challenges** | Native EdgeX Foundry integration, seamless integration |
| **Slow data query response** | LevelDB backend, millisecond-level local query response |
| **Cloud dependency** | Can run independently, no central system dependency |

---

## 📋 Product Introduction

**sfsEdgeStore** is a lightweight data storage adapter designed specifically for industrial IoT edge scenarios, serving as a bridge between EdgeX Foundry and sfsDb database, providing efficient local data read/write and caching capabilities.

### Core Values

- ✅ **Pure Go Implementation**：No CGO dependencies, simple cross-platform compilation, worry-free deployment
- ✅ **Ultra-lightweight**：Extremely low resource consumption, can run on any edge device
- ✅ **Highly Reliable**：Local storage, network interruptions don't affect data collection
- ✅ **Easy Integration**：Native EdgeX Foundry integration, ready to use out of the box
- ✅ **High Performance**：LevelDB backend, millisecond-level local query response
- ✅ **Open Source & Free**：Full functionality, unlimited usage

---

## ✨ Core Features

- 📡 **MQTT Data Ingestion**：Subscribe to EdgeX Foundry event topics
- 💾 **Local Data Storage**：Efficient edge data storage using sfsDb/LevelDB
- 🔄 **Data Queue**：Power failure recovery and data retry mechanism to ensure no data loss
- 📊 **Real-time Monitoring**：Built-in system and business metrics monitoring
- ⚠️ **Intelligent Alerts**：Threshold-based alerts and anomaly detection
- 📈 **Data Analysis**：Built-in time window aggregation and prediction
- 🔐 **Authentication & Authorization**：API Key and RBAC permission control
- 🌐 **HTTP API**：RESTful interface for external queries
- 🔄 **Data Synchronization**：Optional cloud data synchronization
- 🗑️ **Data Retention**：Automatic cleanup of expired data

---

## 🚀 5-Minute Quick Start

### Prerequisites

- Go 1.25+ (for source code compilation)
- EdgeX Foundry (optional, for data source)
- MQTT Broker (e.g., Mosquitto)

### Method 1: Binary Deployment (Recommended, High Performance)

For极致 performance? Use binary files with systemd (Linux) or Windows services, zero virtualization overhead!

**Linux (systemd):**
```bash
# Download binary files for your platform from GitHub Releases
# https://github.com/your-username/sfsEdgeStore/releases

# Run directly (for testing)
./sfsedgestore

# For production, recommended to use systemd daemon (auto-start on boot, crash restart)
```

**Windows:**
```bash
# Use Windows services or tools like NSSM to configure as system service
```

### Method 2: Docker Deployment (Convenient, Quick Experience)

Docker is suitable for quick testing and deployment, **comes with daemon functionality (auto-start on boot, crash restart), but has a slight performance overhead (about 5-10%).

```bash
# Pull image
docker pull sfsedgestore/sfsedgestore:latest

# Run
docker run -d \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  sfsedgestore/sfsedgestore:latest
```

### Method 3: Compile from Source

```bash
# Clone repository
git clone https://github.com/your-username/sfsEdgeStore.git
cd sfsEdgeStore

# Install dependencies
go mod download

# Compile
go build -o sfsedgestore

# Run
./sfsedgestore
```

### Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# View metrics
curl http://localhost:8080/metrics
```

---

## 🏗️ Architecture Design

### Lightweight Edge Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Edge Node (Resource-Constrained)        │
│                                                           │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              EdgeX Foundry                              │ │
│  │  (Data Collection, Device Management)                │ │
│  └────────────────────┬──────────────────────────────────┘ │
│                       │ MQTT                              │
│                       ▼                                   │
│  ┌───────────────────────────────────────────────────────┐ │
│  │         sfsEdgeStore (Lightweight Adapter)            │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  │ │
│  │  │  MQTT Client  │→ │  Data Queue   │→ │  sfsDb   │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────┘  │ │
│  │  ┌──────────────┐  ┌──────────────┐                  │ │
│  │  │  HTTP Server  │  │  Monitoring & Alerts │        │ │
│  │  └──────────────┘  └──────────────┘                  │ │
│  └───────────────────────────────────────────────────────┘ │
│                       │ HTTP API                         │
└───────────────────────┼─────────────────────────────────────┘
                        ▼
                  External Query/Monitoring
```

### Design Principles

- **Small and Beautiful**：Do one thing and do it extremely well
- **Edge-First**：All features prioritize edge scenarios
- **Zero Dependencies**：No heavy components other than sfsDb
- **High Availability**：Power failure recovery, data retry, local storage

---

## 📡 API Interface

### Health Check

```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "1h23m45s"
}
```

### Get Metrics

```bash
GET /metrics
```

### Query Data

```bash
# Query device data
GET /query?deviceName=Device001

# Query by time range
GET /query?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-12-31T23:59:59Z
```

### View Alerts

```bash
GET /alerts
```

For complete API documentation, please see [API Documentation](./docs/api.md)

---

### Commercial Services

| Service Type | Description |
|-------------|-------------|
| **Consulting Services** | Architecture design, technical consulting, performance optimization advice |
| **Custom Development** | Feature customization, system integration, plugin development |

For detailed service descriptions, please see [Commercial Services Documentation](./docs/pricing/SERVICES.md).

---

## 📚 Documentation Center

For complete documentation, please visit the [Documentation Center](./docs/README.md), including:

- 📖 **User Guide** - User manual, admin guide, FAQ
- 🔧 **Technical Documentation** - API reference, architecture design, design decisions
- 🚀 **Best Practices** - Deployment, monitoring, security, backup & recovery, encryption configuration
- 💼 **Commercial Services** - Technical support, implementation consulting, custom development
- 💰 **Sales Plans** - Sales strategy, pricing plans, success metrics

---

## 📖 E-Books

### Training Manual
Complete training tutorial suitable for beginners and quick start.

📚 [EdgeX Foundry and sfsDb: Industrial IoT Edge Computing Data Storage Practice](./book/README.md)

### Technical Deep Dive
Technical book for Go language and EdgeX Foundry developers, including source code analysis, architecture design, performance optimization, etc.

📚 [sfsEdgeStore Technical Deep Dive](./tech-book/00-前言与目录.md)

---

## 🤝 Contributing

Welcome to submit Issues and Pull Requests!

Please see [CONTRIBUTING.md](./CONTRIBUTING.md) to learn how to contribute.

## 🔒 Security

Please see [SECURITY.md](./SECURITY.md) to learn about security policies and vulnerability reporting methods.

## 📄 License

Apache License 2.0

## 🙏 Acknowledgments

- [EdgeX Foundry](https://www.edgexfoundry.org/)
- [sfsDb](https://github.com/liaoran123/sfsDb)
- [Eclipse Paho MQTT](https://www.eclipse.org/paho/)
