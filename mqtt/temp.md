
          
## 数据库表结构 vs EdgeX 消息标准对比

### 📊 字段映射对比

| EdgeX 原始字段 | messageProcessor.go 存储字段 | createDb.go 数据库字段 | 状态 |
|---------------|---------------------------|---------------------|------|
| `reading.ID` | `id` | `id` | ✅ 一致 |
| `event.DeviceName` | `deviceName` | `deviceName` | ✅ 一致 |
| EdgeX 无此字段 | `deviceType` | `deviceType` | ⚠️ 多余字段 |
| `reading.ResourceName` | `reading` | `reading` | ✅ 一致 |
| `ParseValue(reading.Value)` | `value` | `value` (float64) | ✅ 一致 |
| `reading.ValueType` | `valueType` | `valueType` | ✅ 一致 |
| `reading.BaseType` | `baseType` | `baseType` | ✅ 一致 |
| `reading.Origin` | `timestamp` | `timestamp` (int64) | ✅ 一致 |
| `reading.Metadata` | `metadata` | `metadata` | ✅ 一致 |
| `reading.DeviceName` | 未存储 | 未定义 | ℹ️ 已在 event 中 |
| `reading.ProfileName` | 未存储 | 未定义 | ⚠️ 缺失 |

### ⚠️ 问题

1. **多余字段**：数据库有 `deviceType` 字段，但 EdgeX Reading 中没有此字段
2. **缺失字段**：EdgeX Reading 有 `ProfileName`，但未存储到数据库

### 💡 建议

**移除无用字段**：`deviceType` 在 EdgeX 中不存在，代码中也从未写入值（始终为空字符串）。