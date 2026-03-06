# sfsEdgeStore 完整功能测试脚本

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "sfsEdgeStore 完整功能测试" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 健康检查
Write-Host "1. 测试健康检查端点..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method Get
    Write-Host "   ✓ /health 正常" -ForegroundColor Green
    Write-Host "     Status: $($health.status)"
    Write-Host "     Version: $($health.version)"
    
    $ready = Invoke-RestMethod -Uri "http://localhost:8081/ready" -Method Get
    Write-Host "   ✓ /ready 正常" -ForegroundColor Green
} catch {
    Write-Host "   ✗ 健康检查失败: $_" -ForegroundColor Red
}
Write-Host ""

# 2. 创建 API Key
Write-Host "2. 测试 API Key 管理..." -ForegroundColor Yellow
try {
    $createKeyBody = @{
        user_id = "full-test-admin"
        role = "admin"
        expires_in = 24
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/create-key" `
        -Method Post -Body $createKeyBody -ContentType "application/json"
    
    $apiKey = $response.api_key
    Write-Host "   ✓ API Key 创建成功" -ForegroundColor Green
    Write-Host "     API Key: $apiKey"
    
    $headers = @{ "X-API-Key" = $apiKey }
    
    $listKeys = Invoke-RestMethod -Uri "http://localhost:8081/api/auth/list-keys" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 列出 API Keys 成功 ($($listKeys.api_keys.Count) 个)" -ForegroundColor Green
} catch {
    Write-Host "   ✗ API Key 管理失败: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 3. 数据查询
Write-Host "3. 测试数据查询功能..." -ForegroundColor Yellow
try {
    $data = Invoke-RestMethod -Uri "http://localhost:8081/api/readings" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 查询所有数据成功" -ForegroundColor Green
    Write-Host "     总记录数: $($data.count)"
    
    $deviceData = Invoke-RestMethod -Uri "http://localhost:8081/api/readings?deviceName=temperature-sensor-001" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 按设备查询成功" -ForegroundColor Green
    Write-Host "     temperature-sensor-001: $($deviceData.count) 条记录"
} catch {
    Write-Host "   ✗ 数据查询失败: $_" -ForegroundColor Red
}
Write-Host ""

# 4. 数据导出
Write-Host "4. 测试数据导出功能..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "http://localhost:8081/api/data/export?format=json" `
        -Method Get -Headers $headers -OutFile "test_export.json"
    Write-Host "   ✓ JSON 导出成功 (test_export.json)" -ForegroundColor Green
    
    Invoke-RestMethod -Uri "http://localhost:8081/api/export/csv?path=test_export.csv" `
        -Method Get -Headers $headers
    Write-Host "   ✓ CSV 导出成功 (test_export.csv)" -ForegroundColor Green
} catch {
    Write-Host "   ✗ 数据导出失败: $_" -ForegroundColor Red
}
Write-Host ""

# 5. 数据备份
Write-Host "5. 测试数据备份..." -ForegroundColor Yellow
try {
    $backup = Invoke-RestMethod -Uri "http://localhost:8081/api/backup" `
        -Method Post -Headers $headers
    Write-Host "   ✓ 数据备份成功" -ForegroundColor Green
    Write-Host "     备份文件: $($backup.backupFile)"
} catch {
    Write-Host "   ✗ 数据备份失败: $_" -ForegroundColor Red
}
Write-Host ""

# 6. 告警测试
Write-Host "6. 测试告警功能..." -ForegroundColor Yellow
try {
    $testAlertBody = @{
        type = "full_test_alert"
        message = "Full test alert message"
        severity = "warning"
    } | ConvertTo-Json
    
    $alertResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/alerts/test" `
        -Method Post -Body $testAlertBody -Headers $headers -ContentType "application/json"
    Write-Host "   ✓ 测试告警发送成功" -ForegroundColor Green
    
    $notifierStatus = Invoke-RestMethod -Uri "http://localhost:8081/api/alerts/notifier/status" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 告警通知器状态查询成功" -ForegroundColor Green
} catch {
    Write-Host "   ✗ 告警功能测试失败: $_" -ForegroundColor Red
}
Write-Host ""

# 7. 数据保留策略
Write-Host "7. 测试数据保留策略..." -ForegroundColor Yellow
try {
    $retentionStatus = Invoke-RestMethod -Uri "http://localhost:8081/api/retention/status" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 保留策略状态查询成功" -ForegroundColor Green
    Write-Host "     已启用: $($retentionStatus.data.enabled)"
    Write-Host "     保留天数: $($retentionStatus.data.retention_days)"
} catch {
    Write-Host "   ✗ 数据保留策略测试失败: $_" -ForegroundColor Red
}
Write-Host ""

# 8. 资源监控
Write-Host "8. 测试资源监控..." -ForegroundColor Yellow
try {
    $resourceStatus = Invoke-RestMethod -Uri "http://localhost:8081/api/resources/status" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 资源状态查询成功" -ForegroundColor Green
    Write-Host "     内存使用: $([math]::Round($resourceStatus.data.memory_mb, 2)) MB"
    Write-Host "     CPU 使用: $($resourceStatus.data.cpu_percent)%"
    Write-Host "     Goroutines: $($resourceStatus.data.goroutines)"
} catch {
    Write-Host "   ✗ 资源监控测试失败: $_" -ForegroundColor Red
}
Write-Host ""

# 9. 配置管理
Write-Host "9. 测试配置管理..." -ForegroundColor Yellow
try {
    $config = Invoke-RestMethod -Uri "http://localhost:8081/api/config/get" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 配置查询成功" -ForegroundColor Green
    Write-Host "     HTTP 端口: $($config.data.http_port)"
    Write-Host "     数据库路径: $($config.data.db_path)"
    Write-Host "     模拟器已启用: $($config.data.enable_simulator)"
} catch {
    Write-Host "   ✗ 配置管理测试失败: $_" -ForegroundColor Red
}
Write-Host ""

# 10. 数据同步
Write-Host "10. 测试数据同步功能..." -ForegroundColor Yellow
try {
    $syncStatus = Invoke-RestMethod -Uri "http://localhost:8081/api/sync/status" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 同步状态查询成功" -ForegroundColor Green
    Write-Host "     已启用: $($syncStatus.data.enabled)"
} catch {
    Write-Host "   ✗ 数据同步测试失败: $_" -ForegroundColor Red
}
Write-Host ""

# 11. 最终数据统计
Write-Host "11. 最终数据统计..." -ForegroundColor Yellow
try {
    $finalData = Invoke-RestMethod -Uri "http://localhost:8081/api/readings" `
        -Method Get -Headers $headers
    Write-Host "   ✓ 最终数据统计" -ForegroundColor Green
    Write-Host "     总记录数: $($finalData.count)"
} catch {
    Write-Host "   ✗ 数据统计失败: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "完整测试完成！" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "测试 API Key: $apiKey" -ForegroundColor Cyan
Write-Host "服务器地址: http://localhost:8081" -ForegroundColor Cyan
Write-Host "模拟器运行中: 33 个设备持续生成数据" -ForegroundColor Cyan
