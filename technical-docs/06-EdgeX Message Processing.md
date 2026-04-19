# EdgeX Message Processing

## Overview

sfsEdgeStore receives data from EdgeX Foundry via MQTT protocol, parses EdgeX message format, and converts it to a format suitable for storage.

## Data Structures

### EdgeXMessage Structure

```go
// edgex/models.go:6-11
type EdgeXMessage struct {
	CorrelationID string          `json:"correlationId,omitempty"`
	MessageType   string          `json:"messageType,omitempty"`
	Origin        int64           `json:"origin,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}
```

**Field Description:**
- `CorrelationID`: Correlation ID for message tracking
- `MessageType`: Message type, must be "event" or "Event"
- `Origin`: Message timestamp (nanoseconds)
- `Payload`: Actual event data (JSON format)

### EdgeXEvent Structure

```go
// edgex/models.go:14-21
type EdgeXEvent struct {
	ID          string         `json:"id"`
	DeviceName  string         `json:"deviceName"`
	Readings    []EdgeXReading `json:"readings"`
	Origin      int64          `json:"origin"`
	ProfileName string         `json:"profileName,omitempty"`
	SourceName  string         `json:"sourceName,omitempty"`
}
```

**Field Description:**
- `ID`: Unique event identifier
- `DeviceName`: Device name (formatted to 64 characters)
- `Readings`: List of readings
- `Origin`: Event timestamp (nanoseconds)

### EdgeXReading Structure

```go
// edgex/models.go:24-33
type EdgeXReading struct {
	ID           string          `json:"id"`
	ResourceName string          `json:"resourceName"`
	Value        string          `json:"value"`
	ValueType    string          `json:"valueType,omitempty"`
	Origin       int64           `json:"origin"`
	ProfileName  string          `json:"profileName,omitempty"`
	DeviceName   string          `json:"deviceName,omitempty"`
	BaseType     string          `json:"baseType,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}
```

**Field Description:**
- `ResourceName`: Resource name (corresponds to reading name)
- `Value`: Reading value (string format)
- `ValueType`: Value type (e.g., Float32, Int64, String, etc.)
- `BaseType`: Base type
- `Metadata`: Metadata (JSON format)

## Message Processing

### ProcessMessage Function

```go
// edgex/processor.go:11-33
func ProcessMessage(payload []byte) (*EdgeXEvent, error) {
	var edgexMsg EdgeXMessage
	if err := json.Unmarshal(payload, &edgexMsg); err != nil {
		return nil, err
	}

	// Check if MessageType is "event"
	if edgexMsg.MessageType != "event" && edgexMsg.MessageType != "Event" {
		log.Printf("Ignoring message with type: %s", edgexMsg.MessageType)
		return nil, nil
	}

	// Parse event from payload
	var event EdgeXEvent
	if err := json.Unmarshal(edgexMsg.Payload, &event); err != nil {
		return nil, err
	}

	// Format device name from source to ensure 64-character length
	event.DeviceName = common.FormatDeviceName(event.DeviceName)

	return &event, nil
}
```

**Processing Flow:**
1. **Parse Outer Message**: Parse raw MQTT message into `EdgeXMessage` structure
2. **Validate Message Type**: Only process messages with MessageType "event" or "Event"
3. **Parse Event Data**: Parse `EdgeXEvent` from the Payload field
4. **Format Device Name**: Ensure device name is 64 characters for consistent indexing

## Device Name Formatting

### Formatting Logic

```go
// common/utils.go
func FormatDeviceName(deviceName string) string {
	const maxLength = 64
	if len(deviceName) >= maxLength {
		return deviceName[:maxLength]
	}
	
	// Pad with spaces to 64 characters
	return deviceName + strings.Repeat(" ", maxLength-len(deviceName))
}
```

**Design Reasons:**
- Using fixed-length device names ensures consistent primary key indexing
- Facilitates range queries by device name
- Optimizes database query performance

## Data Conversion

### EdgeX Reading to Storage Record

```go
func ConvertReadingsToRecords(event *EdgeXEvent) []*map[string]any {
	var records []*map[string]any

	for _, reading := range event.Readings {
		value, err := parseValue(reading.Value, reading.ValueType)
		if err != nil {
			log.Printf("Failed to parse value %s: %v", reading.Value, err)
			continue
		}

		record := map[string]any{
			"id":         reading.ID,
			"deviceName": event.DeviceName,
			"reading":    reading.ResourceName,
			"value":      value,
			"valueType":  reading.ValueType,
			"baseType":   reading.BaseType,
			"timestamp":  reading.Origin,
			"metadata":   string(reading.Metadata),
		}
		records = append(records, &record)
	}

	return records
}
```

## Examples

### Complete EdgeX Message Example

```json
{
  "correlationId": "abc-123",
  "messageType": "event",
  "origin": 1704067200000000000,
  "payload": {
    "id": "event-001",
    "deviceName": "Device001",
    "readings": [
      {
        "id": "reading-001",
        "resourceName": "temperature",
        "value": "25.5",
        "valueType": "Float32",
        "origin": 1704067200000000000,
        "baseType": "Float"
      },
      {
        "id": "reading-002",
        "resourceName": "humidity",
        "value": "60",
        "valueType": "Int32",
        "origin": 1704067200000000000,
        "baseType": "Int"
      }
    ]
  }
}
```

### Converted Storage Records

```go
[]*map[string]any{
  {
    "id":         "reading-001",
    "deviceName": "Device001                                                       ",
    "reading":    "temperature",
    "value":      25.5,
    "valueType":  "Float32",
    "baseType":   "Float",
    "timestamp":  1704067200000000000,
    "metadata":   "",
  },
  {
    "id":         "reading-002",
    "deviceName": "Device001                                                       ",
    "reading":    "humidity",
    "value":      60,
    "valueType":  "Int32",
    "baseType":   "Int",
    "timestamp":  1704067200000000000,
    "metadata":   "",
  },
}
```

## API Interface

### ProcessMessage Process EdgeX Message

```go
func ProcessMessage(payload []byte) (*EdgeXEvent, error)
```

**Parameters:**
- `payload`: Raw MQTT message byte array

**Return Values:**
- `*EdgeXEvent`: Parsed event object
- `error`: Error information

**Processing Flow:**
1. Parse JSON message
2. Validate message type
3. Parse event data
4. Format device name
5. Return event object