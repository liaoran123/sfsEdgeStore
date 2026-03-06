package simulator

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/common"
	"sfsEdgeStore/database"
	"sfsEdgeStore/edgex"
	"sfsEdgeStore/monitor"
)

type Simulator struct {
	monitor    *monitor.Monitor
	analyzer   *analyzer.Analyzer
	config     *SimulatorConfig
	stopChan   chan struct{}
	running    bool
}

type SimulatorConfig struct {
	Enabled     bool
	Devices     []DeviceConfig
	IntervalMin time.Duration
	IntervalMax time.Duration
	Topic       string
}

type DeviceConfig struct {
	Name        string
	ProfileName string
	Sensors     []SensorConfig
}

type SensorConfig struct {
	ResourceName string
	ValueType    string
	BaseType     string
	MinValue     float64
	MaxValue     float64
}

func NewSimulator(monitor *monitor.Monitor, analyzer *analyzer.Analyzer, config *SimulatorConfig) *Simulator {
	return &Simulator{
		monitor:  monitor,
		analyzer: analyzer,
		config:   config,
		stopChan: make(chan struct{}),
		running:  false,
	}
}

func (s *Simulator) Start() error {
	if !s.config.Enabled {
		log.Println("Simulator disabled, skipping start")
		return nil
	}

	if s.running {
		return fmt.Errorf("simulator already running")
	}

	s.running = true
	go s.runLoop()
	log.Printf("Simulator started with %d devices", len(s.config.Devices))
	return nil
}

func (s *Simulator) Stop() {
	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
	log.Println("Simulator stopped")
}

func (s *Simulator) runLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			for _, device := range s.config.Devices {
				s.publishDeviceData(device)
			}

			interval := s.config.IntervalMin + time.Duration(rand.Int63n(int64(s.config.IntervalMax-s.config.IntervalMin+1)))
			time.Sleep(interval)
		}
	}
}

