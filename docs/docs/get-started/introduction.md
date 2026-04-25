---
sidebar_position: 1
---

# Introduction

## What is sfsEdgeStore?

sfsEdgeStore is a lightweight, high-performance edge data storage adapter designed for Industrial IoT (IIoT) applications. It seamlessly integrates with EdgeX Foundry to collect, process, and store IoT sensor data locally at the edge.

## Core Architecture

```
IoT Sensors ──→ EdgeX Foundry ──→ MQTT ──→ sfsEdgeStore ──→ sfsDb (LevelDB)
                                                          │
                                                          └──→ Web Dashboard (WebSocket)
```

## Key Characteristics

| Feature | Description |
|---------|-------------|
| **Memory Footprint** | Less than 50MB RAM - suitable for resource-constrained edge devices |
| **Storage Engine** | Embedded LevelDB (sfsDb) with compression and encryption support |
| **Protocol Support** | MQTT 3.1.1/5.0, HTTP/HTTPS, WebSocket |
| **Data Format** | EdgeX Foundry native JSON format |
| **Deployment** | Single binary, no external dependencies |

## Use Cases

- **Factory Automation**: Real-time sensor data collection and storage
- **Environmental Monitoring**: Temperature, humidity, air quality tracking
- **Energy Management**: Power consumption monitoring and analysis
- **Smart Agriculture**: Soil moisture, weather station data collection

## License Editions

| Edition | Price | Device Limit | Features |
|---------|-------|-------------|----------|
| **Community** | Free | 5 devices | Core features, web dashboard |
| **Business** | $199/year | 50 devices | All community features + advanced analytics |
| **Enterprise** | $799/year | Unlimited | Full feature set, priority support, cloud sync |

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Database | sfsDb (LevelDB wrapper) |
| Message Queue | MQTT (Paho) |
| Web Framework | Go net/http |
| Frontend | Vanilla JS + Bootstrap + ECharts |
