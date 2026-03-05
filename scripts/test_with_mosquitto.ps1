<#
简化版测试脚本，使用指定的 Mosquitto 执行文件测试项目
#>

param(
    [string]$BrokerExePath = 'C:\Program Files\Mosquitto\mosquitto.exe'
)

# 检查 Mosquitto 执行文件是否存在
if (-not (Test-Path -Path $BrokerExePath)) {
    Write-Error "Mosquitto 执行文件未找到: $BrokerExePath`n请确认 Mosquitto 已安装，或使用 -BrokerExePath 指定正确路径。"
    exit 1
}

# 启动 Mosquitto 后台进程
Write-Host "启动 Mosquitto 后台进程..."
$proc = Start-Process -FilePath $BrokerExePath -ArgumentList '-v' -PassThru -WindowStyle Hidden
$pid = $proc.Id
"$pid" | Out-File -FilePath "$PSScriptRoot\..\mosquitto.pid" -Encoding ascii
Write-Host "Mosquitto 已启动，PID: $pid"

# 等待几秒钟让 Mosquitto 完全启动
Start-Sleep -Seconds 2

# 设置环境变量
$env:EDGEX_MQTT_BROKER = 'tcp://localhost:1883'
$env:EDGEX_MQTT_TOPIC  = 'edgex/events/core/#'
$env:EDGEX_CLIENT_ID   = 'local-test-client'

Write-Host "环境变量已设置: EDGEX_MQTT_BROKER=$env:EDGEX_MQTT_BROKER, EDGEX_MQTT_TOPIC=$env:EDGEX_MQTT_TOPIC"

# 进入项目目录并运行测试
Push-Location "$PSScriptRoot\.."
try {
    Write-Host "运行项目测试: go test ./..."
    & go test ./...
} finally {
    Pop-Location
}

# 停止 Mosquitto 进程
Write-Host "停止 Mosquitto 进程..."
if (Test-Path "$PSScriptRoot\..\mosquitto.pid") {
    try {
        $pid = Get-Content "$PSScriptRoot\..\mosquitto.pid" | Select-Object -First 1
        if ($pid -and (Get-Process -Id $pid -ErrorAction SilentlyContinue)) {
            Stop-Process -Id $pid -Force
            Write-Host "已停止 Mosquitto (PID $pid)"
        }
        Remove-Item "$PSScriptRoot\..\mosquitto.pid" -ErrorAction SilentlyContinue
    } catch {
        Write-Warning "通过 PID 停止失败: $_. 尝试通过进程名停止。"
        Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
    }
} else {
    Write-Host "PID 文件未找到，尝试通过进程名停止 mosquitto..."
    Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
}

Write-Host "测试完成，Mosquitto 已停止"
