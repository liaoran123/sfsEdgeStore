# Architecture

System architecture overview of sfsEdgeStore.

## Lightweight Edge Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Edge Node (Resource-Constrained)       │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              EdgeX Foundry                             │ │
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

## Core Components

### 1. MQTT Client (`core/mqtt/`)

Subscribes to EdgeX Foundry event topics and receives IoT sensor data.

- Protocol: MQTT 3.1.1/5.0
- TLS support with mutual authentication
- Automatic reconnection with configurable timeout
- Topic pattern matching (`edgex/events/#`)

### 2. Data Queue (`queue/`)

Reliable message buffering between MQTT ingestion and database storage.

- Persistent queue using embedded storage
- Automatic retry on failure (up to 3 attempts)
- Data persistence across restarts
- Backpressure handling

### 3. Database (`core/database/`)

Efficient data storage using sfsDb (LevelDB wrapper).

- Embedded, zero external dependencies
- WAL (Write-Ahead Logging) for crash safety
- Compression and optional encryption
- Multiple performance scenarios (embedded, IoT, edge, game)

### 4. HTTP Server (`server/`)

RESTful API for data querying and system management.

- Health check endpoints
- Data query with filtering (device, time range, limit)
- Backup and restore operations
- WebSocket for real-time data streaming
- Prometheus metrics export

### 5. Monitoring (`monitor/`)

System metrics collection and alerting.

- MQTT connection status tracking
- HTTP request counting
- Device monitoring with offline detection
- Automatic cleanup of expired alerts

### 6. Alert System (`alert/`)

Threshold-based alerting with multiple notification channels.

- MQTT topic publishing
- Webhook integration
- Configurable severity levels
- Automatic alert expiration

### 7. Analyzer (`analyzer/`)

Data analysis engine for anomaly detection and trend analysis.

- Configurable thresholds per device/reading
- Low-frequency operation (minimal resource usage)
- Time-based trend analysis

### 8. Resource Monitor (`resource/`)

System resource tracking and protection.

- Memory usage monitoring
- CPU usage tracking
- Configurable thresholds with alerts

### 9. Data Retention (`retention/`)

Automatic cleanup of expired data.

- Configurable retention period
- Batch processing for efficiency
- Low-frequency scheduled execution

### 10. Data Sync (`cloudsync/sync/`) (Enterprise)

Cloud synchronization for distributed deployments.

- Token bucket rate limiting
- Persistent sync queue
- Automatic retry with configurable intervals

## Design Principles

1. **Small and Beautiful** - Does one thing, does it perfectly
2. **Edge First** - All features prioritize edge scenarios
3. **Data Sovereignty** - Data stays local. No cloud dependency, no cross-border transfer.
4. **Zero Dependencies** - Only depends on sfsDb, no heavy components
5. **High Availability** - Power-failure recovery, data retry, local storage

### Why Local-First Matters

Unlike cloud-based IoT solutions, sfsEdgeStore stores all data on the edge device:

- **GDPR Compliance**: Data never leaves your premises
- **EU Cyber Resilience Act**: No dependency on external services
- **Zero Data Leakage Risk**: No telemetry, no analytics sent to third parties
- **Works Offline**: Fully operational without internet connection
- **Cost Predictability**: No cloud storage or bandwidth fees

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Database | sfsDb (LevelDB wrapper) |
| Message Queue | MQTT (Paho) |
| Web Framework | Go net/http |
| Frontend | Vanilla JS + Bootstrap + ECharts |
