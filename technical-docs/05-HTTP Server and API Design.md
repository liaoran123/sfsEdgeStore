# HTTP Server and API Design

## Overview

sfsEdgeStore provides RESTful HTTP API for data query, management operations, and system monitoring.

## Server Structure

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

## Create Server

### NewServer Function

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

## Start Server

### Start Function

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

## Middleware

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

## API Endpoints

### Health Check

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/healthz` | GET | Health check |
| `/ready` | GET | Readiness check |

### Data Query

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/readings` | GET | Query readings | read |

### Backup and Restore

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/backup` | POST | Backup data | backup |
| `/api/restore` | POST | Restore data | restore |

### Data Export/Import

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/export/csv` | GET | Export CSV | backup |
| `/api/export/json` | GET | Export JSON | backup |
| `/api/export/sql` | GET | Export SQL | backup |
| `/api/import/csv` | POST | Import CSV | restore |
| `/api/import/json` | POST | Import JSON | restore |
| `/api/data/export` | GET | Parameterized export | backup |

### Authentication Management

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/auth/create-key` | POST | Create API Key | - |
| `/api/auth/list-keys` | GET | List API Keys | admin |
| `/api/auth/revoke-key` | POST | Revoke API Key | admin |

### Encryption Management

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/encryption/rotate-key` | POST | Rotate encryption key | admin |
| `/api/encryption/status` | GET | Encryption status | admin |

### Retention Policy

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/retention/status` | GET | Retention policy status | admin |
| `/api/retention/cleanup` | POST | Manual cleanup | admin |

### Alert Notification

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/alerts/notifier/status` | GET | Notifier status | admin |
| `/api/alerts/test` | POST | Test alert | admin |

### Data Synchronization

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/sync/status` | GET | Sync status | admin |
| `/api/sync/start` | POST | Start sync | admin |
| `/api/sync/database` | POST | Sync from database | admin |

### Configuration Management

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/config/get` | GET | Get configuration | admin |
| `/api/config/update` | POST | Update configuration | admin |
| `/api/config/reload` | POST | Reload configuration | admin |

### Resource Monitoring

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/api/resources/status` | GET | Resource status | admin |

## API Examples

### Query Readings

```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/readings?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-02T00:00:00Z"
```

### Create API Key

```bash
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "role": "user",
    "expires_in": 8760
  }'
```

### Backup Data

```bash
curl -X POST -H "X-API-Key: your-api-key" \
  "http://localhost:8081/api/backup?path=./backups"
```

### Health Check

```bash
curl http://localhost:8081/health
```