# Chapter 3: sfsEdgeStore Quick Start

## 3.1 5-Minute Deployment Guide

This chapter gets you from zero to running sfsEdgeStore in 5 minutes. Whether you're deploying to a Raspberry Pi, an industrial gateway, or a development machine, these steps will have you up and running quickly.

### Prerequisites

Before starting, ensure you have:

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **OS** | Linux (Raspberry Pi OS), Windows 10+, macOS 10.14+ | Linux (Ubuntu 20.04+ or Raspberry Pi OS 64-bit) |
| **CPU** | 1 core | 2+ cores |
| **RAM** | 512 MB | 2 GB |
| **Disk** | 1 GB free | 10 GB SSD |
| **Go** | 1.21+ (if compiling) | 1.25+ |

### Installation Methods

Choose the method that best fits your environment:

#### Method 1: Download Pre-built Binary (Recommended)

Pre-built binaries are the fastest way to get started:

**Linux (x86-64):**
```bash
# Download the latest release
wget https://github.com/liaoran123/sfsEdgeStore/releases/latest/sfsedgestore-linux-amd64.tar.gz

# Extract
tar -xzf sfsedgestore-linux-amd64.tar.gz
cd sfsedgestore-linux-amd64

# Make executable
chmod +x sfsedgestore

# Quick test
./sfsedgestore --version
```

**Linux (ARM64 - Raspberry Pi 4, Rock Pi):**
```bash
wget https://github.com/liaoran123/sfsEdgeStore/releases/latest/sfsedgestore-linux-arm64.tar.gz
tar -xzf sfsedgestore-linux-arm64.tar.gz
cd sfsedgestore-linux-arm64
chmod +x sfsedgestore
./sfsedgestore --version
```

**Windows:**
```powershell
# Download and extract the Windows binary
# Run from Command Prompt or PowerShell
.\sfsedgestore.exe --version
```

**macOS:**
```bash
# Download and extract
tar -xzf sfsedgestore-darwin-amd64.tar.gz
cd sfsedgestore-darwin-amd64
chmod +x sfsedgestore
./sfsedgestore --version
```

#### Method 2: Docker Deployment (Recommended for Testing)

Docker provides an isolated environment that's perfect for quick testing:

```bash
# Pull the official image
docker pull sfsedgestore/sfsedgestore:latest

# Run with sample configuration
docker run -d \
  --name sfsedgestore \
  -p 8081:8081 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.json:/app/config.json \
  sfsedgestore/sfsedgestore:latest

# Check logs
docker logs sfsedgestore

# Verify it's running
curl http://localhost:8081/health
```

#### Method 3: Compile from Source

If you need to customize or run on an unsupported platform:

```bash
# Clone the repository
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore

# Install dependencies
go mod download

# Build
go build -o sfsedgestore main.go

# Run
./sfsedgestore
```

**Cross-compilation Example (from Linux to Raspberry Pi):**
```bash
# Set target architecture
export GOOS=linux
export GOARCH=arm64

# Compile
go build -o sfsedgestore-linux-arm64 main.go

# Copy to Raspberry Pi
scp sfsedgestore-linux-arm64 pi@raspberrypi:/home/pi/
```

## 3.2 Configuration Templates for Different Scenarios

sfsEdgeStore uses a JSON configuration file. Let's start with a minimal configuration and then explore advanced options.

### Minimal Configuration

Create a file named `config.json`:

```json
{
  "HTTPPort": "8081",
  "DBPath": "./data",
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-001"
}
```

**Required Fields:**

| Field | Description | Example |
|-------|-------------|---------|
| `HTTPPort` | HTTP server port | `"8081"` |
| `DBPath` | Path to database directory | `"./data"` |
| `MQTTBroker` | MQTT broker URL | `"tcp://localhost:1883"` |
| `MQTTTopic` | EdgeX event topic | `"edgex/events"` |
| `ClientID` | Unique client identifier | `"sfsedgestore-001"` |

### Configuration for Raspberry Pi

Optimized for resource-constrained environments:

