# MQTT Client Implementation

## Overview

The MQTT client is the core module for sfsEdgeStore to communicate with EdgeX Foundry, responsible for subscribing to EdgeX event messages, parsing data, and storing it in the local database.

## Core Structure

### Client Struct

```go
// mqtt/client.go:27-37
type Client struct {
	client        mqtt.Client
	config        *config.Config
	dataQueue     *queue.Queue
	monitor       *monitor.Monitor
	analyzer      *analyzer.Analyzer
	batchMessages []map[string]interface{}
	batchSize     int
	batchInterval time.Duration
	lastBatchTime time.Time
}
```

**Field Description:**
- `client`: Underlying MQTT client (from paho.mqtt.golang)
- `config`: Configuration management
- `dataQueue`: Data queue for fault recovery
- `monitor`: Monitoring integration
- `analyzer`: Data analysis integration
- `batchMessages`: Batch message buffer
- `batchSize/batchInterval`: Batch control parameters

## Creating Client

### NewClient Function

```go
// mqtt/client.go:40-136
func NewClient(cfg *config.Config, dataQueue *queue.Queue, monitor *monitor.Monitor, analyzer *analyzer.Analyzer) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.ClientID)
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Minute * 5)

	// Set will message
	willTopic := cfg.MQTTTopic + "/status"
	willMessage := map[string]interface{}{
		"status":    "offline",
		"clientId":  cfg.ClientID,
		"timestamp": time.Now().UnixNano(),
	}
	willPayload, _ := json.Marshal(willMessage)
	opts.SetWill(willTopic, string(willPayload), 1, false)

	// Add TLS support
	if cfg.MQTTUseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		if cfg.MQTTCACert != "" {
			caCert, err := os.ReadFile(cfg.MQTTCACert)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %v", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		if cfg.MQTTClientCert != "" && cfg.MQTTClientKey != "" {
			cert, err := tls.LoadX509KeyPair(cfg.MQTTClientCert, cfg.MQTTClientKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client cert: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		opts.SetTLSConfig(tlsConfig)
	}

	client := &Client{
		config:        cfg,
		dataQueue:     dataQueue,
		monitor:       monitor,
		analyzer:      analyzer,
		batchMessages: make([]map[string]interface{}, 0),
		batchSize:     100,
		batchInterval: 5 * time.Second,
		lastBatchTime: time.Now(),
	}

	// Set connection handler
	opts.SetOnConnectHandler(func(mqttClient mqtt.Client) {
		log.Println("MQTT broker connected")
		onlineTopic := cfg.MQTTTopic + "/status"
		onlineMessage := map[string]interface{}{
			"status":    "online",
			"clientId":  cfg.ClientID,
			"timestamp": time.Now().UnixNano(),
		}
		onlinePayload, _ := json.Marshal(onlineMessage)
		token := mqttClient.Publish(onlineTopic, 1, false, onlinePayload)
		token.Wait()
		token = mqttClient.Subscribe(cfg.MQTTTopic, 1, client.messageHandler())
		token.Wait()
	})

	opts.SetConnectionLostHandler(func(mqttClient mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	client.client = mqttClient
	log.Println("Connected to MQTT broker for agent")
	return client, nil
}
```

**Key Features:**
- Persistent session: `CleanSession=false`
- Auto-reconnect: `SetAutoReconnect(true)`
- Will message: Notification on abnormal disconnection
- TLS support: Two-way authentication

## Message Processing

### Message Handler Function

