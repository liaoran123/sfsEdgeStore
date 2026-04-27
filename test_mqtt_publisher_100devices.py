import paho.mqtt.client as mqtt
import json
import time
import random
import threading
import uuid
import sys

MQTT_BROKER = "localhost"
MQTT_PORT = 1883

DEVICE_TYPES = [
    {"type": "temperature", "topic": "edgex/events/device/temperature-sensor-{:03d}", "min": -40, "max": 125, "unit": "°C", "valueType": "Float64", "interval": (3, 10)},
    {"type": "humidity", "topic": "edgex/events/device/humidity-sensor-{:03d}", "min": 0, "max": 100, "unit": "%", "valueType": "Float64", "interval": (5, 15)},
    {"type": "pressure", "topic": "edgex/events/device/pressure-sensor-{:03d}", "min": 900, "max": 1100, "unit": "hPa", "valueType": "Float64", "interval": (5, 10)},
    {"type": "vibration", "topic": "edgex/events/device/vibration-sensor-{:03d}", "min": 0, "max": 10, "unit": "mm/s", "valueType": "Float64", "interval": (1, 3)},
    {"type": "energy", "topic": "edgex/events/device/energy-meter-{:03d}", "min": 0, "max": 100, "unit": "kWh", "valueType": "Float64", "interval": (10, 60)},
    {"type": "flow", "topic": "edgex/events/device/flow-sensor-{:03d}", "min": 0, "max": 100, "unit": "L/min", "valueType": "Float64", "interval": (3, 8)},
    {"type": "level", "topic": "edgex/events/device/level-sensor-{:03d}", "min": 0, "max": 100, "unit": "%", "valueType": "Float64", "interval": (5, 15)},
    {"type": "ph", "topic": "edgex/events/device/ph-sensor-{:03d}", "min": 0, "max": 14, "unit": "pH", "valueType": "Float64", "interval": (5, 10)},
    {"type": "co2", "topic": "edgex/events/device/co2-sensor-{:03d}", "min": 0, "max": 5000, "unit": "ppm", "valueType": "Float64", "interval": (5, 15)},
    {"type": "light", "topic": "edgex/events/device/light-sensor-{:03d}", "min": 0, "max": 10000, "unit": "lux", "valueType": "Float64", "interval": (3, 10)}
]

TOTAL_DEVICES = 100
ANOMALY_PROBABILITY = 0.05
INVALID_DATA_PROBABILITY = 0.02

device_states = {}
stats = {"sent": 0, "anomalies": 0, "invalid": 0, "errors": 0}
stats_lock = threading.Lock()

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
    if device_type['type'] in device_states:
        prev = device_states[device_type['type']]
        change = random.gauss(0, (device_type['max'] - device_type['min']) * 0.02)
        value = prev + change
    else:
        mid = (device_type['min'] + device_type['max']) / 2
        value = mid + random.gauss(0, (device_type['max'] - device_type['min']) * 0.1)
    
    return round(max(device_type['min'], min(device_type['max'], value)), 2)

def generate_anomaly_value(device_type, device_id):
    anomaly_type = random.choice(["out_of_range_high", "sudden_jump", "out_of_range_low"])
    
    if anomaly_type == "out_of_range_high":
        return round(device_type['max'] * random.uniform(1.2, 1.5), 2)
    elif anomaly_type == "out_of_range_low":
        return round(device_type['min'] * random.uniform(0.5, 0.8), 2)
    elif anomaly_type == "sudden_jump":
        if device_type['type'] in device_states:
            return round(device_states[device_type['type']] * random.uniform(2, 5), 2)
        else:
            return round(device_type['max'] * 2, 2)
    return generate_normal_value(device_type)

def generate_mqtt_message(device_type, device_id):
    global stats
    
    is_invalid = random.random() < INVALID_DATA_PROBABILITY
    is_anomaly = random.random() < ANOMALY_PROBABILITY
    
    if is_invalid:
        value_type = random.choice(["empty", "nan", "string"])
        if value_type == "empty":
            value = ""
        elif value_type == "nan":
            value = "NaN"
        else:
            value = "invalid_data"
        
        with stats_lock:
            stats["invalid"] += 1
    elif is_anomaly:
        value = generate_anomaly_value(device_type, device_id)
        with stats_lock:
            stats["anomalies"] += 1
    else:
        value = generate_normal_value(device_type)
    
    device_states[device_type['type']] = value if isinstance(value, (int, float)) else 0.0
    
    now_ns = int(time.time() * 1000000000)
    
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
    
    with stats_lock:
        stats["sent"] += 1
    
    return message

def device_thread(device_type, device_id):
    global stats
    
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id=f"publisher-{device_id}")
    
    try:
        client.connect(MQTT_BROKER, MQTT_PORT, 60)
        client.loop_start()
        
        topic = device_type["topic"].format(int(device_id.split("-")[-1]))
        min_interval, max_interval = device_type.get("interval", (1, 5))
        
        while True:
            try:
                message = generate_mqtt_message(device_type, device_id)
                result = client.publish(topic, json.dumps(message), qos=1)
                if result.rc != mqtt.MQTT_ERR_SUCCESS:
                    with stats_lock:
                        stats["errors"] += 1
            except Exception as e:
                with stats_lock:
                    stats["errors"] += 1
            
            time.sleep(random.uniform(min_interval, max_interval))
            
    except Exception as e:
        print(f"Error for device {device_id}: {e}")
    finally:
        try:
            client.loop_stop()
            client.disconnect()
        except:
            pass

def stats_printer():
    global stats
    start_time = time.time()
    
    while True:
        time.sleep(60)
        elapsed = time.time() - start_time
        with stats_lock:
            sent = stats["sent"]
            anomalies = stats["anomalies"]
            invalid = stats["invalid"]
            errors = stats["errors"]
        
        rate = sent / elapsed if elapsed > 0 else 0
        print(f"\n{'='*50}")
        print(f"  Statistics (after {elapsed/60:.1f} minutes)")
        print(f"  Total sent:     {sent}")
        print(f"  Anomalies:      {anomalies} ({anomalies/max(sent,1)*100:.1f}%)")
        print(f"  Invalid data:   {invalid} ({invalid/max(sent,1)*100:.1f}%)")
        print(f"  Errors:         {errors}")
        print(f"  Rate:           {rate:.1f} messages/sec")
        print(f"{'='*50}\n")

def main():
    print("=" * 60)
    print("  sfsEdgeStore Production-Ready MQTT Publisher")
    print("=" * 60)
    print(f"  MQTT Broker: {MQTT_BROKER}:{MQTT_PORT}")
    print(f"  Devices:     {TOTAL_DEVICES}")
    print(f"  Anomaly rate: {ANOMALY_PROBABILITY*100:.0f}%")
    print(f"  Invalid rate: {INVALID_DATA_PROBABILITY*100:.0f}%")
    print("=" * 60)
    print()
    
    devices = generate_device_ids()
    threads = []
    
    for device_type, device_id in devices:
        thread = threading.Thread(target=device_thread, args=(device_type, device_id))
        threads.append(thread)
        thread.start()
        time.sleep(0.05)
    
    stats_thread = threading.Thread(target=stats_printer, daemon=True)
    stats_thread.start()
    
    print(f"All {TOTAL_DEVICES} devices started successfully!")
    print(f"Press Ctrl+C to stop.\n")
    
    try:
        for thread in threads:
            thread.join()
    except KeyboardInterrupt:
        print("\nStopping all devices...")

if __name__ == "__main__":
    main()
