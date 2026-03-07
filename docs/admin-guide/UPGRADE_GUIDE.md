# 升级指南

## 概述

本文档详细介绍 sfsEdgeStore 的版本升级流程，包括升级准备、升级步骤、回滚策略等内容。

## 版本兼容性

### 版本号格式

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: 不兼容的 API 变更
- **MINOR**: 向下兼容的功能新增
- **PATCH**: 向下兼容的问题修复

### 兼容性矩阵

| 从版本 | 到版本 | 兼容性 | 说明 |
|--------|--------|--------|------|
| 1.0.x | 1.0.y | ✅ 完全兼容 | 直接升级 |
| 1.0.x | 1.1.0 | ✅ 兼容 | 可能需要配置更新 |
| 1.0.x | 2.0.0 | ⚠️ 不兼容 | 需要迁移数据 |

## 升级前准备

### 1. 阅读发布说明

在升级前，务必阅读目标版本的 [CHANGELOG.md](../../CHANGELOG.md)：

- 新功能介绍
- 破坏性变更
- 配置变更
- 已知问题

### 2. 备份数据

**重要！** 在升级前必须进行完整备份：

```bash
# 使用备份脚本
./scripts/backup.sh --type full --dest /backups/pre-upgrade

# 或手动备份
sudo systemctl stop sfsedgestore
cp -r data/ data.bak.pre-upgrade/
sudo systemctl start sfsedgestore
```

### 3. 备份配置

```bash
# 备份配置文件
cp config.json config.json.bak.pre-upgrade
cp .env .env.bak.pre-upgrade
```

### 4. 检查系统要求

验证目标版本的系统要求：

```bash
# 检查 Go 版本
go version

# 检查依赖
./sfsedgestore --version
```

### 5. 准备测试环境

建议先在测试环境验证升级：

```bash
# 在测试环境部署当前版本
# 复制生产数据到测试环境
# 在测试环境执行升级
# 验证功能和性能
```

### 6. 通知相关方

- 通知用户计划停机时间
- 告知支持团队准备响应
- 确认回滚计划

## 升级步骤

### 场景 1: PATCH 版本升级 (1.0.0 → 1.0.1)

#### 1. 下载新版本

```bash
# 从 GitHub Releases 下载
wget https://github.com/username/sfsedgestore/releases/download/v1.0.1/sfsedgestore-v1.0.1-linux-amd64.tar.gz

# 或使用 Go 安装
go install sfsedgestore@v1.0.1
```

#### 2. 解压文件

```bash
tar -xzf sfsedgestore-v1.0.1-linux-amd64.tar.gz
cd sfsedgestore-v1.0.1
```

#### 3. 停止服务

```bash
sudo systemctl stop sfsedgestore
```

#### 4. 替换二进制文件

```bash
# 备份旧版本
cp /usr/local/bin/sfsedgestore /usr/local/bin/sfsedgestore.bak

# 替换新版本
cp sfsedgestore /usr/local/bin/sfsedgestore

# 设置执行权限
chmod +x /usr/local/bin/sfsedgestore
```

#### 5. 验证版本

```bash
sfsedgestore --version
# 输出: sfsedgestore version 1.0.1
```

#### 6. 启动服务

```bash
sudo systemctl start sfsedgestore
```

#### 7. 验证升级

```bash
# 检查服务状态
sudo systemctl status sfsedgestore

# 检查日志
tail -f logs/sfsedgestore.log

# 健康检查
curl http://localhost:8080/health
```

### 场景 2: MINOR 版本升级 (1.0.x → 1.1.0)

#### 1-5. 同 PATCH 升级步骤

#### 6. 检查配置变更

```bash
# 比较配置文件
diff config.json config.example.json

# 根据发布说明更新配置
vim config.json
```

#### 7. 执行数据库迁移

```bash
# 检查迁移状态
sfsedgestore --migrate-status

# 执行迁移
sfsedgestore --migrate
```

#### 8-9. 启动服务并验证

同 PATCH 升级步骤

### 场景 3: MAJOR 版本升级 (1.x.x → 2.0.0)

#### 1. 完整备份（额外备份）

```bash
sudo systemctl stop sfsedgestore
cp -r data/ data.bak.major-upgrade/
cp config.json config.json.bak.major-upgrade/
```

#### 2. 阅读迁移指南

```bash
# 查看迁移文档
cat docs/migration/MIGRATION_1_x_to_2_0.md
```

#### 3. 使用迁移工具

```bash
# 运行数据迁移工具
sfsedgestore-migrate --from 1.x --to 2.0 --data-dir ./data
```

#### 4. 更新配置文件

