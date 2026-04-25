package sync

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"sfsEdgeStore/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Uploader 上传器接口
type Uploader interface {
	Upload(data []byte) error
}

// NewUploader 创建上传器
func NewUploader(cfg *config.Config) (Uploader, error) {
	// 根据配置选择上传器
	switch cfg.DataSyncMQTTTopic {
	case "":
		return NewHTTPUploader(cfg), nil
	default:
		return NewMQTTUploader(cfg), nil
	}
}

// HTTPUploader HTTP上传器
type HTTPUploader struct {
	config *config.Config
	client *http.Client
}

// NewHTTPUploader 创建HTTP上传器
func NewHTTPUploader(cfg *config.Config) *HTTPUploader {
	return &HTTPUploader{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Upload 上传数据
func (u *HTTPUploader) Upload(data []byte) error {
	// 默认上传地址
	url := "http://localhost:8080/api/sync"
	if u.config.DataSyncMQTTTopic != "" {
		url = u.config.DataSyncMQTTTopic
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-SfsEdgeStore-Sync", "true")

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP upload failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// MQTTUploader MQTT上传器
type MQTTUploader struct {
	config *config.Config
	client mqtt.Client
}

// NewMQTTUploader 创建MQTT上传器
func NewMQTTUploader(cfg *config.Config) *MQTTUploader {
	// 初始化MQTT客户端
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.ClientID + "-sync")
	opts.SetCleanSession(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		// 连接丢失处理
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		// 连接失败不返回错误，在Upload时处理
	}

	return &MQTTUploader{
		config: cfg,
		client: client,
	}
}

// Upload 上传数据
func (u *MQTTUploader) Upload(data []byte) error {
	// 确保客户端已连接
	if !u.client.IsConnected() {
		if token := u.client.Connect(); token.Wait() && token.Error() != nil {
			return fmt.Errorf("MQTT connection failed: %w", token.Error())
		}
	}

	// 上传数据到MQTT主题
	topic := u.config.DataSyncMQTTTopic
	if topic == "" {
		topic = "sfsEdgeStore/sync"
	}

	token := u.client.Publish(topic, 1, false, data)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("MQTT publish failed: %w", token.Error())
	}

	return nil
}

/*
- License验证 ：只有 LicenseType == "enterprise" 时才会启用
- 配置开关 ：需要设置 EnableDataSync: true
- 免费版限制 ：开源版用户无法使用此功能
*/
