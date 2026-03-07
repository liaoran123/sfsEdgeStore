# sfsEdgeStore 常见问题解答

> **文档版本**: v1.0.0  
> **最后更新**: 2026-03-07  
> **适用版本**: sfsEdgeStore v1.x

---

## 目录

1. [通用问题](#通用问题)
2. [安装部署](#安装部署)
3. [配置管理](#配置管理)
4. [性能相关](#性能相关)
5. [安全相关](#安全相关)
6. [故障排除](#故障排除)

---

## 通用问题

### Q1: sfsEdgeStore 是什么？

**A**: sfsEdgeStore 是一个轻量级边缘计算数据存储适配器，专为边缘设备设计，作为 EdgeX Foundry 和 sfsDb 数据库之间的桥梁。它提供高效的本地数据读写和缓存能力，支持断网运行，非常适合工业物联网边缘数据采集场景。

### Q2: sfsEdgeStore 适合哪些场景？

**A**: sfsEdgeStore 特别适合以下场景：
- 工业物联网边缘数据采集
- 智能设备本地数据缓存
- 边缘节点数据预处理
- 离线数据存储和同步
- 需要低延迟、高可靠性的边缘计算应用

### Q3: sfsEdgeStore 和其他解决方案相比有什么优势？

**A**: 主要优势包括：
- **超轻量**: 内存占用 < 50MB，CPU 使用率 < 5%
- **极速启动**: 平均启动时间仅 0.187 秒
- **本地优先**: 网络中断不影响本地数据采集
- **易于部署**: 单二进制文件，无外部依赖
- **企业级**: 内置认证、监控、告警、备份等功能

### Q4: sfsEdgeStore 支持哪些操作系统？

**A**: 支持以下操作系统：
- **Linux** (推荐): Ubuntu 20.04+, Debian 11+, CentOS 8+, RHEL 8+
- **Windows**: Windows 10+, Windows Server 2019+
- **macOS**: macOS 11+ (仅用于开发和测试)

### Q5: sfsEdgeStore 是开源的吗？

**A**: sfsEdgeStore 提供多种许可证选项：
- **社区版**: 开源免费，Apache License 2.0
- **专业版**: 商业许可证，包含技术支持
- **企业版**: 商业许可证，包含企业级功能和 SLA

详细信息请参考 [定价文档](../pricing/PRICING.md)。

---

## 安装部署

### Q6: 如何快速安装 sfsEdgeStore？

**A**: 最简单的方式是使用预编译二进制：

```bash
# 下载
wget https://github.com/your-org/sfsEdgeStore/releases/latest/download/sfsedgestore-linux-amd64
chmod +x sfsedgestore-linux-amd64

# 配置
cp config.example.json config.json
# 编辑 config.json

# 运行
./sfsedgestore-linux-amd64
```

详细步骤请参考 [用户手册](./USER_MANUAL.md)。

### Q7: 可以在 Docker 中运行吗？

**A**: 当然可以！我们提供官方 Docker 镜像：

```bash
docker pull your-org/sfsedgestore:latest
docker run -d -p 8080:8080 -v ./data:/app/data -v ./config.json:/app/config.json:ro your-org/sfsedgestore:latest
```

### Q8: 支持 Kubernetes 部署吗？

**A**: 支持！我们提供完整的 Kubernetes 部署清单，包括 Deployment、Service、Ingress、ConfigMap、Secret、HPA 等。

详细配置请参考 [管理员指南](../admin-guide/ADMIN_GUIDE.md)。

### Q9: 如何配置 Systemd 服务？

**A**: 请参考 [管理员指南](../admin-guide/ADMIN_GUIDE.md) 中的 Systemd 服务部署章节，里面有完整的服务文件示例和配置步骤。

---

## 配置管理

### Q10: 配置文件在哪里？

**A**: 默认情况下，sfsEdgeStore 会按以下顺序查找配置：
1. 命令行参数指定的路径: `-config /path/to/config.json`
2. 当前目录下的 `config.json`
3. 环境变量

### Q11: 如何使用环境变量配置？

**A**: 所有配置项都可以通过环境变量设置，格式为 `EDGEX_配置项名`（大写）：

```bash
export EDGEX_MQTT_BROKER="tcp://mqtt.example.com:1883"
export EDGEX_HTTP_PORT="8081"
export EDGEX_DB_PATH="/var/lib/sfsedgestore/data"
export EDGEX_LOG_LEVEL="debug"
```

### Q12: 配置变更需要重启吗？

**A**: 大部分配置变更需要重启服务。部分配置支持热重载，可以通过发送 SIGHUP 信号来重载：

```bash
sudo systemctl reload sfsedgestore
# 或
kill -HUP $(pidof sfsedgestore)
```

### Q13: 如何验证配置文件？

**A**: 使用内置的验证命令：

```bash
./sfsedgestore -config config.json -validate
```

---

## 性能相关

### Q14: sfsEdgeStore 的性能如何？

**A**: sfsEdgeStore 专为边缘计算优化，性能表现优秀：

| 指标 | 实际值 | 状态 |
|------|--------|------|
| 内存占用 | 20.85 MB | ✅ 优秀 |
| CPU 使用率 | 2.9% | ✅ 良好 |
| 启动时间 | 0.187 秒（平均） | ✅ 极快 |
| 数据库大小 | 0.25 MB（18,681 条记录） | ✅ 高效 |

详细性能报告请参考 [性能报告](../../PERFORMANCE_REPORT.md)。

### Q15: 如何优化性能？

**A**: 以下是一些性能优化建议：

1. **数据库优化**:
   ```json
   {
     "DBCompression": true,
     "DBMaxOpenFiles": 2000
   }
   ```

2. **MQTT 优化**:
   ```json
   {
     "MQTTQoS": 1,
     "QueueSize": 50000
   }
   ```

3. **资源限制**: 使用 systemd 或容器限制资源使用

更多优化建议请参考 [管理员指南](../admin-guide/ADMIN_GUIDE.md)。

### Q16: 支持多少并发连接？

**A**: 默认支持 100 个并发 HTTP 连接，可以通过配置调整：

```json
{
  "HTTPMaxConnections": 1000
}
```

实际并发能力取决于硬件配置。

### Q17: 数据库能存储多少数据？

**A**: sfsDb 数据库理论上可以存储 TB 级别的数据，但实际受限于：
- 磁盘空间
- 内存大小（影响查询性能）
- 数据保留策略配置

建议定期清理过期数据或配置数据保留策略。

---

## 安全相关

### Q18: 数据是加密存储的吗？

**A**: sfsEdgeStore 支持数据库加密：

```json
{
  "DBUseEncryption": true,
  "DBEncryptionKey": "your-strong-encryption-key"
}
```

注意：请妥善保管加密密钥，丢失密钥将无法恢复数据。

### Q19: 如何保护 API 访问？

**A**: 可以通过以下方式保护 API：

1. **启用 API Key 认证**:
   ```json
   {
     "EnableAuth": true,
     "AuthAPIKeyRequired": true
   }
   ```

2. **使用 HTTPS**:
   ```json
   {
     "HTTPUseTLS": true,
     "HTTPCert": "cert.pem",
     "HTTPKey": "key.pem"
   }
   ```

3. **配置 IP 白名单**（如支持）
4. **使用防火墙限制访问**

### Q20: API Key 泄露了怎么办？

**A**: 如果 API Key 泄露，请立即：

1. **撤销泄露的 Key**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/revoke-key \
     -H "X-API-Key: <admin-key>" \
     -H "Content-Type: application/json" \
     -d '{"api_key": "<leaked-key>"}'
   ```

2. **创建新的 Key**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/create-key \
     -H "X-API-Key: <admin-key>" \
     -H "Content-Type: application/json" \
     -d '{"user_id": "app1", "role": "user", "expires_in": 720}'
   ```

3. **更新应用使用新 Key**
4. **审计日志，检查是否有未授权访问**

### Q21: sfsEdgeStore 会访问我的业务数据吗？

**A**: **不会**。sfsEdgeStore 完全运行在本地，不会将任何数据发送到外部服务器，除非您显式配置了数据同步功能。详细信息请参考 [隐私政策](../security/PRIVACY_POLICY.md)。

---

## 故障排除

### Q22: 服务无法启动怎么办？

**A**: 请按以下步骤排查：

1. **查看日志**:
   ```bash
   sudo journalctl -u sfsedgestore -n 200
   ```

2. **前台运行查看错误**:
   ```bash
   sudo -u sfsedgestore /opt/sfsedgestore/sfsedgestore -config /etc/sfsedgestore/config.json
   ```

3. **检查配置**:
   ```bash
   ./sfsedgestore -config /etc/sfsedgestore/config.json -validate
   ```

更多详细步骤请参考 [故障排除手册](./TROUBLESHOOTING.md)。

### Q23: MQTT 连接失败怎么办？

**A**: 常见原因和解决方案：

| 原因 | 检查方法 |
|------|----------|
| Broker 地址错误 | 检查 `MQTTBroker` 配置 |
| 端口被阻止 | `telnet mqtt-broker 1883` |
| 认证失败 | 检查用户名密码 |
| TLS 证书问题 | 验证证书路径和有效期 |

### Q24: 如何获取技术支持？

**A**: 根据您的许可证类型，有多种支持渠道：

- **社区版**: GitHub Issues、社区论坛
- **专业版**: 邮件支持、在线工单
- **企业版**: 7x24 电话支持、专属技术经理

详细信息请参考 [支持文档](../support/SUPPORT.md)。

### Q25: 在哪里可以找到更多文档？

**A**: 完整的文档中心请访问 [docs/README.md](../README.md)，包含：

- 用户手册
- 管理员指南
- 故障排除手册
- API 文档
- 安全策略
- 支持政策
- 等等

---

**文档结束**

如果您有其他问题，请参考 [文档中心](../README.md) 或联系技术支持。
