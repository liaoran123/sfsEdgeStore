package broadcast

// Broadcaster 定义广播功能接口
type Broadcaster interface {
	// Broadcast 广播消息给所有客户端
	Broadcast(message []byte)
}
