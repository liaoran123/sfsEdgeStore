package mqtt

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/broadcast"
	"sfsEdgeStore/config"
	"sfsEdgeStore/filter"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/pool"
	"sfsEdgeStore/queue"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	batchSize = 100
	batchTime = 100
)

type Client struct {
	client        mqtt.Client
	config        *config.Config
	dataQueue     *queue.Queue
	monitor       *monitor.Monitor
	analyzer      *analyzer.Analyzer
	filterManager *filter.FilterManager
	broadcaster   broadcast.Broadcaster
	writePool     *pool.Pool // 复用固定数量协程，防止泄漏

	mu             sync.Mutex
	pendingRecords []*map[string]any
	lastBatchTime  time.Time
	batchTimer     *time.Timer

	registeredDevices map[string]bool
	muDevices         sync.Mutex
}

var standardTopics = []string{
	"edgex/events/#",
	"devices/+/data",
	"edgex/events/core/#",
}

func NewClient(cfg *config.Config, dataQueue *queue.Queue, monitor *monitor.Monitor, analyzer *analyzer.Analyzer, broadcaster broadcast.Broadcaster) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Minute * 5)
	opts.SetResumeSubs(true)

	willTopic := "edgex/events/status"
	willMessage := map[string]any{
		"status":    "offline",
		"clientId":  cfg.ClientID,
		"timestamp": time.Now().UnixNano(),
	}
	willPayload, _ := json.Marshal(willMessage)
	opts.SetWill(willTopic, string(willPayload), 0, false)

	if cfg.MQTTUsername != "" {
		opts.SetUsername(cfg.MQTTUsername)
		if cfg.MQTTPassword != "" {
			opts.SetPassword(cfg.MQTTPassword)
		}
	}

	client := &Client{
		config:            cfg,
		dataQueue:         dataQueue,
		monitor:           monitor,
		analyzer:          analyzer,
		filterManager:     filter.NewFilterManager(),
		broadcaster:       broadcaster,
		writePool:         pool.NewPoolForIO(),
		pendingRecords:    make([]*map[string]any, 0, batchSize),
		lastBatchTime:     time.Now(),
		registeredDevices: make(map[string]bool),
	}

	client.performSecurityCheck(cfg)

	opts.SetOnConnectHandler(func(mqttClient mqtt.Client) {
		log.Println("MQTT broker connected")
		if client.monitor != nil {
			client.monitor.SetMQTTConnectionStatus(true)
		}
		onlineTopic := "edgex/events/status"
		onlineMessage := map[string]interface{}{
			"status":    "online",
			"clientId":  cfg.ClientID,
			"timestamp": time.Now().UnixNano(),
		}
		onlinePayload, _ := json.Marshal(onlineMessage)
		token := mqttClient.Publish(onlineTopic, 1, false, onlinePayload)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to publish online status: %v", token.Error())
		}
		if err := client.subscribeTopics(mqttClient); err != nil {
			log.Printf("Failed to resubscribe to topics: %v", err)
		}
	})

	opts.SetConnectionLostHandler(func(mqttClient mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
		if client.monitor != nil {
			client.monitor.SetMQTTConnectionStatus(false)
		}
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

func (c *Client) subscribeTopics(mqttClient mqtt.Client) error {
	topics := c.getTopicsToSubscribe()
	for _, topic := range topics {
		token := mqttClient.Subscribe(topic, 1, c.messageHandler())
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to subscribe to topic %s: %v", topic, token.Error())
			continue
		}
		log.Printf("Subscribed to topic: %s", topic)
	}
	return nil
}

func (c *Client) getTopicsToSubscribe() []string {
	topics := []string{}
	if c.config.MQTTTopic != "" {
		topics = append(topics, c.config.MQTTTopic)
	}
	if c.config.AutoSubscribe {
		version, err := c.detectEdgeXVersion()
		if err != nil {
			log.Printf("Failed to detect EdgeX version: %v, using standard topics", err)
			topics = append(topics, standardTopics...)
		} else {
			versionTopics := c.getTopicsByVersion(version)
			log.Printf("Using EdgeX %s topics: %v", version, versionTopics)
			topics = append(topics, versionTopics...)
		}
	}
	topicMap := make(map[string]bool)
	for _, topic := range topics {
		topicMap[topic] = true
	}
	uniqueTopics := []string{}
	for topic := range topicMap {
		uniqueTopics = append(uniqueTopics, topic)
	}
	return uniqueTopics
}

func (c *Client) Subscribe() error {
	return c.subscribeTopics(c.client)
}

const (
	EdgeXVersionUnknown = "unknown"
	EdgeXVersionV1      = "v1"
	EdgeXVersionV2      = "v2"
	EdgeXVersionLatest  = "latest"
)

var edgeXVersionTopics = map[string][]string{
	EdgeXVersionV1: {
		"edgex/events/core/#",
		"devices/+/data",
	},
	EdgeXVersionV2: {
		"edgex/events/#",
		"devices/+/data",
		"edgex/v2/events/#",
	},
	EdgeXVersionLatest: {
		"edgex/events/#",
		"devices/+/data",
		"edgex/v2/events/#",
	},
}

func (c *Client) detectEdgeXVersion() (string, error) {
	log.Println("Detecting EdgeX version...")
	if !c.client.IsConnected() {
		log.Println("MQTT not connected, cannot detect EdgeX version")
		return EdgeXVersionLatest, nil
	}
	versions := []string{EdgeXVersionV2, EdgeXVersionV1}
	for _, version := range versions {
		topics := edgeXVersionTopics[version]
		for _, topic := range topics {
			token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
				log.Printf("Received message on %s topic, assuming EdgeX %s", topic, version)
			})
			token.Wait()
			if token.Error() == nil {
				log.Printf("Successfully subscribed to %s topic, detected EdgeX %s", topic, version)
				c.client.Unsubscribe(topic)
				return version, nil
			}
		}
	}
	log.Println("Could not detect EdgeX version, using latest version")
	return EdgeXVersionLatest, nil
}

