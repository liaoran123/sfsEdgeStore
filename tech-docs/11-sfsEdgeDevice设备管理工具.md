# sfsEdgeDevice 设备管理命令行工具

## 1. 概述

`sfsEdgeDevice` 是 sfsEdgeStore 项目中的设备管理命令行工具，提供设备配置管理和向 EdgeX Foundry 配置设备的功能。该工具支持两种管理模式：YAML 模式（GitOps 模式）和 CSV 模式（Excel 模式），方便用户根据实际需求选择合适的配置方式。

## 2. 功能特性

- **配置管理**：管理设备配置和设备模板
- **设备配置**：将设备配置到 EdgeX Foundry
- **批量操作**：支持从 CSV 文件批量配置设备
- **多协议支持**：支持 Modbus-TCP、Modbus-RTU、MQTT 等协议
- **灵活配置**：支持自定义采集间隔、触发模式等参数
- **配置验证**：验证配置的正确性和 EdgeX 连接

## 3. 命令结构

```
sfsedgedevice <command> [options]
```

### 3.1 主要命令

| 命令 | 说明 |
|------|------|
| `config` | 管理配置和模板 |
| `provision` | 配置设备到 EdgeX |
| `help` | 显示帮助信息 |

## 4. config 子命令

### 4.1 列出模板

```bash
sfsedgedevice config list-templates
```

列出所有可用的设备模板。

### 4.2 列出设备

```bash
sfsedgedevice config list-devices
```

列出所有已配置的设备，显示当前配置模式（YAML 或 CSV）。

### 4.3 添加设备

```bash
sfsedgedevice config add <csv-file>
```

从 CSV 文件添加设备配置。

### 4.4 删除设备

```bash
sfsedgedevice config remove <device-name>
```

从配置中删除指定设备。

### 4.5 验证配置

```bash
sfsedgedevice config validate
```

验证配置文件的正确性。

## 5. provision 子命令

### 5.1 添加设备

```bash
sfsedgedevice provision add <name> <ip> [options]
```

向 EdgeX 添加单个设备。

**参数：**
- `<name>` - 设备名称
- `<ip>` - 设备 IP 地址

**选项：**

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--protocol` | string | `modbus-tcp` | 设备协议（modbus-tcp, modbus-rtu, mqtt） |
| `--template` | string | `modbus/temperature` | 设备模板 |
| `--interval` | string | `15s` | AutoEvent 间隔（如 1s, 5s, 1m） |
| `--onChange` | bool | `false` | 是否仅在值变化时触发 |
| `--minInterval` | string | `` | 最小触发间隔 |
| `--subscriptionTopic` | string | `` | MQTT 订阅主题 |

**示例：**

```bash
# 添加一个 Modbus-TCP 温度传感器
sfsedgedevice provision add temp-sensor-001 192.168.1.100 \
  --protocol modbus-tcp \
  --template modbus/temperature \
  --interval 1s

# 添加一个 MQTT 设备
sfsedgedevice provision add mqtt-device-001 192.168.1.101 \
  --protocol mqtt \
  --template mqtt/generic \
  --subscriptionTopic "devices/+/messages"

# 添加设备并在值变化时触发
sfsedgedevice provision add pressure-sensor-001 192.168.1.102 \
  --template modbus/pressure \
  --interval 5s \
  --onChange
```

### 5.2 配置所有设备

```bash
sfsedgedevice provision all
```

从配置文件批量配置所有设备到 EdgeX。

### 5.3 列出设备

```bash
sfsedgedevice provision list
```

列出 EdgeX 中配置的所有设备。

### 5.4 删除设备

```bash
sfsedgedevice provision remove <name>
```

从 EdgeX 删除指定设备。

### 5.5 验证连接

```bash
sfsedgedevice provision validate
```

验证与 EdgeX Foundry 的连接是否正常。

## 6. 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SFSCONFIG_DIR` | `./devconfig` | 配置文件目录 |

## 7. 配置模式

### 7.1 YAML 模式（GitOps 模式）

使用 YAML 格式的配置文件，适合版本控制和自动化部署。

```yaml
devices:
  - name: "temp-sensor-001"
    ip: "192.168.1.100"
    protocol: "modbus-tcp"
    template: "modbus/temperature"
    interval: "1s"
```

### 7.2 CSV 模式（Excel 模式）

使用 CSV 格式的配置文件，适合大规模设备管理和 Excel 编辑。

```csv
Name,IP,Protocol,Template,Interval
temp-sensor-001,192.168.1.100,modbus-tcp,modbus/temperature,1s
pressure-sensor-001,192.168.1.101,modbus-tcp,modbus/pressure,5s
```

## 8. 使用示例

### 8.1 快速开始

```bash
# 1. 验证配置
sfsedgedevice config validate

# 2. 列出可用的模板
sfsedgedevice config list-templates

# 3. 列出已配置的设备
sfsedgedevice config list-devices

# 4. 验证 EdgeX 连接
sfsedgedevice provision validate

# 5. 配置单个设备
sfsedgedevice provision add temp-sensor-001 192.168.1.100 \
  --template modbus/temperature \
  --interval 1s

# 6. 批量配置所有设备
sfsedgedevice provision all
```

### 8.2 完整工作流程

```bash
# 设置配置目录
export SFSCONFIG_DIR=/path/to/config

# 验证配置正确性
sfsedgedevice config validate

# 查看设备列表
sfsedgedevice config list-devices

# 验证 EdgeX 连接
sfsedgedevice provision validate

# 批量配置设备
sfsedgedevice provision all

# 查看已配置的设备
sfsedgedevice provision list
```

## 9. 设备模板

### 9.1 内置模板

| 模板 | 说明 |
|------|------|
| `modbus/temperature` | Modbus 温度传感器 |
| `modbus/pressure` | Modbus 压力传感器 |
| `modbus/humidity` | Modbus 湿度传感器 |
| `modbus/flow` | Modbus 流量传感器 |
| `modbus/power` | Modbus 功率传感器 |
| `mqtt/generic` | 通用 MQTT 设备 |

### 9.2 自定义模板

用户可以在 `templates` 目录下创建自定义设备模板。

## 10. 故障排除

### 10.1 常见问题

**Q: 配置验证失败怎么办？**
```bash
# 查看详细的验证错误
sfsedgedevice config validate
```

**Q: 无法连接到 EdgeX 怎么办？**
```bash
# 验证 EdgeX 连接
sfsedgedevice provision validate

# 检查 EdgeX 服务状态
curl http://localhost:59880/api/v2/ping
```

**Q: 设备配置失败怎么办？**
```bash
# 查看 EdgeX 日志
# 检查设备服务是否正常运行
# 验证设备 IP 和协议配置是否正确
```

### 10.2 调试模式

在命令行中添加 `-v` 或 `--verbose` 选项（如支持）以查看详细日志。

## 11. 相关文件

- `devconfig/` - 设备配置文件目录
- `provision/` - 设备配置逻辑
- `templates/` - 设备模板目录