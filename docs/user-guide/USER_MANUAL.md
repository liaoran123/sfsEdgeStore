# sfsEdgeStore 用户手册

> **文档版本**: v1.0.0  
> **最后更新**: 2026-03-07  
> **适用版本**: sfsEdgeStore v1.x

---

## 目录

1. [概述](#概述)
2. [快速开始](#快速开始)
3. [安装部署](#安装部署)
4. [配置说明](#配置说明)
5. [基本使用](#基本使用)
6. [高级功能](#高级功能)
7. [API 使用指南](#api-使用指南)
8. [最佳实践](#最佳实践)

---

## 概述

### 什么是 sfsEdgeStore？

sfsEdgeStore 是一个轻量级边缘计算数据存储适配器，专为边缘设备设计，作为 EdgeX Foundry 和 sfsDb 数据库之间的桥梁。

### 核心特性

- **轻量级部署**: 内存占用 &lt; 50MB，CPU 使用率 &lt; 5%
- **毫秒级启动**: 平均启动时间仅 0.187 秒
- **本地数据存储**: 使用 sfsDb 高效存储边缘数据
- **MQTT 数据接入**: 无缝集成 EdgeX Foundry
- **实时监控**: 内置系统指标和业务指标监控
- **智能告警**: 阈值告警和异常检测
- **认证授权**: API Key 和 RBAC 权限控制
- **断网运行**: 网络中断不影响本地数据采集

### 适用场景

- 工业物联网边缘数据采集
- 智能设备本地数据缓存
- 边缘节点数据预处理
- 离线数据存储和同步

---

## 快速开始

### 前置条件

- **操作系统**: Linux, Windows, macOS
- **Go 版本**: 1.25+ (仅编译时需要)
- **MQTT Broker**: Mosquitto 或其他兼容的 MQTT 服务器
- **EdgeX Foundry**: (可选) 用于数据源

### 5 分钟快速上手

#### 1. 获取程序

```bash
# 从源码编译
git clone https://github.com/your-org/sfsEdgeStore.git
cd sfsEdgeStore
go build -o sfsedgestore

# 或从 Releases 下载预编译二进制
# https://github.com/your-org/sfsEdgeStore/releases
```

#### 2. 配置

```bash
# 复制配置模板
cp config.example.json config.json

# 编辑配置文件
vim config.json
```

基本配置示例：

```json
{
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTClientID": "sfsedgestore",
  "MQTTTopic": "edgex/events/core/#",
  "DBPath": "./data",
  "DBUseEncryption": false,
  "HTTPPort": 8080
}
```

#### 3. 启动 MQTT Broker

```bash
# 使用 Mosquitto
mosquitto -p 1883
```

#### 4. 运行 sfsEdgeStore

```bash
./sfsedgestore
```

#### 5. 验证安装

```bash
# 健康检查
curl http://localhost:8080/health

# 查看指标
curl http://localhost:8080/metrics
```

---

## 安装部署

### 二进制部署

#### Linux

```bash
# 下载
wget https://github.com/your-org/sfsEdgeStore/releases/download/v1.0.0/sfsedgestore-linux-amd64
chmod +x sfsedgestore-linux-amd64
mv sfsedgestore-linux-amd64 /usr/local/bin/sfsedgestore

# 创建数据目录
mkdir -p /var/lib/sfsedgestore/data
mkdir -p /etc/sfsedgestore

# 配置
cp config.example.json /etc/sfsedgestore/config.json

# 运行
sfsedgestore -config /etc/sfsedgestore/config.json
```

#### Windows

```powershell
# 下载解压
# 编辑 config.json

# 运行
.\sfsedgestore.exe
```

#### macOS

```bash
# 下载
wget https://github.com/your-org/sfsEdgeStore/releases/download/v1.0.0/sfsedgestore-darwin-arm64
chmod +x sfsedgestore-darwin-arm64

# 运行
./sfsedgestore-darwin-arm64
```

### Docker 部署

#### 使用官方镜像

```bash
docker pull your-org/sfsedgestore:latest
```

#### 运行容器

```bash
docker run -d \
  --name sfsedgestore \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./config.json:/app/config.json \
  --restart unless-stopped \
  your-org/sfsedgestore:latest
```

#### Docker Compose

创建 `docker-compose.yml`:

```yaml
version: '3'
services:
  sfsedgestore:
    image: your-org/sfsedgestore:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./config.json:/app/config.json
    restart: unless-stopped
    depends_on:
      - mosquitto

  mosquitto:
    image: eclipse-mosquitto:latest
    ports:
      - "1883:1883"
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf
```

### Systemd 服务部署

#### 创建服务文件

```ini
# /etc/systemd/system/sfsedgestore.service
[Unit]
Description=sfsEdgeStore Edge Data Adapter
After=network.target mosquitto.service

[Service]
Type=simple
User=sfsedgestore
Group=sfsedgestore
WorkingDirectory=/var/lib/sfsedgestore
ExecStart=/usr/local/bin/sfsedgestore -config /etc/sfsedgestore/config.json
Restart=always
RestartSec=10
Environment="GOMAXPROCS=2"

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sfsedgestore

[Install]
WantedBy=multi-user.target
```

#### 创建用户和目录

```bash
# 创建用户
useradd -r -s /bin/false sfsedgestore

# 创建目录
mkdir -p /var/lib/sfsedgestore/data
mkdir -p /etc/sfsedgestore
chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore
chown -R sfsedgestore:sfsedgestore /etc/sfsedgestore
```

#### 启用服务

```bash
# 重载 systemd
systemctl daemon-reload

# 启用开机自启
systemctl enable sfsedgestore

# 启动服务
systemctl start sfsedgestore

# 查看状态
systemctl status sfsedgestore

# 查看日志
journalctl -u sfsedgestore -f
```

### Kubernetes 部署

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sfsedgestore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sfsedgestore
  template:
    metadata:
      labels:
        app: sfsedgestore
    spec:
      containers:
      - name: sfsedgestore
        image: your-org/sfsedgestore:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: data
          mountPath: /app/data
        - name: config
          mountPath: /app/config.json
          subPath: config.json
        resources:
          requests:
            cpu: "100m"
            memory: "64Mi"
          limits:
            cpu: "500m"
            memory: "128Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: sfsedgestore-data
      - name: config
        configMap:
          name: sfsedgestore-config
---
apiVersion: v1
kind: Service
metadata:
  name: sfsedgestore
spec:
  selector:
    app: sfsedgestore
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sfsedgestore-data
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

---

## 配置说明

### 配置文件结构

sfsEdgeStore 使用 JSON 格式的配置文件，默认读取当前目录下的 `config.json`。

### 完整配置示例

```json
{
  "MQTTBroker": "tcp://localhost:1883",
  "MQTTClientID": "sfsedgestore",
  "MQTTUsername": "",
  "MQTTPassword": "",
  "MQTTTopic": "edgex/events/core/#",
  "MQTTQoS": 1,
  "MQTTUseTLS": false,
  "MQTTCACert": "",
  "MQTTClientCert": "",
  "MQTTClientKey": "",

  "DBPath": "./data",
  "DBUseEncryption": false,
  "DBEncryptionKey": "",
  "DBCompression": true,
  "DBMaxOpenFiles": 1000,

  "HTTPPort": 8080,
  "HTTPHost": "0.0.0.0",
  "HTTPUseTLS": false,
  "HTTPCert": "",
  "HTTPKey": "",
  "HTTPReadTimeout": 30,
  "HTTPWriteTimeout": 30,
  "HTTPMaxConnections": 100,

  "EnableAuth": true,
  "AuthAPIKeyRequired": true,

  "EnableMonitor": true,
  "MonitorInterval": 60,
  "EnableAlert": true,
  "AlertThresholds": {
    "cpu": 80,
    "memory": 85,
    "disk": 90
  },

  "EnableRetention": true,
  "RetentionDays": 30,
  "RetentionCheckInterval": 3600,

  "EnableSync": false,
  "SyncEndpoint": "",
  "SyncInterval": 300,
  "SyncAPIKey": "",

  "LogLevel": "info",
  "LogFormat": "text",
  "LogOutput": "stdout",

  "QueueSize": 10000,
  "QueueMaxRetries": 3,
  "QueueRetryInterval": 5,

  "EnableSimulator": false,
  "SimulatorInterval": 1000
}
```

### 配置项详解

#### MQTT 配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `MQTTBroker` | MQTT Broker 地址 | `tcp://localhost:1883` | 是 |
| `MQTTClientID` | MQTT 客户端 ID | `sfsedgestore` | 是 |
| `MQTTUsername` | MQTT 用户名 | `""` | 否 |
| `MQTTPassword` | MQTT 密码 | `""` | 否 |
| `MQTTTopic` | 订阅主题 | `edgex/events/core/#` | 是 |
| `MQTTQoS` | 消息 QoS 级别 | `1` | 否 |
| `MQTTUseTLS` | 是否启用 TLS | `false` | 否 |
| `MQTTCACert` | CA 证书路径 | `""` | 否 |
| `MQTTClientCert` | 客户端证书路径 | `""` | 否 |
| `MQTTClientKey` | 客户端密钥路径 | `""` | 否 |

#### 数据库配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `DBPath` | 数据库存储路径 | `./data` | 是 |
| `DBUseEncryption` | 是否启用加密 | `false` | 否 |
| `DBEncryptionKey` | 加密密钥 | `""` | 否 |
| `DBCompression` | 是否启用压缩 | `true` | 否 |
| `DBMaxOpenFiles` | 最大打开文件数 | `1000` | 否 |

#### HTTP 服务配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `HTTPPort` | HTTP 服务端口 | `8080` | 是 |
| `HTTPHost` | 监听地址 | `0.0.0.0` | 否 |
| `HTTPUseTLS` | 是否启用 HTTPS | `false` | 否 |
| `HTTPCert` | TLS 证书路径 | `""` | 否 |
| `HTTPKey` | TLS 密钥路径 | `""` | 否 |
| `HTTPReadTimeout` | 读取超时(秒) | `30` | 否 |
| `HTTPWriteTimeout` | 写入超时(秒) | `30` | 否 |
| `HTTPMaxConnections` | 最大连接数 | `100` | 否 |

#### 认证配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `EnableAuth` | 是否启用认证 | `true` | 否 |
| `AuthAPIKeyRequired` | 是否要求 API Key | `true` | 否 |

#### 监控告警配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `EnableMonitor` | 是否启用监控 | `true` | 否 |
| `MonitorInterval` | 监控间隔(秒) | `60` | 否 |
| `EnableAlert` | 是否启用告警 | `true` | 否 |
| `AlertThresholds.cpu` | CPU 告警阈值(%) | `80` | 否 |
| `AlertThresholds.memory` | 内存告警阈值(%) | `85` | 否 |
| `AlertThresholds.disk` | 磁盘告警阈值(%) | `90` | 否 |

#### 数据保留配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `EnableRetention` | 是否启用数据保留 | `true` | 否 |
| `RetentionDays` | 数据保留天数 | `30` | 否 |
| `RetentionCheckInterval` | 检查间隔(秒) | `3600` | 否 |

#### 数据同步配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `EnableSync` | 是否启用数据同步 | `false` | 否 |
| `SyncEndpoint` | 同步端点 | `""` | 否 |
| `SyncInterval` | 同步间隔(秒) | `300` | 否 |
| `SyncAPIKey` | 同步 API Key | `""` | 否 |

#### 日志配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `LogLevel` | 日志级别 | `info` | 否 |
| `LogFormat` | 日志格式 | `text` | 否 |
| `LogOutput` | 日志输出 | `stdout` | 否 |

日志级别: `debug`, `info`, `warn`, `error`  
日志格式: `text`, `json`

#### 队列配置

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `QueueSize` | 队列大小 | `10000` | 否 |
| `QueueMaxRetries` | 最大重试次数 | `3` | 否 |
| `QueueRetryInterval` | 重试间隔(秒) | `5` | 否 |

### 环境变量配置

所有配置项都可以通过环境变量设置，格式为 `EDGEX_配置项名`（大写）：

```bash
export EDGEX_MQTT_BROKER="tcp://mqtt.example.com:1883"
export EDGEX_HTTP_PORT="8081"
export EDGEX_DB_PATH="/var/lib/sfsedgestore/data"
export EDGEX_LOG_LEVEL="debug"
```

### 命令行参数

```bash
./sfsedgestore -config /path/to/config.json -port 8081 -debug
```

---

## 基本使用

### 启动和停止

#### 前台运行

```bash
./sfsedgestore
```

#### 后台运行（Linux/macOS）

```bash
nohup ./sfsedgestore > sfsedgestore.log 2>&1 &
```

#### 停止服务

```bash
# 查找进程
ps aux | grep sfsedgestore

# 优雅停止
kill -TERM <pid>

# 强制停止
kill -9 <pid>
```

### 数据查询

#### 查询所有数据

```bash
curl "http://localhost:8080/api/readings"
```

#### 按设备查询

```bash
curl "http://localhost:8080/api/readings?deviceName=Device001"
```

#### 按时间范围查询

```bash
curl "http://localhost:8080/api/readings?startTime=2026-03-01T00:00:00Z&endTime=2026-03-07T23:59:59Z"
```

#### 分页查询

```bash
curl "http://localhost:8080/api/readings?limit=100&offset=0"
```

#### 组合查询

```bash
curl "http://localhost:8080/api/readings?deviceName=Device001&startTime=2026-03-01T00:00:00Z&limit=50"
```

### 数据导出

#### 导出 JSON 格式

```bash
curl "http://localhost:8080/api/readings?format=json" > data.json
```

#### 导出 CSV 格式

```bash
curl "http://localhost:8080/api/readings?format=csv" > data.csv
```

### 健康检查和监控

#### 健康检查

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": "2026-03-07T10:00:00Z",
  "version": "1.0.0",
  "uptime": "3600s"
}
```

#### 获取指标

```bash
curl http://localhost:8080/metrics
```

响应示例：

```json
{
  "system": {
    "cpu_usage": 2.5,
    "memory_usage": 20.8,
    "disk_usage": 45.2,
    "goroutines": 15
  },
  "business": {
    "total_readings": 18681,
    "mqtt_messages_received": 25000,
    "mqtt_messages_processed": 24980,
    "queue_length": 0
  },
  "database": {
    "size_mb": 0.25,
    "read_count": 150000,
    "write_count": 25000
  }
}
```

#### 查看告警

```bash
curl http://localhost:8080/alerts
```

---

## 高级功能

### 认证授权

#### 创建 API Key

首次启动时，需要创建管理员 API Key：

```bash
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 8760
  }'
```

响应：

```json
{
  "status": "success",
  "api_key": "sk_abc123def456...",
  "user_id": "admin",
  "role": "admin",
  "expires_at": "2027-03-07T10:00:00Z"
}
```

#### 使用 API Key

```bash
curl -H "X-API-Key: sk_abc123def456..." \
  "http://localhost:8080/api/readings"
```

#### 角色权限

| 角色 | 权限 |
|------|------|
| `admin` | 所有权限 |
| `user` | 读写数据、备份 |
| `readonly` | 只读数据 |

### 数据备份和恢复

#### 手动备份

```bash
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=./backups"
```

#### 自动备份

配置 cron 任务：

```bash
# 每天凌晨 2 点备份
0 2 * * * curl -X POST -H "X-API-Key: <key>" "http://localhost:8080/api/backup?path=/backup"
```

#### 恢复数据

```bash
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/restore?file=./backups/backup-20260307.db"
```

### 数据保留策略

#### 配置保留策略

在 `config.json` 中设置：

```json
{
  "EnableRetention": true,
  "RetentionDays": 90,
  "RetentionCheckInterval": 3600
}
```

这将保留最近 90 天的数据，每小时检查一次。

### TLS/SSL 配置

#### 生成自签名证书

```bash
# 生成私钥
openssl genrsa -out key.pem 2048

# 生成证书
openssl req -new -x509 -days 365 -key key.pem -out cert.pem
```

#### 配置 HTTPS

```json
{
  "HTTPUseTLS": true,
  "HTTPCert": "cert.pem",
  "HTTPKey": "key.pem"
}
```

#### 配置 MQTT TLS

```json
{
  "MQTTUseTLS": true,
  "MQTTCACert": "ca.pem",
  "MQTTClientCert": "client.pem",
  "MQTTClientKey": "client.key"
}
```

---

## API 使用指南

完整的 API 参考文档请查看 [API 文档](../api.md)。

### 数据查询 API

#### GET /api/readings

查询读数数据。

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `deviceName` | string | 设备名称过滤 |
| `resourceName` | string | 资源名称过滤 |
| `startTime` | string | 开始时间 (RFC3339) |
| `endTime` | string | 结束时间 (RFC3339) |
| `limit` | int | 返回数量限制 |
| `offset` | int | 偏移量 |
| `format` | string | 输出格式 (json/csv) |

**示例：**

```bash
# 最近 100 条记录
curl "http://localhost:8080/api/readings?limit=100"

# 指定设备和时间范围
curl "http://localhost:8080/api/readings?deviceName=TempSensor&startTime=2026-03-01T00:00:00Z"
```

### 监控 API

#### GET /health

健康检查。

#### GET /metrics

获取系统指标。

#### GET /alerts

获取告警信息。

### 认证 API

#### POST /api/auth/create-key

创建 API Key。

**请求体：**

```json
{
  "user_id": "user1",
  "role": "user",
  "expires_in": 720
}
```

#### GET /api/auth/list-keys

列出所有 API Key（需要 admin 权限）。

#### POST /api/auth/revoke-key

撤销 API Key（需要 admin 权限）。

**请求体：**

```json
{
  "api_key": "sk_..."
}
```

### 备份恢复 API

#### POST /api/backup

创建备份。

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `path` | string | 备份目录路径 |

#### POST /api/restore

恢复备份。

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `file` | string | 备份文件路径 |

---

## 最佳实践

### 部署最佳实践

1. **资源限制**

```bash
# 使用 systemd 限制资源
[Service]
MemoryMax=128M
CPUQuota=50%
```

2. **数据持久化**
   - 使用独立的数据分区
   - 配置定期备份
   - 监控磁盘空间

3. **高可用性**
   - 配置 MQTT Broker 集群
   - 使用负载均衡（多节点部署）
   - 配置自动故障转移

### 性能优化

1. **数据库优化**

```json
{
  "DBCompression": true,
  "DBMaxOpenFiles": 2000
}
```

2. **MQTT 优化**

```json
{
  "MQTTQoS": 1,
  "QueueSize": 50000
}
```

3. **HTTP 优化**

```json
{
  "HTTPMaxConnections": 200,
  "HTTPReadTimeout": 60
}
```

### 安全最佳实践

1. **启用认证**

```json
{
  "EnableAuth": true,
  "AuthAPIKeyRequired": true
}
```

2. **使用 TLS**

```json
{
  "HTTPUseTLS": true,
  "MQTTUseTLS": true
}
```

3. **API Key 管理**
   - 定期轮换 API Key
   - 使用不同的 Key 给不同的应用
   - 设置合理的过期时间
   - 及时撤销不再使用的 Key

### 监控和维护

1. **关键指标监控**
   - CPU 使用率
   - 内存使用率
   - 磁盘使用率
   - MQTT 消息处理速率
   - 数据库大小

2. **日志管理**

```json
{
  "LogLevel": "info",
  "LogFormat": "json",
  "LogOutput": "/var/log/sfsedgestore/app.log"
}
```

3. **定期维护**
   - 每日检查日志
   - 每周检查备份
   - 每月检查磁盘空间
   - 每季度进行性能评估

### EdgeX 集成最佳实践

1. **主题配置**

```json
{
  "MQTTTopic": "edgex/events/core/device/+/+/+/+"
}
```

2. **数据过滤**
   - 在 EdgeX 中配置过滤器
   - 只转发需要的数据
   - 使用资源级别过滤

3. **错误处理**
   - 监控 EdgeX 连接状态
   - 配置重试机制
   - 记录失败的消息

---

## 附录

### 术语表

| 术语 | 说明 |
|------|------|
| EdgeX Foundry | 边缘计算开源框架 |
| sfsDb | 轻量级嵌入式数据库 |
| MQTT | 消息队列遥测传输协议 |
| QoS | 服务质量级别 |
| API Key | 应用程序接口密钥 |
| RBAC | 基于角色的访问控制 |
| TLS | 传输层安全协议 |

### 常见数据格式

#### EdgeX 事件格式

```json
{
  "apiVersion": "v3",
  "id": "uuid",
  "deviceName": "Device001",
  "profileName": "Profile001",
  "sourceName": "Source001",
  "origin": 1234567890,
  "readings": [
    {
      "id": "uuid",
      "origin": 1234567890,
      "deviceName": "Device001",
      "resourceName": "Temperature",
      "profileName": "Profile001",
      "value": "25.5",
      "valueType": "Float32"
    }
  ]
}
```

### 参考资料

- [EdgeX Foundry 官方文档](https://docs.edgexfoundry.org/)
- [MQTT 协议规范](http://mqtt.org/)
- [sfsDb 文档](https://github.com/liaoran123/sfsDb)
- [项目 README](../../README.md)
- [性能报告](../../PERFORMANCE_REPORT.md)
- [故障排除指南](./TROUBLESHOOTING.md)
- [常见问题](./FAQ.md)

---

**文档结束**

如需帮助，请参考 [支持文档](../support/SUPPORT.md) 或联系技术支持。
