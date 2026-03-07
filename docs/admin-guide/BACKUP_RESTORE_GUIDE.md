# 备份恢复指南

## 概述

本文档详细介绍 sfsEdgeStore 的数据备份和恢复策略，包括备份方法、恢复流程、验证步骤等内容。

## 备份策略

### 备份类型

| 备份类型 | 说明 | 适用场景 | 频率建议 |
|----------|------|----------|----------|
| **完整备份** | 备份整个数据库 | 重大变更前、定期基准 | 每日/每周 |
| **增量备份** | 备份自上次备份的变更 | 日常备份 | 每小时 |
| **实时备份** | 连续复制数据 | 高可用性要求 | 持续 |

### 保留策略

```json
{
  "retention": {
    "daily": 7,
    "weekly": 4,
    "monthly": 12,
    "yearly": 3
  }
}
```

### 备份计划建议

- **每日**: 每日凌晨 2:00 完整备份
- **每小时**: 每小时增量备份
- **每周**: 每周日完整备份
- **每月**: 每月1日完整备份

## 备份方法

### 1. 使用内置备份 API

#### 创建备份

```bash
curl -X POST http://localhost:8080/api/admin/backup \
  -H "Content-Type: application/json" \
  -d '{
    "type": "full",
    "destination": "/backups/sfsedgestore-$(date +%Y%m%d-%H%M%S)"
  }'
```

响应示例：

```json
{
  "id": "backup-12345",
  "status": "in_progress",
  "type": "full",
  "startedAt": 1620000000000
}
```

#### 查询备份状态

```bash
curl http://localhost:8080/api/admin/backup/backup-12345
```

#### 列出备份

```bash
curl http://localhost:8080/api/admin/backups
```

### 2. 使用脚本备份

项目提供了备份脚本：

#### Linux/macOS

```bash
cd scripts
./backup.sh
```

#### Windows

```powershell
cd scripts
.\backup.ps1
```

#### 脚本参数

```bash
# 完整备份
./backup.sh --type full --dest /backups

# 增量备份
./backup.sh --type incremental --dest /backups

# 指定保留天数
./backup.sh --retention 30
```

### 3. 手动备份（文件系统）

#### 停止服务

```bash
# Linux/macOS
sudo systemctl stop sfsedgestore

# Windows
sc stop sfsedgestore
```

#### 复制数据目录

```bash
# 复制整个数据目录
cp -r data/ backups/sfsedgestore-$(date +%Y%m%d-%H%M%S)/

# 使用 tar 打包
tar -czf backups/sfsedgestore-$(date +%Y%m%d-%H%M%S).tar.gz data/
```

#### 启动服务

```bash
# Linux/macOS
sudo systemctl start sfsedgestore

# Windows
sc start sfsedgestore
```

### 4. Docker 备份

#### 备份数据卷

```bash
# 创建备份容器
docker run --rm \
  -v sfsedgestore_data:/data \
  -v $(pwd)/backups:/backups \
  alpine tar czf /backups/sfsedgestore-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .
```

#### 使用 docker-compose

```bash
docker-compose exec -T sfsedgestore tar czf - /data > backups/sfsedgestore-$(date +%Y%m%d-%H%M%S).tar.gz
```

## 恢复流程

### 恢复前准备

1. **停止服务**
   ```bash
   sudo systemctl stop sfsedgestore
   ```

2. **备份当前数据（如果需要）**
   ```bash
   mv data/ data.bak.$(date +%Y%m%d-%H%M%S)/
   ```

3. **验证备份文件完整性**
   ```bash
   # 检查文件是否存在
   ls -lh backups/sfsedgestore-20260307-020000.tar.gz

   # 检查 tar 包完整性
   tar -tzf backups/sfsedgestore-20260307-020000.tar.gz
   ```

### 恢复步骤

#### 使用内置恢复 API

