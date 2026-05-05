package mqtt

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// BroadcastMessage 广播消息结构体
type BroadcastMessage struct {
	Type      string `json:"type"`
	Data      any    `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

// broadcastPool 广播消息对象池，减少 GC 压力
var broadcastPool = sync.Pool{
	New: func() any {
		return &BroadcastMessage{}
	},
}

// GetBroadcastMessage 从对象池获取消息对象，并保证数据干净
func GetBroadcastMessage() *BroadcastMessage {
	msg := broadcastPool.Get().(*BroadcastMessage)
	msg.Type = ""
	msg.Data = nil
	msg.Timestamp = time.Now().UnixNano()
	return msg
}

// PutBroadcastMessage 将消息对象放回对象池
func PutBroadcastMessage(msg *BroadcastMessage) {
	msg.Type = ""
	msg.Data = nil
	msg.Timestamp = 0
	broadcastPool.Put(msg)
}

// MarshalJSON 序列化为 JSON 字节
func (m *BroadcastMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(m)
}

// BroadcastData 广播数据给所有 WebSocket 客户端
func (w *BatchWriter) BroadcastData(dataType string, data any) {
	if w.broadcaster == nil {
		return
	}
	msg := GetBroadcastMessage() // 从对象池获取消息对象，并保证数据干净
	msg.Type = dataType
	msg.Data = data

	defer PutBroadcastMessage(msg) // 放回对象池，确保数据干净

	jsonData, err := msg.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal broadcast data: %v", err)
		return
	}
	// 广播数据到所有 WebSocket 连接的 Web 客户端
	w.broadcaster.Broadcast(jsonData)
}
