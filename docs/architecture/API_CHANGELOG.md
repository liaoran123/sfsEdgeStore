# API 变更日志

## 概述

本文档记录 sfsEdgeStore API 的所有变更，包括新增功能、破坏性变更、bug 修复等。

## 版本控制

### 版本号格式

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向下兼容的功能新增
- **PATCH**: 向下兼容的问题修复

## [1.1.0] - 2026-04-01

### 新增

- `GET /api/readings/export` - 数据导出端点，支持 CSV 和 JSON 格式
- `POST /api/readings/import` - 数据导入端点
- `GET /api/devices/{deviceName}/resources` - 获取设备资源列表
- `GET /api/aggregate/statistics` - 综合统计信息
- `GET /api/health/detailed` - 详细健康检查端点
- `WebSocket /ws/readings` - 实时读数推送

### 改进

- 优化 `GET /api/readings` 查询性能，支持更大的 limit
- 增加 `GET /api/devices` 的过滤选项
- 改进错误响应格式，包含更多调试信息

### 弃用

- `GET /api/readings?deprecated_param=xxx` - 某些旧参数将在 2.0 移除

## [1.0.0] - 2026-03-01

### 初始发布

#### 读数 API

- `GET /api/readings` - 查询读数
- `POST /api/readings` - 写入单个读数
- `POST /api/readings/batch` - 批量写入读数
- `DELETE /api/readings` - 删除读数

#### 设备 API

- `GET /api/devices` - 列出所有设备
- `GET /api/devices/{deviceName}` - 获取设备详情
- `POST /api/devices` - 注册新设备
- `PUT /api/devices/{deviceName}` - 更新设备信息
- `DELETE /api/devices/{deviceName}` - 删除设备

#### 聚合 API

- `GET /api/aggregate` - 聚合查询
- `GET /api/aggregate/series` - 时间序列聚合
- `GET /api/aggregate/latest` - 最新值聚合

#### 管理 API

- `GET /api/admin/stats` - 系统统计
- `POST /api/admin/backup` - 创建备份
- `POST /api/admin/restore` - 恢复备份
- `POST /api/admin/compact` - 压缩数据库
- `POST /api/admin/cleanup` - 清理过期数据

#### 健康检查

- `GET /health` - 基本健康检查
- `GET /metrics` - Prometheus 格式指标
- `GET /metrics/json` - JSON 格式指标

## 0.x.x 版本

### [0.9.0] - 2026-02-15

- 首次公开预览版
- 基础读数 API
- 基础设备 API
- MQTT 集成

### [0.8.0] - 2026-02-01

- 内部测试版
- LevelDB 存储
- 批量处理

## 破坏性变更

### 1.0.0 → 2.0.0 (计划)

预计变更：
- 移除所有 1.x 中标记为弃用的端点
- 重构认证机制
- 统一分页参数格式

## 迁移指南

### 从 0.9.0 升级到 1.0.0

#### 读数查询参数变更

**旧版本**:
```bash
curl "http://localhost:8080/api/readings?device=device001&start_time=1620000000000&end_time=1620003600000"
```

**新版本**:
```bash
curl "http://localhost:8080/api/readings?device=device001&start=1620000000000&end=1620003600000"
```

变更：
- `start_time` → `start`
- `end_time` → `end`

#### 错误响应格式变更

**旧版本**:
```json
{
  "error": "Invalid parameter"
}
```

**新版本**:
```json
{
  "code": 400,
  "message": "Invalid parameter",
  "details": {
    "parameter": "start",
    "reason": "must be a valid timestamp"
  },
  "requestId": "req-12345"
}
```

## 兼容性承诺

### 稳定 API

以下 API 在 MAJOR 版本内保持稳定：
- `/api/readings`
- `/api/devices`
- `/api/aggregate`
- `/health`
- `/metrics`

### 实验性 API

以下 API 可能在 MINOR 版本中变更：
- `/api/admin/*` (管理 API)
- `/ws/*` (WebSocket)
- 任何标记为 `experimental` 的端点

## 更多资源

- [API 文档](../api.md)
- [升级指南](../admin-guide/UPGRADE_GUIDE.md)
- [架构文档](./ARCHITECTURE.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
