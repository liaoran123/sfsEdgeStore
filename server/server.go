package server

import (
	"log"
	"net/http"

	"sfsEdgeStore/alert"
	"sfsEdgeStore/config"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/resource"
	"sfsEdgeStore/retention"

	"github.com/liaoran123/sfsDb/engine"
)

// Server 结构体
type Server struct {
	Table             *engine.Table
	Config            *config.Config
	Monitor           *monitor.Monitor
	RetentionMgr      *retention.RetentionManager
	AlertNotifier     *alert.Notifier
	ResourceMonitor   *resource.ResourceMonitor
	wsManager         *WSManager
	deviceStatusCache *deviceStatusCache
}

// NewServer 创建一个新的服务器实例
func NewServer(table *engine.Table, cfg *config.Config, monitor *monitor.Monitor, retentionMgr *retention.RetentionManager, alertNotifier *alert.Notifier, resourceMonitor *resource.ResourceMonitor) *Server {
	wsManager := NewWSManager()
	go wsManager.Run()

	return &Server{
		Table:             table,
		Config:            cfg,
		Monitor:           monitor,
		RetentionMgr:      retentionMgr,
		AlertNotifier:     alertNotifier,
		ResourceMonitor:   resourceMonitor,
		wsManager:         wsManager,
		deviceStatusCache: newDeviceStatusCache(),
	}
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	s.registerRoutes()

	go func() {
		port := s.Config.HTTPPort
		if port == "" {
			port = "8081"
		}
		log.Printf("Starting HTTP server for health checks on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// GetWSManager 获取 WebSocket 管理器
func (s *Server) GetWSManager() *WSManager {
	return s.wsManager
}

// Broadcast 实现 broadcast.Broadcaster 接口
func (s *Server) Broadcast(message []byte) {
	if s.wsManager != nil {
		s.wsManager.Broadcast(message)
	}
}