```bash
# 使用配置迁移工具
sfsedgestore-config-migrate --input config.json --output config-new.json

# 检查并应用新配置
mv config.json config.json.old
mv config-new.json config.json
```

#### 5. 部署新版本

同 MINOR 升级步骤

#### 6. 全面验证

```bash
# 运行完整测试套件
./scripts/full_test.ps1

# 验证 API 兼容性
./scripts/api_compat_check.sh

# 性能基准测试
./scripts/performance_test.sh
```

## Docker 升级

### 使用 docker-compose

```bash
# 更新镜像版本
vim docker-compose.yml
# 修改 image: sfsedgestore:v1.0.1

# 拉取新镜像
docker-compose pull

# 停止旧容器
docker-compose stop sfsedgestore

# 启动新容器
docker-compose up -d sfsedgestore

# 查看日志
docker-compose logs -f sfsedgestore
```

### 使用 Docker 命令

```bash
# 拉取新镜像
docker pull sfsedgestore:v1.0.1

# 停止旧容器
docker stop sfsedgestore

# 重命名旧容器
docker rename sfsedgestore sfsedgestore-old

# 启动新容器
docker run -d \
  --name sfsedgestore \
  -v sfsedgestore_data:/data \
  -p 8080:8080 \
  sfsedgestore:v1.0.1

# 验证后删除旧容器
docker rm sfsedgestore-old
```

## Kubernetes 升级

### 使用 Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sfsedgestore
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    spec:
      containers:
      - name: sfsedgestore
        image: sfsedgestore:v1.0.1  # 更新这里
```

```bash
# 应用更新
kubectl apply -f deployment.yaml

# 监控升级
kubectl rollout status deployment/sfsedgestore

# 查看历史
kubectl rollout history deployment/sfsedgestore
```

## 回滚策略

### 何时回滚

出现以下情况时考虑回滚：
- 服务无法启动
- 关键功能不可用
- 数据损坏
- 性能严重下降
- 安全问题

### 回滚步骤

#### 1. 停止服务

```bash
sudo systemctl stop sfsedgestore
```

#### 2. 恢复旧版本二进制

```bash
cp /usr/local/bin/sfsedgestore.bak /usr/local/bin/sfsedgestore
```

#### 3. 恢复配置

```bash
cp config.json.bak.pre-upgrade config.json
```

#### 4. 恢复数据（如果需要）

```bash
# 只有数据损坏时才执行
rm -rf data/
cp -r data.bak.pre-upgrade/ data/
```

#### 5. 启动旧版本

```bash
sudo systemctl start sfsedgestore
```

#### 6. 验证回滚

```bash
# 检查版本
sfsedgestore --version

# 检查服务状态
sudo systemctl status sfsedgestore

# 健康检查
curl http://localhost:8080/health
```

## 升级检查清单

### 升级前

- [ ] 阅读发布说明
- [ ] 备份数据
- [ ] 备份配置
- [ ] 检查系统要求
- [ ] 准备测试环境
- [ ] 通知相关方
- [ ] 制定回滚计划

### 升级中

- [ ] 停止服务
- [ ] 替换二进制/镜像
- [ ] 更新配置
- [ ] 执行数据迁移
- [ ] 启动服务

### 升级后

- [ ] 检查服务状态
- [ ] 查看日志
- [ ] 健康检查
- [ ] 功能验证
- [ ] 性能测试
- [ ] 通知相关方
- [ ] 文档更新

## 零停机升级

### 使用负载均衡器

```yaml
# 配置健康检查
healthCheck:
  path: /health
  interval: 5s
  timeout: 3s
```

### 滚动升级

1. 部署新版本实例
2. 等待健康检查通过
3. 逐步将流量切换到新版本
4. 下线旧版本实例

## 常见问题

### Q: 升级后服务无法启动？

A: 检查：
1. 配置文件格式
2. 日志中的错误信息
3. 端口是否被占用
4. 数据目录权限

### Q: 数据迁移失败怎么办？

A: 
1. 停止迁移
2. 从备份恢复数据
3. 查看迁移日志
4. 联系技术支持

### Q: 可以跳过版本升级吗？

A: 不建议跳过 MINOR 或 MAJOR 版本，建议按顺序升级。

## 更多资源

- [运维操作手册](./OPERATIONS_GUIDE.md)
- [备份恢复指南](./BACKUP_RESTORE_GUIDE.md)
- [CHANGELOG.md](../../CHANGELOG.md)
- [架构文档](../architecture/ARCHITECTURE.md)

---

**文档版本**: 1.0  
**最后更新**: 2026-03-07
