package mqtt

// groupKey 分析分组键：设备名 + 读数类型
type groupKey struct {
	device  string
	reading string
}

/*
analyzeData 数据分析
- 按 deviceName + reading 二维分组
- 分析引擎需要按 单一设备 + 单一读数 进行分析（如阈值检测、趋势分析）
- 告警消息直接推入广播通道
*/
func (w *BatchWriter) analyzeData(records []*map[string]any) {
	groups := make(map[groupKey][]map[string]any, len(records))

	for _, record := range records {
		r := *record
		deviceName, ok := r["deviceName"].(string)
		if !ok || deviceName == "" {
			continue
		}
		readingName, ok := r["reading"].(string)
		if !ok || readingName == "" {
			continue
		}
		key := groupKey{deviceName, readingName}
		groups[key] = append(groups[key], r)
	}

	for key, data := range groups {
		alerts := w.analyzer.Analyze(data, key.device, key.reading)
		if len(alerts) > 0 {
			for _, alert := range alerts {
				w.monitor.RecordError(alert.AlertType, alert.Message)
			}
			NewBroadcastMessage(w.broadcastChan, "alerts", map[string]any{
				"deviceName": key.device,
				"alerts":     alerts,
			})
		}
	}
}
