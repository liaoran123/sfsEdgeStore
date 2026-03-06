
# sfsEdgeStore 负载测试脚本
# 使用方法：.\load_test.ps1

param(
    [string]$Url = "http://localhost:8081",
    [int]$Concurrency = 10,
    [int]$Requests = 100,
    [string]$TestType = "health",
    [string]$DeviceName = "TestDevice-001",
    [string]$OutputJson = "",
    [switch]$UseGoTool = $false
)

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  sfsEdgeStore 负载测试工具" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

if ($UseGoTool) {
    Write-Host "使用 Go 负载测试工具..." -ForegroundColor Yellow
    $goToolPath = ".\cmd\loadtest\loadtest.exe"
    
    if (-not (Test-Path $goToolPath)) {
        Write-Host "正在编译 Go 负载测试工具..." -ForegroundColor Yellow
        go build -o $goToolPath .\cmd\loadtest\main.go
        if ($LASTEXITCODE -ne 0) {
            Write-Host "编译失败！" -ForegroundColor Red
            exit 1
        }
    }
    
    $args = @("-url", $Url, "-c", $Concurrency, "-n", $Requests, "-type", $TestType)
    if ($DeviceName) { $args += "-device", $DeviceName }
    if ($OutputJson) { $args += "-json", $OutputJson }
    
    &amp; $goToolPath @args
    exit $LASTEXITCODE
} else {
    Write-Host "使用 PowerShell 简单负载测试..." -ForegroundColor Yellow
    Write-Host "目标 URL: $Url" -ForegroundColor White
    Write-Host "并发数: $Concurrency" -ForegroundColor White
    Write-Host "总请求数: $Requests" -ForegroundColor White
    Write-Host "测试类型: $TestType" -ForegroundColor White
    Write-Host ""
    
    $successCount = 0
    $failCount = 0
    $latencies = @()
    $errors = @{}
    $startTime = Get-Date
    
    $semaphore = [System.Threading.SemaphoreSlim]::new($Concurrency, $Concurrency)
    $jobs = @()
    
    for ($i = 0; $i -lt $Requests; $i++) {
        $semaphore.Wait() | Out-Null
        
        $job = Start-Job -ScriptBlock {
            param($Url, $TestType, $DeviceName)
            
            $testUrl = $Url
            switch ($TestType) {
                "health" { $testUrl = "$Url/health" }
                "ready" { $testUrl = "$Url/ready" }
                "query" { $testUrl = "$Url/api/readings?deviceName=$DeviceName" }
            }
            
            $reqStart = Get-Date
            try {
                $response = Invoke-WebRequest -Uri $testUrl -Method Get -TimeoutSec 30 -UseBasicParsing
                $latency = (Get-Date) - $reqStart
                return @{
                    Success = $true
                    Latency = $latency.TotalMilliseconds
                    StatusCode = $response.StatusCode
                }
            } catch {
                $latency = (Get-Date) - $reqStart
                return @{
                    Success = $false
                    Latency = $latency.TotalMilliseconds
                    Error = $_.Exception.Message
                }
            }
        } -ArgumentList $Url, $TestType, $DeviceName
        
        $jobs += $job
        
        Register-ObjectEvent -InputObject $job -EventName StateChanged -Action {
            $semaphore.Release() | Out-Null
        } | Out-Null
    }
    
    Write-Host "等待所有请求完成..." -ForegroundColor Yellow
    
    $jobs | Wait-Job | Out-Null
    $endTime = Get-Date
    $totalDuration = $endTime - $startTime
    
    foreach ($job in $jobs) {
        $result = Receive-Job -Job $job
        Remove-Job -Job $job
        
        if ($result.Success) {
            $successCount++
            $latencies += $result.Latency
        } else {
            $failCount++
            $errMsg = $result.Error
            if ($errors.ContainsKey($errMsg)) {
                $errors[$errMsg]++
            } else {
                $errors[$errMsg] = 1
            }
        }
    }
    
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "  负载测试报告" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host ""
    
    $totalRequests = $successCount + $failCount
    Write-Host "总请求数:        $totalRequests" -ForegroundColor White
    Write-Host "成功:            $successCount" -ForegroundColor Green
    Write-Host "失败:            $failCount" -ForegroundColor Red
    
    if ($totalRequests -gt 0) {
        $successRate = [math]::Round(($successCount / $totalRequests) * 100, 2)
        Write-Host "成功率:          $successRate%" -ForegroundColor White
    }
    
    Write-Host "总耗时:          $($totalDuration.ToString('hh\:mm\:ss\.fff'))" -ForegroundColor White
    
    if ($latencies.Count -gt 0) {
        $avgLatency = [math]::Round(($latencies | Measure-Object -Average).Average, 2)
        $minLatency = [math]::Round(($latencies | Measure-Object -Minimum).Minimum, 2)
        $maxLatency = [math]::Round(($latencies | Measure-Object -Maximum).Maximum, 2)
        $rps = [math]::Round($totalRequests / $totalDuration.TotalSeconds, 2)
        
        Write-Host "平均延迟:        $avgLatency ms" -ForegroundColor White
        Write-Host "最小延迟:        $minLatency ms" -ForegroundColor White
        Write-Host "最大延迟:        $maxLatency ms" -ForegroundColor White
        Write-Host "请求/秒:         $rps" -ForegroundColor White
    }
    
    if ($errors.Count -gt 0) {
        Write-Host ""
        Write-Host "错误统计:" -ForegroundColor Red
        foreach ($err in $errors.Keys) {
            Write-Host "  - $err : $($errors[$err])" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    
    if ($OutputJson) {
        $jsonResult = @{
            total_requests = $totalRequests
            success_requests = $successCount
            failed_requests = $failCount
            success_rate = if ($totalRequests -gt 0) { [math]::Round(($successCount / $totalRequests) * 100, 2) } else { 0 }
            total_duration_ms = [math]::Round($totalDuration.TotalMilliseconds, 2)
            average_latency_ms = if ($latencies.Count -gt 0) { $avgLatency } else { 0 }
            min_latency_ms = if ($latencies.Count -gt 0) { $minLatency } else { 0 }
            max_latency_ms = if ($latencies.Count -gt 0) { $maxLatency } else { 0 }
            requests_per_second = $rps
            error_counts = $errors
        }
        $jsonResult | ConvertTo-Json -Depth 10 | Out-File -FilePath $OutputJson -Encoding utf8
        Write-Host "结果已保存到: $OutputJson" -ForegroundColor Green
    }
}
