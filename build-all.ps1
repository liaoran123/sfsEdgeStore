# sfsEdgeStore 多平台构建脚本
# 用于本地构建所有平台的执行文件

Write-Host "=== sfsEdgeStore 多平台构建 ===" -ForegroundColor Green

# 定义需要构建的平台
$platforms = @(
    @{GOOS="linux"; GOARCH="amd64"}
    @{GOOS="linux"; GOARCH="arm64"}
    @{GOOS="windows"; GOARCH="amd64"}
    @{GOOS="windows"; GOARCH="arm64"}
    @{GOOS="darwin"; GOARCH="amd64"}
    @{GOOS="darwin"; GOARCH="arm64"}
)

# 创建输出目录
$outputDir = "release"
if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir | Out-Null
}

# 构建每个平台
foreach ($platform in $platforms) {
    $goos = $platform.GOOS
    $goarch = $platform.GOARCH

    $binaryName = "sfsedgestore-$goos-$goarch"
    if ($goos -eq "windows") {
        $binaryName += ".exe"
    }

    Write-Host "`n构建: $binaryName" -ForegroundColor Yellow

    $env:GOOS = $goos
    $env:GOARCH = $goarch
    $env:CGO_ENABLED = "0"

    go build -ldflags "-s -w" -o "$outputDir/$binaryName"

    if ($LASTEXITCODE -eq 0) {
        $size = (Get-Item "$outputDir/$binaryName").Length / 1MB
        Write-Host "  ✅ 成功 - $([math]::Round($size, 2)) MB" -ForegroundColor Green
    } else {
        Write-Host "  ❌ 失败" -ForegroundColor Red
    }
}

Write-Host "`n=== 构建完成 ===" -ForegroundColor Green
Write-Host "输出目录: $((Get-Item $outputDir).FullName)"
Write-Host "文件列表:"
Get-ChildItem $outputDir | ForEach-Object {
    $size = $_.Length / 1MB
    Write-Host "  - $($_.Name) - $([math]::Round($size, 2)) MB"
}
