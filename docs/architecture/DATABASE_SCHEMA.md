# 数据库 Schema

## 概述

本文档详细描述 sfsEdgeStore 的数据库 schema 和数据存储格式。

## 数据库引擎

sfsEdgeStore 使用 LevelDB 作为底层存储引擎，这是一个键值存储数据库。

## 键设计

### 键格式

所有键都遵循以下格式：

```
{Prefix}{DeviceName}{Separator}{ResourceName}{Separator}{Timestamp}
```

### 前缀说明

| 前缀 | 说明 | 示例 |
|------|------|------|
| `R:` | 读数数据 | `R:device001:temperature:1620000000000` |
| `D:` | 设备元数据 | `D:device001` |
| `M:` | 元数据索引 | `M:device001:temperature` |
| `A:` | 聚合数据 | `A:device001:temperature:1h:1620000000000` |
| `I:` | 倒排索引 | `I:timestamp:1620000000000` |

### 设备名格式化

设备名统一格式化为 64 字符：
- 不足 64 字符：右侧补零
- 超过 64 字符：截断至 64 字符

```go
func FormatDeviceName(name string) string {
    if len(name) > 64 {
        return name[:64]
    }
    return fmt.Sprintf("%-64s", name)
}
```

## 读数数据 (Readings)

### 键格式

```
R:{DeviceName}:{ResourceName}:{Timestamp}
```

### 值格式 (JSON)

```json
{
  "id": "uuid-string",
  "deviceName": "device001",
  "resourceName": "temperature",
  "value": "25.5",
  "valueType": "Float64",
  "timestamp": 1620000000000,
  "origin": 1620000000000,
  "profileName": "temperature-sensor",
  "metadata": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

### 值类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `Bool` | 布尔值 | "true", "false" |
| `String` | 字符串 | "hello" |
| `Uint8` | 8位无符号整数 | "255" |
| `Uint16` | 16位无符号整数 | "65535" |
| `Uint32` | 32位无符号整数 | "4294967295" |
| `Uint64` | 64位无符号整数 | "18446744073709551615" |
| `Int8` | 8位有符号整数 | "-128" |
| `Int16` | 16位有符号整数 | "-32768" |
| `Int32` | 32位有符号整数 | "-2147483648" |
| `Int64` | 64位有符号整数 | "-9223372036854775808" |
| `Float32` | 32位浮点数 | "3.14" |
| `Float64` | 64位浮点数 | "3.1415926535" |
| `Binary` | 二进制数据 (base64) | "SGVsbG8=" |
| `Object` | JSON 对象 | "{\"key\":\"value\"}" |
| `Array` | JSON 数组 | "[1,2,3]" |

### 示例

#### 键

```
R:device001                                                    :temperature:1620000000000
```

#### 值

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "deviceName": "device001",
  "resourceName": "temperature",
  "value": "25.5",
  "valueType": "Float64",
  "timestamp": 1620000000000,
  "origin": 1620000000000,
  "profileName": "temperature-sensor",
  "metadata": {
    "unit": "celsius",
    "accuracy": "0.1"
  }
}
```

## 设备元数据 (Device Metadata)

### 键格式

```
D:{DeviceName}
```

### 值格式 (JSON)

```json
{
  "deviceName": "device001",
  "profileName": "temperature-sensor",
  "description": "Temperature sensor in room 101",
  "location": "room-101",
  "resources": [
    {
      "name": "temperature",
      "valueType": "Float64",
      "unit": "celsius"
    },
    {
      "name": "humidity",
      "valueType": "Float64",
      "unit": "percent"
    }
  ],
  "lastSeen": 1620000000000,
  "createdAt": 1620000000000,
  "metadata": {
    "vendor": "Acme Corp",
    "model": "TH-100"
  }
}
```

## 元数据索引 (Metadata Index)

### 键格式

```
M:{DeviceName}:{ResourceName}
```

### 值格式 (JSON)

```json
{
  "deviceName": "device001",
  "resourceName": "temperature",
  "firstTimestamp": 1620000000000,
  "lastTimestamp": 1620003600000,
  "count": 3600,
  "valueType": "Float64",
  "minValue": "20.0",
  "maxValue": "30.0",
  "avgValue": "25.0"
}
```

## 聚合数据 (Aggregated Data)

### 键格式

```
A:{DeviceName}:{ResourceName}:{Granularity}:{Timestamp}
```

### 粒度说明

| 粒度 | 说明 | 时间戳对齐 |
|------|------|------------|
| `1m` | 1分钟 | 分钟开始 |
| `5m` | 5分钟 | 5分钟边界 |
| `15m` | 15分钟 | 15分钟边界 |
| `1h` | 1小时 | 小时开始 |
| `6h` | 6小时 | 6小时边界 |
| `1d` | 1天 | 天开始 (UTC) |

### 值格式 (JSON)

```json
{
  "deviceName": "device001",
  "resourceName": "temperature",
  "granularity": "1h",
  "timestamp": 1620000000000,
  "count": 3600,
  "min": 20.0,
  "max": 30.0,
  "avg": 25.0,
  "sum": 90000.0,
  "first": 22.5,
  "last": 27.5,
  "stddev": 2.5
}
```

## 索引策略

### 主键索引 (Primary Index)

```
R:{DeviceName}:{ResourceName}:{Timestamp}
```

支持的查询：
- 按设备 + 资源 + 时间范围查询
- 按设备 + 资源查询最新数据
- 按设备 + 资源 + 时间点精确查询

### 时间范围索引 (Time Range Index)

利用 LevelDB 的有序存储特性，支持高效的时间范围扫描。

### 设备索引 (Device Index)

通过键前缀扫描，可以快速获取设备的所有数据。

## 数据压缩

### LevelDB 压缩

LevelDB 默认使用 Snappy 压缩算法，可以通过配置调整：

```json
{
  "database": {
    "compression": "snappy",
    "blockSize": 4096,
    "writeBufferSize": 4194304
  }
}
```

### 压缩选项

| 选项 | 说明 |
|------|------|
| `none` | 不压缩 |
| `snappy` | Snappy 压缩（默认，平衡速度和压缩率） |
| `zlib` | zlib 压缩（更高压缩率，更慢） |

## 数据加密

### 加密选项

可以启用透明数据加密 (TDE)：

```json
{
  "database": {
    "encryption": {
      "enabled": true,
      "key": "base64-encoded-32-byte-key",
      "algorithm": "AES-256-GCM"
    }
  }
}
```

### 加密范围

- 所有读数数据值
- 设备元数据
- 聚合数据

注意：键不加密，以保持索引功能。

## 备份格式

### 完整备份

完整备份包含整个 LevelDB 目录：

```
backups/
└── sfsedgestore-20260307-020000/
    ├── CURRENT
    ├── MANIFEST-000001
    ├── 000001.log
    ├── 000002.sst
    └── ...
```

### 增量备份

增量备份只包含自上次备份以来的变更：

```
backups/
├── sfsedgestore-20260307-020000-full/
├── sfsedgestore-20260307-030000-incr/
└── sfsedgestore-20260307-040000-incr/
```

## 迁移指南

### 从 1.0 迁移到 1.1

1. 停止服务
2. 备份数据库
3. 运行迁移工具
4. 验证数据完整性
5. 启动服务

```bash
sfsedgestore --migrate --from 1.0 --to 1.1
```

## 更多资源

- [架构文档](./ARCHITECTURE.md)
- [设计决策记录](./DESIGN_DECISIONS.md)
- [运维操作手册](../admin-guide/OPERATIONS_GUIDE.md)
- [备份恢复指南](../admin-guide/BACKUP_RESTORE_GUIDE.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
