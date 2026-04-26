# Backup & Restore

Data backup and recovery procedures.

## Backup

### Via API

```bash
curl -X POST "http://localhost:8081/api/backup?path=./backups"
```

Response:
```json
{
  "status": "success",
  "backupFile": "./backups/backup_2024-01-01_12-00-00.db"
}
```

### Via Script

Linux:
```bash
./scripts/backup.sh
```

Windows:
```powershell
.\scripts\backup.ps1
```

## Restore

### Via API

```bash
curl -X POST "http://localhost:8081/api/restore?file=./backups/backup_2024-01-01_12-00-00.db"
```

Response:
```json
{
  "status": "success",
  "message": "Database restored successfully"
}
```

## Automated Backup

Enable automatic backups in configuration:

```json
{
  "enable_auto_backup": true,
  "backup_interval_hours": 24,
  "backup_path": "./backups",
  "max_backup_count": 7
}
```

## Backup Strategy

| Scenario | Frequency | Retention |
|----------|-----------|-----------|
| Production | Daily | 7 days |
| Development | Weekly | 4 weeks |
| Testing | Manual | As needed |

## Best Practices

1. **Regular Backups** - Schedule automated daily backups
2. **Off-Site Storage** - Copy backups to separate location
3. **Test Restores** - Periodically verify backup integrity
4. **Monitor Backup Size** - Ensure sufficient disk space
