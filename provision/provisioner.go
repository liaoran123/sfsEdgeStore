package provision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"sfsEdgeStore/devconfig"
)

const (
	EdgeXCoreMetadataAPI = "http://localhost:59881/api/v1/device"
	EdgeXProfileAPI      = "http://localhost:59880/api/v1/deviceprofile"
)

type Provisioner struct {
	configManager *devconfig.ConfigManager
	edgeXEndpoint string
	httpClient    *http.Client
}

type DeviceProfile struct {
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	Labels       []string `json:"labels"`
	Description  string   `json:"description"`
}

type Device struct {
	Name        string                       `json:"name"`
	ServiceName string                       `json:"serviceName"`
	ProfileName string                       `json:"profileName"`
	Protocols   map[string]map[string]string `json:"protocols"`
	AutoEvents  []AutoEvent                  `json:"autoEvents,omitempty"`
}

type AutoEvent struct {
	Interval string `json:"interval"`
	OnChange bool   `json:"onChange"`
	Resource string `json:"resource"`
}

type AutoEventV2 struct {
	Interval string `json:"interval"`
	OnChange bool   `json:"onChange"`
	Resource string `json:"resource"`
}

type DeviceV2 struct {
	Name        string                       `json:"name"`
	ServiceName string                       `json:"serviceName"`
	ProfileName string                       `json:"profileName"`
	Protocols   map[string]map[string]string `json:"protocols"`
	AutoEvents  []AutoEventV2                `json:"autoEvents,omitempty"`
}

func NewProvisioner(cm *devconfig.ConfigManager, edgeXEndpoint string) *Provisioner {
	if edgeXEndpoint == "" {
		edgeXEndpoint = EdgeXCoreMetadataAPI
	}

	return &Provisioner{
		configManager: cm,
		edgeXEndpoint: edgeXEndpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *Provisioner) ProvisionDevice(name, ip, protocol, templateName, interval string, onChange bool, minInterval string, subscriptionTopic string) error {
	if name == "" {
		return fmt.Errorf("device name is required")
	}
	if ip == "" {
		return fmt.Errorf("device IP is required")
	}
	if protocol == "" {
		protocol = "modbus"
	}
	if interval == "" {
		interval = "15s"
	}

	profileName := p.generateProfileName(templateName)

	protocolConfig := p.buildProtocolConfig(protocol, ip)

	// 使用提供的interval或minInterval
	eventInterval := interval
	if minInterval != "" {
		eventInterval = minInterval
	}

	autoEvents := []AutoEventV2{
		{
			Interval: eventInterval,
			OnChange: onChange,
			Resource: "GetValue",
		},
	}

	device := DeviceV2{
		Name:        name,
		ServiceName: "device-modbus",
		ProfileName: profileName,
		Protocols:   protocolConfig,
		AutoEvents:  autoEvents,
	}

	// 注册设备到EdgeX
	err := p.registerDevice(device)
	if err != nil {
		return err
	}

	// 如果设置了订阅主题，记录日志
	if subscriptionTopic != "" {
		log.Printf("Device %s configured with subscription topic: %s", name, subscriptionTopic)
	}

	return nil
}

func (p *Provisioner) ProvisionAll() error {
	devices, err := p.configManager.LoadDevices()
	if err != nil {
		return fmt.Errorf("failed to load devices: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No devices found in devices.csv")
		return nil
	}

	var errors []string
	successCount := 0

	for _, device := range devices {
		fmt.Printf("Provisioning device: %s (IP: %s, Protocol: %s)...\n", device.Name, device.IP, device.Protocol)

		err := p.ProvisionDevice(device.Name, device.IP, device.Protocol, device.Template, device.Interval, device.OnChange, device.MinInterval, device.SubscriptionTopic)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to provision %s: %v", device.Name, err))
			continue
		}

		fmt.Printf("Successfully provisioned: %s\n", device.Name)
		successCount++
	}

	fmt.Printf("\nProvisioned %d/%d devices successfully\n", successCount, len(devices))

	if len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("some devices failed to provision")
	}

	return nil
}

func (p *Provisioner) RemoveDevice(name string) error {
	if name == "" {
		return fmt.Errorf("device name is required")
	}

	url := fmt.Sprintf("%s/name/%s", p.edgeXEndpoint, name)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete device: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *Provisioner) ListDevices() error {
	resp, err := p.httpClient.Get(p.edgeXEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get devices: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to list devices: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var devices []map[string]interface{}
	if err := json.Unmarshal(body, &devices); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return nil
	}

	fmt.Printf("Found %d devices:\n", len(devices))
	for _, d := range devices {
		name, _ := d["name"].(string)
		profile, _ := d["profileName"].(string)
		service, _ := d["serviceName"].(string)
		fmt.Printf("  - Name: %s, Profile: %s, Service: %s\n", name, profile, service)
	}

	return nil
}

func (p *Provisioner) ValidateConnection() error {
	testURL := "http://localhost:59881/api/v1/health"

	resp, err := p.httpClient.Get(testURL)
	if err != nil {
		return fmt.Errorf("cannot connect to EdgeX Core Metadata: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("EdgeX Core Metadata returned status: %d", resp.StatusCode)
	}

	fmt.Println("Successfully connected to EdgeX Core Metadata")
	return nil
}

func (p *Provisioner) UploadProfile(profilePath string) error {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to read profile file: %v", err)
	}

	req, err := http.NewRequest("POST", EdgeXProfileAPI, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/yaml")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload profile: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *Provisioner) generateProfileName(templateName string) string {
	parts := strings.Split(templateName, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".yaml")
	return name + "-profile"
}

func (p *Provisioner) buildProtocolConfig(protocol, ip string) map[string]map[string]string {
	switch strings.ToLower(protocol) {
	case "modbus", "modbus-tcp":
		return map[string]map[string]string{
			"modbus": {
				"Address": ip,
				"Port":    "502",
				"UnitID":  "1",
			},
		}
	case "modbus-rtu":
		return map[string]map[string]string{
			"modbus-rtu": {
				"Address":  ip,
				"BaudRate": "9600",
				"DataBits": "8",
				"StopBits": "1",
				"Parity":   "none",
				"UnitID":   "1",
			},
		}
	case "mqtt":
		return map[string]map[string]string{
			"mqtt": {
				"BrokerURL": ip,
				"ClientID":  "device-mqtt",
				"Topic":     "devices/" + ip + "/events",
			},
		}
	default:
		return map[string]map[string]string{
			"generic": {
				"Address": ip,
			},
		}
	}
}

func (p *Provisioner) registerDevice(device DeviceV2) error {
	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}

	req, err := http.NewRequest("POST", p.edgeXEndpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register device: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *Provisioner) GetHTTPClient() *http.Client {
	return p.httpClient
}
