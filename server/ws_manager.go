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
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 启动 WebSocket 管理器
func (manager *WSManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.clients[client] = true
			log.Printf("WebSocket client connected, total: %d", len(manager.clients))
		case client := <-manager.unregister:
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
				log.Printf("WebSocket client disconnected, total: %d", len(manager.clients))
			}
		case message := <-manager.broadcast:
			for client := range manager.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(manager.clients, client)
				}
			}
		}
	}
}

// Broadcast 广播消息给所有客户端
func (manager *WSManager) Broadcast(message []byte) {
	manager.broadcast <- message
}
