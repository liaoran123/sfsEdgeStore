---
sidebar_position: 2
---

# Quick Start

Get sfsEdgeStore running in 5 minutes.

## Prerequisites

- Go 1.21 or later installed
- MQTT Broker (Mosquitto recommended)
- EdgeX Foundry running (optional, for testing use simulator)

## Step 1: Clone the Repository

```bash
git clone https://github.com/your-org/sfsEdgeStore.git
cd sfsEdgeStore
```

## Step 2: Build the Application

```bash
go build -o sfsEdgeStore .
```

## Step 3: Configure

Edit `config.json` with your settings:

```json
{
  "db_path": "./data",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/#",
  "client_id": "sfsdb-edgex-adapter",
  "http_port": "8081",
  "auto_subscribe": true,
  "license_type": "community"
}
```

## Step 4: Run

```bash
./sfsEdgeStore
```

You should see:

```
╔══════════════════════════════════════════╗
║         sfsEdgeStore Started             ║
╠══════════════════════════════════════════╣
║  Lightweight Industrial IoT Edge Adapter ║
║  Memory: <50MB | Ultra-light | Reliable  ║
╠══════════════════════════════════════════╣
║  Web Dashboard: http://localhost:8081    ║
║  Stop: Ctrl+C                            ║
╚══════════════════════════════════════════╝
```

## Step 5: Access the Dashboard

Open your browser and navigate to:

```
http://localhost:8081
```

## Next Steps

- [Installation Guide](./installation.md) - Detailed installation for production
- [Configure MQTT](../how-to/configure-mqtt.md) - Connect to your EdgeX instance
- [API Reference](../reference/api/overview.md) - Explore the REST API
