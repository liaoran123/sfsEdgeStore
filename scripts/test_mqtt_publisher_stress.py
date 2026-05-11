import paho.mqtt.client as mqtt
import json
import time
import random
import threading
from datetime import datetime

BROKER = "localhost"
PORT = 1883
CLIENT_ID_PREFIX = "sfsedgestore-stress-test"

devices = [
    "temperature-sensor-001",
    "temperature-sensor-002",
    "temperature-sensor-003",
    "humidity-sensor-001",
    "humidity-sensor-002",
    "humidity-sensor-003",
    "pressure-sensor-001",
    "pressure-sensor-002",
    "pressure-sensor-003",
    "vibration-sensor-001",
    "vibration-sensor-002",
    "vibration-sensor-003",
    "energy-meter-001",
    "energy-meter-002",
    "energy-meter-003",
    "flow-sensor-001",
    "flow-sensor-002",
    "level-sensor-001",
    "level-sensor-002",
    "ph-sensor-001",
]

resources = {
    "temperature-sensor": {"name": "Temperature", "min": -20.0, "max": 85.0, "unit": "°C"},
    "humidity-sensor": {"name": "Humidity", "min": 10.0, "max": 95.0, "unit": "%"},
    "pressure-sensor": {"name": "Pressure", "min": 800.0, "max": 1200.0, "unit": "hPa"},
    "vibration-sensor": {"name": "Vibration", "min": 0.0, "max": 50.0, "unit": "mm/s"},
    "energy-meter": {"name": "Energy", "min": 0.0, "max": 1000.0, "unit": "kWh"},
    "flow-sensor": {"name": "FlowRate", "min": 0.0, "max": 500.0, "unit": "L/min"},
    "level-sensor": {"name": "Level", "min": 0.0, "max": 10.0, "unit": "m"},
    "ph-sensor": {"name": "pH", "min": 0.0, "max": 14.0, "unit": "pH"},
}

exception_types = [
    "out_of_range_high",
    "out_of_range_low",
    "sudden_jump",
    "missing_value",
    "invalid_format",
    "duplicate_reading",
    "late_arrival",
]

def get_resource_info(device_name):
    for key, info in resources.items():
        if key in device_name:
            return info
    return {"name": "Unknown", "min": 0, "max": 100, "unit": ""}

def create_edgex_event(device_name, readings_data):
    timestamp = int(time.time() * 1000000000)
    readings = []
    for reading_name, value, is_exception, exception_type in readings_data:
        if is_exception:
            if exception_type == "out_of_range_high":
                range_info = next((r for r in resources.values() if r["name"] == reading_name), {})
                value = range_info.get("max", 100) * random.uniform(1.5, 3.0)
            elif exception_type == "out_of_range_low":
                range_info = next((r for r in resources.values() if r["name"] == reading_name), {})
                value = range_info.get("min", 0) * random.uniform(-1.0, 0.5)
            elif exception_type == "sudden_jump":
                value = random.uniform(1000, 50000)
            elif exception_type == "missing_value":
                value = ""
            elif exception_type == "invalid_format":
                value = "NaN"
            elif exception_type == "duplicate_reading":
                pass
            elif exception_type == "late_arrival":
                timestamp = timestamp - random.randint(3600, 86400) * 1000000000

        reading = {
            "id": f"reading-{random.randint(100000, 999999)}",
            "origin": timestamp,
            "deviceName": device_name,
            "resourceName": reading_name,
            "profileName": "device-profile",
            "valueType": "Float64" if not is_exception or exception_type not in ["missing_value", "invalid_format"] else "String",
            "value": str(value),
        }
        readings.append(reading)

    event = {
        "apiVersion": "v2",
        "id": f"event-{random.randint(100000, 999999)}",
        "deviceName": device_name,
        "profileName": "device-profile",
        "sourceName": readings[0]["resourceName"] if readings else "unknown",
        "origin": timestamp,
        "readings": readings,
    }
    return event

connected_flag = False
message_count = 0
exception_count = 0
lock = threading.Lock()

def on_connect(client, userdata, flags, rc, properties=None):
    global connected_flag
    if rc == 0:
        connected_flag = True
        print(f"客户端 {client._client_id.decode()} 已连接")
    else:
        print(f"连接失败，返回码: {rc}")

def publisher_worker(worker_id, interval_min, interval_max, exception_rate):
    global message_count, exception_count, connected_flag
    client_id = f"{CLIENT_ID_PREFIX}-worker-{worker_id}"
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id)
    client.on_connect = on_connect
    try:
        client.connect(BROKER, PORT, 60)
        client.loop_start()
        for _ in range(30):
            if connected_flag:
                break
            time.sleep(0.2)
        local_count = 0
        local_exceptions = 0
        while True:
            if not client.is_connected():
                try:
                    client.reconnect()
                    time.sleep(1)
                except:
                    time.sleep(2)
                    continue
            device = random.choice(devices)
            resource_info = get_resource_info(device)
            reading_name = resource_info["name"]
            num_readings = random.choices([1, 2, 3], weights=[60, 30, 10])[0]
            readings_data = []
            for _ in range(num_readings):
                is_exception = random.random() < exception_rate
                exception_type = random.choice(exception_types) if is_exception else None
                if not is_exception:
                    value = round(random.uniform(resource_info["min"], resource_info["max"]), 2)
                else:
                    value = resource_info["min"]
                    local_exceptions += 1
                readings_data.append((reading_name, value, is_exception, exception_type))
            event = create_edgex_event(device, readings_data)
            topic = f"edgex/events/core/{device}"
            payload = json.dumps({
                "messageType": "event",
                "payload": event
            })
            client.publish(topic, payload, qos=0)
            local_count += 1
            if local_count % 200 == 0:
                print(f"[Worker-{worker_id}] 已发送 {local_count} 条, 异常 {local_exceptions} 条")
            time.sleep(random.uniform(interval_min, interval_max))
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(f"Worker-{worker_id} 错误: {e}")
    finally:
        with lock:
            message_count += local_count
            exception_count += local_exceptions
        client.loop_stop()
        client.disconnect()

def main():
    global message_count, exception_count
    print("=" * 80)
    print("sfsEdgeStore 高压测试发布器")
    print("=" * 80)
    print(f"Broker: {BROKER}:{PORT}")
    print(f"设备数: {len(devices)}")
    print(f"工作线程数: 5")
    print(f"异常率: 8%")
    print(f"预计吞吐量: ~500-2000 条/分钟")
    print("=" * 80)
    print("")
    workers_config = [
        (1, 0.05, 0.15, 0.08),
        (2, 0.1, 0.25, 0.08),
        (3, 0.15, 0.35, 0.08),
        (4, 0.2, 0.5, 0.08),
        (5, 0.3, 0.8, 0.08),
    ]
    threads = []
    for config in workers_config:
        t = threading.Thread(target=publisher_worker, args=config, daemon=True)
        threads.append(t)
        t.start()
        time.sleep(0.1)
    try:
        while True:
            time.sleep(30)
            print(f"\n{'='*40} 全局统计 {'='*40}")
            print(f"总发送: {message_count} 条")
            print(f"异常数据: {exception_count} 条 ({exception_count/max(message_count,1)*100:.1f}%)")
            print(f"设备数: {len(devices)}")
            print(f"工作线程: {len(threads)}")
            print(f"{'='*90}\n")
    except KeyboardInterrupt:
        print("\n收到停止信号，等待工作线程结束...")
        time.sleep(2)
        print(f"\n测试结束，共发送 {message_count} 条消息，其中 {exception_count} 条异常数据")

if __name__ == "__main__":
    main()