```json
{
  "HTTPPort": "8081",
  "HTTPUseTLS": false,

  "DBPath": "/var/lib/sfsedgestore/data",
  "DBScenario": "edge",
  "DBUseEncryption": false,
  "DBCompression": true,

  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-pi-001",
  "MQTTQoS": 1,
  "MQTTKeepAlive": 30,
  "MQTTUseTLS": false,

  "EnableSimulator": false,
  "EnableAnalyzer": false,
  "EnableRetention": true,
  "RetentionDays": 30,

  "EnableAlertNotifications": false,
  "EnableSync": false,

  "LogLevel": "info"
}
```

### Configuration for Industrial Gateway

Higher performance for demanding workloads:

```json
{
  "HTTPPort": "8081",
  "HTTPUseTLS": true,
  "HTTPCert": "/etc/sfsedgestore/certs/server.crt",
  "HTTPKey": "/etc/sfsedgestore/certs/server.key",

  "DBPath": "/var/lib/sfsedgestore/data",
  "DBScenario": "standard",
  "DBUseEncryption": true,
  "DBEncryptionKey": "your-32-character-encryption-key!",
  "DBEncryptionAlgorithm": "AES-256",
  "DBCompression": true,

  "MQTTBroker": "tcp://mqtt.factory.local:1883",
  "MQTTTopic": "factory/edgex/events",
  "ClientID": "sfsedgestore-gateway-001",
  "MQTTQoS": 2,
  "MQTTKeepAlive": 60,
  "MQTTUseTLS": true,
  "MQTTCACert": "/etc/sfsedgestore/certs/ca.crt",
  "MQTTClientCert": "/etc/sfsedgestore/certs/client.crt",
  "MQTTClientKey": "/etc/sfsedgestore/certs/client.key",

  "EnableSimulator": false,
  "EnableAnalyzer": true,
  "EnableRetention": true,
  "RetentionDays": 90,

  "EnableAlertNotifications": true,
  "AlertNotificationChannels": ["mqtt", "webhook"],
  "AlertMinSeverity": "warning",
  "AlertMQTTTopic": "factory/alerts",
  "AlertWebhookURL": "https://monitoring.factory.local/webhook",

  "EnableSync": true,
  "SyncEndpoint": "https://cloud.factory.local/api/sync",

  "LogLevel": "info"
}
```

### Configuration Options Reference

| Category | Field | Type | Default | Description |
|----------|-------|------|---------|-------------|
| **HTTP Server** | `HTTPPort` | string | `"8081"` | Server port |
| | `HTTPUseTLS` | bool | `false` | Enable HTTPS |
| | `HTTPCert` | string | - | TLS certificate path |
| | `HTTPKey` | string | - | TLS key path |
| **Database** | `DBPath` | string | `"./data"` | Database directory |
| | `DBScenario` | string | `"edge"` | Scenario preset |
| | `DBUseEncryption` | bool | `false` | Enable encryption |
| | `DBEncryptionKey` | string | - | Encryption key (32 chars) |
| | `DBEncryptionAlgorithm` | string | `"AES-256"` | Encryption algorithm |
| | `DBCompression` | bool | `true` | Enable compression |
| **MQTT** | `MQTTBroker` | string | - | Broker URL |
| | `MQTTTopic` | string | - | Subscription topic |
| | `ClientID` | string | - | Unique client ID |
| | `MQTTQoS` | int | `1` | QoS level (0, 1, or 2) |
| | `MQTTKeepAlive` | int | `30` | Keep-alive interval (seconds) |
| | `MQTTUseTLS` | bool | `false` | Enable TLS |
| | `MQTTCACert` | string | - | CA certificate path |
| | `MQTTClientCert` | string | - | Client certificate path |
| | `MQTTClientKey` | string | - | Client key path |
| **Features** | `EnableSimulator` | bool | `false` | Enable data simulator |
| | `EnableAnalyzer` | bool | `false` | Enable data analysis |
| | `EnableRetention` | bool | `true` | Enable data retention |
| | `RetentionDays` | int | `30` | Data retention period |
| **Alerts** | `EnableAlertNotifications` | bool | `false` | Enable alerts |
| | `AlertNotificationChannels` | array | `[]` | Channel list |
| | `AlertMinSeverity` | string | `"warning"` | Minimum severity |
| **Sync** | `EnableSync` | bool | `false` | Enable cloud sync |
| | `SyncEndpoint` | string | - | Sync endpoint URL |

## 3.3 Integration with EdgeX Foundry

sfsEdgeStore is designed to work seamlessly with EdgeX Foundry. This section covers the integration process.

