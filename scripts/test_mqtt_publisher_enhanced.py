import paho.mqtt.client as mqtt
import json
import time
import random
from datetime import datetime

BROKER = "localhost"
PORT = 1883
CLIENT_ID = "sfsedgestore-test-publisher-enhanced"

devices = [
    "temperature-sensor-001",
    "humidity-sensor-001",
    "pressure-sensor-001",
    "vibration-sensor-001",
    "energy-meter-001"
]

resources = {
    "Temperature": {"min": 15.0, "max": 35.0, "unit": "°C"},
    "Humidity": {"min": 30.0, "max": 80.0, "unit": "%"},
    "Pressure": {"min": 980.0, "max": 1060.0, "unit": "hPa"},
    "Vibration": {"min": 0.0, "max": 10.0, "unit": "mm/s"},
    "Energy": {"min": 50.0, "max": 600.0, "unit": "kWh"}
}

exception_types = [
    "out_of_range_high",
    "out_of_range_low",
    "sudden_jump",
    "missing_value",
    "invalid_format"
]

def create_edgex_event(device_name, resource_name, value, is_exception=False, exception_type=None):
    timestamp = int(time.time() * 1000000000)
    if is_exception:
        if exception_type == "out_of_range_high":
            range_info = resources[resource_name]
            value = range_info["max"] * 2
        elif exception_type == "out_of_range_low":
            range_info = resources[resource_name]
            value = range_info["min"] * 0.5
        elif exception_type == "sudden_jump":
            value = random.uniform(1000, 10000)
        elif exception_type == "missing_value":
            value = ""
        elif exception_type == "invalid_format":
            value = "invalid_value"
    event = {
        "apiVersion": "v2",
        "id": f"event-{random.randint(100000, 999999)}",
        "deviceName": device_name,
        "profileName": "device-profile",
        "sourceName": resource_name,
        "origin": timestamp,
        "readings": [
            {
                "id": f"reading-{random.randint(100000, 999999)}",
                "origin": timestamp,
                "deviceName": device_name,
                "resourceName": resource_name,
                "profileName": "device-profile",
                "valueType": "Float64" if not is_exception or exception_type not in ["missing_value", "invalid_format"] else "String",
                "value": str(value)
            }
        ]
    }
    return event

connected_flag = False

def on_connect(client, userdata, flags, rc, properties=None):
    global connected_flag
    if rc == 0:
        connected_flag = True
        print(f"已连接到 MQTT Broker")
    else:
        print(f"连接失败，返回码: {rc}")

def main():
    global connected_flag
    print("启动增强版 MQTT 测试发布器...")
    print(f"Broker: {BROKER}:{PORT}")
    print("发送主题: edgex/events/core/{device}")
    print("包含正常数据和异常数据")
    print("")
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, CLIENT_ID)
    client.on_connect = on_connect
    try:
        client.connect(BROKER, PORT, 60)
        client.loop_start()
        for _ in range(30):
            if connected_flag:
                break
            time.sleep(0.2)
        if not connected_flag:
            print("连接 MQTT Broker 超时")
            return
        print("开始发送测试数据... (按 Ctrl+C 停止)")
        print("-" * 80)
        message_count = 0
        exception_count = 0
        while True:
            if not client.is_connected():
                print("MQTT 连接断开，尝试重新连接...")
                client.reconnect()
                time.sleep(2)
                continue
            device = random.choice(devices)
            if "temperature" in device:
                resource = "Temperature"
            elif "humidity" in device:
                resource = "Humidity"
            elif "pressure" in device:
                resource = "Pressure"
            elif "vibration" in device:
                resource = "Vibration"
            elif "energy" in device:
                resource = "Energy"
            else:
                resource = random.choice(list(resources.keys()))
            is_exception = random.random() < 0.1
            exception_type = None
            if is_exception:
                exception_type = random.choice(exception_types)
                exception_count += 1
            if not is_exception:
                range_info = resources[resource]
                value = round(random.uniform(range_info["min"], range_info["max"]), 2)
            else:
                value = 0
            event = create_edgex_event(device, resource, value, is_exception, exception_type)
            topic = f"edgex/events/core/{device}"
            wrapped_message = {
                "messageType": "event",
                "payload": event
            }
            payload = json.dumps(wrapped_message)
            client.publish(topic, payload, qos=0)
            message_count += 1
            if is_exception:
                print(f"[{message_count}] {device} - {resource}: {event['readings'][0]['value']} (异常: {exception_type})")
            else:
                print(f"[{message_count}] {device} - {resource}: {value} {resources[resource]['unit']}")
            time.sleep(random.uniform(2.0, 6.0))
            if message_count % 50 == 0:
                print("-" * 80)
                print(f"统计: 已发送 {message_count} 条消息，其中 {exception_count} 条异常数据")
                print("-" * 80)
    except KeyboardInterrupt:
        print("")
        print("收到停止信号")
    except Exception as e:
        print(f"错误: {e}")
    finally:
        client.loop_stop()
        client.disconnect()
        print(f"已断开连接，共发送 {message_count} 条消息，其中 {exception_count} 条异常数据")

if __name__ == "__main__":
    main()