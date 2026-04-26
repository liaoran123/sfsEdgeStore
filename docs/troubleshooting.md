# Troubleshooting

Common issues and solutions for sfsEdgeStore.

## MQTT Connection Issues

### Connection Refused

**Symptom:** MQTT client fails to connect to broker.

**Checklist:**
1. Verify MQTT broker is running:
   ```bash
   mosquitto -v
   ```
2. Check broker address in configuration:
   ```json
   "mqtt_broker": "tcp://localhost:1883"
   ```
3. Verify network connectivity:
   ```bash
   telnet localhost 1883
   ```
4. Check authentication credentials (if required):
   ```json
   "mqtt_username": "user",
   "mqtt_password": "password"
   ```

### TLS Connection Failure

**Symptom:** MQTT TLS handshake fails.

**Checklist:**
1. Verify certificate files exist and are readable:
   ```bash
   ls -la /etc/ssl/certs/ca.pem
   ls -la /etc/ssl/certs/server.pem
   ls -la /etc/ssl/private/server.key
   ```
2. Check certificate validity:
   ```bash
   openssl x509 -in /etc/ssl/certs/server.pem -text -noout
   ```

### No Data Received

**Symptom:** Connected but no readings stored.

**Checklist:**
1. Verify topic subscription matches EdgeX publication:
   ```bash
   curl http://localhost:8081/api/subscription/status
   ```
2. Test with sample EdgeX payload:
   ```bash
   curl -X POST http://localhost:8081/api/test-edgex -d @sample_edgex_payload.json
   ```
3. Check device configuration matches incoming data format

## Database Issues

### Database Initialization Failure

**Symptom:** Application exits with database error.

**Checklist:**
1. Verify directory permissions:
   ```bash
   ls -la data/sfs.db
   ```
2. Check disk space:
   ```bash
   df -h
   ```
3. Try clearing corrupted database:
   ```bash
   rm -rf data/
   ```

### Database Locked

**Note:** sfsEdgeStore uses sfsDb (LevelDB-based), which does not have SQLite's "database locked" issue. If you see access errors, check file permissions.

### Disk Space Full

**Symptom:** Database write failures.

**Solution:**
1. Check data retention policy:
   ```bash
   curl http://localhost:8081/api/retention/status
   ```
2. Manually trigger cleanup:
   ```bash
   curl -X POST http://localhost:8081/api/retention/cleanup
   ```
3. Enable automatic retention policy in configuration

## Memory Issues

### High Memory Usage

**Symptom:** Memory exceeds configured limit.

**Solution:**
1. Check current resource usage:
   ```bash
   curl http://localhost:8081/api/resources/status
   ```
2. Adjust memory limits in configuration:
   ```json
   "max_memory_mb": 512
   ```
3. Review data retention policy to prevent data accumulation
4. Consider reducing analyzer memory usage:
   ```json
   "analyzer_max_memory": 128
   ```

## HTTP Server Issues

### Port Already in Use

**Symptom:** HTTP server fails to start.

**Solution:**
1. Check if port is occupied:
   ```bash
   # Linux
   lsof -i :8081
   # Windows
   netstat -ano | findstr :8081
   ```
2. Change HTTP port in configuration:
   ```json
   "http_port": "8082"
   ```

### HTTPS Certificate Error

**Symptom:** HTTPS server fails to start.

**Checklist:**
1. Verify certificate and key paths are correct
2. Check certificate/key format:
   ```bash
   openssl x509 -in server.pem -text -noout
   openssl rsa -in server.key -check
   ```
3. Ensure key matches certificate

## Performance Issues

### Slow Query Response

**Solution:**
1. Use appropriate database scenario for your workload
2. Enable data retention to limit data size
3. Use time range filters in queries:
   ```bash
   curl "http://localhost:8081/api/readings?deviceName=Device001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-02T00:00:00Z"
   ```
4. Use limit parameter to restrict result size

### High CPU Usage

**Solution:**
1. Disable analyzer if not needed:
   ```json
   "enable_analyzer": false
   ```
2. Reduce monitoring frequency
3. Check for alert storm causing excessive notifications

## Logs

### View Logs

```bash
# Linux systemd
journalctl -u sfsedgestore -n 100

# Docker
docker logs sfsedgestore --tail 100

# Direct execution
./sfsedgestore 2>&1 | tee sfsedgestore.log
```

### Enable Verbose Logging

Run with debug output:
```bash
export GODEBUG=1
./sfsedgestore
```

## Getting Help

If the issue persists:

1. Check [GitHub Issues](https://github.com/liaoran123/sfsEdgeStore/issues) for similar problems
2. Create a new issue with:
   - sfsEdgeStore version
   - Operating system and architecture
   - Configuration (sensitive data redacted)
   - Log output
   - Steps to reproduce
