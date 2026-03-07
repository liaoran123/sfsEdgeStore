# 运维操作手册

## 概述

本文档提供 sfsEdgeStore 的日常运维操作指南，包括服务管理、配置管理、性能调优等内容。

## 服务管理

### 启动服务

#### Linux/macOS

```bash
# 直接启动
./sfsedgestore

# 后台启动
nohup ./sfsedgestore > sfsedgestore.log 2>&1 &

# 使用 systemd 服务
sudo systemctl start sfsedgestore
```

#### Windows

```powershell
# 直接启动
.\sfsedgestore.exe

# 后台启动（使用 PowerShell）
Start-Process -FilePath ".\sfsedgestore.exe" -WindowStyle Hidden

# 使用 Windows 服务
sc start sfsedgestore
```

### 停止服务

#### Linux/macOS

```bash
# 查找进程
ps aux | grep sfsedgestore

# 优雅停止
kill -TERM <pid>

# 使用 systemd
sudo systemctl stop sfsedgestore
```

#### Windows

```powershell
# 查找进程
Get-Process sfsedgestore

# 停止进程
Stop-Process -Name sfsedgestore

# 使用 Windows 服务
sc stop sfsedgestore
```

### 重启服务

```bash
# Linux/macOS
sudo systemctl restart sfsedgestore

# Windows
sc restart sfsedgestore
```

### 查看服务状态

```bash
# Linux/macOS
sudo systemctl status sfsedgestore

# Windows
sc query sfsedgestore
```

## 配置管理

### 配置文件位置

- 主配置文件：`config.json`
- 环境变量示例：`.env.example`
- 配置示例：`config.example.json`

### 配置热重载

sfsEdgeStore 支持部分配置的热重载：

```bash
# 发送 SIGHUP 信号触发重载
kill -HUP <pid>
```

支持热重载的配置项：
- 日志级别
- 监控采样率
- 告警阈值
- 数据保留策略

### 配置验证

```bash
# 验证配置文件语法
./sfsedgestore --validate-config
```

## 日志管理

### 日志级别

- `debug` - 调试信息
- `info` - 一般信息（默认）
- `warn` - 警告信息
- `error` - 错误信息

### 日志位置

- 默认日志目录：`logs/`
- 主日志文件：`sfsedgestore.log`
- 错误日志：`error.log`

### 日志轮转

配置日志轮转策略（config.json）：

```json
{
  "logging": {
    "maxSize": 100,
    "maxBackups": 10,
    "maxAge": 30,
    "compress": true
  }
}
```

### 日志查询

```bash
# 实时查看日志
tail -f logs/sfsedgestore.log

# 查看错误日志
grep "ERROR" logs/sfsedgestore.log

# 按时间范围查询
grep "2026-03-07" logs/sfsedgestore.log
```

## 性能监控

### 健康检查

```bash
# HTTP 健康检查
curl http://localhost:8080/health

# 详细健康状态
curl http://localhost:8080/health/detailed
```

### 性能指标

访问监控端点获取性能指标：

```bash
curl http://localhost:8080/metrics
```

关键指标：
- `request_count` - 请求总数
- `request_duration` - 请求耗时
- `queue_length` - 队列长度
- `database_ops` - 数据库操作数
- `memory_usage` - 内存使用量

### 系统资源监控

```bash
# CPU 使用
top -p <pid>

# 内存使用
ps -p <pid> -o %mem,rss

# 磁盘 I/O
iostat -x 1
```

## 数据库管理

### 数据库备份

参考 [备份恢复指南](./BACKUP_RESTORE_GUIDE.md)

### 数据库维护

#### 压缩数据库

```bash
# 通过 API 触发压缩
curl -X POST http://localhost:8080/api/admin/compact
```

#### 数据清理

```bash
# 清理过期数据（根据保留策略）
curl -X POST http://localhost:8080/api/admin/cleanup
```

### 数据库迁移

```bash
# 检查迁移状态
./sfsedgestore --migrate-status

# 执行迁移
./sfsedgestore --migrate
```

## 安全管理

### 用户管理

#### 创建用户

```bash
curl -X POST http://localhost:8080/api/admin/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "securepassword",
    "role": "user"
  }'
```

#### 修改密码

```bash
curl -X PUT http://localhost:8080/api/admin/users/newuser/password \
  -H "Content-Type: application/json" \
  -d '{"newPassword": "newsecurepassword"}'
```

#### 删除用户

```bash
curl -X DELETE http://localhost:8080/api/admin/users/newuser
```

### API Key 管理

#### 创建 API Key

```bash
curl -X POST http://localhost:8080/api/admin/api-keys \
  -H "Content-Type: application/json" \
  -d '{
    "name": "integration-key",
    "role": "user",
    "expiresAt": "2026-12-31T23:59:59Z"
  }'
```

#### 撤销 API Key

```bash
curl -X DELETE http://localhost:8080/api/admin/api-keys/<key-id>
```

### 证书管理

#### 更新 TLS 证书

```bash
# 替换证书文件
cp new-cert.pem certs/cert.pem
cp new-key.pem certs/key.pem

# 重启服务
sudo systemctl restart sfsedgestore
```

## 日常检查清单

### 每日检查

- [ ] 服务运行状态
- [ ] 磁盘空间使用
- [ ] 错误日志检查
- [ ] 告警通知检查

### 每周检查

- [ ] 性能趋势分析
- [ ] 数据库大小检查
- [ ] 备份完整性验证
- [ ] 安全审计日志

### 每月检查

- [ ] 全面性能测试
- [ ] 数据归档清理
- [ ] 配置审核
- [ ] 文档更新

## 应急响应

### 服务不可用

1. 检查服务状态
2. 查看错误日志
3. 检查系统资源
4. 尝试重启服务
5. 回滚最近变更

### 性能下降

1. 查看性能指标
2. 检查队列长度
3. 分析慢查询
4. 优化配置参数
5. 增加系统资源

### 数据损坏

1. 停止服务
2. 从备份恢复
3. 验证数据完整性
4. 分析损坏原因
5. 实施预防措施

## 更多资源

- [管理员指南](./ADMIN_GUIDE.md)
- [监控指南](./MONITORING_GUIDE.md)
- [备份恢复指南](./BACKUP_RESTORE_GUIDE.md)
- [升级指南](./UPGRADE_GUIDE.md)
- [故障排除](../user-guide/TROUBLESHOOTING.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
