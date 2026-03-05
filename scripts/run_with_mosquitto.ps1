<#
一键在 Windows 上启动 Mosquitto（可后台）、设置环境变量并运行/测试本项目。
用法示例：
  # 后台启动 mosquitto 并以 go run 启动项目
  .\run_with_mosquitto.ps1 -StartBroker -Background -Action run -ProjectPath 'D:\MyGo\src\sfsdb-edgex-adapter-enterprise -1'

  # 只运行项目（假定已手动启动 broker）
  .\run_with_mosquitto.ps1 -Action run

  # 启动 broker（前台，便于查看日志）
  .\run_with_mosquitto.ps1 -StartBroker

  # 停止后台启动的 broker（会尝试读取 mosquitto.pid）
  .\run_with_mosquitto.ps1 -StopBroker

说明：脚本会在当前 PowerShell 会话设置环境变量 EDGEX_MQTT_BROKER/EDGEX_MQTT_TOPIC/EDGEX_CLIENT_ID。
#>

param(
    [switch]$StartBroker,
    [switch]$StopBroker,
    [string]$BrokerExePath = 'C:\Program Files\Mosquitto\mosquitto.exe',
    [switch]$Background,
    [ValidateSet('run','build','test','none')]
    [string]$Action = 'run',
    [string]$ProjectPath = (Split-Path -Parent $MyInvocation.MyCommand.Definition)
)

function Start-MyBroker {
    param([string]$exePath, [switch]$bg)

    if (-not (Test-Path -Path $exePath)) {
        Write-Error "mosquitto executable not found at: $exePath`n请确认 Mosquitto 已安装，或用 -BrokerExePath 指定正确路径。"
        return $false
    }

    if ($bg) {
        $proc = Start-Process -FilePath $exePath -ArgumentList '-v' -PassThru -WindowStyle Hidden
        $pid = $proc.Id
        "$pid" | Out-File -FilePath "$ProjectPath\mosquitto.pid" -Encoding ascii
        Write-Host "Started mosquitto in background (PID $pid). PID saved to $ProjectPath\mosquitto.pid"
    } else {
        Write-Host "Starting mosquitto in foreground (verbose). Use Ctrl+C to stop."
        & "$exePath" -v
    }
    return $true
}

function Stop-MyBroker {
    param([string]$projectPath)

    $pidFile = Join-Path $projectPath 'mosquitto.pid'
    if (Test-Path $pidFile) {
        try {
            $pid = Get-Content $pidFile | Select-Object -First 1
            if ($pid -and (Get-Process -Id $pid -ErrorAction SilentlyContinue)) {
                Stop-Process -Id $pid -Force
                Write-Host "Stopped mosquitto (PID $pid)."
            } else {
                Write-Warning "PID $pid not running. Attempting to stop by process name."
                Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
            }
            Remove-Item $pidFile -ErrorAction SilentlyContinue
        } catch {
            Write-Warning "Failed to stop by PID: $_. Trying to stop by process name."
            Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
        }
    } else {
        Write-Host "PID file not found, attempting to stop mosquitto by process name..."
        Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
    }
}

# 处理 StopBroker 优先
if ($StopBroker) {
    Stop-MyBroker -projectPath $ProjectPath
    return
}

if ($StartBroker) {
    $ok = Start-MyBroker -exePath $BrokerExePath -bg:$Background
    if (-not $ok) { return }
}

# 设置环境变量（仅当前 PowerShell 会话）
$env:EDGEX_MQTT_BROKER = 'tcp://localhost:1883'
$env:EDGEX_MQTT_TOPIC  = 'edgex/events/core/#'
$env:EDGEX_CLIENT_ID   = 'local-test-client'

Write-Host "Environment variables set: EDGEX_MQTT_BROKER=$env:EDGEX_MQTT_BROKER, EDGEX_MQTT_TOPIC=$env:EDGEX_MQTT_TOPIC"

# 进入项目路径并执行动作
Push-Location $ProjectPath
try {
    switch ($Action) {
        'run' {
            Write-Host "Running project: go run main.go"
            & go run main.go
        }
        'build' {
            Write-Host "Building project: go build -o sfsdb-adapter.exe main.go"
            & go build -o sfsdb-adapter.exe main.go
            if (Test-Path './sfsdb-adapter.exe') {
                Write-Host "Running built binary..."
                & .\sfsdb-adapter.exe
            }
        }
        'test' {
            Write-Host "Running go test ./..."
            & go test ./...
        }
        'none' {
            Write-Host "No project action requested. Broker may be running in background."
        }
    }
} finally {
    Pop-Location
}

if ($StartBroker -and $Background -and $Action -ne 'none') {
    Write-Host "NOTE: mosquitto is still running in background. To stop it later run this script with -StopBroker, or use:`n  Get-Process mosquitto | Stop-Process -Force`n  or`n  taskkill /IM mosquitto.exe /F"
}
