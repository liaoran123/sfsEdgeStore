package server

import (
	"log"

	"github.com/gorilla/websocket"
)

// WSManager 管理 WebSocket 连接
type WSManager struct {
	clients    map[*websocket.Conn]bool
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

// Run 启动 WebSocket 管理器，处理注册、注销和广播消息
// 消费者 - 处理 register channel 中的连接，添加到 clients 中
// 消费者 - 处理 unregister channel 中的连接，从 clients 中移除并关闭连接
// 消费者 - 处理 broadcast channel 中的消息，发送给所有连接的客户端
// 由于边缘计算场景，大多数情况下只有一个客户端，所以使用单线程处理即可
func (manager *WSManager) Run() {
	for {
		select {
		case conn := <-manager.register: //注册连接
			manager.clients[conn] = true
			log.Printf("WebSocket client connected, total: %d", len(manager.clients))
		case conn := <-manager.unregister: //注销连接
			if _, ok := manager.clients[conn]; ok {
				delete(manager.clients, conn)
				conn.Close()
				log.Printf("WebSocket client disconnected, total: %d", len(manager.clients))
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					// 注销连接 - 发送失败时关闭连接
					conn.Close()                  //关闭连接
					delete(manager.clients, conn) //从 clients 中移除
				}
			}
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
