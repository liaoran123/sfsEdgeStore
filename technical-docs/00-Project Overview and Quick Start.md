# sfsEdgeStore Technical Documentation

## Project Overview

sfsEdgeStore is a lightweight data storage adapter designed specifically for industrial IoT edge scenarios, serving as a bridge between EdgeX Foundry and sfsDb database, providing efficient local data read/write and caching capabilities.

## Product Features

### Core Features
- Pure Go implementation: No CGO dependencies, simple cross-platform compilation, worry-free deployment
- Ultra-lightweight: Extremely low resource consumption, can run on any edge device
- Highly reliable: Local storage, network interruptions don't affect data collection
- Easy integration: Native EdgeX Foundry integration, ready to use out of the box
- High performance: LevelDB backend, millisecond-level local query response
- Open source and free: Full functionality, unlimited usage

### Performance Highlights
| Metric | Measured Value | Description |
|--------|---------------|-------------|
| **Memory Usage** | 20.85 MB | Ultra-lightweight, suitable for resource-constrained devices |
| **CPU Usage** | 2.9% | Almost no resource consumption when running in background |
| **Startup Time** | 0.187 seconds | Lightning-fast startup, millisecond-level response |
| **Database Size** | 0.25 MB | 18,681 records take only 0.25 MB |

### Feature Highlights
- 📡 **MQTT Data Ingestion**: Subscribe to EdgeX Foundry event topics
- 💾 **Local Data Storage**: Efficient edge data storage using sfsDb/LevelDB
- 🔄 **Data Queue**: Power failure recovery and data retry mechanism to ensure no data loss
- 📊 **Real-time Monitoring**: Built-in system and business metrics monitoring
- ⚠️ **Intelligent Alerts**: Threshold-based alerts and anomaly detection
- 📈 **Data Analysis**: Built-in time window aggregation and prediction
- 🔐 **Authentication & Authorization**: API Key and RBAC permission control
- 🌐 **HTTP API**: RESTful interface for external queries
- 🔄 **Data Synchronization**: Optional cloud data synchronization
- 🗑️ **Data Retention**: Automatic cleanup of expired data

## Version Changelog

### v1.0.0 (2026-03-08)
- ✅ Initial release
- ✅ MQTT client implementation
- ✅ Database encapsulation and index design
- ✅ Data queue and retry mechanism
- ✅ Authentication and RBAC
- ✅ HTTP server and API design
- ✅ EdgeX message processing
- ✅ Monitoring and alerting
- ✅ Performance optimization and best practices

## Customer Cases

### Case 1: Smart Factory
**Customer**: An automotive parts manufacturer
**Scenario**: Workshop equipment data collection and storage
**Challenge**: Unstable network, data loss risk
**Solution**: Deploy sfsEdgeStore on edge devices to ensure local data storage
**Result**: Data collection rate improved to 99.9%, query response time reduced to milliseconds

### Case 2: Smart Building
**Customer**: Commercial building management company
**Scenario**: HVAC system data monitoring
**Challenge**: Unified management of multi-device data
**Solution**: Use sfsEdgeStore as edge data center
**Result**: Real-time energy consumption analysis, 20% energy savings

## Technical Support

