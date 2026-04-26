# Quick Start

Get sfsEdgeStore running in 5 minutes.

## Prerequisites

- Go 1.21 or later (for building from source)
- MQTT Broker (e.g., Mosquitto)
- EdgeX Foundry (optional, for real data source)

## Step 1: Get the Code

```bash
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore
```

## Step 2: Build

```bash
go build -ldflags="-s -w" -o sfsedgestore .
```

## Step 3: Run

```bash
./sfsedgestore
```

sfsEdgeStore uses intelligent defaults. No configuration required to get started:

| Setting | Default Value |
|---------|---------------|
| MQTT Broker | `tcp://localhost:1883` |
| MQTT Topic | `edgex/events/#` |
| HTTP Port | `8081` |
| Database Path | `data/sfs.db` |

## Step 4: Verify

```bash
# Health check
curl http://localhost:8081/health

# View metrics
curl http://localhost:8081/metrics
```

## Step 5: Query Data

```bash
# Query all readings
curl http://localhost:8081/api/readings

# Query by device
curl "http://localhost:8081/api/readings?deviceName=Device001&limit=10"
```

## Next Steps

- [Installation](./installation.md) - Production deployment guide
- [Configuration](./configuration.md) - All available options
- [API Reference](./api-reference.md) - REST API documentation
