# Security

Authentication, TLS, and authorization for sfsEdgeStore.

## Data Sovereignty

sfsEdgeStore follows a **local-first architecture** that ensures **full data sovereignty**:

- **No Cloud Dependency**: All data is stored locally on the edge device
- **No Cross-Border Transfer**: Data never leaves your premises, ensuring compliance with GDPR, EU Cyber Resilience Act, and other data residency regulations
- **Zero Third-Party Exposure**: No data is sent to external services unless explicitly configured
- **Complete Control**: You own the hardware, the software, and the data

This is a **key differentiator** for industries handling sensitive industrial data in Europe and other privacy-regulated regions.

## Encryption at Rest

### Database Encryption

sfsEdgeStore supports AES-256 encryption for data stored on disk:

```json
{
  "db_use_encryption": true,
  "db_encryption_key": "your-encryption-key",
  "db_encryption_algorithm": "aes-256-gcm"
}
```

**Why this matters for GDPR compliance:**
- Data is encrypted at rest on the edge device
- Even if the physical device is stolen, data remains unreadable
- You control the encryption keys - no third-party key management
- Encryption is optional and can be disabled for non-sensitive deployments

### Key Rotation

Rotate encryption key:

### API Key Authentication

sfsEdgeStore supports API key-based authentication for protecting sensitive endpoints.

#### Create API Key

```bash
curl -X POST http://localhost:8081/api/auth/create-key
```

#### List API Keys

```bash
curl http://localhost:8081/api/auth/list-keys
```

#### Revoke API Key

```bash
curl -X POST http://localhost:8081/api/auth/revoke-key
```

### RBAC (Role-Based Access Control)

sfsEdgeStore implements role-based access control with the following roles:

| Role | Permissions |
|------|-------------|
| Admin | Full access to all APIs |
| Viewer | Read-only access to data and metrics |
| Operator | Read access + backup/restore operations |

## TLS/SSL

### MQTT TLS

Secure MQTT communication:

```json
{
  "mqtt_use_tls": true,
  "mqtt_ca_cert": "/etc/ssl/certs/ca.pem",
  "mqtt_client_cert": "/etc/ssl/certs/client.pem",
  "mqtt_client_key": "/etc/ssl/private/client.key"
}
```

### HTTP TLS (HTTPS)

Secure HTTP communication:

```json
{
  "http_use_tls": true,
  "http_cert": "/etc/ssl/certs/server.pem",
  "http_key": "/etc/ssl/private/server.key"
}
```

## Database Encryption

Encrypt data at rest:

```json
{
  "db_use_encryption": true,
  "db_encryption_key": "your-encryption-key",
  "db_encryption_algorithm": "aes-256-gcm"
}
```

### Key Rotation

Rotate encryption key:

```bash
curl -X POST http://localhost:8081/api/encryption/rotate-key
```

### Check Encryption Status

```bash
curl http://localhost:8081/api/encryption/status
```

## Security Best Practices

1. **Enable TLS** - Use TLS for both MQTT and HTTP communication
2. **Use API Keys** - Protect sensitive API endpoints
3. **Encrypt Database** - Enable database encryption for sensitive data (required for GDPR compliance with industrial data)
4. **Rotate Keys** - Regularly rotate encryption and API keys
5. **Limit Access** - Use RBAC to restrict access based on role
6. **Monitor Logs** - Review logs for suspicious activity
7. **Keep Updated** - Apply security updates promptly
8. **Local-First Deployment** - Keep data on edge devices, avoid unnecessary cloud sync

## GDPR Compliance Checklist

For European deployments, ensure the following:

- [ ] Enable database encryption (`db_use_encryption: true`)
- [ ] Secure encryption keys in a safe location
- [ ] Enable TLS for all communication (MQTT + HTTP)
- [ ] Configure data retention policies (`enable_retention_policy: true`)
- [ ] Limit access to edge device hardware and network
- [ ] Document data processing activities (device types, data collected)
- [ ] Configure backup encryption for data at rest
- [ ] Review and restrict data sync to approved destinations only

## Network Security

### Firewall Rules

Only expose necessary ports:

| Port | Protocol | Purpose |
|------|----------|---------|
| 8081 | TCP | HTTP/HTTPS API |
| 1883 | TCP | MQTT (if running broker locally) |

### Production Checklist

- [ ] Enable HTTPS for HTTP server
- [ ] Enable TLS for MQTT connection
- [ ] Configure API keys for sensitive endpoints
- [ ] Enable database encryption
- [ ] Set resource limits (memory, CPU)
- [ ] Configure firewall rules
- [ ] Enable audit logging
