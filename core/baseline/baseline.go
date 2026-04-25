package baseline

import (
	"math"
	"sfsEdgeStore/core/database"
	"time"

	"github.com/liaoran123/sfsDb/record"
)

// Baseline 基线结构
type Baseline struct {
	DeviceName     string
	ReadingName    string
	Average        float64
	StdDev         float64
	MinValue       float64
	MaxValue       float64
	SampleCount    int
	LastUpdated    time.Time
	LearningPeriod int // 学习期（天）
	Enabled        bool
}

// Manager 基线管理器
type Manager struct {
	baselines      map[string]Baseline
	learningPeriod int
	enabled        bool
}

// NewManager 创建基线管理器
func NewManager(learningPeriod int, enabled bool) *Manager {
	return &Manager{
		baselines:      make(map[string]Baseline),
		learningPeriod: learningPeriod,
		enabled:        enabled,
	}
}

// CalculateBaseline 计算设备读数的基线
func (m *Manager) CalculateBaseline(deviceName, readingName string) (Baseline, error) {
	if !m.enabled {
		return Baseline{}, nil
	}

	// 生成基线键
	key := deviceName + ":" + readingName

	// 检查是否已有基线
	if baseline, exists := m.baselines[key]; exists {
		// 检查基线是否过期
		if time.Since(baseline.LastUpdated) < 24*time.Hour {
			return baseline, nil
		}
	}

	// 计算时间范围
	// 不需要时间范围，因为 QueryRecords 会处理所有数据

	// 查询历史数据
	records, err := database.QueryRecords(
		database.Table,
		deviceName,
		readingName,
		"",
		false,
	)
	if err != nil {
		return Baseline{}, err
	}

	// 计算基线
	baseline, err := m.calculateStatistics(records, deviceName, readingName)
	if err != nil {
		return Baseline{}, err
	}

	// 保存基线
	m.baselines[key] = baseline

	return baseline, nil
}

// calculateStatistics 计算统计数据
func (m *Manager) calculateStatistics(records record.Records, deviceName, readingName string) (Baseline, error) {
	if len(records) == 0 {
		return Baseline{
			DeviceName:     deviceName,
			ReadingName:    readingName,
			Average:        0,
			StdDev:         0,
			MinValue:       0,
			MaxValue:       0,
			SampleCount:    0,
			LastUpdated:    time.Now(),
			LearningPeriod: m.learningPeriod,
			Enabled:        m.enabled,
		}, nil
	}

	// 计算平均值
	var sum float64
	var minValue float64 = math.MaxFloat64
	var maxValue float64 = -math.MaxFloat64

	for _, record := range records {
		if value, ok := record["value"].(float64); ok {
			sum += value
			if value < minValue {
				minValue = value
			}
			if value > maxValue {
				maxValue = value
			}
		}
	}

	average := sum / float64(len(records))

	// 计算标准差
	var variance float64
	for _, record := range records {
		if value, ok := record["value"].(float64); ok {
			variance += math.Pow(value-average, 2)
		}
	}

	stdDev := math.Sqrt(variance / float64(len(records)))

	return Baseline{
		DeviceName:     deviceName,
		ReadingName:    readingName,
		Average:        average,
		StdDev:         stdDev,
		MinValue:       minValue,
		MaxValue:       maxValue,
		SampleCount:    len(records),
		LastUpdated:    time.Now(),
		LearningPeriod: m.learningPeriod,
		Enabled:        m.enabled,
	}, nil
}

// GetDynamicThreshold 获取动态阈值
func (m *Manager) GetDynamicThreshold(deviceName, readingName string, stdMultiplier float64) (float64, float64, error) {
	baseline, err := m.CalculateBaseline(deviceName, readingName)
	if err != nil {
		return 0, 0, err
	}

	// 计算动态阈值
	upperThreshold := baseline.Average + (stdMultiplier * baseline.StdDev)
	lowerThreshold := baseline.Average - (stdMultiplier * baseline.StdDev)

	return lowerThreshold, upperThreshold, nil
}

// CheckAnomaly 检查是否异常
func (m *Manager) CheckAnomaly(deviceName, readingName string, value float64, stdMultiplier float64) (bool, error) {
	lowerThreshold, upperThreshold, err := m.GetDynamicThreshold(deviceName, readingName, stdMultiplier)
	if err != nil {
		return false, err
	}

	// 检查是否超出阈值
	return value < lowerThreshold || value > upperThreshold, nil
}

// GetBaseline 获取指定设备和读数的基线
func (m *Manager) GetBaseline(deviceName, readingName string) (Baseline, bool) {
	key := deviceName + ":" + readingName
	baseline, exists := m.baselines[key]
	return baseline, exists
}

// UpdateBaseline 更新基线
func (m *Manager) UpdateBaseline(baseline Baseline) {
	key := baseline.DeviceName + ":" + baseline.ReadingName
	m.baselines[key] = baseline
}

// ListBaselines 列出所有基线
func (m *Manager) ListBaselines() map[string]Baseline {
	return m.baselines
}
