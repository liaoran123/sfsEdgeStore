# 测试 sfsEdgeStore 启动时间
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "sfsEdgeStore 启动时间测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 清理旧进程
Get-Process -Name sfsEdgeStore -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1

Write-Host "启动程序..." -ForegroundColor Yellow

# 记录启动时间
$startTime = Get-Date

# 启动程序
$process = Start-Process -FilePath ".\sfsEdgeStore.exe" -PassThru -WindowStyle Hidden

Write-Host "程序已启动 (PID: $($process.Id))" -ForegroundColor Green

# 等待健康检查端点响应
$ready = $false
$maxWait = 30
$waitCount = 0

Write-Host "等待服务就绪..." -ForegroundColor Yellow

while (-not $ready -and $waitCount -lt $maxWait) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing -TimeoutSec 2 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            $ready = $true
        }
    } catch {
        # 忽略错误
    }
    
    if (-not $ready) {
        Start-Sleep -Milliseconds 200
        $waitCount++
    }
}

$endTime = Get-Date
$totalTime = ($endTime - $startTime).TotalSeconds

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "启动时间测试结果" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "启动时间: $([math]::Round($totalTime, 3)) 秒" -ForegroundColor Green
Write-Host "服务状态: $(if ($ready) { "✅ 就绪" } else { "❌ 超时" })" -ForegroundColor $(if ($ready) { "Green" } else { "Red" })
Write-Host "========================================" -ForegroundColor Cyan

# 保持程序运行
Write-Host ""
Write-Host "程序继续运行中..." -ForegroundColor Cyan
