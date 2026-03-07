# sfsEdgeStore 故障排除手册

> **文档版本**: v1.0.0  
> **最后更新**: 2026-03-07  
> **适用版本**: sfsEdgeStore v1.x

---

## 目录

1. [故障排除流程](#故障排除流程)
2. [常见问题](#常见问题)
3. [日志分析](#日志分析)
4. [诊断工具](#诊断工具)
5. [联系支持](#联系支持)

---

## 故障排除流程

### 初步诊断步骤

1. **检查服务状态**
   ```bash
   # Systemd
   sudo systemctl status sfsedgestore

   # Docker
   docker ps | grep sfsedgestore

   # 进程
   ps aux | grep sfsedgestore
   ```

2. **查看日志**
   ```bash
   # Systemd 日志
   sudo journalctl -u sfsedgestore -n 100 --no-pager

   # 应用日志
   tail -100 /var/log/sfsedgestore/app.log

   # Docker 日志
   docker logs sfsedgestore --tail 100
   ```

3. **健康检查**
   ```bash
   curl http://localhost:8080/health?full=true
   ```

4. **检查资源使用**
   ```bash
   # CPU 和内存
   top -p $(pidof sfsedgestore)

   # 磁盘空间
   df -h /var/lib/sfsedgestore

   # 网络连接
   netstat -tulpn | grep sfsedgestore
   ```

---

## 常见问题

### 启动问题

#### 问题 1: 服务无法启动

**症状**: 服务启动后立即退出

**诊断步骤**:
```bash
# 1. 查看详细日志
sudo journalctl -u sfsedgestore -n 200

# 2. 前台运行查看错误
sudo -u sfsedgestore /opt/sfsedgestore/sfsedgestore -config /etc/sfsedgestore/config.json

# 3. 检查配置文件
./sfsedgestore -config /etc/sfsedgestore/config.json -validate
```

**可能原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| 配置文件语法错误 | 检查 `config.json` 语法，使用 JSON 验证工具 |
| 数据库路径权限问题 | `sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore/data` |
| 端口被占用 | `netstat -tulpn | grep 8080`，更改端口或停止占用进程 |
| MQTT Broker 不可达 | 检查 MQTT Broker 状态和网络连接 |
| 数据库损坏 | 从备份恢复或重建数据库 |

---

### MQTT 连接问题

#### 问题 2: MQTT 连接失败

**症状**: 无法连接到 MQTT Broker

**诊断步骤**:
```bash
# 1. 测试网络连接
telnet mqtt-broker.example.com 1883
# 或
nc -zv mqtt-broker.example.com 1883

# 2. 使用 mosquitto_sub 测试
mosquitto_sub -h mqtt-broker.example.com -t 'test' -v

# 3. 检查防火墙
sudo ufw status
# 或
sudo firewall-cmd --list-all
```

**可能原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| Broker 地址错误 | 检查 `MQTTBroker` 配置 |
| 端口被防火墙阻止 | 开放 1883（非 TLS）或 8883（TLS）端口 |
| 认证失败 | 检查 `MQTTUsername` 和 `MQTTPassword` |
| TLS 证书问题 | 验证证书路径和有效期 |
| Client ID 冲突 | 确保 `MQTTClientID` 唯一 |

---

### 数据库问题

#### 问题 3: 数据库读写错误

**症状**: 无法写入或读取数据

**诊断步骤**:
```bash
# 1. 检查磁盘空间
df -h /var/lib/sfsedgestore

# 2. 检查文件权限
ls -la /var/lib/sfsedgestore/data

# 3. 检查数据库完整性
./sfsedgestore -checkdb -config /etc/sfsedgestore/config.json
```

**可能原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| 磁盘空间不足 | 清理磁盘或扩容 |
| 权限不足 | `sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore/data` |
| 数据库损坏 | 从备份恢复或重建数据库 |
| 加密密钥不匹配 | 检查 `DBEncryptionKey` 配置 |

---

### 性能问题

#### 问题 4: 系统响应缓慢

**症状**: API 响应慢，MQTT 消息处理延迟

**诊断步骤**:
```bash
# 1. 查看资源使用
top -p $(pidof sfsedgestore)

# 2. 查看指标
curl http://localhost:8080/metrics

# 3. 检查队列长度
curl http://localhost:8080/metrics | grep queue_length
```

**可能原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| CPU 使用率过高 | 减少并发连接，优化查询 |
| 内存不足 | 增加内存，调整 `DBBlockCacheSize` |
| 队列积压 | 增加 `QueueSize`，优化处理速度 |
| 磁盘 I/O 瓶颈 | 使用 SSD，优化数据库配置 |

---

### API 认证问题

#### 问题 5: API 认证失败

**症状**: 收到 401 Unauthorized 错误

**诊断步骤**:
```bash
# 1. 检查认证是否启用
curl http://localhost:8080/health | jq '.checks'

# 2. 验证 API Key
curl -H "X-API-Key: <your-key>" http://localhost:8080/health

# 3. 列出有效 API Key（需要 admin 权限）
curl -H "X-API-Key: <admin-key>" http://localhost:8080/api/auth/list-keys
```

**可能原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| API Key 缺失 | 确保请求头包含 `X-API-Key` |
| API Key 无效 | 检查 Key 是否正确，是否被撤销 |
| API Key 过期 | 创建新的 API Key |
| 权限不足 | 使用具有适当角色的 API Key |

---

### 备份恢复问题

#### 问题 6: 备份或恢复失败

**症状**: 无法创建备份或恢复数据

**诊断步骤**:
```bash
# 1. 检查备份目录权限
ls -la /var/lib/sfsedgestore/backups

# 2. 检查磁盘空间
df -h /var/lib/sfsedgestore/backups

# 3. 验证备份文件
/opt/sfsedgestore/scripts/verify-backup.sh /path/to/backup.tar.gz
```

**可能原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| 备份目录不存在 | `mkdir -p /var/lib/sfsedgestore/backups` |
| 权限不足 | `sudo chown -R sfsedgestore:sfsedgestore /var/lib/sfsedgestore/backups` |
| 磁盘空间不足 | 清理旧备份或扩容 |
| 备份文件损坏 | 使用其他备份文件 |

---

## 日志分析

### 日志级别

| 级别 | 说明 |
|------|------|
| `debug` | 详细调试信息 |
| `info` | 一般信息 |
| `warn` | 警告信息 |
| `error` | 错误信息 |

### 关键日志模式

#### 启动日志

```
[INFO] Starting sfsEdgeStore v1.0.0
[INFO] Configuration loaded successfully
[INFO] Connecting to MQTT broker...
[INFO] MQTT connected successfully
[INFO] HTTP server started on :8080
[INFO] sfsEdgeStore is running
```

#### MQTT 日志

```
[INFO] Received MQTT message from topic: edgex/events/core/...
[ERROR] Failed to process MQTT message: <error>
[WARN] MQTT connection lost, reconnecting...
```

#### 错误日志

```
[ERROR] Database error: <error>
[ERROR] API request failed: <error>
[FATAL] Critical error, shutting down: <error>
```

### 日志查询技巧

```bash
# 查看最近 100 条错误日志
sudo journalctl -u sfsedgestore -n 100 -p err

# 查看特定时间段的日志
sudo journalctl -u sfsedgestore --since "2026-03-07 10:00:00" --until "2026-03-07 12:00:00"

# 搜索特定关键词
sudo journalctl -u sfsedgestore | grep "MQTT"

# 实时跟踪日志
sudo journalctl -u sfsedgestore -f
```

---

## 诊断工具

### 内置诊断命令

```bash
# 验证配置
./sfsedgestore -config config.json -validate

# 检查数据库
./sfsedgestore -checkdb -config config.json

# 优化数据库
./sfsedgestore -optimize -config config.json

# 性能基准测试
./sfsedgestore -benchmark -config config.json

# 测试配置并退出
./sfsedgestore -test -config config.json
```

### 健康检查脚本

```bash
#!/bin/bash
# /opt/sfsedgestore/scripts/health-check.sh

echo "=== sfsEdgeStore Health Check ==="
echo ""

# 1. 检查进程
if pgrep -x "sfsedgestore" > /dev/null; then
    echo "✓ Service is running"
else
    echo "✗ Service is NOT running"
    exit 1
fi

# 2. 检查端口
if nc -zv localhost 8080 2>&1 | grep -q succeeded; then
    echo "✓ Port 8080 is open"
else
    echo "✗ Port 8080 is NOT open"
fi

# 3. 健康检查 API
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed"
fi

# 4. 检查磁盘空间
DISK_USAGE=$(df -P /var/lib/sfsedgestore | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 90 ]; then
    echo "✓ Disk usage: ${DISK_USAGE}%"
else
    echo "✗ Disk usage critical: ${DISK_USAGE}%"
fi

echo ""
echo "=== Check completed ==="
```

---

## 联系支持

### 收集诊断信息

在联系支持前，请收集以下信息：

```bash
# 1. 系统信息
uname -a
cat /etc/os-release

# 2. sfsEdgeStore 版本
./sfsedgestore -version

# 3. 配置文件（脱敏）
grep -vE 'password|key|secret' /etc/sfsedgestore/config.json

# 4. 最近日志
sudo journalctl -u sfsedgestore -n 500 --no-pager > sfsedgestore-logs.txt

# 5. 健康检查
curl http://localhost:8080/health?full=true > health-check.json

# 6. 指标
curl http://localhost:8080/metrics > metrics.txt
```

### 支持渠道

- **文档**: [文档中心](../README.md)
- **支持政策**: [SUPPORT.md](../support/SUPPORT.md)
- **问题升级**: [ESCALATION.md](../support/ESCALATION.md)

---

**文档结束**
