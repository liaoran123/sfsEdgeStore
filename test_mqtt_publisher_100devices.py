import paho.mqtt.client as mqtt
import json
import time
import random
import threading
import uuid

MQTT_BROKER = "localhost"
MQTT_PORT = 1883

DEVICE_TYPES = [
    {"type": "temperature", "topic": "edgex/events/device/temperature-sensor-{:03d}", "min": -40, "max": 125, "unit": "°C", "valueType": "Float64"},
    {"type": "humidity", "topic": "edgex/events/device/humidity-sensor-{:03d}", "min": 0, "max": 100, "unit": "%", "valueType": "Float64"},
    {"type": "pressure", "topic": "edgex/events/device/pressure-sensor-{:03d}", "min": 900, "max": 1100, "unit": "hPa", "valueType": "Float64"},
    {"type": "vibration", "topic": "edgex/events/device/vibration-sensor-{:03d}", "min": 0, "max": 10, "unit": "mm/s", "valueType": "Float64"},
    {"type": "energy", "topic": "edgex/events/device/energy-meter-{:03d}", "min": 0, "max": 100, "unit": "kWh", "valueType": "Float64"},
    {"type": "flow", "topic": "edgex/events/device/flow-sensor-{:03d}", "min": 0, "max": 100, "unit": "L/min", "valueType": "Float64"},
    {"type": "level", "topic": "edgex/events/device/level-sensor-{:03d}", "min": 0, "max": 100, "unit": "%", "valueType": "Float64"},
    {"type": "ph", "topic": "edgex/events/device/ph-sensor-{:03d}", "min": 0, "max": 14, "unit": "pH", "valueType": "Float64"},
    {"type": "co2", "topic": "edgex/events/device/co2-sensor-{:03d}", "min": 0, "max": 5000, "unit": "ppm", "valueType": "Float64"},
    {"type": "light", "topic": "edgex/events/device/light-sensor-{:03d}", "min": 0, "max": 10000, "unit": "lux", "valueType": "Float64"}
]

TOTAL_DEVICES = 100
ANOMALY_PROBABILITY = 0.1

device_states = {}
clients = []

def generate_device_ids():
    devices = []
    device_count_per_type = TOTAL_DEVICES // len(DEVICE_TYPES)
    remaining_devices = TOTAL_DEVICES % len(DEVICE_TYPES)
    
    for i, device_type in enumerate(DEVICE_TYPES):
        count = device_count_per_type + (1 if i < remaining_devices else 0)
        for j in range(1, count + 1):
            device_id = f"{device_type['type']}-sensor-{j:03d}"
            devices.append((device_type, device_id))
    
    return devices

def generate_normal_value(device_type):
    return round(random.uniform(device_type["min"], device_type["max"]), 2)

def generate_anomaly_value(device_type, device_id):
    anomaly_type = random.choice(["out_of_range_high", "out_of_range_low", "sudden_jump"])
    
    if anomaly_type == "out_of_range_high":
        return round(device_type["max"] * 1.5, 2)
    elif anomaly_type == "out_of_range_low":
        return round(device_type["min"] * 1.5, 2)
    elif anomaly_type == "sudden_jump":
        if device_id in device_states:
            return round(device_states[device_id] * 10, 2)
        else:
            return round(device_type["max"] * 2, 2)
    return generate_normal_value(device_type)

def generate_mqtt_message(device_type, device_id):
    is_anomaly = random.random() < ANOMALY_PROBABILITY
    
    if is_anomaly:
        value = generate_anomaly_value(device_type, device_id)
    else:
        value = generate_normal_value(device_type)
    
    if not isinstance(value, (int, float)):
        value = 0.0
    
    device_states[device_id] = value
    
    now_ns = time.time_ns()
    
    reading = {
        "id": str(uuid.uuid4()),
        "origin": now_ns,
        "deviceName": device_id,
        "resourceName": device_type["type"],
        "profileName": f"{device_type['type']}-profile",
        "valueType": "Float64",
        "value": str(value),
        "baseType": "Simple"
    }
    
    event = {
        "id": str(uuid.uuid4()),
        "deviceName": device_id,
        "profileName": f"{device_type['type']}-profile",
        "sourceName": f"{device_type['type']}-source",
        "origin": now_ns,
        "readings": [reading]
    }
    
    message = {
        "apiVersion": "v2",
        "correlationId": str(uuid.uuid4()),
        "messageType": "event",
        "origin": now_ns,
        "payload": event
    }
    
    return message

def device_thread(device_type, device_id):
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION1, client_id=f"publisher-{device_id}")
    clients.append(client)
    
    try:
        client.connect(MQTT_BROKER, MQTT_PORT, 60)
        client.loop_start()
        
        topic = device_type["topic"].format(int(device_id.split("-")[-1]))
        
        print(f"Started device: {device_id} on topic: {topic}")
        
        while True:
            message = generate_mqtt_message(device_type, device_id)
            client.publish(topic, json.dumps(message), qos=1)
            time.sleep(random.uniform(0.1, 0.5))
            
    except Exception as e:
        print(f"Error for device {device_id}: {e}")
    finally:
        client.loop_stop()
        client.disconnect()

def main():
    print(f"Starting EdgeX MQTT publisher for {TOTAL_DEVICES} devices...")
    print(f"MQTT Broker: {MQTT_BROKER}:{MQTT_PORT}")
    
    devices = generate_device_ids()
    threads = []
    
    for device_type, device_id in devices:
        thread = threading.Thread(target=device_thread, args=(device_type, device_id))
        threads.append(thread)
        thread.start()
        time.sleep(0.05)
    
    print(f"All {TOTAL_DEVICES} devices started successfully!")
    
    try:
        for thread in threads:
            thread.join()
    except KeyboardInterrupt:
        print("Stopping all devices...")
        for client in clients:
            try:
                client.loop_stop()
                client.disconnect()
            except:
                pass

if __name__ == "__main__":
    main()
