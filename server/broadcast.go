package server

import (
	"sfsEdgeStore/mqtt"
)

// StartBroadcast 启动广播监听，直接监听读取 MQTT 的广播通道，将消息推送到 WebSocket通道
func (s *Server) StartBroadcast(ch <-chan *mqtt.BroadcastMessage) {
	go s.broadcastLoop(ch) //这是一个 单通道消费 + 顺序广播 模型，不适用 Pool 模型，因为广播消息的顺序是重要的
}

// broadcastLoop 直接读取广播通道并推送到 WebSocket
func (s *Server) broadcastLoop(ch <-chan *mqtt.BroadcastMessage) {
	for msg := range ch {
		jsonData, err := msg.MarshalJSON()
		if err != nil {
			mqtt.PutBroadcastMessage(msg)
			continue
		}
		s.Broadcast(jsonData)
		mqtt.PutBroadcastMessage(msg)
	}
}
