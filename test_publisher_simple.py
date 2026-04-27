#!/usr/bin/env python3
"""Simple MQTT publisher for testing sfsEdgeStore"""

import json
import time
import random
import uuid

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("Error: paho-mqtt not installed. Run: pip install paho-mqtt")
    exit(1)

MQTT_BROKER = "localhost"
MQTT_PORT = 1883
MQTT_TOPIC = "edgex/events/device"

DEVICES = [
    {"name": "TemperatureSensor01", "type": "temperature", "unit": "C", "min": 20, "max": 85, "baseline": 45},
    {"name": "HumiditySensor01", "type": "humidity", "unit": "RH", "min": 30, "max": 95, "baseline": 60},
    {"name": "PressureSensor01", "type": "pressure", "unit": "hPa", "min": 950, "max": 1050, "baseline": 1013},
    {"name": "VibrationSensor01", "type": "vibration", "unit": "mm_s", "min": 0, "max": 50, "baseline": 5},
    {"name": "PowerMeter01", "type": "power", "unit": "kW", "min": 0, "max": 100, "baseline": 45},
]

def generate_message(device):
    value = device["baseline"] + random.uniform(-device["baseline"] * 0.1, device["baseline"] * 0.1)
    value = round(max(device["min"], min(device["max"], value)), 2)
    
    event_data = {
        "id": str(uuid.uuid4()),
        "deviceName": device["name"],
        "profileName": f"{device['type'].title()}Profile",
        "sourceName": device["type"],
        "origin": int(time.time() * 1000),
        "readings": [{
            "id": str(uuid.uuid4()),
            "origin": int(time.time() * 1000),
            "deviceName": device["name"],
            "resourceName": device["type"],
            "profileName": f"{device['type'].title()}Profile",
            "valueType": "Float64",
            "value": str(value),
            "units": device["unit"]
        }]
    }
    
    return json.dumps({
        "messageType": "event",
        "origin": int(time.time() * 1000),
        "payload": event_data,
        "contentType": "application/json"
    })

def main():
    print("=== sfsEdgeStore Simple Publisher ===")
    print(f"Broker: {MQTT_BROKER}:{MQTT_PORT}")
    print(f"Topic: {MQTT_TOPIC}")
    print(f"Devices: {len(DEVICES)}")
    print("Starting... (Ctrl+C to stop)\n")
    
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2)
    
    try:
        client.connect(MQTT_BROKER, MQTT_PORT, 60)
        client.loop_start()
        print("Connected to MQTT broker")
    except Exception as e:
        print(f"Failed to connect: {e}")
        return
    
    count = 0
    try:
        while True:
            device = random.choice(DEVICES)
            payload = generate_message(device)
            client.publish(MQTT_TOPIC, payload, qos=0)
            count += 1
            
            if count % 10 == 0:
                print(f"Sent {count} messages")
            
            time.sleep(random.uniform(0.5, 1.5))
    except KeyboardInterrupt:
        print(f"\nStopped. Total: {count} messages")
    finally:
        client.loop_stop()
        client.disconnect()
        print("Disconnected")

if __name__ == "__main__":
    main()
