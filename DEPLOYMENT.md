# sfsEdgeStore 部署文档

## 目录
1. [系统要求](#系统要求)
2. [快速开始](#快速开始)
3. [配置说明](#配置说明)
4. [Linux 部署](#linux-部署)
5. [Windows 部署](#windows-部署)
6. [Docker 部署](#docker-部署)
7. [健康检查](#健康检查)
8. [故障排查](#故障排查)

---

## 系统要求

### 硬件要求
- **CPU**: 1 核或以上
- **内存**: 128MB 或以上（推荐 256MB）
- **存储**: 1GB 或以上可用空间

### 软件要求
- **操作系统**: Linux (推荐) / Windows / macOS
- **Go**: 1.25.3 或以上（仅编译需要）
- **MQTT Broker**: EdgeX Foundry 或其他 MQTT Broker

---

## 快速开始

### 1. 编译项目

```bash
# 克隆项目
git clone <repository-url>
cd sfsEdgeStore

# 编译
go build -o sfsEdgeStore
```

### 2. 配置

```bash
# 复制示例配置
cp config.example.json config.json
# 或
cp .env.example .env

# 编辑配置文件
# 修改为实际值
```

### 3. 运行

```bash
# Linux/macOS
./sfsEdgeStore

# Windows
sfsEdgeStore.exe
```

---

## 配置说明

### 配置方式优先级
1. **环境变量**（最高优先级）
2. **配置文件**（`config.json`）
3. **默认值**

### 主要配置项

#### 数据库配置
- `db_path`: 数据库存储路径
- `db_use_encryption`: 是否启用数据库加密
- `db_encryption_key`: 加密密钥
- `db_scenario`: 数据库场景（embedded/iot/edge/game/default）

#### MQTT 配置
- `mqtt_broker`: MQTT Broker 地址
- `mqtt_topic`: 订阅主题
- `mqtt_use_tls`: 是否启用 TLS
- `client_id`: 客户端 ID

#### HTTP 服务配置
- `http_port`: HTTP 服务端口
- `http_use_tls`: 是否启用 HTTPS

#### 数据保留策略
- `enable_retention_policy`: 启用数据保留策略
- `retention_days`: 数据保留天数
- `cleanup_interval_hours`: 清理间隔（小时）

#### 资源监控
- `enable_resource_monitoring`: 启用资源监控
- `max_memory_mb`: 最大内存（MB）
- `max_cpu_percent`: 最大 CPU（%）

---

## Linux 部署

### Systemd 服务部署

#### 1. 创建服务文件

```bash
sudo tee /etc/systemd/system/sfsEdgeStore.service << 'EOF'
[Unit]
Description=sfsEdgeStore - EdgeX 边缘存储适配器
After=network.target

[Service]
Type=simple
User=edgex
Group=edgex
WorkingDirectory=/opt/sfsEdgeStore
ExecStart=/opt/sfsEdgeStore/sfsEdgeStore
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sfsEdgeStore

[Install]
WantedBy=multi-user.target
EOF
```

#### 2. 创建用户和目录

```bash
# 创建用户
sudo useradd -r -s /bin/false edgex

# 创建目录
sudo mkdir -p /opt/sfsEdgeStore
sudo mkdir -p /opt/sfsEdgeStore/edgex_data
sudo mkdir -p /opt/sfsEdgeStore/data_sync_queue

# 设置权限
sudo chown -R edgex:edgex /opt/sfsEdgeStore
```

#### 3. 安装二进制文件

```bash
# 复制二进制文件
sudo cp sfsEdgeStore /opt/sfsEdgeStore/
sudo chmod +x /opt/sfsEdgeStore/sfsEdgeStore

# 复制配置文件
sudo cp config.json /opt/sfsEdgeStore/
sudo chown edgex:edgex /opt/sfsEdgeStore/config.json
```

#### 4. 启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start sfsEdgeStore

# 设置开机自启
sudo systemctl enable sfsEdgeStore

# 查看状态
sudo systemctl status sfsEdgeStore

# 查看日志
sudo journalctl -u sfsEdgeStore -f
```

---

## Windows 部署

### Windows 服务（使用 NSSM）

#### 1. 下载 NSSM

从 https://nssm.cc/download 下载 NSSM

#### 2. 安装服务

```powershell
# 以管理员身份运行 PowerShell
nssm install sfsEdgeStore
```

在弹出的窗口中配置：
- Path: `C:\sfsEdgeStore\sfsEdgeStore.exe
- Startup directory: `C:\sfsEdgeStore`

#### 3. 启动服务

```powershell
nssm start sfsEdgeStore
```

---

## Docker 部署

### Dockerfile

```dockerfile
FROM golang:1.25.3-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o sfsEdgeStore

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/sfsEdgeStore .
COPY config.example.json config.json

EXPOSE 8081

CMD ["./sfsEdgeStore"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  sfsedgestore:
    build: .
    container_name: sfsedgestore
    ports:
      - "8081:8081"
    volumes:
      - ./edgex_data:/app/edgex_data
      - ./data_sync_queue:/app/data_sync_queue
      - ./config.json:/app/config.json:ro
    environment:
      - EDGEX_MQTT_BROKER=tcp://mqtt-broker:1883
    restart: unless-stopped
    depends_on:
      - mqtt-broker

  mqtt-broker:
    image: eclipse-mosquitto:latest
    ports:
      - "1883:1883"
    volumes:
      - mosquitto-data:/mosquitto/data

volumes:
  edgex_data:
  data_sync_queue:
  mosquitto-data:
```

### 启动 Docker Compose

```bash
docker-compose up -d
```

---

## 健康检查

### HTTP 健康检查端点

- **健康检查**: `GET /health`
- **响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### 资源状态端点

- **资源状态**: `GET /api/resources/status`

---

## 故障排查

### 常见问题

#### 1. MQTT 连接失败
- 检查 MQTT Broker 是否运行
- 检查网络连接
- 检查认证配置

#### 2. 数据库错误
- 检查数据库路径权限
- 检查磁盘空间
- 查看日志文件

#### 3. 内存超限
- 检查资源监控配置
- 调整 `max_memory_mb` 配置
- 检查数据保留策略

### 日志查看

```bash
# Linux systemd 日志
journalctl -u sfsEdgeStore -n 100

# Docker 日志
docker logs sfsedgestore --tail 100
```

---

## 附录

### 支持的数据库场景

| 场景 | 写缓冲 | 块缓存 | 适用场景 |
|--------|----------|----------|----------|
| embedded | 2MB | 4MB | 嵌入式设备 |
| iot | 4MB | 8MB | IoT 设备 |
| **edge** (默认） | 16MB | 32MB | 边缘计算 |
| game | 64MB | 128MB | 游戏场景 |