func (c *Client) getTopicsByVersion(version string) []string {
	if topics, ok := edgeXVersionTopics[version]; ok {
		return topics
	}
	return edgeXVersionTopics[EdgeXVersionLatest]
}

func (c *Client) createTLSConfig(cfg *config.Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	if cfg.MQTTCACert != "" {
		caCert, err := os.ReadFile(cfg.MQTTCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %v", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to add CA cert to pool")
		}
		tlsConfig.RootCAs = caCertPool
		log.Println("TLS: CA certificate loaded successfully")
	} else {
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			log.Printf("Warning: Failed to load system cert pool: %v", err)
			caCertPool = x509.NewCertPool()
		}
		tlsConfig.RootCAs = caCertPool
		log.Println("TLS: Using system default certificate pool")
	}
	if cfg.MQTTClientCert != "" && cfg.MQTTClientKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.MQTTClientCert, cfg.MQTTClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		log.Println("TLS: Client certificate loaded successfully (mutual TLS enabled)")
	}
	return tlsConfig, nil
}

func (c *Client) performSecurityCheck(cfg *config.Config) {
	log.Println("Performing security check...")
	if cfg.MQTTUsername == "" {
		log.Println("Security: No username set, using anonymous authentication")
	} else {
		log.Println("Security: Username/password authentication enabled")
	}
	if cfg.MQTTUseTLS {
		log.Println("Security: TLS encryption enabled")
		if cfg.MQTTCACert != "" {
			log.Println("Security: Custom CA certificate configured")
		} else {
			log.Println("Security: Using system default CA certificates")
		}
		if cfg.MQTTClientCert != "" && cfg.MQTTClientKey != "" {
			log.Println("Security: Mutual TLS authentication enabled")
		}
	} else {
		log.Println("Security: WARNING - TLS encryption disabled, connection is not secure")
	}
	if cfg.MQTTPassword != "" && len(cfg.MQTTPassword) < 8 {
		log.Println("Security: WARNING - Password is too short (less than 8 characters)")
	}
}

func (c *Client) Disconnect() {
	c.mu.Lock()
	if len(c.pendingRecords) > 0 {
		records := c.pendingRecords
		c.pendingRecords = make([]*map[string]any, 0, batchSize)
		c.mu.Unlock()
		c.flushRecords(records)
	} else {
		c.mu.Unlock()
	}
	c.writePool.Stop()
	c.client.Disconnect(250)
}

func (c *Client) Publish(topic string, qos byte, retained bool, payload any) error {
	token := c.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

func (c *Client) PublishBatch(topic string, qos byte, messages []map[string]any) error {
	compressedPayload, err := c.compressMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to compress messages: %v", err)
	}
	return c.Publish(topic, qos, false, compressedPayload)
}

func (c *Client) compressMessages(messages []map[string]any) ([]byte, error) {
	jsonData, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gzw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) isDeviceAllowed(deviceName string) bool {
	if c.config.LicenseType == "enterprise" {
		return true
	}
	if c.registeredDevices[deviceName] {
		return true
	}
	maxDevices := 5
	if c.config.LicenseType == "business" {
		maxDevices = 50
	}
	if c.config.EnterpriseFeatures.MaxDevices >= 0 {
		maxDevices = c.config.EnterpriseFeatures.MaxDevices
	}
	if maxDevices > 0 && len(c.registeredDevices) >= maxDevices {
		return false
	}
	c.registeredDevices[deviceName] = true
	log.Printf("New device registered: %s (total: %d/%d, license: %s)", deviceName, len(c.registeredDevices), maxDevices, c.config.LicenseType)
	return true
}

func (c *Client) GetRegisteredDeviceCount() int {
	return len(c.registeredDevices)
}

func (c *Client) AddSubscription(topic string) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}
	token := c.client.Subscribe(topic, 1, c.messageHandler())
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %v", topic, token.Error())
	}
	log.Printf("Successfully added subscription to topic: %s", topic)
	return nil
}

func (c *Client) RemoveSubscription(topic string) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}
	token := c.client.Unsubscribe(topic)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %v", topic, token.Error())
	}
	log.Printf("Successfully removed subscription from topic: %s", topic)
	return nil
}

func (c *Client) GetSubscriptions() []string {
	topics := []string{}
	if c.config.MQTTTopic != "" {
		topics = append(topics, c.config.MQTTTopic)
	}
	if c.config.AutoSubscribe {
		topics = append(topics, standardTopics...)
	}
	return topics
}
