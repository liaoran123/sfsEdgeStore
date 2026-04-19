# sfsEdgeStore Device Configuration Management

## 1. Configuration System Overview

sfsEdgeStore device configuration system is a lightweight device management solution based on declarative configuration and templated design, aimed at simplifying EdgeX Foundry device onboarding.

### Core Components
- **devconfig package**: Configuration management core, responsible for managing device inventory and templates
- **provision package**: Device provisioning tool, responsible for interacting with EdgeX API
- **sfsedgedevice CLI**: Command-line tool, providing a user-friendly operation interface

### Design Principles
- **Small and Beautiful**: Focus on device configuration and provisioning functions
- **Edge-First**: Lightweight design, suitable for edge devices
- **Configuration as Code**: Use YAML and CSV files to manage configuration
- **Zero Dependency**: No additional dependencies except standard library

## 2. Dual-Mode Configuration Support

sfsEdgeStore supports **two configuration modes** to meet different market needs:

### Configuration Priority
```
YAML (GitOps Mode) > CSV (Excel Mode)
```

If both `devices.yaml` and `devices.csv` exist, YAML takes precedence.

### Target Markets

| Mode | File | Target User | Use Case |
|------|------|-------------|----------|
| **GitOps Mode** | devices.yaml | International developers | Version control, CI/CD, Git workflow |
| **Excel Mode** | devices.csv | Domestic engineers | Excel batch import, field deployment |

## 3. Configuration Directory Structure

```
devconfig/
├── templates/                    # Template library (core asset)
│   ├── modbus/                   # Modbus protocol templates
│   │   ├── temperature.yaml      # Temperature sensor template
│   │   └── pressure.yaml         # Pressure sensor template
│   ├── mqtt/                     # MQTT protocol templates
│   └── opcua/                    # OPC UA protocol templates
├── devices.yaml                  # 【Core】Declarative device configuration (GitOps)
├── devices.yaml.example          # YAML configuration example
├── devices.csv                    # 【Compatibility】Batch import mode (Excel)
└── devices.csv.example           # CSV example file
```

## 4. YAML Configuration (GitOps Mode)

### Why YAML?

- **GitOps Friendly**: Can be submitted to Git repository for version control
- **Templated**: Use `template` field to reference templates under `templates/`, reflecting reusability
- **Declarative**: Clear structure, easy to understand and maintain
- **IDE Support**: Full autocomplete and validation support in modern IDEs

### YAML Format (devices.yaml)

```yaml
# sfsEdgeStore Device Configuration (GitOps Mode)
devices:
  - name: "temp-sensor-001"
    protocol: "modbus-tcp"
    address: "192.168.1.100:502"
    unitId: 1
    template: "modbus/temperature.yaml"
    interval: "1s"
    tags:
      - "production"
      - "zone-a"

  - name: "pressure-sensor-001"
    protocol: "modbus-tcp"
    address: "192.168.1.110:502"
    unitId: 1
    template: "modbus/pressure.yaml"
    interval: "2s"
    tags:
      - "production"
      - "zone-a"
```

### YAML Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| name | string | Yes | Device name | temp-sensor-001 |
| protocol | string | Yes | Communication protocol | modbus-tcp, mqtt |
| address | string | Yes | Device address (IP:Port) | 192.168.1.100:502 |
| unitId | int | No | Modbus Unit ID (default: 1) | 1 |
| topic | string | No | MQTT topic (for MQTT protocol) | factory/hvac/status |
| template | string | Yes | Template path (relative to templates/) | modbus/temperature.yaml |
| interval | string | No | Collection interval (default: 15s) | 1s, 5s, 1m |
| tags | []string | No | Device tags for organization | ["production", "zone-a"] |

## 5. CSV Configuration (Excel Mode)

### Why CSV?

- **Domestic Engineers Favorite**: Field implementation personnel often use Excel for batch processing
- **Simple Operation**: Open Excel, copy-paste, fill in IP and port, save
- **Mass Deployment**: A factory may have 500 sensors, Excel batch processing is the king

### CSV Format (devices.csv)

