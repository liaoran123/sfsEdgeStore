import paho.mqtt.client as mqtt
import json
import time
import random

# MQTT 代理设置
MQTT_BROKER = "localhost"
MQTT_PORT = 1883

# 设备类型和配置
DEVICE_TYPES = [
    {"type": "temperature", "topic": "devices/temperature-sensor-{:03d}/data", "min": -40, "max": 125, "unit": "°C"},
    {"type": "humidity", "topic": "devices/humidity-sensor-{:03d}/data", "min": 0, "max": 100, "unit": "%"},
    {"type": "pressure", "topic": "devices/pressure-sensor-{:03d}/data", "min": 900, "max": 1100, "unit": "hPa"},
    {"type": "vibration", "topic": "devices/vibration-sensor-{:03d}/data", "min": 0, "max": 10, "unit": "mm/s"},
    {"type": "energy", "topic": "devices/energy-meter-{:03d}/data", "min": 0, "max": 100, "unit": "kWh"}
]

# 设备数量
TOTAL_DEVICES = 100

# 异常数据类型
ANOMALY_TYPES = [
    "out_of_range_high",  # 超出范围（高）
    "out_of_range_low",   # 超出范围（低）
    "sudden_jump",        # 突然跳变
    "missing_value",      # 缺失值
    "invalid_format"      # 无效格式
]

# 异常数据概率
ANOMALY_PROBABILITY = 0.1  # 10%

# 设备状态记录
device_states = {}

# 生成设备列表
def generate_devices():
    devices = []
    device_count_per_type = TOTAL_DEVICES // len(DEVICE_TYPES)
    remaining_devices = TOTAL_DEVICES % len(DEVICE_TYPES)

    for i, device_type in enumerate(DEVICE_TYPES):
        count = device_count_per_type + (1 if i < remaining_devices else 0)
        for j in range(1, count + 1):
            device_id = f"{device_type['type']}-sensor-{j:03d}"
            devices.append((device_type, device_id))

    return devices

# 生成正常数据
def generate_normal_value(device_type):
    min_val = device_type["min"]
    max_val = device_type["max"]
    return round(random.uniform(min_val, max_val), 2)

# 生成异常数据
def generate_anomaly_value(device_type, device_id):
    anomaly_type = random.choice(ANOMALY_TYPES)

    if anomaly_type == "out_of_range_high":
        return device_type["max"] * 2
    elif anomaly_type == "out_of_range_low":
        return device_type["min"] * 2
    elif anomaly_type == "sudden_jump":
        # 突然跳变 10 倍
        if device_id in device_states:
            return device_states[device_id] * 10
        else:
            return device_type["max"] * 2
    elif anomaly_type == "missing_value":
        return None
    elif anomaly_type == "invalid_format":
        return "invalid_value"

# 生成 MQTT 消息
def generate_mqtt_message(device_type, device_id):
    # 决定是否生成异常数据
    is_anomaly = random.random() < ANOMALY_PROBABILITY

    if is_anomaly:
        value = generate_anomaly_value(device_type, device_id)
    else:
        value = generate_normal_value(device_type)

    # 更新设备状态
    if value is not None and isinstance(value, (int, float)):
        device_states[device_id] = value

    # 生成 EdgeX 格式的消息（符合 EdgeXMessage 结构）
    event = {
        "id": f"event-{int(time.time() * 1000000)}",
        "deviceName": device_id,
        "profileName": f"{device_type['type']}-profile",
        "sourceName": f"{device_type['type']}-source",
        "origin": int(time.time() * 1000000000),
        "readings": [
            {
                "id": f"reading-{int(time.time() * 1000000)}",
                "origin": int(time.time() * 1000000000),
                "deviceName": device_id,
                "resourceName": device_type["type"],
                "profileName": f"{device_type['type']}-profile",
                "valueType": "Float64" if value is not None and isinstance(value, (int, float)) else "String",
                "value": str(value) if value is not None else ""
            }
        ]
    }

    # 外层 EdgeXMessage 结构
    message = {
        "messageType": "event",
        "origin": int(time.time() * 1000000000),
        "payload": event
    }

    return message

# 主函数
def main():
    print(f"Starting MQTT publisher for {TOTAL_DEVICES} devices...")
    print(f"MQTT Broker: {MQTT_BROKER}:{MQTT_PORT}")

    # 连接 MQTT 客户端
    client = mqtt.Client(client_id="publisher-100devices", callback_api_version=mqtt.CallbackAPIVersion.VERSION2)

    try:
        client.connect(MQTT_BROKER, MQTT_PORT, 60)
        client.loop_start()

        devices = generate_devices()
        print(f"All {TOTAL_DEVICES} devices initialized successfully!")
        print(f"PRODUCTION MODE: ~200 msg/s (real-world high load)")
        print(f"Simulating 100 devices with 500ms reporting interval")
        print(f"Sending data from {TOTAL_DEVICES} devices...")

        # 预生成消息模板（减少 CPU 消耗）
        message_templates = []
        for device_type, device_id in devices:
            topic = device_type["topic"].format(int(device_id.split("-")[-1]))
            # 生成基础消息模板
            value = round(random.uniform(device_type["min"], device_type["max"]), 2)
            message_template = {
                "messageType": "event",
                "origin": int(time.time() * 1000000000),
                "payload": {
                    "id": f"event-{device_id}",
                    "deviceName": device_id,
                    "profileName": f"{device_type['type']}-profile",
                    "sourceName": f"{device_type['type']}-source",
                    "origin": int(time.time() * 1000000000),
                    "readings": [
                        {
                            "id": f"reading-{device_id}",
                            "origin": int(time.time() * 1000000000),
                            "deviceName": device_id,
                            "resourceName": device_type["type"],
                            "profileName": f"{device_type['type']}-profile",
                            "valueType": "Float64",
                            "value": str(value)
                        }
                    ]
                }
            }
            message_templates.append((topic, message_template, value))

        batch_count = 0
        start_time = time.time()

        while True:
            # 生产环境：所有设备每 500ms 上报一次（模拟真实场景的高负载）
            for topic, template, base_value in message_templates:
                # 简单更新 value（±5% 波动）
                new_value = round(base_value * (1 + random.uniform(-0.05, 0.05)), 2)
                template["payload"]["readings"][0]["value"] = str(new_value)
                template["payload"]["origin"] = int(time.time() * 1000000000)
                template["origin"] = int(time.time() * 1000000000)
                
                client.publish(topic, json.dumps(template), qos=1)
            
            batch_count += 1
            elapsed = time.time() - start_time
            
            # 每 10 批显示一次统计（约每 5 秒显示一次）
            if batch_count % 10 == 0:
                msg_count = batch_count * TOTAL_DEVICES
                rate = msg_count / elapsed
                print(f"  Sent {msg_count} messages in {elapsed:.1f}s ({rate:.0f} msg/s)")

            # 生产环境：500ms 间隔，约 200 msg/s
            time.sleep(0.5)

    except Exception as e:
        print(f"Error: {e}")
    finally:
        client.loop_stop()
        client.disconnect()

if __name__ == "__main__":
    main()