# sfsEdgeStore 自动备份脚本 (Windows PowerShell)
# 使用方式: .\backup.ps1 [-BackupDir <备份目录>] [-RetentionDays <保留天数>]

param(
    [string]$BackupDir = ".\backups",
    [int]$RetentionDays = 30
)

$DBPath = ".\edgex_data"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir

Set-Location $ProjectDir

function Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message"
}

Log "开始 sfsEdgeStore 备份..."

# 创建备份目录
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

# 备份文件名
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$BackupFilename = "sfsedgestore_backup_$timestamp.zip"
$BackupPath = Join-Path $BackupDir $BackupFilename

Log "备份文件: $BackupPath"

# 检查数据库目录是否存在
if (-not (Test-Path $DBPath)) {
    Log "错误: 数据库目录不存在: $DBPath"
    exit 1
}

# 创建备份
try {
    Compress-Archive -Path $DBPath -DestinationPath $BackupPath -Force
    Log "备份创建成功"
} catch {
    Log "错误: 备份创建失败 - $_"
    exit 1
}

# 显示备份大小
$BackupSize = (Get-Item $BackupPath).Length / 1MB
Log "备份完成，大小: $([math]::Round($BackupSize, 2)) MB"

# 清理旧备份
Log "清理 $RetentionDays 天前的备份..."
$cutoffDate = (Get-Date).AddDays(-$RetentionDays)
Get-ChildItem -Path $BackupDir -Filter "sfsedgestore_backup_*.zip" | 
    Where-Object { $_.LastWriteTime -lt $cutoffDate } | 
    Remove-Item -Force

Log "备份任务完成!"
