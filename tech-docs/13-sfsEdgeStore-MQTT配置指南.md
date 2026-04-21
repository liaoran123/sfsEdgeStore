# sfsEdgeStore MQTT 配置指南

## 1. 概述

本文档详细介绍 sfsEdgeStore 的 MQTT 配置选项，包括连接参数、TLS/SSL 加密、用户名/密码认证等，帮助用户在各种部署场景下正确配置 MQTT 客户端。

## 2. 配置方式

sfsEdgeStore 支持两种配置方式：

### 2.1 配置文件方式（config.json）

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsedgedgestore-device",
  "mqtt_use_tls": false,
  "mqtt_ca_cert": "",
  "mqtt_client_cert": "",
  "mqtt_client_key": "",
  "mqtt_username": "",
  "mqtt_password": ""
}
```

### 2.2 环境变量方式

```bash
export EDGEX_MQTT_BROKER=tcp://localhost:1883
export EDGEX_MQTT_TOPIC=edgex/events/core/#
export EDGEX_CLIENT_ID=sfsedgedgestore-device
export EDGEX_MQTT_USE_TLS=false
export EDGEX_MQTT_CA_CERT=/path/to/ca.crt
export EDGEX_MQTT_CLIENT_CERT=/path/to/client.crt
export EDGEX_MQTT_CLIENT_KEY=/path/to/client.key
export EDGEX_MQTT_USERNAME=your_username
export EDGEX_MQTT_PASSWORD=your_password
```

## 3. 完整配置参数

### 3.1 基础连接参数

| 参数 | JSON 字段 | 环境变量 | 默认值 | 说明 |
|------|-----------|----------|--------|------|
| Broker 地址 | `mqtt_broker` | `EDGEX_MQTT_BROKER` | `tcp://localhost:1883` | MQTT Broker 连接地址 |
| 订阅主题 | `mqtt_topic` | `EDGEX_MQTT_TOPIC` | `edgex/events/core/#` | 要订阅的 MQTT 主题 |
| 客户端 ID | `client_id` | `EDGEX_CLIENT_ID` | 自动生成 | MQTT 客户端标识符 |

### 3.2 TLS/SSL 参数

| 参数 | JSON 字段 | 环境变量 | 默认值 | 说明 |
|------|-----------|----------|--------|------|
| 启用 TLS | `mqtt_use_tls` | `EDGEX_MQTT_USE_TLS` | `false` | 是否启用 TLS 加密 |
| CA 证书 | `mqtt_ca_cert` | `EDGEX_MQTT_CA_CERT` | `` | CA 证书文件路径 |
| 客户端证书 | `mqtt_client_cert` | `EDGEX_MQTT_CLIENT_CERT` | `` | 客户端证书文件路径 |
| 客户端私钥 | `mqtt_client_key` | `EDGEX_MQTT_CLIENT_KEY` | `` | 客户端私钥文件路径 |

### 3.3 认证参数

| 参数 | JSON 字段 | 环境变量 | 默认值 | 说明 |
|------|-----------|----------|--------|------|
| 用户名 | `mqtt_username` | `EDGEX_MQTT_USERNAME` | `` | MQTT 认证用户名 |
| 密码 | `mqtt_password` | `EDGEX_MQTT_PASSWORD` | `` | MQTT 认证密码 |

## 4. 配置场景

### 4.1 本地开发环境（无加密）

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsedgedgestore-dev",
  "mqtt_use_tls": false
}
```

### 4.2 生产环境（TLS + 认证）

```json
{
  "mqtt_broker": "mqtts://mqtt.example.com:8883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsedgedgestore-prod",
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "/etc/sfsEdgeStore/certs/ca.crt",
  "mqtt_client_cert": "/etc/sfsEdgeStore/certs/client.crt",
  "mqtt_client_key": "/etc/sfsEdgeStore/certs/client.key",
  "mqtt_username": "sfsedgestore",
  "mqtt_password": "your_secure_password"
}
```

### 4.3 使用环境变量配置

```bash
# 基础配置
export EDGEX_MQTT_BROKER=mqtts://mqtt.example.com:8883
export EDGEX_MQTT_TOPIC=edgex/events/core/#
export EDGEX_CLIENT_ID=sfsedgedgestore-prod

# TLS 配置
export EDGEX_MQTT_USE_TLS=true
export EDGEX_MQTT_CA_CERT=/etc/sfsEdgeStore/certs/ca.crt
export EDGEX_MQTT_CLIENT_CERT=/etc/sfsEdgeStore/certs/client.crt
export EDGEX_MQTT_CLIENT_KEY=/etc/sfsEdgeStore/certs/client.key