### Understanding the Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      EdgeX Foundry                           │
│                                                              │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│  │ Device  │───▶│ Device  │───▶│   App   │───▶│  MQTT   │  │
│  │Service  │    │Service  │    │ Service │    │ Export  │  │
│  └─────────┘    └─────────┘    └─────────┘    └────┬────┘  │
│                                                      │       │
│                     MQTT Topic: "edgex/events"       │       │
└─────────────────────────────────────────────────────│───────┘
                                                  │ MQTT
                                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                     sfsEdgeStore                             │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │    MQTT      │───▶│    Data      │───▶│    sfsDb     │  │
│  │   Client     │    │    Queue     │    │   Database   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐                       │
│  │     HTTP     │───▶│   Query &   │                       │
│  │    Server    │    │   Export    │                       │
│  └──────────────┘    └──────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

### EdgeX Configuration

**Step 1: Configure MQTT Export Service**

In your EdgeX `docker-compose.yml` or configuration files:

```yaml
# Add to app-service-config-mqtt.yaml or similar
# This configures EdgeX to publish events to MQTT

Service:
  Host: localhost
  Port: 8080

MQTTProtocol:
  Scheme: tcp
  Broker: localhost:1883
  Topic: edgex/events

PublishTopic: edgex/events  # sfsEdgeStore subscribes to this

# Optional: Filter by device
DeviceNames: "Random-Boolean-Device,Random-Integer-Device,Random-Float-Device"
```

**Step 2: Start EdgeX Services**

```bash
# Using Docker Compose (recommended)
cd edgex-compose
docker-compose up -d edgex-core-data edgex-core-command edgex-app-mqtt-export

# Or for native EdgeX (Edinburgh or later)
export MQTTBROKER=localhost
export MQTTCLIENTID=edgex-mqtt
edgexFoundry &
```

**Step 3: Verify EdgeX is Publishing**

```bash
# Subscribe to the MQTT topic to see EdgeX events
mosquitto_sub -t "edgex/events" -v

# You should see JSON messages like:
# edgex/events {"id": "...", "deviceName": "...", "readings": [...]}
```

### Connecting sfsEdgeStore to EdgeX

**Step 1: Create Configuration**

```bash
# Create config directory
mkdir -p /etc/sfsedgestore

# Create configuration file
cat > /etc/sfsedgestore/config.json << 'EOF'
{
  "HTTPPort": "8081",
  "DBPath": "/var/lib/sfsedgestore/data",
  "DBScenario": "edge",

  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-001",

  "EnableSimulator": false,
  "EnableAnalyzer": false,
  "EnableRetention": true,
  "RetentionDays": 30
}
EOF
```

**Step 2: Create Data Directory**

```bash
# Create data directory with proper permissions
mkdir -p /var/lib/sfsedgestore/data
chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore

# Or for development, use current directory
mkdir -p ./data
```

**Step 3: Start sfsEdgeStore**

```bash
# Run directly (development)
./sfsedgestore

# Or as a systemd service (production)
sudo cp sfsedgestore.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable sfsedgestore
sudo systemctl start sfsedgestore

# Check status
sudo systemctl status sfsedgestore
```

### Verifying the Integration

**Step 1: Check Health**

```bash
curl http://localhost:8081/health

# Expected response:
# {"status":"healthy","uptime":"1h23m45s","version":"1.0.0"}
```

**Step 2: Check Metrics**

```bash
curl http://localhost:8081/metrics

# Expected response includes:
# - mqtt_messages_received
# - mqtt_messages_processed
# - database_operations
```

**Step 3: Query Data**

```bash
# After EdgeX sends some events, query for readings
curl "http://localhost:8081/api/readings?deviceName=Random-Integer-Device"

# Expected response:
# {"records":[...], "count": 100}
```

**Step 4: Check Logs**

```bash
# View recent logs
journalctl -u sfsedgestore -n 50 --no-pager

# Or for direct output
tail -f /var/log/sfsedgestore/sfsedgestore.log
```

## 3.4 End-to-End Testing

Let's create a complete test to verify everything is working:

### Using the Built-in Simulator

If you don't have EdgeX running, use the simulator to test:

**Enable Simulator in Config:**

```json
{
  "HTTPPort": "8081",
  "DBPath": "./data",
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTTopic": "edgex/events",
  "ClientID": "sfsedgestore-test",
  "EnableSimulator": true
}
```