```csv
name,ip,protocol,template,interval
temp-sensor-001,192.168.1.100,modbus-tcp,modbus/temperature,1s
temp-sensor-002,192.168.1.101,modbus-tcp,modbus/temperature,1s
pressure-sensor-001,192.168.1.110,modbus-tcp,modbus/pressure,2s
```

### CSV Fields

| Field | Description | Example | Required |
|-------|-------------|---------|----------|
| name | Device name | temp-sensor-001 | Yes |
| ip | Device IP address | 192.168.1.100 | Yes |
| protocol | Communication protocol | modbus-tcp | Yes |
| template | Template path | modbus/temperature | Yes |
| interval | Collection interval | 1s | No (default: 15s) |

## 6. Device Template Format

Device templates use YAML format, following EdgeX Foundry Device Profile specification.

### Temperature Sensor Template (temperature.yaml)

```yaml
name: "Temperature-Sensor"
manufacturer: "Generic"
model: "Temp-V1"
labels:
  - "modbus"
  - "temperature"
description: "Generic Modbus Temperature Sensor"

deviceResources:
  - name: "CurrentTemp"
    description: "Current temperature reading"
    properties:
      valueType: "Float32"
      readWrite: "R"
      defaultValue: "0"
      units:
        type: "String"
        readWrite: "R"
        defaultValue: "Celsius"
      attributes:
        PrimaryRegisterAddress: "100"
        PrimaryRegisterType: "InputRegister"

deviceCommands:
  - name: "GetTemp"
    readWrite: "R"
    resourceOperations:
      - resource: "CurrentTemp"
        index: "0"
        operation: "get"
```

### Pressure Sensor Template (pressure.yaml)

```yaml
name: "Pressure-Sensor"
manufacturer: "Generic"
model: "Pressure-V1"
labels:
  - "modbus"
  - "pressure"
description: "Generic Modbus Pressure Sensor"

deviceResources:
  - name: "CurrentPressure"
    description: "Current pressure reading"
    properties:
      valueType: "Float32"
      readWrite: "R"
      defaultValue: "0"
      units:
        type: "String"
        readWrite: "R"
        defaultValue: "Bar"
      attributes:
        PrimaryRegisterAddress: "200"
        PrimaryRegisterType: "InputRegister"

deviceCommands:
  - name: "GetPressure"
    readWrite: "R"
    resourceOperations:
      - resource: "CurrentPressure"
        index: "0"
        operation: "get"
```

## 7. CLI Command-Line Tool

### Basic Usage

```bash
# Display help information
sfsedgedevice help

# Environment variable settings (optional)
export SFSCONFIG_DIR="./devconfig"
```

### Configuration Management Commands

#### List Available Templates

```bash
sfsedgedevice config list-templates
```

#### List Configured Devices

```bash
sfsedgedevice config list-devices

# Output:
# 📄 Mode: YAML (GitOps Mode)
# Configured devices (3):
#   - temp-sensor-001 (IP: 192.168.1.100:502, Protocol: modbus-tcp, Template: modbus/temperature.yaml, Interval: 1s)
```

#### Validate Configuration

```bash
sfsedgedevice config validate
```

#### Remove Device

```bash
sfsedgedevice config remove <device-name>
```

### Device Provisioning Commands

#### Add Single Device

```bash
sfsedgedevice provision add <name> <ip> [options]

# Example
sfsedgedevice provision add temp-sensor-001 192.168.1.100 \
  --protocol modbus-tcp \
  --template modbus/temperature.yaml \
  --interval 1s
```

#### Batch Provision All Devices

```bash
sfsedgedevice provision all
```

#### List Devices in EdgeX

```bash
sfsedgedevice provision list
```

#### Remove Device from EdgeX

```bash
sfsedgedevice provision remove <device-name>
```

#### Validate EdgeX Connection

```bash
sfsedgedevice provision validate
```

## 8. Supported Protocols

| Protocol | Config Value | Description |
|----------|--------------|-------------|
| Modbus TCP | modbus-tcp | Modbus protocol based on TCP |
| Modbus RTU | modbus-rtu | Modbus protocol based on serial |
| MQTT | mqtt | Message Queue Telemetry Transport |
| Generic | generic | Generic protocol template |

## 9. Configuration Validation

Configuration validation checks the following:

1. Whether configuration directory exists
2. Whether template directory exists
3. Whether template files exist
4. Whether device configuration format is correct
5. Whether device-referenced templates exist

