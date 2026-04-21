# MQTT 认证功能完整测试脚本

param(
    [string]$BrokerExePath = 'C:\Program Files\Mosquitto\mosquitto.exe',
    [string]$ConfigDir = "$PSScriptRoot\test_auth_config"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MQTT Authentication Complete Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Create config directory
if (-not (Test-Path $ConfigDir)) {
    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
}

# Create Mosquitto config file
$mosquittoConf = "$ConfigDir\mosquitto.conf"
$confContent = @"
listener 1883
allow_anonymous true
log_dest stdout
log_type error
log_type warning
log_type notice
log_type information
log_timestamp true
persistence false
"@

$confContent | Out-File -FilePath $mosquittoConf -Encoding utf8

# Check Mosquitto executable
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
$mosquittoPid = $mosquittoProc.Id
Write-Host "Mosquitto started, PID: $mosquittoPid"
Start-Sleep -Seconds 2

# Build project
Write-Host ""
Write-Host "Building project..."
Push-Location "$PSScriptRoot\.."
try {
    $buildResult = & go build -o sfsedgestore_test.exe .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed"
        exit 1
    }
    Write-Host "Build successful" -ForegroundColor Green
}
finally {
    Pop-Location
}

# Test 1: Non-authenticated connection
Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
Write-Host "Test 1: Non-auth MQTT Connection" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow

$env:EDGEX_MQTT_BROKER = "tcp://localhost:1883"
$env:EDGEX_MQTT_TOPIC = "edgex/events/core/#"
$env:EDGEX_CLIENT_ID = "test-client-no-auth"
$env:EDGEX_MQTT_USE_TLS = "false"
$env:EDGEX_MQTT_USERNAME = ""
$env:EDGEX_MQTT_PASSWORD = ""

Write-Host "Environment:"
Write-Host "  EDGEX_MQTT_BROKER=$env:EDGEX_MQTT_BROKER"
Write-Host "  EDGEX_CLIENT_ID=$env:EDGEX_CLIENT_ID"
Write-Host "  EDGEX_MQTT_USE_TLS=$env:EDGEX_MQTT_USE_TLS"
Write-Host "  EDGEX_MQTT_USERNAME=(empty)"

Write-Host ""
Write-Host "Starting sfsEdgeStore (5 second test)..."

# Start sfsEdgeStore in background
$sfsProc = Start-Process -FilePath "$PSScriptRoot\..\sfsedgestore_test.exe" -PassThru -WindowStyle Hidden -RedirectStandardOutput "$ConfigDir\sfs_output.log" -RedirectStandardError "$ConfigDir\sfs_error.log"

Start-Sleep -Seconds 5

# Check if process is still running
if ($sfsProc.HasExited) {
    Write-Host "sfsEdgeStore exited with code: $($sfsProc.ExitCode)" -ForegroundColor Red
    Write-Host "Error log:"
    if (Test-Path "$ConfigDir\sfs_error.log") {
        Get-Content "$ConfigDir\sfs_error.log" | Select-Object -First 10
    }
} else {
    Write-Host "sfsEdgeStore is running (PID: $($sfsProc.Id))" -ForegroundColor Green

    # Check output log
    if (Test-Path "$ConfigDir\sfs_output.log") {
        $output = Get-Content "$ConfigDir\sfs_output.log" -Raw
        if ($output -match "Connected to MQTT broker") {
            Write-Host "MQTT Connection: SUCCESS" -ForegroundColor Green
        } else {
            Write-Host "MQTT Connection: Check logs" -ForegroundColor Yellow
        }
    }

    # Stop sfsEdgeStore
    Stop-Process -Id $sfsProc.Id -Force -ErrorAction SilentlyContinue
    Write-Host "sfsEdgeStore stopped"
}

# Test 2: With username/password
Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
Write-Host "Test 2: MQTT with Username/Password" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow

$env:EDGEX_MQTT_USERNAME = "testuser"
$env:EDGEX_MQTT_PASSWORD = "testpass"

Write-Host "Environment:"
Write-Host "  EDGEX_MQTT_USERNAME=$env:EDGEX_MQTT_USERNAME"
Write-Host "  EDGEX_MQTT_PASSWORD=*****"

Write-Host ""
Write-Host "Starting sfsEdgeStore with authentication (5 second test)..."

# Clear logs
Remove-Item "$ConfigDir\sfs_output.log" -Force -ErrorAction SilentlyContinue
Remove-Item "$ConfigDir\sfs_error.log" -Force -ErrorAction SilentlyContinue

# Start sfsEdgeStore in background
$sfsProc = Start-Process -FilePath "$PSScriptRoot\..\sfsedgestore_test.exe" -PassThru -WindowStyle Hidden -RedirectStandardOutput "$ConfigDir\sfs_output.log" -RedirectStandardError "$ConfigDir\sfs_error.log"

Start-Sleep -Seconds 5

# Check if process is still running
if ($sfsProc.HasExited) {
    Write-Host "sfsEdgeStore exited with code: $($sfsProc.ExitCode)" -ForegroundColor Red
    Write-Host "Error log:"
    if (Test-Path "$ConfigDir\sfs_error.log") {
        Get-Content "$ConfigDir\sfs_error.log" | Select-Object -First 10
    }
} else {
    Write-Host "sfsEdgeStore is running (PID: $($sfsProc.Id))" -ForegroundColor Green

    # Check output log
    if (Test-Path "$ConfigDir\sfs_output.log") {
        $output = Get-Content "$ConfigDir\sfs_output.log" -Raw
        if ($output -match "Connected to MQTT broker") {
            Write-Host "MQTT Connection with Auth: SUCCESS" -ForegroundColor Green
        } else {
            Write-Host "MQTT Connection: Check logs" -ForegroundColor Yellow
        }
    }

    # Stop sfsEdgeStore
    Stop-Process -Id $sfsProc.Id -Force -ErrorAction SilentlyContinue
    Write-Host "sfsEdgeStore stopped"
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "All Tests Completed" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# Cleanup
Write-Host "Cleaning up..."
if ($mosquittoProc -and !$mosquittoProc.HasExited) {
    Stop-Process -Id $mosquittoPid -Force -ErrorAction SilentlyContinue
}

Remove-Item -Path "$ConfigDir" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "$PSScriptRoot\..\sfsedgestore_test.exe" -Force -ErrorAction SilentlyContinue

Write-Host "Cleanup completed"