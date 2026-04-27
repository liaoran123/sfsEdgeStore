#!/usr/bin/env python3
"""
sfsEdgeStore MQTT 模拟数据发布器
用于向 Mosquitto MQTT 代理发送模拟的设备数据，测试 sfsEdgeStore 的数据接收和处理能力。
"""

import json
import time
import random
import uuid
import sys
from datetime import datetime, timezone

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("Error: paho-mqtt package not installed.")
    print("Install it with: pip install paho-mqtt")
    exit(1)

# MQTT 配置
MQTT_BROKER = "localhost"
MQTT_PORT = 1883
MQTT_TOPIC = "edgex/events/device"

# 设备配置
DEVICES = [
    {"name": "TemperatureSensor01", "type": "temperature", "unit": "°C", "min": 20.0, "max": 85.0, "baseline": 45.0},
    {"name": "TemperatureSensor02", "type": "temperature", "unit": "°C", "min": 18.0, "max": 90.0, "baseline": 50.0},
    {"name": "HumiditySensor01", "type": "humidity", "unit": "%RH", "min": 30.0, "max": 95.0, "baseline": 60.0},
    {"name": "PressureSensor01", "type": "pressure", "unit": "hPa", "min": 950.0, "max": 1050.0, "baseline": 1013.0},
    {"name": "VibrationSensor01", "type": "vibration", "unit": "mm/s", "min": 0.0, "max": 50.0, "baseline": 5.0},
    {"name": "PowerMeter01", "type": "power", "unit": "kW", "min": 0.0, "max": 100.0, "baseline": 45.0},
]

# 异常数据类型
ANOMALY_TYPES = [
    {"name": "spike_high", "probability": 0.05, "multiplier": 2.5},
    {"name": "spike_low", "probability": 0.03, "multiplier": 0.1},
    {"name": "flatline", "probability": 0.02, "value": 0.0},
]

def generate_reading(device, anomaly=None):
    """生成传感器读数"""
    baseline = device["baseline"]
    
    if anomaly == "spike_high":
        value = baseline * 2.5
    elif anomaly == "spike_low":
        value = baseline * 0.1
    elif anomaly == "flatline":
        value = 0.0
    else:
        # 正常波动
        value = baseline + random.uniform(-baseline * 0.1, baseline * 0.1)
    
    value = max(device["min"], min(device["max"], value))
    
    return {
        "name": device["type"],
        "value": round(value, 2),
        "unit": device["unit"]
    }

def generate_edgex_event(device, anomaly=None):
    """生成 EdgeX V2 格式的事件（包装为 MessageEnvelope）"""
    reading = generate_reading(device, anomaly)
    
    # 内部事件数据
    event_data = {
        "apiVersion": "v2",
        "id": str(uuid.uuid4()),
        "deviceName": device["name"],
        "profileName": f"{device['type'].title()}Profile",
        "sourceName": device["type"],
        "origin": int(time.time() * 1000),
        "readings": [
            {
                "id": str(uuid.uuid4()),
                "origin": int(time.time() * 1000),
                "deviceName": device["name"],
                "resourceName": device["type"],
                "profileName": f"{device['type'].title()}Profile",
                "valueType": "Float64",
                "value": str(reading["value"]),
                "units": reading["unit"]
            }
        ],
        "tags": {
            "gateway": "sfsEdgeStore",
            "location": "factory-floor-1"
        }
    }
    
    # 包装为 MessageEnvelope 格式
    envelope = {
        "apiVersion": "v2",
        "correlationId": str(uuid.uuid4()),
        "messageType": "event",
        "origin": int(time.time() * 1000),
        "payload": event_data,
        "contentType": "application/json"
    }
    
    return envelope

def on_connect(client, userdata, flags, rc):
    """连接回调"""
    if rc == 0:
        print(f"✓ Connected to MQTT broker at {MQTT_BROKER}:{MQTT_PORT}")
    else:
        print(f"✗ Failed to connect, return code {rc}")

def on_publish(client, userdata, mid):
    """发布回调"""
    pass

def main():
    """主函数"""
    # Windows 设置控制台编码
    if sys.platform == 'win32':
        import io
        sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')
        sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='replace')
    
    print("=" * 60)
    print("  sfsEdgeStore MQTT 模拟数据发布器")
    print("=" * 60)
    print(f"  MQTT Broker: {MQTT_BROKER}:{MQTT_PORT}")
    print(f"  Topic: {MQTT_TOPIC}")
    print(f"  Devices: {len(DEVICES)}")
    print("=" * 60)
    
    # 创建 MQTT 客户端
    client = mqtt.Client()
    client.on_connect = on_connect
    client.on_publish = on_publish
    
    try:
        client.connect(MQTT_BROKER, MQTT_PORT, 60)
        client.loop_start()
    except Exception as e:
        print(f"✗ Failed to connect to MQTT broker: {e}")
        print("  Make sure Mosquitto is running: \"C:\\Program Files\\Mosquitto\\mosquitto.exe\"")
        return
    
    print("\n开始发送模拟数据... (按 Ctrl+C 停止)\n")
    
    message_count = 0
    anomaly_count = 0
    
    try:
        while True:
            # 随机选择设备
            device = random.choice(DEVICES)
            
            # 决定是否生成异常数据
            anomaly = None
            rand = random.random()
            cumulative = 0
            for atype in ANOMALY_TYPES:
                cumulative += atype["probability"]
                if rand < cumulative:
                    anomaly = atype["name"]
                    break
            
            # 生成事件
            event = generate_edgex_event(device, anomaly)
            
            # 发布消息
            payload = json.dumps(event)
            info = client.publish(MQTT_TOPIC, payload, qos=0)
            
            if anomaly:
                anomaly_count += 1
                print(f"  [Anomaly] {device['name']} = {event['payload']['readings'][0]['value']} {device['unit']}")
            else:
                print(f"  [Normal] {device['name']} = {event['payload']['readings'][0]['value']} {device['unit']}")
            
            message_count += 1
            
            # 随机间隔 0.5-2 秒
            time.sleep(random.uniform(0.5, 2.0))
            
            # 每 50 条消息显示统计信息
            if message_count % 50 == 0:
                print(f"\n--- 统计: 总消息={message_count}, 异常={anomaly_count} ({anomaly_count/message_count*100:.1f}%) ---\n")
    
    except KeyboardInterrupt:
        print(f"\n\n停止发送...")
        print(f"最终统计: 总消息={message_count}, 异常={anomaly_count}")
    finally:
        client.loop_stop()
        client.disconnect()
        print("已断开 MQTT 连接")

if __name__ == "__main__":
    main()
