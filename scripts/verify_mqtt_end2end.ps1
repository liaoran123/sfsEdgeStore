<#
End-to-end MQTT verification script for Windows PowerShell.

What it does:
- Starts mosquitto_sub to listen on a topic and redirects output to a temp file.
- Publishes a sample EdgeX MessageEnvelope JSON using mosquitto_pub.
- Waits up to TimeoutSec for the subscriber output to contain the published message.
- Stops the subscriber process and reports PASS/FAIL.

Usage example:
  # Basic run with defaults (assumes mosquitto_pub/sub in C:\Program Files\Mosquitto)
  .\verify_mqtt_end2end.ps1

  # Custom topic and timeout
  .\verify_mqtt_end2end.ps1 -Topic 'edgex/events/core/test' -TimeoutSec 10

Parameters:
- BrokerHost, BrokerPort: broker address (default localhost:1883)
- SubExePath, PubExePath: paths to mosquitto_sub.exe and mosquitto_pub.exe
- Topic: topic to publish/subscribe
- PayloadFile: optional path to a JSON payload file. If omitted, the script will create a temporary sample payload.
- TimeoutSec: how many seconds to wait for the subscriber to receive the message
#>

param(
    [string]$BrokerHost = 'localhost',
    [int]$BrokerPort = 1883,
    [string]$SubExePath = 'C:\Program Files\Mosquitto\mosquitto_sub.exe',
    [string]$PubExePath = 'C:\Program Files\Mosquitto\mosquitto_pub.exe',
    [string]$Topic = 'edgex/events/core/test',
    [string]$PayloadFile = '',
    [int]$TimeoutSec = 8
)

function Write-Info { Write-Host "[INFO]" $args }
function Write-Err  { Write-Host "[ERROR]" $args }

# Verify executables
if (-not (Test-Path -Path $SubExePath)) { Write-Err "mosquitto_sub not found at: $SubExePath"; exit 2 }
if (-not (Test-Path -Path $PubExePath)) { Write-Err "mosquitto_pub not found at: $PubExePath"; exit 2 }

# Prepare payload
$cleanupPayload = $false
if ($PayloadFile -and (Test-Path $PayloadFile)) {
    $payloadPath = $PayloadFile
    Write-Info "Using payload file: $payloadPath"
} else {
    $payloadPath = Join-Path $env:TEMP "payload_edgex_$(Get-Random).json"
    $nowNano = [int64]([DateTime]::UtcNow - [DateTime]'1970-01-01').TotalMilliseconds * 1000000

    $json = @'
{
  "correlationId": "test-$(Get-Random)",
  "messageType": "event",
  "origin": %ORIGIN%,
  "payload": {
    "id": "event-id-1",
    "deviceName": "device-001",
    "readings": [
      {
        "id": "reading-1",
        "resourceName": "temperature",
        "value": "25",
        "valueType": "Int32",
        "origin": %ORIGIN%,
        "baseType": "Int32"
      }
    ],
    "origin": %ORIGIN%
  }
}
'@
    $json = $json -replace '%ORIGIN%', $nowNano
    $json | Out-File -FilePath $payloadPath -Encoding utf8
    Write-Info "Wrote temporary payload to: $payloadPath"
    $cleanupPayload = $true
}

# Prepare subscriber output file
$outFile = Join-Path $env:TEMP "mosq_sub_out_$(Get-Random).log"

# Start mosquitto_sub and redirect output
$args = @('-h', $BrokerHost, '-p', $BrokerPort.ToString(), '-t', $Topic, '-v')
Write-Info "Starting subscriber: $SubExePath $($args -join ' ')"
$proc = Start-Process -FilePath $SubExePath -ArgumentList $args -RedirectStandardOutput $outFile -NoNewWindow -PassThru

# Give subscriber a moment to start
Start-Sleep -Seconds 1

# Publish payload
Write-Info "Publishing payload to topic '$Topic'"
& $PubExePath -h $BrokerHost -p $BrokerPort -t $Topic -f $payloadPath

# Wait for subscriber output to contain a recognizable piece of the message
$found = $false
for ($i=0; $i -lt $TimeoutSec; $i++) {
    Start-Sleep -Seconds 1
    if (Test-Path $outFile) {
        $content = Get-Content $outFile -Raw -ErrorAction SilentlyContinue
        if ($content -and $content -match 'device-001|reading-1|"temperature"') {
            $found = $true
            break
        }
    }
}

# Stop subscriber
try {
    if ($proc -and (Get-Process -Id $proc.Id -ErrorAction SilentlyContinue)) {
        Write-Info "Stopping subscriber (PID $($proc.Id))"
        Stop-Process -Id $proc.Id -Force
    } else {
        # try by name as fallback
        Get-Process -Name mosquitto_sub -ErrorAction SilentlyContinue | Stop-Process -Force
    }
} catch {
    Write-Err "Failed to stop subscriber: $_"
}

# Report results
if ($found) {
    Write-Host "`n========== RESULT: PASS ==========" -ForegroundColor Green
    Write-Host "Subscriber output (excerpt):`n"
    Get-Content $outFile -Tail 50 | ForEach-Object { Write-Host $_ }
    $exitCode = 0
} else {
    Write-Host "`n========== RESULT: FAIL ==========" -ForegroundColor Red
    Write-Host "Subscriber did not receive the expected message within $TimeoutSec seconds."
    if (Test-Path $outFile) {
        Write-Host "Subscriber output (last lines):"
        Get-Content $outFile -Tail 50 | ForEach-Object { Write-Host $_ }
    } else {
        Write-Host "Subscriber produced no output file. Ensure broker is reachable and topic is correct."
    }
    $exitCode = 3
}

# Cleanup temp files
if ($cleanupPayload -and (Test-Path $payloadPath)) { Remove-Item $payloadPath -ErrorAction SilentlyContinue }
if (Test-Path $outFile) { Remove-Item $outFile -ErrorAction SilentlyContinue }

exit $exitCode
