package server

import (
	"sfsEdgeStore/mqtt"
)

// StartBroadcast 启动广播监听，直接读取 MQTT 的广播通道
func (s *Server) StartBroadcast(ch <-chan *mqtt.BroadcastMessage) {
	go s.broadcastLoop(ch)
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