**Start and Monitor:**

```bash
# Start sfsedgestore
./sfsedgestore

# In another terminal, watch the logs
# You should see simulator generating data

# Query after 10 seconds
sleep 10
curl "http://localhost:8081/api/readings?deviceName=SimulatedDevice"

# Check metrics
curl http://localhost:8081/metrics | jq
```

### Test Scripts

Use the project's test scripts for comprehensive validation:

```bash
# Clone and run test script
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore

# Run with Mosquitto (if you have docker)
./scripts/test_with_mosquitto.sh

# Or run load test
./scripts/load_test.sh
```

### Test MQTT End-to-End

**Publisher Script:**

```bash
# Publish test message
mosquitto_pub -t "edgex/events" -m '{
  "correlationId": "test-001",
  "messageType": "event",
  "origin": 1704067200000000000,
  "payload": {
    "id": "event-001",
    "deviceName": "TestDevice",
    "readings": [
      {
        "id": "reading-001",
        "resourceName": "temperature",
        "value": "25.5",
        "valueType": "Float32",
        "origin": 1704067200000000000
      }
    ]
  }
}'
```

**Verify Reception:**

```bash
# Query the data
curl "http://localhost:8081/api/readings?deviceName=TestDevice"

# Should return the published reading
```

## 3.5 systemd Service Configuration

For production deployments on Linux, configure sfsEdgeStore as a systemd service:

### Service File

Create `/etc/systemd/system/sfsedgestore.service`:

```ini
[Unit]
Description=sfsEdgeStore - Lightweight Edge Data Storage Adapter
After=network.target mosquitto.service
Wants=mosquitto.service

[Service]
Type=simple
User=sfsedgestore
Group=sfsedgestore
WorkingDirectory=/etc/sfsedgestore
ExecStart=/usr/local/bin/sfsedgestore
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

# Resource limits
MemoryMax=256M
CPUQuota=50%

# Environment
Environment=LOG_LEVEL=info

[Install]
WantedBy=multi-user.target
```

### Installation Steps

```bash
# Create service user
sudo useradd -r -s /usr/sbin/nologin sfsedgestore

# Create directories
sudo mkdir -p /etc/sfsedgestore /var/lib/sfsedgestore /var/log/sfsedgestore

# Set permissions
sudo chown -R sfsedgestore:sfsedgestore /etc/sfsedgestore /var/lib/sfsedgestore /var/log/sfsedgestore

# Copy binary
sudo cp sfsedgestore /usr/local/bin/
sudo chmod +x /usr/local/bin/sfsedgestore

# Copy configuration
sudo cp config.json /etc/sfsedgestore/
sudo chown sfsedgestore:sfsedgestore /etc/sfsedgestore/config.json

# Install service
sudo cp sfsedgestore.service /etc/systemd/system/
sudo systemctl daemon-reload

# Start service
sudo systemctl start sfsedgestore
sudo systemctl enable sfsedgestore

# Check status
sudo systemctl status sfsedgestore
```

## 3.6 Chapter Summary

This chapter covered the essential steps to get sfsEdgeStore running in your environment.

**Key Takeaways:**

1. **Three installation methods**: Pre-built binaries (fastest), Docker (isolated), or compile from source (most flexible).

2. **Configuration is JSON-based**: Start with minimal config and add features as needed. Use scenario presets for common environments.

3. **EdgeX integration is seamless**: Point sfsEdgeStore at EdgeX's MQTT topic and data flows automatically.

4. **Systemd for production**: Configure as a systemd service for automatic startup, restart, and resource limits.

**Common Next Steps:**

- Explore configuration options in Chapter 4
- Learn troubleshooting in Chapter 5
- Dive into source code in Chapters 6-7

**What's Next:**

➡️ Next: [Chapter 4: Configuration Deep Dive](./04-Chap4-Configuration-Deep-Dive.md)

---

**Quick Reference Commands:**

```bash
# Start sfsEdgeStore
./sfsedgestore

# Health check
curl http://localhost:8081/health

# Query data
curl "http://localhost:8081/api/readings?deviceName=YourDevice"

# View logs
journalctl -u sfsedgestore -f

# Stop service
sudo systemctl stop sfsedgestore

# Restart service
sudo systemctl restart sfsedgestore
```