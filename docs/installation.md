# Installation

Complete installation guide for sfsEdgeStore.

## Supported Platforms

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | amd64 | Official |
| Linux | arm64 | Official |
| Linux | armv7 (Raspberry Pi) | Official |
| Windows | amd64 | Official |
| macOS | amd64/arm64 | Community |

## System Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 1 core | 2+ cores |
| RAM | 128 MB | 256 MB+ |
| Disk | 1 GB | 2 GB+ |

## Method 1: From Source

```bash
# Clone repository
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore

# Build
go build -ldflags="-s -w" -o sfsedgestore .

# Run
./sfsedgestore
```

## Method 2: Binary Release

Download from [GitHub Releases](https://github.com/liaoran123/sfsEdgeStore/releases):

```bash
# Linux amd64
wget https://github.com/liaoran123/sfsEdgeStore/releases/latest/download/sfsedgestore-linux-amd64.tar.gz
tar -xzf sfsedgestore-linux-amd64.tar.gz
sudo mv sfsedgestore /usr/local/bin/

# Run
sfsedgestore
```

## Method 3: Docker

```bash
docker run -d \
  --name sfsedgestore \
  -p 8081:8081 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  liaoran123/sfsedgestore:latest
```

## Method 4: Linux System Service

### 1. Create systemd service file

```ini
[Unit]
Description=sfsEdgeStore IoT Data Adapter
After=network.target mosquitto.service

[Service]
Type=simple
User=sfsedgestore
WorkingDirectory=/opt/sfsedgestore
ExecStart=/opt/sfsedgestore/sfsedgestore
Restart=on-failure
RestartSec=10
MemoryLimit=256M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
```

### 2. Install and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable sfsedgestore
sudo systemctl start sfsedgestore
sudo systemctl status sfsedgestore
```

## Verify Installation

```bash
# Health check
curl http://localhost:8081/health

# Expected response
{"status":"healthy","timestamp":1234567890}
```

## Next Steps

- [Quick Start](./quick-start.md) - Basic usage
- [Configuration](./configuration.md) - Configure for your environment
- [Troubleshooting](./troubleshooting.md) - Common issues
