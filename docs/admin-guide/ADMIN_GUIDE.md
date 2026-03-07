# sfsEdgeStore 管理员指南

> **文档版本**: v1.0.0  
> **最后更新**: 2026-03-07  
> **适用版本**: sfsEdgeStore v1.x  
> **目标读者**: 系统管理员、运维工程师

---

## 目录

1. [概述](#概述)
2. [系统要求](#系统要求)
3. [部署管理](#部署管理)
4. [配置管理](#配置管理)
5. [用户和权限管理](#用户和权限管理)
6. [监控和告警](#监控和告警)
7. [备份和恢复](#备份和恢复)
8. [升级和维护](#升级和维护)
9. [安全管理](#安全管理)
10. [性能优化](#性能优化)
11. [故障诊断](#故障诊断)

---

## 概述

### 管理员职责

作为 sfsEdgeStore 管理员，您的主要职责包括：

- 系统部署和配置
- 用户和权限管理
- 系统监控和告警处理
- 数据备份和恢复
- 系统升级和维护
- 安全策略实施
- 性能优化
- 故障排查和处理

### 关键概念

| 术语 | 说明 |
|------|------|
| 边缘节点 | 部署 sfsEdgeStore 的设备 |
| MQTT Broker | 消息队列代理服务器 |
| API Key | 用于认证的访问密钥 |
| RBAC | 基于角色的访问控制 |
| 数据保留 | 自动清理过期数据的策略 |

---

## 系统要求

### 硬件要求

#### 最低配置

| 资源 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 1 核 | 2 核或以上 |
| 内存 | 64 MB | 128 MB 或以上 |
| 存储 | 500 MB | 1 GB 或以上 |
| 网络 | 10 Mbps | 100 Mbps 或以上 |

#### 生产环境推荐配置

| 资源 | 小型部署 | 中型部署 | 大型部署 |
|------|----------|----------|----------|
| CPU | 2 核 | 4 核 | 8 核 |
| 内存 | 128 MB | 256 MB | 512 MB |
| 存储 | 1 GB | 10 GB | 100 GB |
| 网络 | 100 Mbps | 1 Gbps | 1 Gbps |
| 并发连接 | 100 | 1000 | 10000 |

### 软件要求

#### 操作系统

- **Linux** (推荐): Ubuntu 20.04+, Debian 11+, CentOS 8+, RHEL 8+
- **Windows**: Windows 10+, Windows Server 2019+
- **macOS**: macOS 11+ (仅用于开发和测试)

#### 依赖软件

- **MQTT Broker**: Eclipse Mosquitto 2.0+, EMQX 5.0+, 或其他兼容 MQTT 3.1.1/5.0 的 Broker
- **Go**: 1.25+ (仅编译时需要)
- **OpenSSL**: 1.1.1+ (如需 TLS 支持)

### 网络要求

#### 端口使用

| 端口 | 协议 | 用途 | 必填 |
|------|------|------|------|
| 8080 | TCP | HTTP API 服务 | 是 |
| 1883 | TCP | MQTT (非 TLS) | 是 |
| 8883 | TCP | MQTT (TLS) | 否 |
| 8443 | TCP | HTTPS API 服务 | 否 |

#### 防火墙规则

```bash
# Linux (ufw)
sudo ufw allow 8080/tcp
sudo ufw allow 1883/tcp

# Linux (firewalld)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=1883/tcp
sudo firewall-cmd --reload

# Windows (PowerShell)
New-NetFirewallRule -DisplayName "sfsEdgeStore HTTP" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "sfsEdgeStore MQTT" -Direction Inbound -LocalPort 1883 -Protocol TCP -Action Allow
```

---

## 部署管理

### 部署前检查清单

- [ ] 硬件资源满足要求
- [ ] 操作系统已更新到最新版本
- [ ] 防火墙规则已配置
- [ ] MQTT Broker 已部署并运行
- [ ] 网络连接测试通过
- [ ] DNS 解析正常（如需要）
- [ ] NTP 时间同步已配置
- [ ] 备份策略已制定
- [ ] 监控方案已准备

### 二进制部署

#### Linux 详细部署

##### 1. 准备系统

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 创建专用用户
sudo useradd -r -s /bin/false -d /var/lib/sfsedgestore sfsedgestore

# 创建目录结构
sudo mkdir -p /opt/sfsedgestore
sudo mkdir -p /var/lib/sfsedgestore/data
sudo mkdir -p /var/lib/sfsedgestore/backups
sudo mkdir -p /var/log/sfsedgestore
sudo mkdir -p /etc/sfsedgestore

# 设置权限
sudo chown -R sfsedgestore:sfsedgestore /opt/sfsedgestore
sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore
sudo chown -R sfsedgestore:sfsedgestore /var/log/sfsedgestore
sudo chown -R sfsedgestore:sfsedgestore /etc/sfsedgestore
```

##### 2. 安装程序

```bash
# 下载或上传二进制文件
cd /tmp
wget https://github.com/your-org/sfsEdgeStore/releases/download/v1.0.0/sfsedgestore-linux-amd64
chmod +x sfsedgestore-linux-amd64
sudo mv sfsedgestore-linux-amd64 /opt/sfsedgestore/sfsedgestore

# 复制配置文件
sudo cp config.example.json /etc/sfsedgestore/config.json
sudo chown sfsedgestore:sfsedgestore /etc/sfsedgestore/config.json

# 编辑配置
sudo -u sfsedgestore vim /etc/sfsedgestore/config.json
```

##### 3. 创建 Systemd 服务

```ini
# /etc/systemd/system/sfsedgestore.service
[Unit]
Description=sfsEdgeStore Edge Data Storage Service
Documentation=https://docs.example.com/sfsedgestore
After=network.target mosquitto.service
Wants=network-online.target

[Service]
Type=simple
User=sfsedgestore
Group=sfsedgestore
WorkingDirectory=/var/lib/sfsedgestore
ExecStart=/opt/sfsedgestore/sfsedgestore -config /etc/sfsedgestore/config.json
Restart=always
RestartSec=10
StartLimitInterval=0

# 资源限制
MemoryMax=256M
CPUQuota=50%
TasksMax=100

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sfsedgestore /var/log/sfsedgestore
PrivateDevices=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
LockPersonality=true
MemoryDenyWriteExecute=true

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sfsedgestore

[Install]
WantedBy=multi-user.target
```

##### 4. 启用并启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable sfsedgestore

# 启动服务
sudo systemctl start sfsedgestore

# 检查状态
sudo systemctl status sfsedgestore

# 查看日志
sudo journalctl -u sfsedgestore -f -n 100
```

#### Windows 详细部署

##### 1. 准备系统

```powershell
# 以管理员身份运行 PowerShell

# 创建目录
New-Item -ItemType Directory -Path "C:\Program Files\sfsEdgeStore" -Force
New-Item -ItemType Directory -Path "C:\ProgramData\sfsEdgeStore\data" -Force
New-Item -ItemType Directory -Path "C:\ProgramData\sfsEdgeStore\backups" -Force
New-Item -ItemType Directory -Path "C:\ProgramData\sfsEdgeStore\logs" -Force
```

##### 2. 安装 NSSM (Non-Sucking Service Manager)

```powershell
# 下载 NSSM
Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile "nssm.zip"
Expand-Archive -Path "nssm.zip" -DestinationPath "."
cd nssm-2.24\win64
```

##### 3. 安装 Windows 服务

```powershell
# 安装服务
.\nssm.exe install sfsEdgeStore

# 配置服务（交互方式）
# Path: C:\Program Files\sfsEdgeStore\sfsedgestore.exe
# Startup directory: C:\ProgramData\sfsEdgeStore
# Arguments: -config C:\ProgramData\sfsEdgeStore\config.json

# 或使用命令行配置
.\nssm.exe set sfsEdgeStore Application "C:\Program Files\sfsEdgeStore\sfsedgestore.exe"
.\nssm.exe set sfsEdgeStore AppDirectory "C:\ProgramData\sfsEdgeStore"
.\nssm.exe set sfsEdgeStore AppParameters "-config C:\ProgramData\sfsEdgeStore\config.json"
.\nssm.exe set sfsEdgeStore DisplayName "sfsEdgeStore Edge Storage"
.\nssm.exe set sfsEdgeStore Description "sfsEdgeStore - Edge Computing Data Storage Service"
.\nssm.exe set sfsEdgeStart SERVICE_AUTO_START
.\nssm.exe set sfsEdgeAppExit DEFAULT RESTART
.\nssm.exe set sfsEdgeAppRestartDelay 10000
```

##### 4. 启动服务

```powershell
# 启动服务
Start-Service sfsEdgeStore

# 检查状态
Get-Service sfsEdgeStore

# 查看事件日志
Get-WinEvent -LogName Application -Source sfsEdgeStore -MaxEvents 50
```

### 容器化部署

#### Docker 单节点部署

```bash
# 创建数据卷
docker volume create sfsedgestore_data
docker volume create sfsedgestore_config

# 复制配置文件
docker run --rm -v sfsedgestore_config:/config -v $(pwd):/tmp alpine cp /tmp/config.json /config/

# 运行容器
docker run -d \
  --name sfsedgestore \
  --hostname sfsedgestore-01 \
  --network sfsedgestore-net \
  -p 8080:8080 \
  -p 1883:1883 \
  -v sfsedgestore_data:/app/data \
  -v sfsedgestore_config:/app/config \
  --memory 256m \
  --cpus 0.5 \
  --restart unless-stopped \
  --health-cmd="curl -f http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  --health-start-period=60s \
  your-org/sfsedgestore:latest
```

#### Docker Compose 高可用部署

```yaml
version: '3.8'

services:
  sfsedgestore:
    image: your-org/sfsedgestore:latest
    hostname: sfsedgestore-01
    networks:
      - sfsedgestore-net
    ports:
      - "8080:8080"
    volumes:
      - sfsedgestore-data:/app/data
      - ./config.json:/app/config.json:ro
      - ./logs:/app/logs
    environment:
      - EDGEX_LOG_LEVEL=info
      - EDGEX_ENABLE_MONITOR=true
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
      restart_policy:
        condition: on-failure
        delay: 10s
        max_attempts: 5
        window: 120s
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    depends_on:
      - mqtt-broker
      - redis

  mqtt-broker:
    image: eclipse-mosquitto:2.0
    networks:
      - sfsedgestore-net
    ports:
      - "1883:1883"
    volumes:
      - mosquitto-config:/mosquitto/config
      - mosquitto-data:/mosquitto/data
      - mosquitto-logs:/mosquitto/logs
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.25'
          memory: 128M

  redis:
    image: redis:7-alpine
    networks:
      - sfsedgestore-net
    volumes:
      - redis-data:/data
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '0.25'
          memory: 64M

  haproxy:
    image: haproxy:2.8-alpine
    networks:
      - sfsedgestore-net
    ports:
      - "80:8080"
      - "8404:8404"
    volumes:
      - ./haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.25'
          memory: 64M
    depends_on:
      - sfsedgestore

networks:
  sfsedgestore-net:
    driver: overlay

volumes:
  sfsedgestore-data:
  mosquitto-config:
  mosquitto-data:
  mosquitto-logs:
  redis-data:
```

### Kubernetes 部署

#### 基础 Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sfsedgestore
  namespace: sfsedgestore
  labels:
    app: sfsedgestore
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: sfsedgestore
  template:
    metadata:
      labels:
        app: sfsedgestore
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - sfsedgestore
              topologyKey: kubernetes.io/hostname
      containers:
      - name: sfsedgestore
        image: your-org/sfsedgestore:1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 8080
          protocol: TCP
        env:
        - name: EDGEX_LOG_LEVEL
          value: "info"
        - name: EDGEX_ENABLE_MONITOR
          value: "true"
        - name: EDGEX_DB_PATH
          value: "/data"
        envFrom:
        - configMapRef:
            name: sfsedgestore-config
        - secretRef:
            name: sfsedgestore-secrets
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /app/config.json
          subPath: config.json
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "256Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
          successThreshold: 1
        readinessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
          successThreshold: 1
        startupProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 30
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: sfsedgestore-data
      - name: config
        configMap:
          name: sfsedgestore-config
          items:
          - key: config.json
            path: config.json
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
      terminationGracePeriodSeconds: 30

---
apiVersion: v1
kind: Service
metadata:
  name: sfsedgestore
  namespace: sfsedgestore
  labels:
    app: sfsedgestore
spec:
  type: ClusterIP
  selector:
    app: sfsedgestore
  ports:
  - name: http
    port: 8080
    targetPort: http
    protocol: TCP

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sfsedgestore-data
  namespace: sfsedgestore
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: sfsedgestore-config
  namespace: sfsedgestore
data:
  config.json: |
    {
      "MQTTBroker": "tcp://mqtt-broker:1883",
      "MQTTClientID": "sfsedgestore",
      "MQTTTopic": "edgex/events/core/#",
      "DBPath": "/data",
      "HTTPPort": 8080,
      "EnableMonitor": true,
      "EnableRetention": true,
      "RetentionDays": 30
    }

---
apiVersion: v1
kind: Secret
metadata:
  name: sfsedgestore-secrets
  namespace: sfsedgestore
type: Opaque
stringData:
  DB_ENCRYPTION_KEY: "your-encryption-key-here"
  MQTT_PASSWORD: "your-mqtt-password-here"

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sfsedgestore
  namespace: sfsedgestore
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - sfsedgestore.example.com
    secretName: sfsedgestore-tls
  rules:
  - host: sfsedgestore.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: sfsedgestore
            port:
              number: 8080

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: sfsedgestore
  namespace: sfsedgestore
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: sfsedgestore
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 配置管理

### 配置文件管理

#### 配置文件模板

```json
{
  "_comment": "sfsEdgeStore Configuration File",
  "_version": "1.0",

  "MQTTBroker": "tcp://localhost:1883",
  "MQTTClientID": "sfsedgestore-${HOSTNAME}",
  "MQTTUsername": "",
  "MQTTPassword": "",
  "MQTTTopic": "edgex/events/core/#",
  "MQTTQoS": 1,
  "MQTTKeepAlive": 60,
  "MQTTCleanSession": true,
  "MQTTUseTLS": false,
  "MQTTCACert": "/etc/sfsedgestore/certs/ca.pem",
  "MQTTClientCert": "/etc/sfsedgestore/certs/client.pem",
  "MQTTClientKey": "/etc/sfsedgestore/certs/client.key",

  "DBPath": "/var/lib/sfsedgestore/data",
  "DBUseEncryption": false,
  "DBEncryptionKey": "${DB_ENCRYPTION_KEY}",
  "DBCompression": true,
  "DBMaxOpenFiles": 1000,
  "DBWriteBufferSize": 16777216,
  "DBBlockCacheSize": 33554432,

  "HTTPPort": 8080,
  "HTTPHost": "0.0.0.0",
  "HTTPUseTLS": false,
  "HTTPCert": "/etc/sfsedgestore/certs/server.pem",
  "HTTPKey": "/etc/sfsedgestore/certs/server.key",
  "HTTPReadTimeout": 30,
  "HTTPWriteTimeout": 30,
  "HTTPIdleTimeout": 60,
  "HTTPMaxConnections": 100,
  "HTTPMaxHeaderBytes": 1048576,

  "EnableAuth": true,
  "AuthAPIKeyRequired": true,
  "AuthRateLimit": 1000,
  "AuthRateLimitWindow": 60,

  "EnableMonitor": true,
  "MonitorInterval": 60,
  "EnableAlert": true,
  "AlertThresholds": {
    "cpu": 80,
    "memory": 85,
    "disk": 90,
    "mqtt_lag": 1000
  },
  "AlertNotifiers": [
    {
      "type": "webhook",
      "url": "https://api.example.com/alerts",
      "secret": "${ALERT_WEBHOOK_SECRET}"
    }
  ],

  "EnableRetention": true,
  "RetentionDays": 30,
  "RetentionCheckInterval": 3600,
  "RetentionBatchSize": 1000,

  "EnableSync": false,
  "SyncEndpoint": "https://cloud.example.com/api/sync",
  "SyncInterval": 300,
  "SyncBatchSize": 100,
  "SyncAPIKey": "${SYNC_API_KEY}",

  "LogLevel": "info",
  "LogFormat": "json",
  "LogOutput": "/var/log/sfsedgestore/app.log",
  "LogMaxSize": 100,
  "LogMaxBackups": 10,
  "LogMaxAge": 30,
  "LogCompress": true,

  "QueueSize": 10000,
  "QueueMaxRetries": 3,
  "QueueRetryInterval": 5,
  "QueueMaxBatchSize": 100,

  "EnableSimulator": false,
  "SimulatorInterval": 1000,
  "SimulatorDevices": 10,
  "SimulatorResources": 5
}
```

#### 配置验证

```bash
# 验证配置文件语法
./sfsedgestore -config config.json -validate

# 测试配置并退出
./sfsedgestore -config config.json -test
```

### 环境变量配置

#### 完整环境变量列表

```bash
# 通用配置
export EDGEX_CONFIG_FILE="/etc/sfsedgestore/config.json"
export EDGEX_LOG_LEVEL="info"
export EDGEX_LOG_FORMAT="json"

# MQTT 配置
export EDGEX_MQTT_BROKER="tcp://mqtt.example.com:1883"
export EDGEX_MQTT_CLIENT_ID="sfsedgestore-01"
export EDGEX_MQTT_USERNAME="sfsedgestore"
export EDGEX_MQTT_PASSWORD="secret"
export EDGEX_MQTT_TOPIC="edgex/events/core/#"
export EDGEX_MQTT_QOS="1"
export EDGEX_MQTT_USE_TLS="false"

# 数据库配置
export EDGEX_DB_PATH="/var/lib/sfsedgestore/data"
export EDGEX_DB_USE_ENCRYPTION="false"
export EDGEX_DB_ENCRYPTION_KEY="your-encryption-key"
export EDGEX_DB_COMPRESSION="true"

# HTTP 配置
export EDGEX_HTTP_PORT="8080"
export EDGEX_HTTP_HOST="0.0.0.0"
export EDGEX_HTTP_USE_TLS="false"

# 认证配置
export EDGEX_ENABLE_AUTH="true"
export EDGEX_AUTH_API_KEY_REQUIRED="true"

# 监控配置
export EDGEX_ENABLE_MONITOR="true"
export EDGEX_MONITOR_INTERVAL="60"
export EDGEX_ENABLE_ALERT="true"

# 数据保留配置
export EDGEX_ENABLE_RETENTION="true"
export EDGEX_RETENTION_DAYS="30"

# 数据同步配置
export EDGEX_ENABLE_SYNC="false"
export EDGEX_SYNC_ENDPOINT="https://cloud.example.com/api/sync"
export EDGEX_SYNC_API_KEY="your-sync-api-key"
```

### 配置热重载

#### 使用信号重载配置

```bash
# 发送 SIGHUP 信号重载配置
kill -HUP $(pidof sfsedgestore)

# 使用 systemd
sudo systemctl reload sfsedgestore
```

#### 配置变更管理

```bash
# 备份当前配置
cp /etc/sfsedgestore/config.json /etc/sfsedgestore/config.json.$(date +%Y%m%d_%H%M%S)

# 验证新配置
./sfsedgestore -config /etc/sfsedgestore/config.json.new -validate

# 应用新配置
mv /etc/sfsedgestore/config.json.new /etc/sfsedgestore/config.json
sudo systemctl reload sfsedgestore

# 检查状态
sudo systemctl status sfsedgestore
sudo journalctl -u sfsedgestore -n 50
```

---

## 用户和权限管理

### API Key 管理

#### 创建管理员 API Key

```bash
# 首次启动时创建（无需认证）
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 8760,
    "description": "System Administrator API Key"
  }'
```

响应示例：

```json
{
  "status": "success",
  "api_key": "sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c",
  "user_id": "admin",
  "role": "admin",
  "expires_at": "2027-03-07T10:00:00Z",
  "created_at": "2026-03-07T10:00:00Z"
}
```

#### 创建普通用户 API Key

```bash
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "X-API-Key: sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "app1",
    "role": "user",
    "expires_in": 720,
    "description": "Application 1 API Key"
  }'
```

#### 创建只读用户 API Key

```bash
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "X-API-Key: sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "monitoring",
    "role": "readonly",
    "expires_in": 0,
    "description": "Monitoring System API Key"
  }'
```

> 注意：`expires_in: 0` 表示永不过期（不推荐用于生产环境）

#### 列出所有 API Key

```bash
curl -H "X-API-Key: sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c" \
  http://localhost:8080/api/auth/list-keys
```

响应示例：

```json
{
  "status": "success",
  "keys": [
    {
      "user_id": "admin",
      "role": "admin",
      "expires_at": "2027-03-07T10:00:00Z",
      "created_at": "2026-03-07T10:00:00Z",
      "last_used_at": "2026-03-07T12:00:00Z",
      "description": "System Administrator API Key"
    },
    {
      "user_id": "app1",
      "role": "user",
      "expires_at": "2026-04-07T10:00:00Z",
      "created_at": "2026-03-07T10:00:00Z",
      "last_used_at": null,
      "description": "Application 1 API Key"
    }
  ]
}
```

#### 撤销 API Key

```bash
curl -X POST http://localhost:8080/api/auth/revoke-key \
  -H "X-API-Key: sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "sk_1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d"
  }'
```

#### 轮换 API Key

```bash
# 1. 创建新的 API Key
curl -X POST http://localhost:8080/api/auth/create-key \
  -H "X-API-Key: sk_8f7d3e2c1b0a9f8e7d6c5b4a3f2e1d0c" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 8760,
    "description": "System Administrator API Key (New)"
  }'

# 2. 更新应用使用新 Key
# ... 更新配置 ...

# 3. 验证新 Key 工作正常
curl -H "X-API-Key: <new-api-key>" http://localhost:8080/health

# 4. 撤销旧 Key
curl -X POST http://localhost:8080/api/auth/revoke-key \
  -H "X-API-Key: <new-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "<old-api-key>"
  }'
```

### 角色和权限

#### 权限矩阵

| 权限 | admin | user | readonly |
|------|-------|------|----------|
| 读取数据 | ✅ | ✅ | ✅ |
| 写入数据 | ✅ | ✅ | ❌ |
| 删除数据 | ✅ | ✅ | ❌ |
| 查询数据 | ✅ | ✅ | ✅ |
| 创建备份 | ✅ | ✅ | ❌ |
| 恢复数据 | ✅ | ❌ | ❌ |
| 管理 API Key | ✅ | ❌ | ❌ |
| 查看监控 | ✅ | ✅ | ✅ |
| 管理告警 | ✅ | ❌ | ❌ |
| 修改配置 | ✅ | ❌ | ❌ |

#### 自定义角色

> 注意：当前版本仅支持预定义角色，自定义角色功能将在未来版本中提供。

### 审计日志

#### 启用审计日志

```json
{
  "EnableAuditLog": true,
  "AuditLogPath": "/var/log/sfsedgestore/audit.log",
  "AuditLogLevel": "info"
}
```

#### 审计日志格式

```json
{
  "timestamp": "2026-03-07T10:00:00Z",
  "event_type": "api_key_create",
  "user_id": "admin",
  "client_ip": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "resource": "/api/auth/create-key",
  "action": "POST",
  "status": "success",
  "details": {
    "target_user_id": "app1",
    "role": "user"
  }
}
```

---

## 监控和告警

### 系统监控

#### 健康检查

```bash
# HTTP 健康检查
curl http://localhost:8080/health

# 详细健康状态
curl http://localhost:8080/health?full=true
```

响应示例：

```json
{
  "status": "healthy",
  "timestamp": "2026-03-07T10:00:00Z",
  "version": "1.0.0",
  "uptime": "86400s",
  "checks": {
    "mqtt": {
      "status": "healthy",
      "connected": true,
      "last_message": "2026-03-07T09:59:59Z"
    },
    "database": {
      "status": "healthy",
      "size_mb": 0.25,
      "read_write_ok": true
    },
    "disk": {
      "status": "healthy",
      "used_percent": 45.2
    },
    "memory": {
      "status": "healthy",
      "used_percent": 20.8
    }
  }
}
```

#### 指标监控

```bash
# 获取所有指标
curl http://localhost:8080/metrics

# Prometheus 格式
curl http://localhost:8080/metrics?format=prometheus
```

指标列表：

| 指标名称 | 类型 | 说明 |
|----------|------|------|
| `system_cpu_usage_percent` | Gauge | CPU 使用率 |
| `system_memory_usage_percent` | Gauge | 内存使用率 |
| `system_disk_usage_percent` | Gauge | 磁盘使用率 |
| `system_goroutines` | Gauge | Go 协程数 |
| `business_total_readings` | Counter | 总读数数量 |
| `business_mqtt_messages_received` | Counter | 接收的 MQTT 消息数 |
| `business_mqtt_messages_processed` | Counter | 处理的 MQTT 消息数 |
| `business_queue_length` | Gauge | 队列长度 |
| `database_size_mb` | Gauge | 数据库大小（MB）|
| `database_read_count` | Counter | 数据库读取次数 |
| `database_write_count` | Counter | 数据库写入次数 |
| `http_requests_total` | Counter | HTTP 请求总数 |
| `http_request_duration_seconds` | Histogram | HTTP 请求耗时 |
| `http_errors_total` | Counter | HTTP 错误总数 |

#### Prometheus 集成

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'sfsedgestore'
    scrape_interval: 15s
    static_configs:
      - targets: ['sfsedgestore:8080']
    metrics_path: '/metrics'
```

### 告警管理

#### 配置告警阈值

```json
{
  "EnableAlert": true,
  "AlertThresholds": {
    "cpu": 80,
    "memory": 85,
    "disk": 90,
    "mqtt_lag": 1000,
    "queue_length": 5000
  },
  "AlertNotifiers": [
    {
      "type": "webhook",
      "url": "https://api.example.com/alerts",
      "secret": "your-webhook-secret",
      "timeout": 10
    },
    {
      "type": "email",
      "smtp_server": "smtp.example.com",
      "smtp_port": 587,
      "smtp_username": "alerts@example.com",
      "smtp_password": "your-smtp-password",
      "from": "alerts@example.com",
      "to": ["admin@example.com"],
      "subject_prefix": "[sfsEdgeStore Alert]"
    },
    {
      "type": "slack",
      "webhook_url": "https://hooks.slack.com/services/xxx/yyy/zzz",
      "channel": "#alerts",
      "username": "sfsEdgeStore Alert Bot"
    },
    {
      "type": "pagerduty",
      "integration_key": "your-pagerduty-integration-key",
      "service_id": "your-service-id"
    }
  ]
}
```

#### 告警级别

| 级别 | 说明 | 响应时间 |
|------|------|----------|
| Critical | 系统不可用 | 立即响应 |
| Warning | 性能降级 | 1 小时内响应 |
| Info | 信息通知 | 24 小时内响应 |

#### 告警示例

```json
{
  "alert_id": "alert-001",
  "level": "critical",
  "title": "Disk Usage Critical",
  "message": "Disk usage has reached 95% on node sfsedgestore-01",
  "timestamp": "2026-03-07T10:00:00Z",
  "node": "sfsedgestore-01",
  "metric": "disk_usage_percent",
  "value": 95.0,
  "threshold": 90.0,
  "status": "firing",
  "duration": "5m"
}
```

#### 查看告警

```bash
# 获取所有活动告警
curl -H "X-API-Key: <api-key>" http://localhost:8080/alerts

# 获取告警历史
curl -H "X-API-Key: <api-key>" "http://localhost:8080/alerts/history?limit=100"

# 确认告警
curl -X POST -H "X-API-Key: <api-key>" \
  http://localhost:8080/alerts/alert-001/acknowledge

# 解决告警
curl -X POST -H "X-API-Key: <api-key>" \
  http://localhost:8080/alerts/alert-001/resolve
```

### 可视化监控

#### Grafana 仪表板

```json
{
  "dashboard": {
    "title": "sfsEdgeStore Monitoring",
    "panels": [
      {
        "title": "System Overview",
        "type": "row",
        "panels": [
          {
            "title": "CPU Usage",
            "type": "stat",
            "targets": [
              {
                "expr": "system_cpu_usage_percent",
                "legendFormat": "CPU"
              }
            ]
          },
          {
            "title": "Memory Usage",
            "type": "stat",
            "targets": [
              {
                "expr": "system_memory_usage_percent",
                "legendFormat": "Memory"
              }
            ]
          },
          {
            "title": "Disk Usage",
            "type": "stat",
            "targets": [
              {
                "expr": "system_disk_usage_percent",
                "legendFormat": "Disk"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

---

## 备份和恢复

### 备份策略

#### 备份类型

| 类型 | 说明 | 频率 | 保留期 |
|------|------|------|--------|
| 完整备份 | 完整数据库备份 | 每日 | 30 天 |
| 增量备份 | 变更数据备份 | 每小时 | 7 天 |
| 快照备份 | 时间点快照 | 每周 | 90 天 |

#### 3-2-1 备份原则

- **3 份数据副本**
- **2 种不同介质**
- **1 份异地备份**

### 手动备份

#### 使用 API 备份

```bash
# 创建备份
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=/var/lib/sfsedgestore/backups"

# 指定备份文件名
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=/var/lib/sfsedgestore/backups&name=backup-$(date +%Y%m%d_%H%M%S)"

# 压缩备份
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup?path=/var/lib/sfsedgestore/backups&compress=true"
```

#### 文件系统备份

```bash
# 停止服务
sudo systemctl stop sfsedgestore

# 创建备份
sudo tar -czf \
  /var/lib/sfsedgestore/backups/sfsedgestore-$(date +%Y%m%d_%H%M%S).tar.gz \
  -C /var/lib/sfsedgestore data \
  -C /etc/sfsedgestore config.json

# 启动服务
sudo systemctl start sfsedgestore

# 验证备份
tar -tzf /var/lib/sfsedgestore/backups/sfsedgestore-*.tar.gz
```

#### 数据库直接备份

```bash
# 使用 sfsDb 工具备份
./sfsdb backup \
  --path /var/lib/sfsedgestore/data \
  --output /var/lib/sfsedgestore/backups/sfsdb-backup-$(date +%Y%m%d_%H%M%S).db
```

### 自动备份

#### Cron 定时备份

```bash
# 编辑 crontab
sudo crontab -e -u sfsedgestore

# 添加备份任务
# 每小时增量备份
0 * * * * /opt/sfsedgestore/scripts/backup.sh incremental >> /var/log/sfsedgestore/backup.log 2>&1

# 每日完整备份
0 2 * * * /opt/sfsedgestore/scripts/backup.sh full >> /var/log/sfsedgestore/backup.log 2>&1

# 每周快照备份
0 3 * * 0 /opt/sfsedgestore/scripts/backup.sh snapshot >> /var/log/sfsedgestore/backup.log 2>&1

# 清理旧备份
0 4 * * * /opt/sfsedgestore/scripts/cleanup-backups.sh >> /var/log/sfsedgestore/backup.log 2>&1
```

#### 备份脚本示例

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/backup.sh

set -e

BACKUP_TYPE=${1:-full}
BACKUP_DIR="/var/lib/sfsedgestore/backups"
DATA_DIR="/var/lib/sfsedgestore/data"
CONFIG_DIR="/etc/sfsedgestore"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

backup_full() {
    log "Starting full backup..."
    
    BACKUP_FILE="$BACKUP_DIR/sfsedgestore-full-$DATE.tar.gz"
    
    # 使用 API 热备份
    curl -X POST \
      -H "X-API-Key: ${BACKUP_API_KEY}" \
      "http://localhost:8080/api/backup?path=$BACKUP_DIR&name=sfsedgestore-full-$DATE&compress=true"
    
    log "Full backup completed: $BACKUP_FILE"
}

backup_incremental() {
    log "Starting incremental backup..."
    
    # 查找上次完整备份
    LAST_FULL=$(find "$BACKUP_DIR" -name "sfsedgestore-full-*.tar.gz" -type f -mtime -1 | sort -r | head -n 1)
    
    if [ -z "$LAST_FULL" ]; then
        log "No recent full backup found, doing full backup instead"
        backup_full
        return
    fi
    
    # 增量备份逻辑
    # ...
    
    log "Incremental backup completed"
}

cleanup_old() {
    log "Cleaning up old backups..."
    
    find "$BACKUP_DIR" -name "sfsedgestore-full-*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete
    find "$BACKUP_DIR" -name "sfsedgestore-incr-*.tar.gz" -type f -mtime +7 -delete
    
    log "Old backups cleaned up"
}

case "$BACKUP_TYPE" in
    full)
        backup_full
        ;;
    incremental)
        backup_incremental
        ;;
    snapshot)
        # LVM 快照或文件系统快照
        ;;
    cleanup)
        cleanup_old
        ;;
    *)
        echo "Usage: $0 {full|incremental|snapshot|cleanup}"
        exit 1
        ;;
esac

log "Backup operation completed successfully"
```

#### Systemd 定时备份

```ini
# /etc/systemd/system/sfsedgestore-backup.service
[Unit]
Description=sfsEdgeStore Backup Service
After=network.target

[Service]
Type=oneshot
User=sfsedgestore
Group=sfsedgestore
ExecStart=/opt/sfsedgestore/scripts/backup.sh full
Environment="BACKUP_API_KEY=your-api-key-here"

# /etc/systemd/system/sfsedgestore-backup.timer
[Unit]
Description=sfsEdgeStore Backup Timer

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

启用定时备份：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now sfsedgestore-backup.timer
sudo systemctl list-timers sfsedgestore-backup.timer
```

### 异地备份

#### Rsync 同步

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/rsync-backup.sh

REMOTE_SERVER="backup-server.example.com"
REMOTE_PATH="/backups/sfsedgestore"
LOCAL_PATH="/var/lib/sfsedgestore/backups"

rsync -avz --delete \
  --exclude='*.tmp' \
  --link-dest="$REMOTE_PATH/latest" \
  "$LOCAL_PATH/" \
  "$REMOTE_SERVER:$REMOTE_PATH/$(date +%Y%m%d_%H%M%S)/"

# 更新 latest 链接
ssh "$REMOTE_SERVER" "rm -f $REMOTE_PATH/latest && ln -s $REMOTE_PATH/$(date +%Y%m%d_%H%M%S) $REMOTE_PATH/latest"
```

#### 云存储备份

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/cloud-backup.sh

# 使用 AWS CLI
aws s3 sync \
  /var/lib/sfsedgestore/backups/ \
  s3://your-bucket/sfsedgestore/backups/ \
  --storage-class STANDARD_IA \
  --delete

# 或使用 Azure CLI
az storage blob sync \
  --source /var/lib/sfsedgestore/backups \
  --container sfsedgestore-backups \
  --account-name your-storage-account

# 或使用 Google Cloud SDK
gsutil -m rsync -r -d \
  /var/lib/sfsedgestore/backups/ \
  gs://your-bucket/sfsedgestore/backups/
```

### 数据恢复

#### 恢复前检查清单

- [ ] 确认备份文件完整性
- [ ] 停止相关服务
- [ ] 备份当前状态
- [ ] 通知相关人员
- [ ] 准备回滚方案
- [ ] 验证恢复环境

#### 使用 API 恢复

```bash
# 列出可用备份
curl -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/backup/list?path=/var/lib/sfsedgestore/backups"

# 恢复备份
curl -X POST \
  -H "X-API-Key: <api-key>" \
  "http://localhost:8080/api/restore?file=/var/lib/sfsedgestore/backups/backup-20260307_020000.db"
```

#### 文件系统恢复

```bash
# 1. 停止服务
sudo systemctl stop sfsedgestore

# 2. 备份当前数据（以防万一）
sudo mv /var/lib/sfsedgestore/data /var/lib/sfsedgestore/data.bak.$(date +%Y%m%d_%H%M%S)

# 3. 解压备份
sudo tar -xzf \
  /var/lib/sfsedgestore/backups/sfsedgestore-full-20260307_020000.tar.gz \
  -C /var/lib/sfsedgestore

# 4. 验证权限
sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore/data

# 5. 启动服务
sudo systemctl start sfsedgestore

# 6. 验证恢复
sudo systemctl status sfsedgestore
curl http://localhost:8080/health
```

#### 数据库直接恢复

```bash
# 停止服务
sudo systemctl stop sfsedgestore

# 使用 sfsDb 工具恢复
./sfsdb restore \
  --input /var/lib/sfsedgestore/backups/sfsdb-backup-20260307_020000.db \
  --path /var/lib/sfsedgestore/data

# 启动服务
sudo systemctl start sfsedgestore
```

#### 时间点恢复

> 注意：时间点恢复需要开启数据库 WAL 日志和定期快照。

```bash
# 1. 恢复最近的完整备份
./sfsdb restore --input backup-full.db --path /var/lib/sfsedgestore/data

# 2. 应用 WAL 日志到指定时间点
./sfsdb recover \
  --path /var/lib/sfsedgestore/data \
  --target-time "2026-03-07T10:00:00Z"
```

### 备份验证

#### 自动验证脚本

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/verify-backup.sh

BACKUP_FILE="$1"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file>"
    exit 1
fi

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# 1. 检查文件存在
if [ ! -f "$BACKUP_FILE" ]; then
    log "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

# 2. 检查文件大小
FILE_SIZE=$(stat -f%z "$BACKUP_FILE" 2>/dev/null || stat -c%s "$BACKUP_FILE")
if [ "$FILE_SIZE" -lt 1024 ]; then
    log "ERROR: Backup file too small: $FILE_SIZE bytes"
    exit 1
fi

# 3. 验证文件完整性（如果有 checksum）
if [ -f "$BACKUP_FILE.sha256" ]; then
    cd "$(dirname "$BACKUP_FILE")"
    if ! sha256sum -c "$(basename "$BACKUP_FILE").sha256"; then
        log "ERROR: Checksum verification failed"
        exit 1
    fi
    log "Checksum verification passed"
fi

# 4. 测试恢复（使用临时目录）
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

log "Testing backup restoration in $TEST_DIR..."

# 解压到临时目录
if tar -tzf "$BACKUP_FILE" >/dev/null 2>&1; then
    tar -xzf "$BACKUP_FILE" -C "$TEST_DIR"
else
    log "ERROR: Invalid backup file format"
    exit 1
fi

# 检查关键文件
if [ ! -d "$TEST_DIR/data" ]; then
    log "ERROR: Data directory not found in backup"
    exit 1
fi

log "Backup verification passed: $BACKUP_FILE"
```

---

## 升级和维护

### 版本管理

#### 版本号规则

sfsEdgeStore 使用语义化版本（Semantic Versioning）：

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向下兼容的功能新增
- **PATCH**: 向下兼容的问题修复

#### 查看版本

```bash
# 查看程序版本
./sfsedgestore -version

# 通过 API 查看
curl http://localhost:8080/health | jq '.version'
```

### 升级前准备

#### 升级检查清单

- [ ] 阅读版本发布说明（Release Notes）
- [ ] 检查兼容性矩阵
- [ ] 备份当前配置和数据
- [ ] 在测试环境验证升级
- [ ] 制定回滚计划
- [ ] 通知相关人员
- [ ] 选择维护窗口
- [ ] 准备升级脚本

#### 兼容性检查

```bash
# 检查当前版本
./sfsedgestore -version

# 检查目标版本兼容性
curl -s https://api.github.com/repos/your-org/sfsEdgeStore/releases/latest \
  | jq -r '.body' | grep -A 10 "Breaking Changes"
```

#### 备份当前状态

```bash
# 1. 备份配置
sudo cp -a /etc/sfsedgestore /etc/sfsedgestore.bak.$(date +%Y%m%d_%H%M%S)

# 2. 备份数据
sudo /opt/sfsedgestore/scripts/backup.sh full

# 3. 备份二进制
sudo cp /opt/sfsedgestore/sfsedgestore /opt/sfsedgestore/sfsedgestore.bak.$(date +%Y%m%d_%H%M%S)
```

### 升级流程

#### 零停机升级（蓝绿部署）

```bash
# 1. 部署新版本（绿色环境）
docker run -d \
  --name sfsedgestore-green \
  -p 8081:8080 \
  -v sfsedgestore-data:/app/data \
  -v /etc/sfsedgestore/config.json:/app/config.json:ro \
  your-org/sfsedgestore:new-version

# 2. 验证新版本
curl http://localhost:8081/health

# 3. 切换流量（更新负载均衡配置）
# ... 更新 haproxy/nginx 配置 ...

# 4. 验证流量切换成功
# ... 监控日志 ...

# 5. 停止旧版本
docker stop sfsedgestore-blue
docker rm sfsedgestore-blue
```

#### 滚动升级（Kubernetes）

```bash
# 更新镜像
kubectl set image deployment/sfsedgestore \
  sfsedgestore=your-org/sfsedgestore:new-version \
  -n sfsedgestore

# 监控升级进度
kubectl rollout status deployment/sfsedgestore -n sfsedgestore

# 查看历史
kubectl rollout history deployment/sfsedgestore -n sfsedgestore

# 如需要，回滚
kubectl rollout undo deployment/sfsedgestore -n sfsedgestore
```

#### 标准升级流程

```bash
# 1. 下载新版本
cd /tmp
wget https://github.com/your-org/sfsEdgeStore/releases/download/v1.1.0/sfsedgestore-linux-amd64
chmod +x sfsedgestore-linux-amd64

# 2. 验证新版本（可选）
./sfsedgestore-linux-amd64 -version
./sfsedgestore-linux-amd64 -config /etc/sfsedgestore/config.json -validate

# 3. 停止服务
sudo systemctl stop sfsedgestore

# 4. 备份旧版本
sudo cp /opt/sfsedgestore/sfsedgestore /opt/sfsedgestore/sfsedgestore.old

# 5. 安装新版本
sudo mv sfsedgestore-linux-amd64 /opt/sfsedgestore/sfsedgestore
sudo chown sfsedgestore:sfsedgestore /opt/sfsedgestore/sfsedgestore
sudo chmod +x /opt/sfsedgestore/sfsedgestore

# 6. 检查配置兼容性
# 如有配置变更，更新配置文件

# 7. 启动服务
sudo systemctl start sfsedgestore

# 8. 验证升级
sudo systemctl status sfsedgestore
sudo journalctl -u sfsedgestore -n 100 --no-pager
curl http://localhost:8080/health

# 9. 清理旧版本（确认一切正常后）
sudo rm /opt/sfsedgestore/sfsedgestore.old
```

### 回滚流程

#### 快速回滚

```bash
# 1. 停止新版本
sudo systemctl stop sfsedgestore

# 2. 恢复旧版本
sudo cp /opt/sfsedgestore/sfsedgestore.old /opt/sfsedgestore/sfsedgestore

# 3. 恢复旧配置（如需要）
sudo cp -a /etc/sfsedgestore.bak /etc/sfsedgestore

# 4. 恢复旧数据（如需要）
# 参考备份恢复章节

# 5. 启动旧版本
sudo systemctl start sfsedgestore

# 6. 验证回滚
sudo systemctl status sfsedgestore
curl http://localhost:8080/health
```

### 数据库迁移

#### 自动迁移

sfsEdgeStore 会在启动时自动检测并执行数据库迁移：

```bash
# 启动新版本，自动执行迁移
sudo systemctl start sfsedgestore

# 查看迁移日志
sudo journalctl -u sfsedgestore -n 200 | grep -i migration
```

#### 手动迁移

```bash
# 1. 停止服务
sudo systemctl stop sfsedgestore

# 2. 备份数据库
sudo cp -a /var/lib/sfsedgestore/data /var/lib/sfsedgestore/data.bak.migration

# 3. 执行迁移工具
./sfsedgestore -migrate -config /etc/sfsedgestore/config.json

# 4. 验证迁移结果
# 检查日志

# 5. 启动服务
sudo systemctl start sfsedgestore
```

### 定期维护

#### 每日维护任务

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/daily-maintenance.sh

# 检查服务状态
if ! systemctl is-active --quiet sfsedgestore; then
    echo "WARNING: sfsEdgeStore is not running"
    systemctl start sfsedgestore
fi

# 检查磁盘空间
DISK_USAGE=$(df -P /var/lib/sfsedgestore | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "WARNING: Disk usage is ${DISK_USAGE}%"
fi

# 清理旧日志
find /var/log/sfsedgestore -name "*.log.*" -mtime +7 -delete

# 检查告警
# ...
```

#### 每周维护任务

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/weekly-maintenance.sh

# 数据完整性检查
./sfsedgestore -checkdb -config /etc/sfsedgestore/config.json

# 优化数据库
./sfsedgestore -optimize -config /etc/sfsedgestore/config.json

# 验证备份
for backup in /var/lib/sfsedgestore/backups/*.tar.gz; do
    /opt/sfsedgestore/scripts/verify-backup.sh "$backup"
done

# 检查安全更新
apt list --upgradable 2>/dev/null | grep -E "(sfsedgestore|golang|openssl)" || true
```

#### 月度维护任务

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/monthly-maintenance.sh

# 性能测试
./sfsedgestore -benchmark -config /etc/sfsedgestore/config.json

# 安全审计
# 检查日志中的异常
grep -i "error\|warning" /var/log/sfsedgestore/app.log | tail -100

# 审查 API Key 使用情况
curl -H "X-API-Key: <admin-key>" http://localhost:8080/api/auth/list-keys

# 更新文档
# ...
```

---

## 安全管理

### 安全基线

#### CIS 基准检查清单

- [ ] 使用专用用户运行
- [ ] 配置文件权限正确（600）
- [ ] 启用认证
- [ ] 使用 HTTPS
- [ ] 使用 MQTT TLS
- [ ] 数据库加密启用
- [ ] 定期轮换 API Key
- [ ] 日志集中化
- [ ] 定期备份
- [ ] 系统更新及时

### 认证和授权

#### API Key 安全最佳实践

1. **使用强密钥**
   - 长度至少 32 字符
   - 包含大小写字母、数字、符号
   - 不要使用可预测的模式

2. **密钥管理**
   ```bash
   # 使用密码管理器存储
   # 定期轮换（建议 90 天）
   # 立即撤销泄露的密钥
   ```

3. **最小权限原则**
   - 为每个应用使用独立的 Key
   - 只授予必要的权限
   - 使用只读 Key 用于监控

#### 强制认证配置

```json
{
  "EnableAuth": true,
  "AuthAPIKeyRequired": true,
  "AuthRateLimit": 1000,
  "AuthRateLimitWindow": 60,
  "AuthIPWhitelist": [
    "192.168.1.0/24",
    "10.0.0.0/8"
  ]
}
```

### 网络安全

#### TLS 配置

##### 生成证书

```bash
# 生成 CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=sfsEdgeStore CA"

# 生成服务器证书
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=sfsedgestore.example.com"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

# 生成客户端证书
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=sfsedgestore-client"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365
```

##### 配置 HTTPS

```json
{
  "HTTPUseTLS": true,
  "HTTPCert": "/etc/sfsedgestore/certs/server.crt",
  "HTTPKey": "/etc/sfsedgestore/certs/server.key",
  "HTTPClientCA": "/etc/sfsedgestore/certs/ca.crt",
  "HTTPRequireClientCert":