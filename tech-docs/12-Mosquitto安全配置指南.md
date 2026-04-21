# Mosquitto MQTT Broker 安全配置指南（客户侧）

## 1. 概述

本文档是 **面向 MQTT Broker 管理员（客户）** 的配置指南，介绍如何在 Mosquitto Broker 端配置 TLS/SSL 加密、用户名/密码认证和 ACL 权限控制。

> **重要说明**：这些安全配置是在 **MQTT Broker 端** 完成的，是客户的基础设施工作。
> sfsEdgeStore 作为 MQTT 客户端，只需配置相应的连接参数即可连接到启用安全特性的 Broker。

## 角色划分

| 配置项 | 负责方 | 说明 |
|--------|--------|------|
| Broker TLS/SSL | 客户（IT/运维） | 在 Broker 端配置证书 |
| Broker 用户认证 | 客户（IT/运维） | 在 Broker 端创建用户 |
| Broker ACL | 客户（IT/运维） | 在 Broker 端设置权限 |
| sfsEdgeStore TLS 配置 | sfsEdgeStore | 配置连接 Broker 的参数 |
| sfsEdgeStore 认证信息 | 客户 | 提供用户名/密码给 sfsEdgeStore |

## 2. 目录结构

```
mosquitto_security/
├── certs/
│   ├── ca.crt              # CA 证书
│   ├── server.crt          # 服务器证书
│   ├── server.key          # 服务器私钥
│   ├── client.crt          # 客户端证书
│   └── client.key          # 客户端私钥
├── config/
│   ├── mosquitto.conf      # 主配置文件
│   ├── pwfile              # 密码文件
│   └── aclfile             # ACL 规则文件
└── logs/
    └── mosquitto.log       # 日志文件
```

## 3. 生成 TLS 证书

### 3.1 创建 CA 证书

```bash
# 创建证书目录
mkdir -p certs
cd certs

# 生成 CA 私钥
openssl genrsa -out ca.key 2048

# 生成 CA 证书（有效期 10 年）
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 \
  -out ca.crt \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=sfsEdgeStore/OU=Security/CN=sfsEdgeStore CA"
```

### 3.2 生成服务器证书

```bash
# 生成服务器私钥
openssl genrsa -out server.key 2048

# 创建服务器证书请求配置
cat > server.cnf << EOF
[req]
distinguished_name = req_distinguished_name
prompt = no
default_md = sha256
x509_extensions = v3_req

[req_distinguished_name]
C = CN
ST = Beijing
L = Beijing
O = sfsEdgeStore
OU = IoT
CN = localhost

[v3_req]
keyUsage = keyEncipherment, digitalSignature
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.sfsedgestore.local
IP.1 = 127.0.0.1
IP.2 = 192.168.1.100
EOF

# 生成服务器证书请求
openssl req -new -key server.key -out server.csr -config server.cnf

# 使用 CA 签发服务器证书（有效期 1 年）
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 365 \
  -sha256 -extfile server.cnf -extensions v3_req
```

### 3.3 生成客户端证书

```bash
# 生成客户端私钥
openssl genrsa -out client.key 2048

# 创建客户端证书请求配置
cat > client.cnf << EOF
[req]
distinguished_name = req_distinguished_name
prompt = no
default_md = sha256
x509_extensions = v3_req

[req_distinguished_name]
C = CN
ST = Beijing
L = Beijing
O = sfsEdgeStore
OU = Device
CN = sfsEdgeStore-device

[v3_req]
keyUsage = digitalSignature
extendedKeyUsage = clientAuth
EOF

# 生成客户端证书请求
openssl req -new -key client.key -out client.csr -config client.cnf

# 使用 CA 签发客户端证书（有效期 1 年）
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 365 \
  -sha256 -extfile client.cnf -extensions v3_req
```

### 3.4 转换证书格式（Java/Go 使用）

```bash
# 转换为 PKCS12 格式（可选）
openssl pkcs12 -export -in client.crt -inkey client.key \
  -certfile ca.crt -out client.p12

# 验证证书
openssl verify -CAfile ca.crt server.crt
openssl verify -CAfile ca.crt client.crt
```

## 4. 配置 Mosquitto（Broker 端）

> **以下配置由客户的 IT/运维团队完成**

### 4.1 主配置文件

创建 `mosquitto.conf`：

