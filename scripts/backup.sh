#!/bin/bash

# sfsEdgeStore 自动备份脚本
# 使用方式: ./backup.sh [备份目录] [保留天数]

BACKUP_DIR=${1:-./backups}
RETENTION_DAYS=${2:-30}
DB_PATH="./edgex_data"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR" || exit 1

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log "开始 sfsEdgeStore 备份..."

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份文件名
BACKUP_FILENAME="sfsedgestore_backup_$(date '+%Y%m%d_%H%M%S').zip"
BACKUP_PATH="$BACKUP_DIR/$BACKUP_FILENAME"

log "备份文件: $BACKUP_PATH"

# 检查数据库目录是否存在
if [ ! -d "$DB_PATH" ]; then
    log "错误: 数据库目录不存在: $DB_PATH"
    exit 1
fi

# 创建备份
if command -v zip &> /dev/null; then
    zip -r "$BACKUP_PATH" "$DB_PATH"
    if [ $? -ne 0 ]; then
        log "错误: 备份创建失败"
        exit 1
    fi
elif command -v tar &> /dev/null; then
    tar -czf "$BACKUP_PATH" "$DB_PATH"
    if [ $? -ne 0 ]; then
        log "错误: 备份创建失败"
        exit 1
    fi
else
    log "错误: 未找到 zip 或 tar 命令"
    exit 1
fi

# 显示备份大小
BACKUP_SIZE=$(du -h "$BACKUP_PATH" | cut -f1)
log "备份完成，大小: $BACKUP_SIZE"

# 清理旧备份
log "清理 $RETENTION_DAYS 天前的备份..."
find "$BACKUP_DIR" -name "sfsedgestore_backup_*.zip" -type f -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "sfsedgestore_backup_*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete

log "备份任务完成!"