```go
// mqtt/client.go:238-396
func (c *Client) messageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		if c.monitor != nil {
			c.monitor.IncrementMQTTMessagesReceived()
		}

		log.Printf("Received message on topic: %s", msg.Topic())

		go func() {
			event, err := edgex.ProcessMessage(msg.Payload())
			if err != nil {
				log.Printf("Failed to process message: %v", err)
				return
			}

			if event == nil {
				return
			}

			records := make([]*map[string]any, 0, len(event.Readings))

			for _, reading := range event.Readings {
				data := objPool.GetMap()

				metadataStr := ""
				if reading.Metadata != nil {
					metadataStr = string(reading.Metadata)
				}

				value := common.ParseValue(reading.Value)

				data["id"] = reading.ID
				data["deviceName"] = event.DeviceName
				data["reading"] = reading.ResourceName
				data["value"] = value
				data["valueType"] = reading.ValueType
				data["baseType"] = reading.BaseType
				data["timestamp"] = reading.Origin
				data["metadata"] = metadataStr

				records = append(records, &data)
			}

			if len(records) > 0 {
				if c.monitor != nil {
					c.monitor.IncrementDatabaseOperations()
				}

				err := database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
				if err != nil {
					log.Printf("Failed to batch store data after retries: %v", err)

					errorMsg := err.Error()
					if strings.Contains(errorMsg, "no space left") ||
						strings.Contains(errorMsg, "disk full") ||
						strings.Contains(errorMsg, "file system") ||
						strings.Contains(errorMsg, "I/O error") {
						log.Printf("Fatal storage error detected: %v", err)
						if c.monitor != nil {
							c.monitor.RecordError("storage_error", errorMsg)
						}
					} else if strings.Contains(errorMsg, "lock") ||
						strings.Contains(errorMsg, "busy") {
						log.Printf("Resource contention error detected: %v", err)
						if c.monitor != nil {
							c.monitor.RecordError("resource_contention", errorMsg)
						}
					} else {
						log.Printf("Other database error: %v", err)
						if c.monitor != nil {
							c.monitor.RecordError("database_error", errorMsg)
						}
					}

					if err := c.dataQueue.Enqueue(records); err != nil {
						log.Printf("Failed to enqueue data: %v", err)
					} else {
						log.Printf("Enqueued %d readings for later processing", len(records))
					}

					for _, data := range records {
						objPool.PutMap(*data)
					}
				} else {
					log.Printf("Batch stored %d readings from %s", len(records), event.DeviceName)
					if c.monitor != nil {
						c.monitor.IncrementMQTTMessagesProcessed()
					}

					if c.analyzer != nil && c.analyzer.IsEnabled() {
						readingDataMap := make(map[string][]map[string]interface{})
						for _, record := range records {
							readingName, ok := (*record)["reading"].(string)
							if !ok {
								continue
							}
							readingDataMap[readingName] = append(readingDataMap[readingName], *record)
						}

						for readingName, analysisData := range readingDataMap {
							results, alerts := c.analyzer.Analyze(analysisData, event.DeviceName, readingName)
							if len(results) > 0 {
								log.Printf("Analysis completed for %s: %d results", readingName, len(results))
							}
							if len(alerts) > 0 {
								log.Printf("Detected %d alerts for %s", len(alerts), readingName)
								for _, alert := range alerts {
									log.Printf("Alert: %s - %s - %s", alert.Severity, alert.Message, alert.Reading)
									if c.monitor != nil {
										c.monitor.RecordError(alert.AlertType, alert.Message)
									}
								}
							}
						}
					}

					for _, data := range records {
						objPool.PutMap(*data)
					}
				}
			}
		}()
	}
}
```

## Object Pool

### MapPool Implementation

```go
// mqtt/mapPool.go
package mqtt

import "sync"

var objPool = NewObjectPool()

type objectPool struct {
	mapPool sync.Pool
}

func NewObjectPool() *objectPool {
	return &objectPool{
		mapPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]any)
			},
		},
	}
}

func (p *objectPool) GetMap() map[string]any {
	m := p.mapPool.Get().(map[string]any)
	for k := range m {
		delete(m, k)
	}
	return m
}

func (p *objectPool) PutMap(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	p.mapPool.Put(m)
}
```

## Testing

### Object Pool Testing

```go
// mqtt/mqtt_test.go
package mqtt

import (
	"testing"
)

func TestObjectPool(t *testing.T) {
	m1 := objPool.GetMap()
	if m1 == nil {
		t.Fatal("Expected non-nil map")
	}
	if len(m1) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(m1))
	}

	m1["key1"] = "value1"
	m1["key2"] = 42

	objPool.PutMap(m1)

	m2 := objPool.GetMap()
	if m2 == nil {
		t.Fatal("Expected non-nil map")
	}
	if len(m2) != 0 {
		t.Errorf("Expected empty map after PutMap, got %d entries", len(m2))
	}

	t.Log("Object pool test passed")
}

func TestObjectPoolMultiple(t *testing.T) {
	maps := make([]map[string]any, 10)

	for i := 0; i < 10; i++ {
		maps[i] = objPool.GetMap()
		if maps[i] == nil {
			t.Fatalf("Expected non-nil map at index %d", i)
		}
		maps[i]["index"] = i
	}

	for i := 0; i < 10; i++ {
		objPool.PutMap(maps[i])
	}

	for i := 0; i < 10; i++ {
		m := objPool.GetMap()
		if len(m) != 0 {
			t.Errorf("Expected empty map at index %d, got %d entries", i, len(m))
		}
		objPool.PutMap(m)
	}

	t.Log("Multiple object pool test passed")
}
```

### Running Tests

```bash
go test ./mqtt -v
```

## API Interface

### Subscribe to Topic

```go
func (c *Client) Subscribe() error
```

### Disconnect

```go
func (c *Client) Disconnect()
```

### Publish Message

```go
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error
```

### Publish Batch

```go
func (c *Client) PublishBatch(topic string, qos byte, messages []map[string]interface{}) error
```