func (s *Simulator) publishDeviceData(device DeviceConfig) {
	origin := time.Now().UnixNano()

	readings := make([]edgex.EdgeXReading, 0, len(device.Sensors))
	for _, sensor := range device.Sensors {
		value := s.generateSensorValue(sensor)
		readingID := fmt.Sprintf("reading-%s-%d", device.Name, time.Now().UnixNano())

		readings = append(readings, edgex.EdgeXReading{
			ID:           readingID,
			ResourceName: sensor.ResourceName,
			Value:        value,
			ValueType:    sensor.ValueType,
			BaseType:     sensor.BaseType,
			Origin:       origin,
			ProfileName:  device.ProfileName,
			DeviceName:   device.Name,
		})
	}

	event := &edgex.EdgeXEvent{
		ID:          fmt.Sprintf("event-%s-%d", device.Name, time.Now().UnixNano()),
		DeviceName:  device.Name,
		Readings:    readings,
		Origin:      origin,
		ProfileName: device.ProfileName,
		SourceName:  device.Name,
	}

	var records []*map[string]any

	for _, reading := range event.Readings {
		metadataStr := ""
		if reading.Metadata != nil {
			metadataStr = string(reading.Metadata)
		}

		value := common.ParseValue(reading.Value)

		data := map[string]any{
			"id":         reading.ID,
			"deviceName": event.DeviceName,
			"reading":    reading.ResourceName,
			"value":      value,
			"valueType":  reading.ValueType,
			"baseType":   reading.BaseType,
			"timestamp":  reading.Origin,
			"metadata":   metadataStr,
		}

		records = append(records, &data)
	}

	if len(records) > 0 {
		if s.monitor != nil {
			s.monitor.IncrementMQTTMessagesReceived()
			s.monitor.IncrementDatabaseOperations()
		}

		err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
		if err != nil {
			log.Printf("Failed to batch store data after retries: %v", err)
			if s.monitor != nil {
				s.monitor.RecordError("database_error", err.Error())
			}
		} else {
			log.Printf("Batch stored %d readings from %s", len(records), event.DeviceName)
			if s.monitor != nil {
				s.monitor.IncrementMQTTMessagesProcessed()
			}

			if s.analyzer != nil && s.analyzer.IsEnabled() {
				readingDataMap := make(map[string][]map[string]interface{})
				for _, record := range records {
					readingName, ok := (*record)["reading"].(string)
					if !ok {
						continue
					}
					readingDataMap[readingName] = append(readingDataMap[readingName], *record)
				}

				for readingName, analysisData := range readingDataMap {
					results, alerts := s.analyzer.Analyze(analysisData, event.DeviceName, readingName)
					if len(results) > 0 {
						log.Printf("Analysis completed for %s: %d results", readingName, len(results))
					}
					if len(alerts) > 0 {
						log.Printf("Detected %d alerts for %s", len(alerts), readingName)
						for _, alert := range alerts {
							log.Printf("Alert: %s - %s - %s", alert.Severity, alert.Message, alert.Reading)
							if s.monitor != nil {
								s.monitor.RecordError(alert.AlertType, alert.Message)
							}
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) generateSensorValue(sensor SensorConfig) string {
	switch sensor.ValueType {
	case "Int32", "Int64", "Int":
		value := int64(sensor.MinValue) + rand.Int63n(int64(sensor.MaxValue-sensor.MinValue+1))
		return fmt.Sprintf("%d", value)
	case "Float32", "Float64", "Float":
		value := sensor.MinValue + rand.Float64()*(sensor.MaxValue-sensor.MinValue)
		return fmt.Sprintf("%.2f", value)
	case "Bool":
		return fmt.Sprintf("%t", rand.Float32() > 0.5)
	default:
		value := sensor.MinValue + rand.Float64()*(sensor.MaxValue-sensor.MinValue)
		return fmt.Sprintf("%.2f", value)
	}
}

func DefaultSimulatorConfig() *SimulatorConfig {
	devices := make([]DeviceConfig, 0)

	for i := 1; i <= 10; i++ {
		devices = append(devices, DeviceConfig{
			Name:        fmt.Sprintf("temperature-sensor-%03d", i),
			ProfileName: "temperature-sensor",
			Sensors: []SensorConfig{
				{
					ResourceName: "temperature",
					ValueType:    "Int32",
					BaseType:     "Int32",
					MinValue:     18.0,
					MaxValue:     35.0,
				},
				{
					ResourceName: "humidity",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     35.0,
					MaxValue:     85.0,
				},
			},
		})
	}

	for i := 1; i <= 5; i++ {
		devices = append(devices, DeviceConfig{
			Name:        fmt.Sprintf("power-meter-%03d", i),
			ProfileName: "power-meter",
			Sensors: []SensorConfig{
				{
					ResourceName: "voltage",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     208.0,
					MaxValue:     242.0,
				},
				{
					ResourceName: "current",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     0.3,
					MaxValue:     15.0,
				},
				{
					ResourceName: "power",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     50.0,
					MaxValue:     3500.0,
				},
				{
					ResourceName: "energy",
					ValueType:    "Float64",
					BaseType:     "Float64",
					MinValue:     0.0,
					MaxValue:     10000.0,
				},
			},
		})
	}

	for i := 1; i <= 8; i++ {
		devices = append(devices, DeviceConfig{
			Name:        fmt.Sprintf("pressure-sensor-%03d", i),
			ProfileName: "pressure-sensor",
			Sensors: []SensorConfig{
				{
					ResourceName: "pressure",
					ValueType:    "Float64",
					BaseType:     "Float64",
					MinValue:     940.0,
					MaxValue:     1060.0,
				},
				{
					ResourceName: "temperature",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     -20.0,
					MaxValue:     60.0,
				},
			},
		})
	}

	for i := 1; i <= 6; i++ {
		devices = append(devices, DeviceConfig{
			Name:        fmt.Sprintf("vibration-sensor-%03d", i),
			ProfileName: "vibration-sensor",
			Sensors: []SensorConfig{
				{
					ResourceName: "vibration_x",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     -10.0,
					MaxValue:     10.0,
				},
				{
					ResourceName: "vibration_y",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     -10.0,
					MaxValue:     10.0,
				},
				{
					ResourceName: "vibration_z",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     -10.0,
					MaxValue:     10.0,
				},
				{
					ResourceName: "amplitude",
					ValueType:    "Float64",
					BaseType:     "Float64",
					MinValue:     0.0,
					MaxValue:     20.0,
				},
			},
		})
	}

	for i := 1; i <= 4; i++ {
		devices = append(devices, DeviceConfig{
			Name:        fmt.Sprintf("flow-meter-%03d", i),
			ProfileName: "flow-meter",
			Sensors: []SensorConfig{
				{
					ResourceName: "flow_rate",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     0.0,
					MaxValue:     100.0,
				},
				{
					ResourceName: "total_volume",
					ValueType:    "Float64",
					BaseType:     "Float64",
					MinValue:     0.0,
					MaxValue:     100000.0,
				},
				{
					ResourceName: "temperature",
					ValueType:    "Float32",
					BaseType:     "Float32",
					MinValue:     5.0,
					MaxValue:     95.0,
				},
			},
		})
	}

	return &SimulatorConfig{
		Enabled:     false,
		IntervalMin: 500 * time.Millisecond,
		IntervalMax: 2 * time.Second,
		Topic:       "",
		Devices:     devices,
	}
}
