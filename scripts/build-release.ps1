# sfsEdgeStore 发布包构建脚本
# 用于创建"绿色免安装"版本

$ErrorActionPreference = "Stop"

# 设置版本号
$version = "1.0.0"
$outputDir = "release"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   sfsEdgeStore 发布包构建工具" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 创建输出目录
if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir | Out-Null
}

Write-Host "`n[1/5] 编译 Windows 版本..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-s -w" -o sfsedgestore.exe
if ($LASTEXITCODE -ne 0) {
    Write-Host "编译失败！" -ForegroundColor Red
    exit 1
}
Write-Host "Windows 版本编译完成" -ForegroundColor Green

Write-Host "`n[2/5] 准备发布文件..." -ForegroundColor Yellow

# 创建临时目录
$tempDir = "sfsEdgeStore_$version"
if (Test-Path $tempDir) {
    Remove-Item -Recurse -Force $tempDir
}
New-Item -ItemType Directory -Path $tempDir | Out-Null

# 复制文件
Copy-Item sfsedgestore.exe $tempDir\
Copy-Item config.simple.json $tempDir\config.json
Copy-Item README.md $tempDir\
Copy-Item 启动.bat $tempDir\
if (Test-Path web) {
    Copy-Item -Recurse web $tempDir\
}

Write-Host "文件准备完成" -ForegroundColor Green

Write-Host "`n[3/5] 压缩发布包..." -ForegroundColor Yellow
$zipFile = "$outputDir\sfsEdgeStore_${version}_windows.zip"
Compress-Archive -Path $tempDir\* -DestinationPath $zipFile -Force

Write-Host "发布包创建完成: $zipFile" -ForegroundColor Green

Write-Host "`n[4/5] 编译 Linux 版本..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags "-s -w" -o sfsedgestore
if ($LASTEXITCODE -ne 0) {
    Write-Host "编译失败！" -ForegroundColor Red
    exit 1
}
Write-Host "Linux 版本编译完成" -ForegroundColor Green

Write-Host "`n[5/5] 准备 Linux 发布包..." -ForegroundColor Yellow
$linuxTempDir = "sfsEdgeStore_linux_$version"
if (Test-Path $linuxTempDir) {
    Remove-Item -Recurse -Force $linuxTempDir
}
New-Item -ItemType Directory -Path $linuxTempDir | Out-Null

# 复制文件
Copy-Item sfsedgestore $linuxTempDir\
Copy-Item config.simple.json $linuxTempDir\config.json
Copy-Item README.md $linuxTempDir\
Copy-Item start.sh $linuxTempDir\
if (Test-Path web) {
    Copy-Item -Recurse web $linuxTempDir\
}

$linuxZipFile = "$outputDir\sfsEdgeStore_${version}_linux.zip"
Compress-Archive -Path $linuxTempDir\* -DestinationPath $linuxZipFile -Force

Write-Host "Linux 发布包创建完成: $linuxZipFile" -ForegroundColor Green

# 清理临时文件
Remove-Item -Recurse -Force $tempDir
Remove-Item -Recurse -Force $linuxTempDir
Remove-Item sfsedgestore.exe
Remove-Item sfsedgestore

Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "   构建完成！" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "`nWindows 包: $zipFile" -ForegroundColor Cyan
Write-Host "Linux 包: $linuxZipFile" -ForegroundColor Cyan
Write-Host "`n使用方法：" -ForegroundColor Yellow
Write-Host " 1. 解压到任意目录"
Write-Host " 2. 编辑 config.json 配置 MQTT"
Write-Host " 3. 双击运行 sfsedgestore.exe (Linux: ./sfsedgestore"
Write-Host " 4. 浏览器打开 http://localhost:8081"