```properties
# ============================================
# 基础配置
# ============================================

# 监听端口（TLS 连接）
listener 8883

# 允许匿名访问（改为 false 启用认证）
allow_anonymous false

# 密码文件
password_file /path/to/config/pwfile

# ACL 文件
acl_file /path/to/config/aclfile

# ============================================
# TLS/SSL 配置
# ============================================

# TLS 版本（推荐 TLSv1.2 或更高）
tls_version tlsv1.2

# CA 证书
cafile /path/to/certs/ca.crt

# 服务器证书
certfile /path/to/certs/server.crt

# 服务器私钥
keyfile /path/to/certs/server.key

# 要求客户端证书
require_certificate true

# 验证客户端证书
use_identity_as_username true

# ============================================
# 日志配置
# ============================================

log_dest file /path/to/logs/mosquitto.log
log_dest stdout
log_type error
log_type warning
log_type notice
log_type information
log_timestamp true
log_timestamp_format %Y-%m-%d %H:%M:%S

# ============================================
# 性能配置
# ============================================

# 最大连接数
max_connections -1

# 消息持久化
persistence true
persistence_location /var/lib/mosquitto/

# 消息超时
message_timeout 300
```

### 4.2 密码文件配置

```bash
# 创建密码文件（每行格式：用户名:密码哈希）
# Mosquitto 1.6+ 使用 argon2 加密

# 交互式创建用户
mosquitto_passwd -c /path/to/config/pwfile admin

# 创建只读用户（用于 sfsEdgeStore）
mosquitto_passwd -c /path/to/config/pwfile sfsreader

# 创建读写用户（用于设备）
mosquitto_passwd /path/to/config/pwfile device

# 创建管理用户
mosquitto_passwd /path/to/config/pwfile manager

# 删除用户
mosquitto_passwd -D /path/to/config/pwfile username
```

### 4.3 ACL 规则配置

创建 `aclfile`：

```properties
# ============================================
# ACL 规则
# ============================================

# 用户 admin 具有所有权限
user admin
topic readwrite #

# 用户 manager 具有管理权限
user manager
topic readwrite device/+/config
topic readwrite device/+/command
topic readwrite edgex/#
topic readwrite sfsEdgeStore/#

# 用户 sfsreader 只有读取权限
user sfsreader
topic read edgex/events/#
topic read sfsEdgeStore/data/#

# 用户 device 只能操作自己的设备
user device
# 设备上传数据
topic write device/{device_id}/data
topic read device/{device_id}/config
topic read device/{device_id}/command
# 设备状态
topic readwrite device/{device_id}/status

# 通配规则
# 模式匹配 - 用户只能访问自己的设备
pattern readwrite device/%u/#
pattern readwrite device/%u/data
pattern readwrite device/%u/status
```

## 5. sfsEdgeStore 端配置（客户端）

> **以下配置由 sfsEdgeStore 用户完成**（客户从 Broker 管理员获取证书和凭据后配置）

### 5.1 config.json 配置

```json
{
  "mqtt": {
    "broker": "mqtts://localhost:8883",
    "client_id": "sfsedgedgestore-device",
    "clean_session": true,
    "auto_reconnect": true,
    "resume_subs": true,
    "tls": {
      "enabled": true,
      "ca_cert": "./certs/ca.crt",
      "client_cert": "./certs/client.crt",
      "client_key": "./certs/client.key",
      "insecure_skip_verify": false
    },
    "auth": {
      "username": "sfsreader",
      "password": "your_password"
    }
  }
}
```

### 5.2 Go 代码中的 TLS 配置

```go
// mqtt/client.go
import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "github.com/eclipse/paho.mqtt.golang"
)

func NewSecureClient(config *Config) (*Client, error) {
    opts := mqtt.NewClientOptions()

    // TLS 配置
    tlsConfig := &tls.Config{
        MinVersion:         tls.VersionTLS12,
        InsecureSkipVerify: false,
    }

    // 加载 CA 证书
    caCert, err := ioutil.ReadFile(config.MQTT.TLS.CACert)
    if err != nil {
        return nil, fmt.Errorf("加载 CA 证书失败: %v", err)
    }
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    tlsConfig.RootCAs = caCertPool

    // 加载客户端证书（双向认证）
    if config.MQTT.TLS.ClientCert != "" && config.MQTT.TLS.ClientKey != "" {
        clientCert, err := tls.LoadX509KeyPair(
            config.MQTT.TLS.ClientCert,
            config.MQTT.TLS.ClientKey,
        )
        if err != nil {
            return nil, fmt.Errorf("加载客户端证书失败: %v", err)
        }
        tlsConfig.Certificates = []tls.Certificate{clientCert}
    }

    opts.SetTLSConfig(tlsConfig)

    // 设置认证
    opts.SetUsername(config.MQTT.Auth.Username)
    opts.SetPassword(config.MQTT.Auth.Password)

    // 连接选项
    opts.SetAutoReconnect(config.MQTT.AutoReconnect)
    opts.SetCleanSession(config.MQTT.CleanSession)
    opts.SetResumeSubs(config.MQTT.ResumeSubs)
    opts.SetConnectTimeout(30 * time.Second)

    return &Client{config: config, opts: opts}, nil
}
```