# 认证配置
export EDGEX_MQTT_USERNAME=sfsedgestore
export EDGEX_MQTT_PASSWORD=your_secure_password
```

## 5. TLS 配置详解

### 5.1 单向 TLS（仅验证服务器）

适用于客户端需要验证服务器身份的场景。

```json
{
  "mqtt_broker": "mqtts://mqtt.example.com:8883",
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "/path/to/ca.crt"
}
```

**说明：**
- 只需要 CA 证书来验证服务器证书
- 服务器不需要客户端证书

### 5.2 双向 TLS（验证服务器 + 客户端证书）

适用于服务器和客户端都需要验证对方身份的场景。

```json
{
  "mqtt_broker": "mqtts://mqtt.example.com:8883",
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "/path/to/ca.crt",
  "mqtt_client_cert": "/path/to/client.crt",
  "mqtt_client_key": "/path/to/client.key"
}
```

**说明：**
- 需要 CA 证书、客户端证书和客户端私钥
- 服务器会验证客户端证书
- 客户端证书由同一个 CA 签发

## 6. 认证配置详解

### 6.1 用户名/密码认证

Mosquitto Broker 配置示例：

```properties
# mosquitto.conf
allow_anonymous false
password_file /path/to/pwfile
```

创建用户：

```bash
# 创建密码文件
mosquitto_passwd -c /path/to/pwfile username

# 添加更多用户
mosquitto_passwd /path/to/pwfile another_user
```

### 6.2 匿名认证

如果 Broker 允许匿名访问，可以不设置用户名密码：

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_use_tls": false,
  "mqtt_username": "",
  "mqtt_password": ""
}
```

## 7. 连接可靠性配置

sfsEdgeStore 的 MQTT 客户端默认配置为高可靠性模式：

| 配置 | 值 | 说明 |
|------|-----|------|
| `CleanSession` | `true` | 每次连接都是新会话，避免旧会话残留问题 |
| `AutoReconnect` | `true` | 网络中断后自动重连 |
| `MaxReconnectInterval` | `5分钟` | 最大重连间隔 |
| `ResumeSubs` | `true` | 重连后自动恢复订阅 |
| `ConnectTimeout` | `30秒` | 连接超时时间 |

## 8. 常见问题

### 8.1 连接失败

**问题**：连接到 Broker 失败

**排查步骤**：
1. 检查 Broker 地址是否正确
2. 检查端口是否正确（默认 1883，TLS 默认 8883）
3. 检查防火墙是否允许连接
4. 如果启用 TLS，检查证书是否有效

**日志示例**：
```
Failed to connect to MQTT broker: connection refused
```

### 8.2 认证失败

**问题**：用户名/密码认证被拒绝

**排查步骤**：
1. 确认用户名密码正确
2. 检查 Broker 是否启用了认证
3. 检查密码文件格式是否正确

**日志示例**：
```
Failed to connect to MQTT broker: connection refused: Bad username or password
```

### 8.3 TLS 证书错误

**问题**：TLS 连接失败

**排查步骤**：
1. 检查 CA 证书文件是否存在且有效
2. 检查客户端证书和私钥是否匹配
3. 检查证书是否由受信任的 CA 签发
4. 如果使用自签名证书，确保 `InsecureSkipVerify` 未启用（生产环境不推荐）

**日志示例**：
```
Failed to connect to MQTT broker: x509: certificate signed by unknown authority
```

### 8.4 主题订阅失败

**问题**：订阅主题失败

**排查步骤**：
1. 检查主题格式是否正确
2. 检查是否有订阅权限
3. 检查 ACL 配置

**日志示例**：
```
Failed to subscribe to topic edgex/events/core/#: Not authorized
```

## 9. 安全建议

### 9.1 生产环境

1. **必须使用 TLS**：所有生产环境都应使用 MQTTS
2. **使用强密码**：密码长度至少 12 位，包含大小写字母、数字和特殊字符
3. **定期轮换证书**：建议每年轮换一次
4. **最小权限原则**：为不同设备分配不同的用户名和密码

### 9.2 证书管理

```bash
# 检查证书有效期
openssl x509 -in client.crt -noout -dates

# 验证证书
openssl verify -CAfile ca.crt client.crt

# 检查私钥匹配
openssl x509 -in client.crt -noout -modulus | md5sum
openssl rsa -in client.key -noout -modulus | md5sum
```

## 10. 配置检查清单

部署前检查：

- [ ] Broker 地址和端口正确
- [ ] TLS 配置正确（如果启用）
- [ ] 证书文件路径正确且可读
- [ ] 用户名密码配置正确
- [ ] 客户端 ID 唯一
- [ ] 订阅主题正确
- [ ] 网络连接正常（防火墙开放端口）