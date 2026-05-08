package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WSManager 管理 WebSocket 连接
type WSManager struct {
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewWSManager 创建新的 WebSocket 管理器
func NewWSManager() *WSManager {
	return &WSManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

/*
Broadcast() 负责把消息放进 channel，
Run() 负责从 channel 取出消息并发送给所有客户端。这是典型的 生产者-消费者 模式。
*/
// Run 启动 WebSocket 管理器，处理注册、注销和广播消息
func (manager *WSManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.clientsMu.Lock()
			manager.clients[client] = true
			manager.clientsMu.Unlock()
			log.Printf("WebSocket client connected, total: %d", len(manager.clients))
		case client := <-manager.unregister:
			manager.clientsMu.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				manager.clientsMu.Unlock()
				client.Close()
				log.Printf("WebSocket client disconnected, total: %d", len(manager.clients))
			} else {
				manager.clientsMu.Unlock()
			}
		case message := <-manager.broadcast:
			manager.clientsMu.RLock()
			for client := range manager.clients {
				go func(c *websocket.Conn) {
					if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
						manager.clientsMu.Lock()
						delete(manager.clients, c)
						manager.clientsMu.Unlock()
						c.Close()
					}
				}(client)
			}
			manager.clientsMu.RUnlock()
		}
	}
}

// Broadcast 广播消息给所有客户端
// 生产者 - 外部调用，将消息推入 broadcast channel
func (manager *WSManager) Broadcast(message []byte) {
	select {
	case manager.broadcast <- message:
	default:
	}
}

// Register 注册新的 WebSocket 连接
// 生产者 - 外部调用，将连接推入 register channel
func (manager *WSManager) Register(conn *websocket.Conn) {
	manager.register <- conn
}

// Unregister 注销 WebSocket 连接
// 生产者 - 外部调用，将连接推入 unregister channel
func (manager *WSManager) Unregister(conn *websocket.Conn) {
	manager.unregister <- conn
}
