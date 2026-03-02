# sfsDb EdgeX Adapter 部署指南

## 1. 本地部署

### 前提条件
- Go 1.25 或更高版本
- MQTT Broker（如 Mosquitto）
- sfsDb 库

### 构建和运行
1. 克隆仓库：
   ```bash
   git clone https://github.com/your-org/sfsdb-edgex-adapter.git
   cd sfsdb-edgex-adapter
   ```

2. 构建应用：
   ```bash
   go build
   ```

3. 运行应用：
   ```bash
   ./sfsdb-edgex-adapter
   ```

### 配置
可以通过环境变量或配置文件进行配置：

#### 环境变量
- `EDGEX_DB_PATH` - 数据库存储路径
- `EDGEX_MQTT_BROKER` - MQTT broker 地址
- `EDGEX_MQTT_TOPIC` - MQTT 订阅主题
- `EDGEX_CLIENT_ID` - MQTT 客户端 ID

#### 配置文件
创建 `config.json` 文件：
```json
{
  "db_path": "./edgex_data",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsdb-edgex-adapter"
}
```

## 2. Docker 部署

### 前提条件
- Docker
- Docker Compose

### 使用 Docker Compose
1. 克隆仓库：
   ```bash
   git clone https://github.com/your-org/sfsdb-edgex-adapter.git
   cd sfsdb-edgex-adapter
   ```

2. 启动服务：
   ```bash
   docker-compose up -d
   ```

3. 查看日志：
   ```bash
   docker-compose logs -f
   ```

4. 停止服务：
   ```bash
   docker-compose down
   ```

### 环境变量配置
在 `docker-compose.yml` 文件中修改环境变量：
```yaml
environment:
  - EDGEX_MQTT_BROKER=tcp://mosquitto:1883
  - EDGEX_MQTT_TOPIC=edgex/events/core/#
  - EDGEX_CLIENT_ID=sfsdb-edgex-adapter
  - EDGEX_DB_PATH=/app/edgex_data
```

## 3. Kubernetes 部署

### 前提条件
- Kubernetes 集群
- kubectl 命令行工具

### 部署步骤
1. 创建命名空间：
   ```bash
   kubectl create namespace edgex
   ```

2. 应用部署配置：
   ```bash
   kubectl apply -f kubernetes/deployment.yaml -n edgex
   ```

3. 查看部署状态：
   ```bash
   kubectl get pods -n edgex
   ```

4. 暴露服务：
   ```bash
   kubectl apply -f kubernetes/service.yaml -n edgex
   ```

## 4. API 使用

### 健康检查
- **端点**：`GET /health`
- **响应**：
  ```json
  {
    "status": "ok"
  }
  ```

### 数据查询
- **端点**：`GET /api/readings`
- **参数**：
  - `deviceName`：设备名称（可选）
  - `startTime`：开始时间（ISO 8601 格式，可选）
  - `endTime`：结束时间（ISO 8601 格式，可选）
- **响应**：
  ```json
  {
    "count": 1,
    "readings": [
      {
        "id": "reading-uuid",
        "deviceName": "test-device",
        "reading": "temperature",
        "value": 25.5,
        "timestamp": 1677721600,
        "metadata": "{}"
      }
    ]
  }
  ```

### 示例查询
1. 查询所有数据：
   ```bash
   curl http://localhost:8081/api/readings
   ```

2. 按设备名称查询：
   ```bash
   curl http://localhost:8081/api/readings?deviceName=test-device
   ```

3. 按时间范围查询：
   ```bash
   curl "http://localhost:8081/api/readings?startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z"
   ```

4. 组合查询：
   ```bash
   curl "http://localhost:8081/api/readings?deviceName=test-device&startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z"
   ```

## 5. 监控和日志

### 日志
- 适配器日志输出到 stdout
- Docker 部署时可通过 `docker-compose logs` 查看
- Kubernetes 部署时可通过 `kubectl logs` 查看

### 健康检查
- 健康检查端点：`/health`
- 可与 Prometheus、Grafana 等监控工具集成

## 6. 故障排除

### 常见问题

1. **MQTT 连接失败**
   - 确保 MQTT broker 正在运行
   - 检查 broker 地址和端口配置
   - 查看网络连接是否正常

2. **数据库初始化失败**
   - 确保数据库目录可写
   - 检查文件系统权限
   - 验证数据库路径配置

3. **消息处理错误**
   - 验证消息格式是否符合 EdgeX MessageEnvelope 格式
   - 查看日志输出获取详细错误信息

4. **API 访问失败**
   - 确保 HTTP 服务器正在运行
   - 检查端口是否正确配置
   - 验证网络连接

### 日志级别
默认日志级别为 info，可通过修改代码中的日志配置进行调整。

## 7. 版本管理

### 语义化版本
遵循语义化版本规范：
- 主版本号：不兼容的 API 变更
- 次版本号：向后兼容的功能添加
- 补丁版本号：向后兼容的 bug 修复

### 发布流程
1. 代码提交和测试
2. 版本号更新
3. 构建和发布 Docker 镜像
4. 发布 GitHub Release

## 8. 安全最佳实践

### 配置安全
- 避免在配置文件中存储敏感信息
- 使用环境变量或密钥管理服务
- 定期更新配置

### 网络安全
- 限制 API 访问
- 使用 TLS 加密（如果需要）
- 配置适当的网络隔离

### 容器安全
- 使用最小化基础镜像
- 定期更新依赖
- 扫描容器镜像中的漏洞

## 9. 性能优化

### 资源使用
- 调整数据库缓存大小
- 优化 MQTT 连接参数
- 合理设置 HTTP 服务器参数

### 并发处理
- 适当增加并发处理能力
- 优化消息处理逻辑
- 考虑使用连接池

### 存储优化
- 定期清理旧数据
- 合理设置数据库索引
- 考虑使用数据压缩

## 10. 扩展和定制

### 功能扩展
- 添加新的 API 端点
- 支持更多 EdgeX 消息类型
- 集成其他存储后端

### 定制配置
- 根据具体需求修改配置
- 调整性能参数
- 添加自定义监控指标

### 集成其他服务
- 与 EdgeX 其他服务集成
- 与外部监控系统集成
- 与告警系统集成