## 6. 启动 Mosquitto

### 6.1 开发/测试环境

```bash
# 使用配置文件启动
mosquitto -c /path/to/config/mosquitto.conf -v

# 或者不使用配置文件，仅启用 TLS（测试用）
mosquitto -p 8883 \
  --cafile ./certs/ca.crt \
  --certfile ./certs/server.crt \
  --keyfile ./certs/server.key \
  --require_certificate false \
  --allow_anonymous false \
  -u test -P test123 \
  -v
```

### 6.2 生产环境（systemd 服务）

创建 systemd 服务文件 `/etc/systemd/system/mosquitto.service`：

```ini
[Unit]
Description=Mosquitto MQTT Broker
After=network.target

[Service]
Type=simple
ExecStart=/usr/sbin/mosquitto -c /path/to/config/mosquitto.conf
Restart=always
RestartSec=10
User=mosquitto
Group=mosquitto

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
# 重新加载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start mosquitto

# 设置开机自启
sudo systemctl enable mosquitto

# 查看状态
sudo systemctl status mosquitto

# 查看日志
sudo journalctl -u mosquitto -f
```

## 7. 测试验证

### 7.1 测试 TLS 连接

```bash
# 测试订阅（使用 TLS）
mosquitto_sub -h localhost -p 8883 \
  --cafile ./certs/ca.crt \
  --cert ./certs/client.crt \
  --key ./certs/client.key \
  -t "edgex/events/#" -v

# 测试发布（使用 TLS）
mosquitto_pub -h localhost -p 8883 \
  --cafile ./certs/ca.crt \
  --cert ./certs/client.crt \
  --key ./certs/client.key \
  -t "test/topic" -m "Hello TLS"
```

### 7.2 测试认证

```bash
# 测试错误密码（应该失败）
mosquitto_pub -h localhost -p 8883 \
  --cafile ./certs/ca.crt \
  -u wronguser -P wrongpass \
  -t "test/topic" -m "Test"

# 测试正确密码（应该成功）
mosquitto_pub -h localhost -p 8883 \
  --cafile ./certs/ca.crt \
  -u sfsreader -P correctpass \
  -t "test/topic" -m "Test"
```

### 7.3 测试 ACL

```bash
# 以 device 用户身份尝试读取其他设备数据（应该失败）
mosquitto_pub -h localhost -p 8883 \
  --cafile ./certs/ca.crt \
  -u device -P devicepass \
  -t "device/other_device/data" -m "Unauthorized"

# 以 device 用户身份读取自己的数据（应该成功）
mosquitto_pub -h localhost -p 8883 \
  --cafile ./certs/ca.crt \
  -u device -P devicepass \
  -t "device/device/data" -m "Authorized"
```

## 8. 常见问题

### 8.1 证书问题

```bash
# 检查证书有效性
openssl x509 -in server.crt -text -noout

# 检查证书链
openssl verify -CAfile ca.crt server.crt

# 检查私钥匹配
openssl x509 -in server.crt -noout -modulus | md5sum
openssl rsa -in server.key -noout -modulus | md5sum
```

### 8.2 连接问题

```bash
# 检查端口监听
netstat -tlnp | grep 8883

# 检查防火墙
sudo iptables -L -n

# 查看 Mosquitto 日志
tail -f /path/to/logs/mosquitto.log
```

### 8.3 性能问题

```bash
# 检查连接数
mosquitto_sub -h localhost -p 8883 -t "$SYS/broker/clients/connected"

# 检查消息数
mosquitto_sub -h localhost -p 8883 -t "$SYS/broker/messages/received"

# 检查内存使用
ps aux | grep mosquitto
```

## 9. 安全检查清单

- [ ] TLS/SSL 已启用
- [ ] 使用 TLSv1.2 或更高版本
- [ ] 双向证书认证已启用
- [ ] 匿名访问已禁用
- [ ] 强密码策略已启用
- [ ] ACL 规则已配置
- [ ] 日志记录已启用
- [ ] 防火墙已配置
- [ ] 证书已过期检查
- [ ] 定期轮换证书和密码

