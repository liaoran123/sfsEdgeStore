toolName: view_files
            
status: success
          
            
filePath: d:\MyGo\src\sfsEdgeStore\mqtt\batchWriter.go
          
针对边缘计算资源受限环境，符合"小而美"原则的必要优化：

## 核心问题

### 1. **Pool 过大导致资源浪费**
`pool.NewPoolForIO()` = `CPU*4` 个协程，边缘设备通常 CPU 较少，且 LevelDB 是串行写入，不需要高并发。

**优化：限制 pool 大小为 1-2**

### 2. **FreeOSMemory 频繁触发导致 STW**
每 50 次 flush 调用 `debug.FreeOSMemory()` 会触发全局 GC STW（Stop The World），影响实时性。

**优化：移除或改为更低频率**

### 3. **pendingRecords 无限增长风险**
如果写入慢于数据到达，`pendingRecords` 可能无限增长导致 OOM。

**优化：增加上限并丢弃最旧数据**

### 4. **`*map[string]any` 冗余指针**
map 本身是引用类型，外层指针无意义。

**优化：改为 `[]map[string]any`**

---

需要我实施这些优化吗？