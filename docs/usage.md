# 认证授权模块使用指南

## 概述

本指南描述了如何使用 sfsdb-edgex-adapter 的认证授权模块，包括 API Key 的创建、使用和管理。

## 快速开始

### 1. 启动服务

首先，启动 sfsdb-edgex-adapter 服务：

```bash
# 启动服务
./sfsdb-edgex-adapter
```

服务默认会在 `8081` 端口启动 HTTP 服务器。

### 2. 创建 API Key

由于默认情况下没有 API Key，你需要先创建一个管理员 API Key：

> 注意：首次启动时，认证管理接口暂时未受保护，以便你可以创建第一个 API Key。

```bash
# 创建管理员 API Key
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin",
    "role": "admin",
    "expires_in": 168
  }'
```

响应示例：

```json
{
  "status": "success",
  "api_key": "5f9a7b3c1d2e4f5a6b7c8d9e0f1a2b3c",
  "user_id": "admin",
  "role": "admin",
  "expires_at": "2026-03-12T15:30:00Z"
}
```

### 3. 使用 API Key

使用生成的 API Key 进行认证：

```bash
# 使用 API Key 查询数据
curl -H "X-API-Key: 5f9a7b3c1d2e4f5a6b7c8d9e0f1a2b3c" \
  "http://localhost:8081/api/readings?deviceName=TestDevice-001"
```

## 详细使用说明

### 管理 API Keys

#### 创建 API Key

```bash
# 创建用户 API Key
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "X-API-Key: <admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "role": "user",
    "expires_in": 24
  }'
```

#### 列出所有 API Keys

```bash
# 列出所有 API Keys
curl -H "X-API-Key: <admin-api-key>" \
  "http://localhost:8081/api/auth/list-keys"
```

#### 撤销 API Key

```bash
# 撤销 API Key
curl -X POST http://localhost:8081/api/auth/revoke-key \
  -H "X-API-Key: <admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "<api-key-to-revoke>"
  }'
```

### 权限管理

#### 角色说明

- **admin**：拥有所有权限，包括管理 API Keys、备份、恢复数据等
- **user**：拥有读、写和备份权限
- **readonly**：只有读权限

#### 权限对应表

| 角色 | 可访问的 API 端点 |
|------|------------------|
| admin | 所有 API 端点 |
| user | /api/readings, /api/backup, /api/test-edgex |
| readonly | /api/readings |

### 数据操作

#### 查询数据

```bash
# 查询所有数据
curl -H "X-API-Key: <api-key>" \
  "http://localhost:8081/api/readings"

# 按设备查询
curl -H "X-API-Key: <api-key>" \
  "http://localhost:8081/api/readings?deviceName=TestDevice-001"

# 按时间范围查询
curl -H "X-API-Key: <api-key>" \
  "http://localhost:8081/api/readings?startTime=2026-03-01T00:00:00Z&endTime=2026-03-02T00:00:00Z"
```

#### 备份数据

```bash
# 备份数据
curl -X POST -H "X-API-Key: <api-key>" \
  "http://localhost:8081/api/backup?path=./backups"
```

#### 恢复数据

```bash
# 恢复数据
curl -X POST -H "X-API-Key: <api-key>" \
  "http://localhost:8081/api/restore?file=./backups/backup-20260305.db"
```

#### 测试 EdgeX 消息

```bash
# 测试 EdgeX 消息
curl -X POST -H "X-API-Key: <api-key>" \
  "http://localhost:8081/api/test-edgex"
```

## 配置

### HTTPS 配置

要启用 HTTPS，需要在配置文件中设置：

```json
{
  "http_use_tls": true,
  "http_cert": "cert.pem",
  "http_key": "key.pem"
}
```

或者通过环境变量设置：

```bash
set EDGEX_HTTP_USE_TLS=true
set EDGEX_HTTP_CERT=cert.pem
set EDGEX_HTTP_KEY=key.pem
```

### MQTT TLS 配置

要启用 MQTT TLS，需要在配置文件中设置：

```json
{
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "ca.pem",
  "mqtt_client_cert": "client.pem",
  "mqtt_client_key": "client.key"
}
```

## 故障排除

### 常见问题

1. **认证失败**
   - 检查 API Key 是否正确
   - 检查 API Key 是否已过期
   - 检查 API Key 是否已被撤销

2. **权限不足**
   - 检查 API Key 的角色是否具有所需权限
   - 确保使用的 API Key 具有足够的权限

3. **API Key 管理接口不可访问**
   - 确保使用的 API Key 具有 `admin` 角色
   - 检查服务是否正常运行

### 日志查看

服务启动后，会在控制台输出日志，可以通过日志了解认证授权相关的操作和错误信息。

## 最佳实践

1. **API Key 管理**
   - 为不同的用户和应用创建不同的 API Key
   - 为 API Key 设置合理的过期时间
   - 定期撤销不再使用的 API Key

2. **权限管理**
   - 遵循最小权限原则，只授予必要的权限
   - 为不同的用户设置适当的角色

3. **安全措施**
   - 启用 HTTPS 保护 API 通信
   - 保护好 API Key，避免泄露
   - 定期更换 API Key

4. **性能优化**
   - 合理使用 API Key，避免频繁创建和撤销
   - 对于高频请求，考虑使用连接池减少认证开销