```bash
curl -X POST http://localhost:8080/api/admin/restore \
  -H "Content-Type: application/json" \
  -d '{
    "backupId": "backup-12345"
  }'
```

#### 从文件系统恢复

```bash
# 解压备份
tar -xzf backups/sfsedgestore-20260307-020000.tar.gz -C ./

# 或复制目录
cp -r backups/sfsedgestore-20260307-020000/data ./data

# 设置正确权限
chown -R sfsedgestore:sfsedgestore data/
```

#### Docker 恢复

```bash
# 停止容器
docker-compose stop sfsedgestore

# 删除旧数据卷
docker volume rm sfsedgestore_data

# 创建新数据卷并恢复
docker run --rm \
  -v sfsedgestore_data:/data \
  -v $(pwd)/backups:/backups \
  alpine tar xzf /backups/sfsedgestore-20260307-020000.tar.gz -C /data

# 启动容器
docker-compose start sfsedgestore
```

### 恢复后验证

#### 1. 检查服务状态

```bash
# Linux/macOS
sudo systemctl status sfsedgestore

# Windows
sc query sfsedgestore
```

#### 2. 检查日志

```bash
tail -f logs/sfsedgestore.log
```

#### 3. 健康检查

```bash
curl http://localhost:8080/health
```

#### 4. 数据验证

```bash
# 查询一些数据验证完整性
curl "http://localhost:8080/api/readings?limit=10"

# 检查数据统计
curl http://localhost:8080/api/stats
```

## 验证备份

### 备份完整性检查

```bash
# 使用脚本验证
./scripts/verify_backup.sh backups/sfsedgestore-20260307-020000.tar.gz
```

### 定期恢复测试

建议每月进行一次恢复测试：

1. 在测试环境恢复最新备份
2. 验证数据完整性
3. 运行功能测试
4. 记录测试结果

## 远程备份

### 上传到云存储

#### AWS S3

```bash
# 使用 AWS CLI
aws s3 cp backups/sfsedgestore-20260307-020000.tar.gz s3://my-bucket/sfsedgestore/

# 加密上传
aws s3 cp backups/sfsedgestore-20260307-020000.tar.gz s3://my-bucket/sfsedgestore/ --sse
```

#### Google Cloud Storage

```bash
gsutil cp backups/sfsedgestore-20260307-020000.tar.gz gs://my-bucket/sfsedgestore/
```

#### Azure Blob Storage

```bash
az storage blob upload \
  --account-name myaccount \
  --container-name sfsedgestore \
  --name sfsedgestore-20260307-020000.tar.gz \
  --file backups/sfsedgestore-20260307-020000.tar.gz
```

### rsync 同步

```bash
# 同步到远程服务器
rsync -avz backups/ user@backup-server:/path/to/backups/sfsedgestore/
```

## 灾难恢复

### 场景 1: 数据损坏

1. 停止服务
2. 从最新完整备份恢复
3. 应用增量备份（如果有）
4. 验证数据完整性
5. 启动服务

### 场景 2: 服务器故障

1. 在新服务器上部署 sfsEdgeStore
2. 从远程备份恢复数据
3. 配置相同的设置
4. 启动并验证服务

### 场景 3: 人为误操作

1. 识别误操作时间点
2. 选择该时间点之前的备份
3. 执行恢复
4. 验证数据

## 安全建议

1. **加密备份**
   - 使用 AES-256 加密备份文件
   - 管理好加密密钥

2. **访问控制**
   - 限制备份文件访问权限
   - 使用最小权限原则

3. **审计日志**
   - 记录所有备份操作
   - 定期审查备份日志

4. **异地备份**
   - 至少保存一份异地备份
   - 考虑跨区域备份

## 更多资源

- [运维操作手册](./OPERATIONS_GUIDE.md)
- [监控指南](./MONITORING_GUIDE.md)
- [管理员指南](./ADMIN_GUIDE.md)
- [架构文档](../architecture/ARCHITECTURE.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
