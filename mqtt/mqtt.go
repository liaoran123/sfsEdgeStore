package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"sfsEdgeStore/config"
	"sfsEdgeStore/monitor"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	batchSize = 100
	batchTime = 100
)

// MQTTClient 纯通信层 - 只负责 MQTT 连接、订阅、发布
type MQTTClient struct {
	client    mqtt.Client
	brokerURL string              // MQTT 代理地址
	clientID  string              // 客户端 ID
	username  string              // 用户名
	password  string              // 密码
	onMessage mqtt.MessageHandler // 消息处理函数
	monitor   *monitor.Monitor
}

func NewMQTTClient(cfg *config.Config, onMessage mqtt.MessageHandler, monitor *monitor.Monitor) (*MQTTClient, error) {
	c := &MQTTClient{
		brokerURL: cfg.MQTTBroker,
		clientID:  cfg.ClientID,
		username:  cfg.MQTTUsername,
		password:  cfg.MQTTPassword,
		onMessage: onMessage,
		monitor:   monitor,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.brokerURL)                   // 添加 MQTT 代理地址
	opts.SetClientID(c.clientID)                  // 设置客户端 ID
	opts.SetCleanSession(false)                   // 持久会话，broker 自动恢复订阅
	opts.SetAutoReconnect(true)                   // 自动重连
	opts.SetMaxReconnectInterval(time.Minute * 5) // 最大重连间隔 5 分钟

	if c.username != "" {
		opts.SetUsername(c.username)
		if c.password != "" {
			opts.SetPassword(c.password)
		}
	}

	// TLS 配置
	if cfg.MQTTUseTLS {
		tlsConfig, err := createTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tlsConfig)
		log.Println("TLS: TLS encryption enabled")
	} else {
		log.Println("Security: WARNING - TLS encryption disabled, connection is not secure")
	}

	// 安全检查
	performSecurityCheck(cfg)

	willPayload, _ := createStatusPayload(c.clientID, "offline")
	/*
		设置离线遗嘱：告诉其他客户端，我离线了
		1 Topic statusTopic 主题： 如：edgex/events/status
		2 QoS 0 消息质量：最多一次
		3 Retained false 不保留：Broker 不存这条消息
	*/
	//上线，离线都是在edgex/events/status主题下发布
	opts.SetWill(statusTopic, string(willPayload), 0, false)

	opts.SetOnConnectHandler(c.onConnect)
	opts.SetConnectionLostHandler(c.onConnectionLost)

	c.client = mqtt.NewClient(opts)
	token := c.client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	return c, nil
}

// onConnect 连接成功回调
func (c *MQTTClient) onConnect(client mqtt.Client) {
	log.Println("MQTT broker connected")
	if c.monitor != nil {
		c.monitor.SetMQTTConnectionStatus(true)
	}
	c.publishOnlineStatus(client)
}

// onConnectionLost 连接丢失回调
func (c *MQTTClient) onConnectionLost(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
	if c.monitor != nil {
		c.monitor.SetMQTTConnectionStatus(false)
	}
}

/*
	IsConnected 检查是否在线。

- monitor不直接使用IsConnected，而是回调onConnect和onConnectionLost更新连接状态，主要为了解耦合。
  - 性能 - atomic.Bool.Load() 比调 paho-mqtt 的 IsConnected() 快得多

- 解耦 - Monitor 不需要知道 MQTT Client 的存在
- 够用 - 连接状态变化不频繁，延迟可以忽略
*/
func (c *MQTTClient) IsConnected() bool {
	return c.client.IsConnected()
}

// publishOnlineStatus 发布上线状态，告诉其他客户端，我上线了
func (c *MQTTClient) publishOnlineStatus(client mqtt.Client) {
	if err := publishStatusMessage(client, c.clientID, "online"); err != nil {
		log.Printf("Failed to publish online status: %v", err)
	}
}

// Subscribe 订阅主题
// 1 QoS 1 消息质量：最多一次
// 2 Retained false 不保留：Broker 不存这条消息
func (c *MQTTClient) Subscribe(topics []string) error {
	for _, topic := range topics {
		/*
		   	topic,         // 参数 1: 要订阅的主题
		       1,             // 1 At least once 至少一次，保证送达
		       c.onMessage    // 参数 3: 消息到达时的回调函数
		*/
		//所有主题共用一个回调函数onMessage。可以为每个主题设置不同的回调函数
		token := c.client.Subscribe(topic, 1, c.onMessage)
		token.Wait() // 等待订阅确认
		if token.Error() != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %v", topic, token.Error())
		}
		log.Printf("Subscribed to topic: %s", topic)
	}
	return nil
}

// Disconnect 断开 MQTT 连接
func (c *MQTTClient) Disconnect() {
	// 发布离线通知
	if c.client.IsConnected() {
		if err := publishStatusMessage(c.client, c.clientID, "offline"); err != nil {
			log.Printf("Failed to publish offline status: %v", err)
		}
	}
	// 断开连接
	c.client.Disconnect(250)
	log.Println("MQTT client disconnected")
}

// createTLSConfig 创建 TLS 配置
func createTLSConfig(cfg *config.Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	// 配置 CA 证书
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

	// 配置客户端证书（双向 TLS）
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

// performSecurityCheck 执行安全检查
func performSecurityCheck(cfg *config.Config) {
	log.Println("Performing security check...")

	// 检查用户名密码认证
	if cfg.MQTTUsername == "" {
		log.Println("Security: No username set, using anonymous authentication")
	} else {
		log.Println("Security: Username/password authentication enabled")
	}

	// 检查 TLS 加密
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
}
