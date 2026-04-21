# 简单 MQTT 连接测试脚本
# 直接测试 MQTT 客户端连接

param(
    [string]$BrokerExePath = 'C:\Program Files\Mosquitto\mosquitto.exe',
    [string]$ConfigDir = "$PSScriptRoot\test_mqtt_simple"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Simple MQTT Connection Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Create config directory
if (-not (Test-Path $ConfigDir)) {
    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
}

# Create Mosquitto config
$mosquittoConf = "$ConfigDir\mosquitto.conf"
$confContent = @"
listener 1883
allow_anonymous true
log_dest stdout
log_type error
log_type warning
log_type notice
log_timestamp true
persistence false
"@
$confContent | Out-File -FilePath $mosquittoConf -Encoding utf8

# Check Mosquitto
if (-not (Test-Path -Path $BrokerExePath)) {
    Write-Error "Mosquitto not found: $BrokerExePath"
    exit 1
}

# Stop existing mosquitto
Write-Host "Stopping existing mosquitto..."
Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1

# Start Mosquitto
Write-Host "Starting Mosquitto..."
$mosquittoProc = Start-Process -FilePath $BrokerExePath -ArgumentList '-c', $mosquittoConf, '-v' -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 2
Write-Host "Mosquitto started" -ForegroundColor Green

# Build project
Write-Host ""
Write-Host "Building project..."
Push-Location "$PSScriptRoot\.."
try {
    $env:GOOS = "windows"
    $buildResult = & go build -o sfsedgestore_simple.exe .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed"
        exit 1
    }
    Write-Host "Build successful" -ForegroundColor Green
}
finally {
    Pop-Location
}

# Create a simple config for testing
$configContent = @"
{
    "mqtt_broker": "tcp://localhost:1883",
    "mqtt_topic": "edgex/events/core/#",
    "client_id": "test-simple-client",
    "mqtt_use_tls": false,
    "mqtt_username": "testuser",
    "mqtt_password": "testpass",
    "http_port": "59881",
    "dev_config_path": "./devconfig",
    "db_path": "./data.db",
    "enable_simulator": false
}
"@
$configContent | Out-File -FilePath "$ConfigDir\config.json" -Encoding utf8

# Test with config file
Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
Write-Host "Test with config file" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow

Write-Host "Starting sfsEdgeStore with config..."

# Start and capture output
$proc = Start-Process -FilePath "$PSScriptRoot\..\sfsedgestore_simple.exe" -ArgumentList "-config=$ConfigDir\config.json" -PassThru -WindowStyle Minimized -RedirectStandardOutput "$ConfigDir\output.log" -RedirectStandardError "$ConfigDir\error.log"

# Wait and check output
Start-Sleep -Seconds 8

# Check process
if ($proc.HasExited) {
    Write-Host "sfsEdgeStore exited with code: $($proc.ExitCode)" -ForegroundColor Red
} else {
    Write-Host "sfsEdgeStore is running (PID: $($proc.Id))" -ForegroundColor Green
}

# Show relevant output
if (Test-Path "$ConfigDir\output.log") {
    $output = Get-Content "$ConfigDir\output.log"
    $mqttLines = $output | Where-Object { $_ -match "MQTT|Mqtt|mqtt|Connected|Broker" }
    if ($mqttLines) {
        Write-Host ""
        Write-Host "MQTT Related Output:" -ForegroundColor Cyan
        $mqttLines | ForEach-Object { Write-Host "  $_" }
    }
}

# Stop if still running
if (-not $proc.HasExited) {
    Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    Write-Host "sfsEdgeStore stopped"
}

# Cleanup
Write-Host ""
Write-Host "Cleaning up..."
if ($mosquittoProc -and !$mosquittoProc.HasExited) {
    Stop-Process -Id $mosquittoProc.Id -Force -ErrorAction SilentlyContinue
}
Remove-Item -Path "$ConfigDir" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "$PSScriptRoot\..\sfsedgestore_simple.exe" -Force -ErrorAction SilentlyContinue

Write-Host "Done" -ForegroundColor Green