### Validation Example

```bash
$ sfsedgedevice config validate
Configuration is valid!
```

## 10. Deployment Process

### Step 1: Prepare Configuration

1. Create `devconfig` directory
2. Prepare device templates (or use default templates)
3. Edit `devices.yaml` (GitOps mode) or `devices.csv` (Excel mode)

### Step 2: Validate Configuration

```bash
sfsedgedevice config validate
```

### Step 3: Start EdgeX Services

```bash
# Start EdgeX using Docker Compose
docker-compose up -d
```

### Step 4: Provision Devices

```bash
# Batch provision all devices
sfsedgedevice provision all

# Or provision device individually
sfsedgedevice provision add <name> <ip> --template <template> --interval <interval>
```

### Step 5: Verify Devices

```bash
# View devices in EdgeX
sfsedgedevice provision list
```

## 11. Configuration Migration

### Migrate from CSV to YAML

```bash
# 1. Export devices from CSV
sfsedgedevice config list-devices

# 2. Create devices.yaml manually based on the output
# 3. Delete devices.csv (YAML takes precedence when both exist)
```

### Migrate from YAML to CSV

```bash
# 1. Delete devices.yaml
rm devconfig/devices.yaml

# 2. Create devices.csv with the same content
```

## 12. Common Issues and Troubleshooting

### Issue 1: Configuration Validation Failed

**Symptom**: `sfsedgedevice config validate` shows error

**Cause**:
- Configuration directory does not exist
- Template file does not exist
- Device configuration format error

**Solution**:
- Check directory structure
- Ensure template files exist
- Validate configuration format

### Issue 2: Device Provisioning Failed

**Symptom**: `sfsedgedevice provision add` shows error

**Cause**:
- EdgeX service is not running
- Network connection problem
- Device configuration error

**Solution**:
- Check EdgeX service status
- Verify network connection
- Check device configuration parameters

### Issue 3: Device Cannot Connect

**Symptom**: Device is registered but no data

**Cause**:
- Device IP address is incorrect
- Device is not powered on
- Network connection problem

**Solution**:
- Verify device IP address
- Check device status
- Test network connection

## 13. Best Practices

### Template Management

- Create dedicated templates for each device type
- Template naming should reflect device functionality
- Regularly update templates to adapt to new devices

### Device Management

- Use unified naming conventions (e.g., `{type}-{location}-{number}`)
- Assign fixed IP addresses to devices
- Set reasonable collection intervals to avoid over-collection

### GitOps Workflow (International Users)

```bash
# Clone configuration repository
git clone https://github.com/your-org/sfsedgestore-config.git
cd sfsedgestore-config

# Edit configuration
vim devices.yaml

# Commit and push
git add devices.yaml
git commit -m "Add new temperature sensors"
git push

# CI/CD pipeline automatically provisions devices
```

### Excel Workflow (Domestic Users)

```bash
# 1. Open devices.csv in Excel
# 2. Copy-paste device information
# 3. Save as CSV format
# 4. Run provisioning command
sfsedgestevice provision all
```

## 14. Summary

sfsEdgeStore device configuration system provides **dual-mode configuration support** to meet the needs of different markets:

- **YAML (GitOps Mode)**: For international developers who prefer configuration as code
- **CSV (Excel Mode)**: For domestic engineers who prefer batch import via Excel

### Key Features

- **Simple and Easy to Use**: YAML and CSV configuration formats
- **Flexible and Scalable**: Supports multiple protocols and device types
- **Lightweight and Efficient**: Suitable for edge device deployment
- **Standardized**: Follows EdgeX Foundry specifications
- **GitOps Ready**: Supports version control and CI/CD workflows

Through this configuration system, you can quickly and reliably manage and provision edge devices, providing a stable data collection foundation for industrial IoT projects.

### Market Strategy

| Market | Configuration Mode | Selling Point |
|--------|---------------------|---------------|
| International ($299) | devices.yaml | GitOps, Zero-Downtime, CI/CD Ready |
| Domestic (Soft + Hard) | devices.csv | Excel Batch Import, Field-Friendly |

One configuration file, two modes of operation. This is the essence of low-cost development.
