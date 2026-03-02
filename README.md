# sfsDb EdgeX Adapter

## Overview
The sfsDb EdgeX Adapter provides seamless integration between EdgeX Foundry and sfsDb embedded database, enabling efficient storage and retrieval of device data at the edge. This adapter follows EdgeX Foundry best practices and standards for edge computing solutions.

## Features
- **EdgeX Foundry Compatible**: Implements EdgeX MessageEnvelope format and MQTT message bus integration
- **Efficient Storage**: Uses sfsDb embedded database for lightweight, high-performance data storage
- **Real-time Processing**: Processes and stores EdgeX events in real-time
- **Automatic Database Management**: Automatically initializes database and creates optimized indexes
- **Configurable**: Supports configuration via environment variables and config file
- **Health Monitoring**: Provides HTTP health check endpoint
- **Data Backup**: Includes backup and restore functionality for data persistence

## Prerequisites
- **EdgeX Foundry**: v3.0.0 or later
- **MQTT Broker**: Mosquitto 2.0+ or compatible MQTT broker
- **Go**: 1.25 or higher
- **sfsDb**: v1.0.0 or later
- **Operating System**: Linux, macOS, or Windows

## Quick Start

### Build from Source
```bash
git clone https://github.com/your-org/sfsdb-edgex-adapter.git
cd sfsdb-edgex-adapter
go build
```

### Run the Adapter
```bash
./sfsdb-edgex-adapter
```

### Default Configuration
By default, the adapter will:
- Connect to MQTT broker at `tcp://localhost:1883`
- Subscribe to topic `edgex/events/core/#`
- Store data in `./edgex_data` directory
- Create table `edgex_readings` with optimized indexes
- Start HTTP health check server on port 8080

## Configuration

### EdgeX Configuration Standards
The adapter follows EdgeX Foundry configuration standards, supporting multiple configuration sources:

1. **Environment Variables** (highest priority)
2. **Configuration File** (`config.json`)
3. **Default Values** (lowest priority)

### Environment Variables
- `EDGEX_DB_PATH` - Database storage path
- `EDGEX_MQTT_BROKER` - MQTT broker address
- `EDGEX_MQTT_TOPIC` - MQTT topic to subscribe to
- `EDGEX_CLIENT_ID` - MQTT client ID

### Configuration File Example
```json
{
  "db_path": "./edgex_data",
  "mqtt_broker": "tcp://localhost:1883",
  "mqtt_topic": "edgex/events/core/#",
  "client_id": "sfsdb-edgex-adapter"
}
```

## EdgeX Integration

### Message Format
The adapter processes EdgeX messages in the standard **MessageEnvelope** format as defined by EdgeX Foundry:

```json
{
  "correlationId": "5e8a3c9d-1b2c-4d5e-8f9a-1b2c3d4e5f6g",
  "messageType": "event",
  "origin": 1677721600000000000,
  "payload": {
    "id": "event-123",
    "deviceName": "Thermostat-001",
    "readings": [
      {
        "id": "reading-123",
        "resourceName": "temperature",
        "value": "25.5",
        "origin": 1677721600000000000,
        "profileName": "ThermostatProfile",
        "deviceName": "Thermostat-001",
        "metadata": {"unit": "Celsius"}
      }
    ],
    "origin": 1677721600000000000,
    "profileName": "ThermostatProfile",
    "sourceName": "ThermostatSource"
  }
}
```

### Data Storage
Data is stored in the `edgex_readings` table with the following schema, optimized for EdgeX data patterns:

| Field | Type | Description |
|-------|------|-------------|
| `id` | String | Unique reading ID |
| `deviceName` | String | Device name (part of composite primary key) |
| `reading` | String | Resource name (e.g., temperature, humidity) |
| `value` | Float | Reading value |
| `timestamp` | Integer | Timestamp in seconds (part of composite primary key) |
| `metadata` | String | JSON metadata |

### Indexes
- **Composite Primary Key**: `(deviceName, timestamp)` for efficient time-range queries
- **Time Index**: For time-based filtering

### Query Optimization
The adapter implements efficient query functionality through the `queryReadings` function, which:
- Supports filtering by device name and time range
- Utilizes the composite primary key for optimized range queries
- Parses RFC3339 format timestamps for time-based filtering
- Returns structured results in a consistent format

**Query Parameters**:
- `deviceName`: Optional device name filter
- `startTime`: Optional start time (RFC3339 format)
- `endTime`: Optional end time (RFC3339 format)

**Example Query**:
```bash
# Query all readings for a specific device
GET /api/readings?deviceName=Thermostat-001

# Query readings for a device within a time range
GET /api/readings?deviceName=Thermostat-001&startTime=2024-01-01T00:00:00Z&endTime=2024-01-02T00:00:00Z
```

## API

### Health Check Endpoint
- **URL**: `/health`
- **Method**: GET
- **Response**: JSON status of the adapter

Example response:
```json
{
  "status": "healthy",
  "components": {
    "database": "connected",
    "mqtt": "connected",
    "adapter": "running"
  }
}
```