### Contact Information
- **GitHub**: [https://github.com/liaoran123/sfsEdgeStore](https://github.com/liaoran123/sfsEdgeStore)
- **GitCode Mirror**: [https://gitcode.com/liuyun258369/sfsEdgeStore](https://gitcode.com/liuyun258369/sfsEdgeStore)
- **Technical Email**: support@sfsedgestore.com
- **Community Forum**: [https://forum.sfsedgestore.com](https://forum.sfsedgestore.com)

### Support Plans
- **Community Edition**: Support through GitHub Issues
- **Professional Edition**: Email support, 24-48 hour response
- **Enterprise Edition**: Dedicated technical support, 1-4 hour response

## Quick Start

### Project Structure

```
sfsEdgeStore/
├── main.go              # Main program entry
├── agent/               # Management agent
├── alert/               # Alert notification
├── analyzer/            # Data analysis engine
├── auth/                # Authentication and authorization
├── common/              # Common utilities
├── config/              # Configuration management
├── database/            # Database encapsulation
├── edgex/               # EdgeX Foundry integration
├── logger/              # Logging
├── monitor/             # Monitoring metrics
├── mqtt/                # MQTT client
├── queue/               # Data queue
├── resource/            # Resource monitoring
├── retention/           # Data retention policy
├── server/              # HTTP server
├── simulator/           # Data simulator
├── sync/                # Data synchronization
└── time/                # Time series analysis
```

### Main Program Entry

The main program `main.go` shows the system startup process:

```go
// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sfsEdgeStore/agent"
	"sfsEdgeStore/alert"
	"sfsEdgeStore/analyzer"
	"sfsEdgeStore/auth"
	"sfsEdgeStore/config"
	"sfsEdgeStore/database"
	"sfsEdgeStore/monitor"
	"sfsEdgeStore/mqtt"
	"sfsEdgeStore/queue"
	"sfsEdgeStore/resource"
	"sfsEdgeStore/retention"
	"sfsEdgeStore/server"
	"sfsEdgeStore/simulator"
	"sfsEdgeStore/sync"
)

var appConfig *config.Config
var dataQueue *queue.Queue
var monitorInstance *monitor.Monitor
var agentInstance *agent.Agent
var analyzerInstance *analyzer.Analyzer
var retentionManager *retention.RetentionManager
var alertNotifier *alert.Notifier
var syncManager *sync.SyncManager
var resourceMonitor *resource.ResourceMonitor
var simulatorInstance *simulator.Simulator

func main() {
	// Load configuration
	var err error
	appConfig, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize monitoring
	monitorInstance = monitor.NewMonitor()

	// Initialize alert notifier
	alertNotifier = alert.NewNotifier(appConfig)
	monitorInstance.SetNotifier(alertNotifier)
	if err := alertNotifier.Start(); err != nil {
		log.Printf("Failed to start alert notifier: %v", err)
	}

	// Initialize analysis engine
	analyzerInstance = analyzer.NewAnalyzer(appConfig)
	if appConfig.EnableAnalyzer {
		log.Println("Analyzer enabled")
	} else {
		log.Println("Analyzer disabled")
	}

	// Connect to sfsDb
	if err = database.Init(appConfig.DBPath, appConfig.DBUseEncryption, appConfig.DBEncryptionKey, appConfig.DBEncryptionAlgorithm); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start authentication cleanup task
	authManager := auth.NewAuthManager()
	authManager.StartCleanupTask(24 * time.Hour)

	// Initialize data queue
	dataQueue, err = queue.NewQueue("./data_queue")
	if err != nil {
		log.Fatalf("Failed to initialize data queue: %v", err)
	}

	var mqttClient *mqtt.Client
	if !appConfig.EnableSimulator {
		mqttClient, err = mqtt.NewClient(appConfig, dataQueue, monitorInstance, analyzerInstance)
		if err != nil {
			log.Fatalf("Failed to initialize MQTT: %v", err)
		}
		defer mqttClient.Disconnect()

		if err := mqttClient.Subscribe(); err != nil {
			log.Fatalf("Failed to subscribe to EdgeX messages: %v", err)
		}
	} else {
		log.Println("Simulator enabled, skipping MQTT connection")
	}

	log.Println("sfsDb EdgeX adapter started successfully")

	// Start queue processing goroutine
	dataQueue.ProcessQueue(func(data interface{}) error {
		records, ok := data.([]*map[string]any)
		if !ok {
			return fmt.Errorf("invalid data type in queue")
		}
		return database.BatchInsertWithRetry(database.Table, records, 3, 2*time.Second)
	})

	// Initialize and start minimalist management agent
	agentInstance, err = agent.NewAgent(appConfig, monitorInstance)
	if err != nil {
		log.Printf("Failed to initialize agent: %v", err)
	} else {
		if err := agentInstance.Start(); err != nil {
			log.Printf("Failed to start agent: %v", err)
		}
	}

	// Initialize and start data retention policy manager
	retentionManager = retention.NewRetentionManager(database.Table, appConfig)
	if err := retentionManager.Start(); err != nil {
		log.Printf("Failed to start retention manager: %v", err)
	}

	// Initialize and start data synchronization manager
	syncManager, err = sync.NewSyncManager(appConfig)
	if err != nil {
		log.Printf("Failed to initialize sync manager: %v", err)
	} else {
		if err := syncManager.Start(); err != nil {
			log.Printf("Failed to start sync manager: %v", err)
		}
	}

	// Initialize and start resource monitor
	resourceMonitor = resource.NewResourceMonitor(appConfig, monitorInstance)
	if err := resourceMonitor.Start(); err != nil {
		log.Printf("Failed to start resource monitor: %v", err)
	}

	// Start HTTP server
	serverInstance := server.NewServer(database.Table, appConfig, monitorInstance, retentionManager, alertNotifier, syncManager, resourceMonitor)
	if err := serverInstance.Start(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down adapter...")

	// Stop all components
	if agentInstance != nil {
		agentInstance.Stop()
	}
	if retentionManager != nil {
		retentionManager.Stop()
	}
	if alertNotifier != nil {
		alertNotifier.Stop()
	}
	if syncManager != nil {
		syncManager.Stop()
	}
	if resourceMonitor != nil {
		resourceMonitor.Stop()
	}

	time.Sleep(5 * time.Second)
	log.Println("Adapter exited")
}
```

### Startup Process

1. **Load Configuration**: Load configuration from config file or environment variables
2. **Initialize Monitoring**: Set up monitoring metrics and alert notifications
3. **Connect Database**: Initialize sfsDb database connection
4. **Initialize Queue**: Create data queue for fault recovery
5. **Connect MQTT**: Connect to EdgeX Foundry's MQTT Broker
6. **Start Queue Processing**: Process data in the queue in the background
7. **Start Components**: Agent, retention policy, synchronization, resource monitoring, etc.
8. **Start HTTP Server**: Provide RESTful API
9. **Wait for Interrupt Signal**: Graceful shutdown

### Compilation and Run

```bash
# Compile
go build -o sfsedgestore main.go

# Run
./sfsedgestore
```

### Dependencies

Main dependencies used by the project:

```go
// go.mod
module sfsEdgeStore

go 1.25.3

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/edgexfoundry/go-mod-configuration/v2 v2.3.0
	github.com/liaoran123/sfsDb v1.9.3
)
```