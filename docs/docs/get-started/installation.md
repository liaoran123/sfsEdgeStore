---
sidebar_position: 3
---

# Installation

Detailed installation guide for production environments.

## Supported Platforms

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | amd64 | ✅ Official |
| Linux | arm64 | ✅ Official |
| Linux | armv7 (Raspberry Pi) | ✅ Official |
| Windows | amd64 | ✅ Official |
| macOS | amd64/arm64 | ✅ Community |

## Method 1: From Source (Recommended)

### Install Go

```bash
# Download Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Clone and Build

```bash
git clone https://github.com/your-org/sfsEdgeStore.git
cd sfsEdgeStore
go build -ldflags="-s -w" -o sfsEdgeStore .
```

## Method 2: Binary Download

```bash
# Linux amd64
wget https://github.com/your-org/sfsEdgeStore/releases/latest/download/sfsEdgeStore-linux-amd64.tar.gz
tar -xzf sfsEdgeStore-linux-amd64.tar.gz
sudo mv sfsEdgeStore /usr/local/bin/
```

## Method 3: Docker

```bash
docker pull yourorg/sfsedgestore:latest
docker run -d \
  --name sfsedgestore \
  -p 8081:8081 \
  -v /path/to/data:/app/data \
  -v /path/to/config.json:/app/config.json \
  yourorg/sfsedgestore:latest
```

## Method 4: System Service (Linux)

Create `/etc/systemd/system/sfsedgestore.service`:

```ini
[Unit]
Description=sfsEdgeStore IoT Data Adapter
After=network.target mosquitto.service

[Service]
Type=simple
User=sfsedgestore
WorkingDirectory=/opt/sfsedgestore
ExecStart=/opt/sfsedgestore/sfsEdgeStore
Restart=on-failure
RestartSec=10

# Resource limits
MemoryLimit=256M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable sfsedgestore
sudo systemctl start sfsedgestore
```

## Verify Installation

```bash
# Check health endpoint
curl http://localhost:8081/health

# Expected response
{"status":"healthy","timestamp":1234567890}
```
