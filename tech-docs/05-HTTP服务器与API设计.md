# HTTP 服务器与 API 设计

## 概述

sfsEdgeStore 提供了 RESTful HTTP API，用于数据查询、管理操作和系统监控。

## Server 结构

```go
// server/server.go:29-36
type Server struct {
	Table           *engine.Table
	Config          *config.Config
	Monitor         *monitor.Monitor
	RetentionMgr    *retention.RetentionManager
	AlertNotifier   *alert.Notifier
	ResourceMonitor *resource.ResourceMonitor
}
```

## 创建服务器

### NewServer 函数

```go
// server/server.go:39-48
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor, retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier, resourceMonitor *resource.ResourceMonitor) *Server {
	return &Server{
		Table:           table,
		Config:          cfg,
		Monitor:         monitor,
		RetentionMgr:    retentionMgr,
		AlertNotifier:   alertNotifier,
		ResourceMonitor: resourceMonitor,
	}
}
```

## 启动服务器

### Start 函数

```go
// server/server.go:52-79
func (s *Server) Start() error {
	// 注册路由
	s.registerRoutes()

	// 在后台启动HTTP服务器
	go func() {
		port := s.Config.HTTPPort
		if port == "" {
			port = "8081" // 默认端口
		}

		if s.Config.HTTPUseTLS && s.Config.HTTPCert != "" && s.Config.HTTPKey != "" {
			// 使用 HTTPS
			log.Printf("Starting HTTPS server for health checks on port %s", port)
			if err := http.ListenAndServeTLS(":"+port, s.Config.HTTPCert, s.Config.HTTPKey, nil); err != nil {
				log.Printf("HTTPS server error: %v", err)
			}
		} else {
			// 使用 HTTP
			log.Printf("Starting HTTP server for health checks on port %s", port)
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}
	}()

	return nil
}
```

## 中间件

### DeviceNameMiddleware

```go
// server/server.go:82-97
func DeviceNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取并格式化deviceName参数
		deviceName := r.URL.Query().Get("deviceName")
		if deviceName != "" {
			formattedDeviceName := common.FormatDeviceName(deviceName)
			// 重写URL参数
			url := *r.URL
			q := url.Query()
			q.Set("deviceName", formattedDeviceName)
			url.RawQuery = q.Encode()
			*r.URL = url
		}
		next(w, r)
	}
}
```

## API 端点

### 健康检查

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/healthz` | GET | 健康检查 |
| `/ready` | GET | 就绪检查 |

### Web界面

| 端点 | 方法 | 说明 |
|------|------|------|
| `/` | GET | Web界面首页 |
| `/dashboard` | GET | Web仪表盘 |
| `/static/` | GET | 静态文件 |

### 许可证信息

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/license` | GET | 许可证信息 |

### 数据查询

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/readings` | GET | 查询读数 | read |

### 数据备份/恢复

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/backup` | POST | 备份数据 | backup |
| `/api/restore` | POST | 恢复数据 | restore |

### 测试端点

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/test-edgex` | POST | 测试EdgeX消息 | write |

### 认证管理

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/auth/create-key` | POST | 创建API Key | - |
| `/api/auth/list-keys` | GET | 列出API Keys | admin |
| `/api/auth/revoke-key` | POST | 撤销API Key | admin |

### 加密管理

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/encryption/rotate-key` | POST | 轮换密钥 | admin |
| `/api/encryption/status` | GET | 加密状态 | admin |

### 数据导出/导入

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/export/csv` | GET | 导出CSV | backup |
| `/api/export/json` | GET | 导出JSON | backup |
| `/api/export/sql` | GET | 导出SQL | backup |
| `/api/import/csv` | POST | 导入CSV | restore |
| `/api/import/json` | POST | 导入JSON | restore |
| `/api/data/export` | GET | 参数化导出 | backup |

### 保留策略

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/retention/status` | GET | 保留策略状态 | admin |
| `/api/retention/cleanup` | POST | 手动清理 | admin |

### 告警通知

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/alerts/notifier/status` | GET | 通知器状态 | admin |
| `/api/alerts/test` | POST | 测试告警 | admin |

### 配置管理

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/config/get` | GET | 获取配置 | admin |
| `/api/config/update` | POST | 更新配置 | admin |
| `/api/config/reload` | POST | 重新加载配置 | admin |

### 资源监控

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/resources/status` | GET | 资源状态 | admin |

## API 示例

### 查询读数

```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/readings?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-02T00:00:00Z"
```

### 创建 API Key

```bash
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "role": "user",
    "expires_in": 8760
  }'
```

### 备份数据

```bash
curl -X POST -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/backup?path=./backups"
```

### 健康检查

```bash
curl http://localhost:8081/health
```

### 测试 EdgeX 消息

```bash
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:8081/api/test-edgex
```

### 导出数据

```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/data/export?format=json&deviceName=Device001"
```

## 认证机制

sfsEdgeStore 使用基于 API Key 的认证机制，通过 `X-API-Key` 请求头传递认证信息。认证中间件会验证 API Key 的有效性和权限。

### 权限控制

每个 API 端点都有对应的权限要求，通过 `PermissionMiddleware` 进行控制：
- `read` - 读取数据权限
- `write` - 写入数据权限
- `admin` - 管理权限
- `backup` - 备份权限
- `restore` - 恢复权限

## 错误处理

API 端点返回标准的 HTTP 状态码：
- `200 OK` - 成功
- `400 Bad Request` - 请求参数错误
- `401 Unauthorized` - 认证失败
- `403 Forbidden` - 权限不足
- `405 Method Not Allowed` - 方法不允许
- `500 Internal Server Error` - 服务器内部错误
- `503 Service Unavailable` - 服务不可用

## 安全建议

1. **使用 HTTPS** - 在生产环境中启用 HTTPS
2. **定期轮换 API Key** - 避免长期使用同一个 API Key
3. **使用强密码** - 生成安全的 API Key
4. **限制权限** - 根据实际需求分配最小权限
5. **监控 API 访问** - 跟踪异常访问模式

## 性能优化

1. **批量操作** - 对于数据导入导出，使用批量操作减少网络开销
2. **缓存** - 对于频繁访问的数据，考虑使用缓存
3. **分页** - 对于大量数据查询，实现分页机制
4. **异步处理** - 对于耗时操作，使用异步处理

## 部署建议

1. **端口配置** - 根据环境需求配置合适的 HTTP 端口
2. **TLS 证书** - 配置有效的 TLS 证书
3. **防火墙** - 配置防火墙规则，只允许必要的端口访问
4. **负载均衡** - 在高流量场景下使用负载均衡
5. **健康检查** - 配置健康检查端点，用于监控系统状态