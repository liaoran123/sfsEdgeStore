package analyzer

import (
	"log"
	"sync"
	"time"

	"sfsdb-edgex-adapter-enterprise/config"
	timeutil "sfsdb-edgex-adapter-enterprise/time"
)

// Analyzer 轻量级分析引擎
type Analyzer struct {
	config        *config.Config
	memoryPool    *sync.Pool
	isEnabled     bool
	maxMemory     int
	maxTimePerRun time.Duration
}

// NewAnalyzer 创建新的分析引擎
func NewAnalyzer(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config: cfg,
		memoryPool: &sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
		isEnabled:     cfg.EnableAnalyzer,
		maxMemory:     cfg.AnalyzerMaxMemory,
		maxTimePerRun: time.Duration(cfg.AnalyzerMaxTimePerRun) * time.Millisecond,
	}
}

// Analyze 分析数据
func (a *Analyzer) Analyze(data []map[string]interface{}, deviceName, reading string) ([]AnalysisResult, []Alert) {
	if !a.isEnabled {
		return nil, nil
	}

	// 限制分析时间
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if elapsed > a.maxTimePerRun {
			log.Printf("Analyzer execution time exceeded limit: %v", elapsed)
		}
	}()

	var results []AnalysisResult
	var alerts []Alert

	// 滑动窗口分析
	slidingResults := a.analyzeSlidingWindow(data, deviceName, reading)
	results = append(results, slidingResults...)

	// 异常检测
	detectionAlerts := a.detectAnomalies(data, deviceName, reading)
	alerts = append(alerts, detectionAlerts...)

	return results, alerts
}

// analyzeSlidingWindow 滑动窗口分析
func (a *Analyzer) analyzeSlidingWindow(data []map[string]interface{}, deviceName, reading string) []AnalysisResult {
	if len(data) < 2 {
		return nil
	}

	// 创建5分钟滑动窗口，步长1分钟
	windowSize := 5 * time.Minute
	stepSize := 1 * time.Minute
	endTime := time.Now()
	startTime := endTime.Add(-30 * time.Minute)

	slidingWindow := timeutil.NewSlidingWindow(startTime, endTime, windowSize, stepSize)

	// 转换数据格式
	var records []map[string]any
	for _, d := range data {
		record := make(map[string]any)
		for k, v := range d {
			record[k] = v
		}
		records = append(records, record)
	}

	// 计算滑动窗口的平均值
	aggregations, err := timeutil.AggregateByWindow(records, "timestamp", "value", slidingWindow, "avg")
	if err != nil {
		log.Printf("Error in sliding window analysis: %v", err)
		return nil
	}

	var results []AnalysisResult
	for _, agg := range aggregations {
		results = append(results, AnalysisResult{
			DeviceName:   deviceName,
			Reading:      reading,
			AnalysisType: "sliding_window_avg",
			Value:        agg.Value,
			WindowStart:  agg.WindowStart,
			WindowEnd:    agg.WindowEnd,
			Timestamp:    time.Now(),
		})
	}

	return results
}

// detectAnomalies 异常检测
func (a *Analyzer) detectAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert {
	if len(data) < 2 {
		return nil
	}

	var alerts []Alert

	// 阈值检测
	thresholdAlerts := a.detectThresholdAnomalies(data, deviceName, reading)
	alerts = append(alerts, thresholdAlerts...)

	// 趋势检测
	trendAlerts := a.detectTrendAnomalies(data, deviceName, reading)
	alerts = append(alerts, trendAlerts...)

	return alerts
}

// detectThresholdAnomalies 阈值异常检测
func (a *Analyzer) detectThresholdAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert {
	var alerts []Alert

	// 获取阈值配置
	thresholds, ok := a.config.AnalyzerThresholds[reading]
	if !ok || (thresholds.Min == 0 && thresholds.Max == 0) {
		return alerts
	}

	for _, d := range data {
		value, ok := d["value"].(float64)
		if !ok {
			continue
		}

		timestamp, ok := d["timestamp"].(time.Time)
		if !ok {
			timestamp = time.Now()
		}

		// 检查上限
		if thresholds.Max > 0 && value > thresholds.Max {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "threshold_exceeded",
				Message:    "Value exceeded maximum threshold",
				Value:      value,
				Threshold:  thresholds.Max,
				Timestamp:  timestamp,
				Severity:   "high",
			})
		}

		// 检查下限
		if thresholds.Min > 0 && value < thresholds.Min {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "threshold_below",
				Message:    "Value below minimum threshold",
				Value:      value,
				Threshold:  thresholds.Min,
				Timestamp:  timestamp,
				Severity:   "medium",
			})
		}
	}

	return alerts
}

// detectTrendAnomalies 趋势异常检测
func (a *Analyzer) detectTrendAnomalies(data []map[string]interface{}, deviceName, reading string) []Alert {
	if len(data) < 3 {
		return nil
	}

	var alerts []Alert

	// 计算趋势
	values := make([]float64, 0, len(data))
	timestamps := make([]time.Time, 0, len(data))

	for _, d := range data {
		value, ok := d["value"].(float64)
		if !ok {
			continue
		}

		timestamp, ok := d["timestamp"].(time.Time)
		if !ok {
			timestamp = time.Now()
		}

		values = append(values, value)
		timestamps = append(timestamps, timestamp)
	}

	if len(values) < 3 {
		return alerts
	}

	// 简单趋势检测：检查最近三个值的变化
	for i := 2; i < len(values); i++ {
		// 计算变化率
		change1 := (values[i-1] - values[i-2]) / values[i-2]
		change2 := (values[i] - values[i-1]) / values[i-1]

		// 检查异常趋势（例如：连续快速增长）
		if change1 > 0.1 && change2 > 0.1 {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "trend_anomaly",
				Message:    "Rapid continuous increase detected",
				Value:      values[i],
				Threshold:  0,
				Timestamp:  timestamps[i],
				Severity:   "medium",
			})
		}

		// 检查异常下降
		if change1 < -0.1 && change2 < -0.1 {
			alerts = append(alerts, Alert{
				DeviceName: deviceName,
				Reading:    reading,
				AlertType:  "trend_anomaly",
				Message:    "Rapid continuous decrease detected",
				Value:      values[i],
				Threshold:  0,
				Timestamp:  timestamps[i],
				Severity:   "medium",
			})
		}
	}

	return alerts
}

// Enable 启用分析引擎
func (a *Analyzer) Enable() {
	a.isEnabled = true
}

// Disable 禁用分析引擎
func (a *Analyzer) Disable() {
	a.isEnabled = false
}

// IsEnabled 检查分析引擎是否启用
func (a *Analyzer) IsEnabled() bool {
	return a.isEnabled
}
