# sfsEdgeStore 功能测试指南

本指南介绍如何启动项目并测试各种功能。

## 1. 启动准备

### 1.1 启动 Mosquitto MQTT Broker

```powershell
# 方式 1: 直接启动（前台）
& "C:\Program Files\Mosquitto\mosquitto.exe" -v

# 方式 2: 后台启动（推荐用于测试）
Start-Process -FilePath "C:\Program Files\Mosquitto\mosquitto.exe" -ArgumentList "-v" -WindowStyle Hidden
```

### 1.2 配置并启动 sfsEdgeStore（带模拟器）

创建 `config.json` 或使用环境变量启用模拟器：

```powershell
# 设置环境变量启用模拟器
$env:EDGEX_ENABLE_SIMULATOR = "true"
$env:EDGEX_SIMULATOR_INTERVAL_MIN = "2"
$env:EDGEX_SIMULATOR_INTERVAL_MAX = "5"

# 启动项目
go run main.go
```

或者，您可以创建一个 `config.json` 文件：

```json
{
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsedgestore-test",
  "db_path": "./edgex_data",
  "http_port": "8081",
  "enable_simulator": true,
  "simulator_interval_min": 2,
  "simulator_interval_max": 5
}
```

## 2. 测试功能清单

### 2.1 健康检查（无需认证）

```powershell
# 健康检查
Invoke-RestMethod -Uri "http://localhost:8081/health" -Method Get

# 就绪检查
Invoke-RestMethod -Uri "http://localhost:8081/ready" -Method Get
```

### 2.2 创建 API Key（用于其他需要认证的 API）

```powershell
# 创建管理员 API Key
$createKeyBody = @{
    user_id = "test-admin"
    role = "admin"
    expires_in = 24
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/create-key" `
    -Method Post -Body $createKeyBody -ContentType "application/json"

$apiKey = $response.api_key
Write-Host "API Key: $apiKey"

# 保存 API Key 以便后续使用
$headers = @{
    "X-API-Key" = $apiKey
}
```

### 2.3 测试数据生成（模拟器）

等待几分钟，模拟器会自动发送数据。或者使用测试端点：

```powershell
# 手动发送测试数据
Invoke-RestMethod -Uri "http://localhost:8081/api/test-edgex" `
    -Method Post -Headers $headers
```

### 2.4 查询数据

```powershell
# 查询所有数据
Invoke-RestMethod -Uri "http://localhost:8081/api/readings" `
    -Method Get -Headers $headers

# 按设备名称查询（使用模拟器中的设备）
Invoke-RestMethod -Uri "http://localhost:8081/api/readings?deviceName=temperature-sensor-001" `
    -Method Get -Headers $headers

# 导出为 CSV
Invoke-RestMethod -Uri "http://localhost:8081/api/data/export?format=csv" `
    -Method Get -Headers $headers -OutFile "data_export.csv"
```

### 2.5 认证和权限管理

```powershell
# 列出所有 API Keys
Invoke-RestMethod -Uri "http://localhost:8081/api/auth/list-keys" `
    -Method Get -Headers $headers

# 创建普通用户 API Key
$userKeyBody = @{
    user_id = "test-user"
    role = "user"
    expires_in = 24
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/auth/create-key" `
    -Method Post -Body $userKeyBody -Headers $headers -ContentType "application/json"
```

### 2.6 数据备份和恢复

```powershell
# 备份数据
Invoke-RestMethod -Uri "http://localhost:8081/api/backup" `
    -Method Post -Headers $headers

# 导出为 JSON
Invoke-RestMethod -Uri "http://localhost:8081/api/export/json?path=./export.json" `
    -Method Get -Headers $headers
```

### 2.7 数据保留策略

```powershell
# 查看保留策略状态
Invoke-RestMethod -Uri "http://localhost:8081/api/retention/status" `
    -Method Get -Headers $headers

# 手动触发数据清理
Invoke-RestMethod -Uri "http://localhost:8081/api/retention/cleanup" `
    -Method Post -Headers $headers
```

### 2.8 告警测试

```powershell
# 发送测试告警
$testAlertBody = @{
    type = "test_alert"
    message = "This is a test alert"
    severity = "warning"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/api/alerts/test" `
    -Method Post -Body $testAlertBody -Headers $headers -ContentType "application/json"

# 查看告警通知器状态
Invoke-RestMethod -Uri "http://localhost:8081/api/alerts/notifier/status" `
    -Method Get -Headers $headers
```

### 2.9 数据同步

```powershell
# 查看同步状态
Invoke-RestMethod -Uri "http://localhost:8081/api/sync/status" `
    -Method Get -Headers $headers

# 启动同步
Invoke-RestMethod -Uri "http://localhost:8081/api/sync/start" `
    -Method Post -Headers $headers
```

### 2.10 资源监控

```powershell
# 查看资源使用状态
Invoke-RestMethod -Uri "http://localhost:8081/api/resources/status" `
    -Method Get -Headers $headers
```

### 2.11 配置管理

```powershell
# 获取当前配置
Invoke-RestMethod -Uri "http://localhost:8081/api/config/get" `
    -Method Get -Headers $headers

# 重新加载配置
Invoke-RestMethod -Uri "http://localhost:8081/api/config/reload" `
    -Method Post -Headers $headers
```

## 3. 使用 curl 测试（跨平台）

如果您更喜欢使用 curl：

```bash
# 健康检查
curl http://localhost:8081/health

# 创建 API Key
curl -X POST http://localhost:8081/api/auth/create-key \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-admin","role":"admin","expires_in":24}'

# 使用 API Key 查询数据（替换 YOUR_API_KEY）
curl -H "X-API-Key: YOUR_API_KEY" http://localhost:8081/api/readings

# 发送测试数据
curl -X POST -H "X-API-Key: YOUR_API_KEY" http://localhost:8081/api/test-edgex
```

## 4. 模拟器说明

默认模拟器包含 3 个模拟设备：

1. **temperature-sensor-001** - 温度和湿度传感器
   - temperature: 18-32°C (Int32)
   - humidity: 40-80% (Float32)

2. **power-meter-001** - 电表
   - voltage: 210-240V (Float32)
   - current: 0.5-10A (Float32)
   - power: 100-2500W (Float32)

3. **pressure-sensor-001** - 压力传感器
   - pressure: 950-1050 hPa (Float64)

数据会以 2-5 秒的随机间隔发送到 MQTT broker。

## 5. 停止服务

```powershell
# 在运行 sfsEdgeStore 的终端按 Ctrl+C

# 如果 Mosquitto 在后台运行，停止它
Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
```

## 6. 故障排查

### MQTT 连接失败
- 确认 Mosquitto 正在运行
- 检查端口 1883 是否被占用
- 查看 sfsEdgeStore 日志

### API 认证失败
- 确认使用了正确的 API Key
- 检查 API Key 是否过期
- 确认角色有足够的权限

### 数据查询为空
- 等待模拟器发送数据（需要几分钟）
- 使用 `/api/test-edgex` 手动触发数据
- 检查数据库路径配置