### Readings Query Endpoint
- **URL**: `/api/readings`
- **Method**: GET
- **Parameters**:
  - `deviceName` (optional): Filter by device name
  - `startTime` (optional): Start time in RFC3339 format
  - `endTime` (optional): End time in RFC3339 format
- **Response**: JSON array of readings

Example response:
```json
[
  {
    "id": "reading-123",
    "deviceName": "Thermostat-001",
    "reading": "temperature",
    "value": 25.5,
    "timestamp": 1677721600,
    "metadata": "{\"unit\": \"Celsius\"}"
  }
]
```

### Backup API Endpoint
- **URL**: `/api/backup`
- **Method**: POST
- **Parameters**:
  - `path` (optional): Backup storage path (default: `./backups`)
- **Response**: JSON object with backup status and file path

Example request:
```bash
# Create backup with default path
POST /api/backup

# Create backup with custom path
POST /api/backup?path=/path/to/backups
```

Example response:
```json
{
  "status": "success",
  "backupFile": "./backups/backup_20240101_120000"
}
```

### Restore API Endpoint
- **URL**: `/api/restore`
- **Method**: POST
- **Parameters**:
  - `file` (required): Backup file path
- **Response**: JSON object with restore status

Example request:
```bash
# Restore from backup file
POST /api/restore?file=./backups/backup_20240101_120000
```

Example response:
```json
{
  "status": "success",
  "message": "Database restored successfully"
}
```

## Monitoring

### Logging
The adapter follows EdgeX Foundry logging standards, providing structured logs for:
- Configuration loading
- Database initialization and operations
- MQTT connection status
- Message processing
- Error conditions and warnings

### Metrics
The adapter exposes the following metrics for monitoring:
- Message processing rate
- Database operation latency
- MQTT connection status
- Storage utilization

## Backup and Restore

### Backup
To create a backup of the database:
```bash
# Backup functionality is integrated into the adapter
# See backup package for details
```

### Restore
To restore from a backup:
```bash
# Restore functionality is integrated into the adapter
# See backup package for details
```

## Troubleshooting

### Common Issues

1. **MQTT Connection Failed**
   - Ensure MQTT broker is running
   - Verify broker address and port configuration
   - Check network connectivity

2. **Database Initialization Failed**
   - Ensure database directory is writable
   - Check file system permissions
   - Verify sufficient disk space

3. **Message Processing Errors**
   - Verify message format matches EdgeX MessageEnvelope standard
   - Check log output for detailed error messages
   - Ensure EdgeX Foundry version compatibility

4. **Performance Issues**
   - Consider increasing MQTT message queue size
   - Verify database directory is on fast storage
   - Monitor system resources (CPU, memory, disk I/O)

## Version Compatibility

| Component | Version |
|-----------|---------|
| EdgeX Foundry | v3.0.0+ |
| Go | 1.25+ |
| sfsDb | v1.0.0+ |
| MQTT Broker | Mosquitto 2.0+ |

## Security

### Best Practices
- Use secure MQTT connections (TLS)
- Implement proper access controls for the database
- Regularly update dependencies
- Follow EdgeX Foundry security guidelines

## Deployment

### Containerization
The adapter can be containerized using Docker:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o sfsdb-edgex-adapter

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sfsdb-edgex-adapter .
COPY config.json .
EXPOSE 8080
CMD ["./sfsdb-edgex-adapter"]
```

### EdgeX Foundry Deployment
The adapter can be deployed as part of an EdgeX Foundry instance using Docker Compose:

```yaml
version: '3.8'
services:
  sfsdb-edgex-adapter:
    image: sfsdb-edgex-adapter:latest
    depends_on:
      - mqtt-broker
    environment:
      - EDGEX_MQTT_BROKER=tcp://mqtt-broker:1883
    volumes:
      - ./edgex_data:/app/edgex_data
    ports:
      - "8080:8080"
```

## Testing

### Unit Tests
```bash
go test -v ./...
```

### Integration Tests
```bash
go test -v -run TestIntegration
```

### EdgeX Compatibility Tests
The adapter has been tested with EdgeX Foundry v3.0.0+ to ensure compatibility with the EdgeX message format and integration patterns.

## Contributing

### EdgeX Contribution Guidelines
This project follows EdgeX Foundry contribution guidelines. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute.

### Code Style
- Follow Go code style standards
- Use EdgeX Foundry recommended patterns
- Include comprehensive test coverage
- Document all public APIs

## License

This project is licensed under the [Apache 2.0 License](LICENSE), consistent with EdgeX Foundry licensing requirements.

## Support

### Community Support
- EdgeX Foundry community forums
- GitHub issues for bug reports and feature requests

### Commercial Support
For commercial support options, please contact your EdgeX Foundry vendor or solution provider.

## Roadmap

### Future Enhancements
- Support for EdgeX Foundry v4.0.0
- Advanced data compression and retention policies
- Integration with EdgeX security services
- Enhanced monitoring and analytics capabilities
- Support for additional database backends