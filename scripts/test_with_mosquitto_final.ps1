<#
Final test script to test the project with Mosquitto
#>

param(
    [string]$BrokerExePath = 'C:\Program Files\Mosquitto\mosquitto.exe'
)

# Check if Mosquitto executable exists
if (-not (Test-Path -Path $BrokerExePath)) {
    Write-Error "Mosquitto executable not found: $BrokerExePath"
    exit 1
}

# Start Mosquitto in background
Write-Host "Starting Mosquitto in background..."
$proc = Start-Process -FilePath $BrokerExePath -ArgumentList '-v' -PassThru -WindowStyle Hidden
$mosquittoPID = $proc.Id
"$mosquittoPID" | Out-File -FilePath "$PSScriptRoot\..\mosquitto.pid" -Encoding ascii
Write-Host "Mosquitto started with PID: $mosquittoPID"

# Wait for Mosquitto to fully start
Start-Sleep -Seconds 2

# Set environment variables
$env:EDGEX_MQTT_BROKER = 'tcp://localhost:1883'
$env:EDGEX_MQTT_TOPIC  = 'edgex/events/core/#'
$env:EDGEX_CLIENT_ID   = 'local-test-client'

Write-Host "Environment variables set:"
Write-Host "  EDGEX_MQTT_BROKER=$env:EDGEX_MQTT_BROKER"
Write-Host "  EDGEX_MQTT_TOPIC=$env:EDGEX_MQTT_TOPIC"
Write-Host "  EDGEX_CLIENT_ID=$env:EDGEX_CLIENT_ID"

# Run tests
Write-Host "Running project tests: go test ./..."
& go test ./...

# Stop Mosquitto process
Write-Host "Stopping Mosquitto process..."
if (Test-Path "$PSScriptRoot\..\mosquitto.pid") {
    try {
        $pid = Get-Content "$PSScriptRoot\..\mosquitto.pid" | Select-Object -First 1
        if ($pid -and (Get-Process -Id $pid -ErrorAction SilentlyContinue)) {
            Stop-Process -Id $pid -Force
            Write-Host "Stopped Mosquitto (PID $pid)"
        }
        Remove-Item "$PSScriptRoot\..\mosquitto.pid" -ErrorAction SilentlyContinue
    } catch {
        Write-Warning "Failed to stop by PID: $_. Trying to stop by process name."
        Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
    }
} else {
    Write-Host "PID file not found, attempting to stop mosquitto by process name..."
    Get-Process mosquitto -ErrorAction SilentlyContinue | Stop-Process -Force
}

Write-Host "Test completed, Mosquitto stopped"
