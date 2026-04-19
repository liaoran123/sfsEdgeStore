# HTTP 服务器与 API 设计

## 概述

sfsEdgeStore 提供了 RESTful HTTP API，用于数据查询、管理操作和系统监控。

## Server 结构

```go
// server/server.go:28-36
type Server struct {
	Table           *engine.Table
	Config          *config.Config
	Monitor         *monitor.Monitor
	RetentionMgr    *retention.RetentionManager
	AlertNotifier   *alert.Notifier
	SyncManager     *sync.SyncManager
	ResourceMonitor *resource.ResourceMonitor
}
```

## 创建服务器

### NewServer 函数

```go
// server/server.go:39-49
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor, retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier, syncManager *sync.SyncManager, resourceMonitor *resource.ResourceMonitor) *Server {
	return &Server{
		Table:           table,
		Config:          cfg,
		Monitor:         monitor,
		RetentionMgr:    retentionMgr,
		AlertNotifier:   alertNotifier,
		SyncManager:     syncManager,
		ResourceMonitor: resourceMonitor,
	}
}
```

## 启动服务器

### Start 函数

```go
// server/server.go:53-80
func (s *Server) Start() error {
	s.registerRoutes()

	go func() {
		port := s.Config.HTTPPort
		if port == "" {
			port = "8081"
		}

		if s.Config.HTTPUseTLS && s.Config.HTTPCert != "" && s.Config.HTTPKey != "" {
			log.Printf("Starting HTTPS server for health checks on port %s", port)
			if err := http.ListenAndServeTLS(":"+port, s.Config.HTTPCert, s.Config.HTTPKey, nil); err != nil {
				log.Printf("HTTPS server error: %v", err)
			}
		} else {
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
// server/server.go:83-98
func DeviceNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceName := r.URL.Query().Get("deviceName")
		if deviceName != "" {
			formattedDeviceName := common.FormatDeviceName(deviceName)
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

### 数据查询

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/readings` | GET | 查询读数 | read |

### 备份恢复

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/backup` | POST | 备份数据 | backup |
| `/api/restore` | POST | 恢复数据 | restore |

### 数据导出/导入

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/export/csv` | GET | 导出 CSV | backup |
| `/api/export/json` | GET | 导出 JSON | backup |
| `/api/export/sql` | GET | 导出 SQL | backup |
| `/api/import/csv` | POST | 导入 CSV | restore |
| `/api/import/json` | POST | 导入 JSON | restore |
| `/api/data/export` | GET | 参数化导出 | backup |

### 认证管理

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/auth/create-key` | POST | 创建 API Key | - |
| `/api/auth/list-keys` | GET | 列出 API Keys | admin |
| `/api/auth/revoke-key` | POST | 撤销 API Key | admin |

### 加密管理

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/encryption/rotate-key` | POST | 轮换密钥 | admin |
| `/api/encryption/status` | GET | 加密状态 | admin |

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

### 数据同步

| 端点 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/sync/status` | GET | 同步状态 | admin |
| `/api/sync/start` | POST | 启动同步 | admin |
| `/api/sync/database` | POST | 从数据库同步 | admin |

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
