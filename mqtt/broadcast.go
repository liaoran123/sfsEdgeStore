package mqtt

import (
	"encoding/json"
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
/*
因为 Alias 是一个新类型，它 不会继承 BroadcastMessage 的 MarshalJSON() 方法。
所以当 json.Marshal() 接收到 *Alias 类型时，发现它没有实现 json.Marshaler 接口，
就会使用默认的 JSON 序列化逻辑（逐个字段序列化），从而 避免了递归调用 。
*/
func (m *BroadcastMessage) MarshalJSON() ([]byte, error) {
	type Alias BroadcastMessage //必须是新类型，否则会递归调用 MarshalJSON() 方法，导致栈溢出
	return json.Marshal((*Alias)(m))
}

// PutTo 将消息推入广播通道，通道满时静默丢弃。
// 通过 通道 broadcastLoop 进行生成和消费消息。
// broadcastLoop 传递信息机制：从通道读取消息，序列化后调用外部传入的 broadcast(jsonData) 函数，推送到 WebSocket 客户端。
/*
完整流程：
1. batchWriter.go:113 或 analyze.go:38 → 把 BroadcastMessage 推入 broadcastChan
2. client.go:61 → broadcastLoop 从通道读取消息
3. 序列化后调用外部传入的 broadcast(jsonData) 函数，推送到 WebSocket 客户端
*/
func (m *BroadcastMessage) PutTo(ch chan *BroadcastMessage) {
	select {
	case ch <- m: // ← 尝试将消息推入通道，若通道满则静默丢弃。生产者模式下，消息丢失风险小。
	default:
		PutBroadcastMessage(m)
	}
}

// NewBroadcastMessage 创建并推送消息到通道
func NewBroadcastMessage(ch chan *BroadcastMessage, msgType string, data any) {
	msg := GetBroadcastMessage()
	msg.Type = msgType
	msg.Data = data
	msg.PutTo(ch)
}
