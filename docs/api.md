# 认证授权 API 文档

## 概述

本文档描述了 sfsdb-edgex-adapter 认证授权模块的 API 接口，包括 API Key 管理和权限控制相关的端点。

## 认证方式

所有 API 接口（除了认证管理接口本身）都需要通过 HTTP Header 进行认证：

```
X-API-Key: <your-api-key>
```

## API 端点

### 1. 创建 API Key

**端点**: `POST /api/auth/create-key`

**权限**: 需要 `admin` 权限

**请求体**:

```json
{
  "user_id": "string",      // 用户 ID
  "role": "string",        // 角色 (admin, user, readonly)
  "expires_in": 24          // 过期时间（小时）
}
```

**响应**:

```json
{
  "status": "success",
  "api_key": "string",      // 生成的 API Key
  "user_id": "string",
  "role": "string",
  "expires_at": "2026-03-06T15:30:00Z"  // 过期时间
}
```

### 2. 列出 API Keys

**端点**: `GET /api/auth/list-keys`

**权限**: 需要 `admin` 权限

**响应**:

```json
{
  "status": "success",
  "api_keys": [
    {
      "id": "string",
      "user_id": "string",
      "role": "string",
      "created_at": "2026-03-05T15:30:00Z",
      "expires_at": "2026-03-06T15:30:00Z",
      "active": true
    }
  ]
}
```

### 3. 撤销 API Key

**端点**: `POST /api/auth/revoke-key`

**权限**: 需要 `admin` 权限

**请求体**:

```json
{
  "api_key": "string"  // 要撤销的 API Key
}
```

**响应**:

```json
{
  "status": "success",
  "message": "API key revoked successfully"
}
```

### 4. 数据查询

**端点**: `GET /api/readings`

**权限**: 需要认证（任何角色）

**查询参数**:

- `deviceName`: 设备名称
- `startTime`: 开始时间（RFC3339 格式）
- `endTime`: 结束时间（RFC3339 格式）

**响应**:

```json
{
  "count": 10,
  "readings": [
    {
      "id": "string",
      "deviceName": "string",
      "reading": "string",
      "value": 25.5,
      "valueType": "Float32",
      "baseType": "Float",
      "timestamp": 1677721600000000000,
      "metadata": "string"
    }
  ]
}
```

### 5. 数据备份

**端点**: `POST /api/backup`

**权限**: 需要 `backup` 权限

**查询参数**:

- `path`: 备份路径（可选，默认为 "./backups"）

**响应**:

```json
{
  "status": "success",
  "backupFile": "string"  // 备份文件路径
}
```

### 6. 数据恢复

**端点**: `POST /api/restore`

**权限**: 需要 `restore` 权限

**查询参数**:

- `file`: 备份文件路径（必需）

**响应**:

```json
{
  "status": "success",
  "message": "Database restored successfully"
}
```

### 7. 测试 EdgeX 消息

**端点**: `POST /api/test-edgex`

**权限**: 需要 `write` 权限

**响应**:

```json
{
  "status": "success",
  "message": "Batch stored X readings from TestDevice-001"
}
```

## 错误响应

所有 API 接口在遇到错误时都会返回统一的错误格式：

```json
{
  "error": "Error message"
}
```

常见错误码:

- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 认证失败
- `403 Forbidden`: 权限不足
- `500 Internal Server Error`: 服务器内部错误

## 角色与权限

| 角色 | 权限 | 可访问的 API 端点 |
|------|------|------------------|
| admin | read, write, admin, backup, restore | 所有 API 端点 |
| user | read, write, backup | /api/readings, /api/backup, /api/test-edgex |
| readonly | read | /api/readings |
