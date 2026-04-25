import paho.mqtt.client as mqtt
import json
import time
import random
from datetime import datetime

# MQTT 配置
BROKER = "localhost"
PORT = 1883
CLIENT_ID = "sfsedgestore-test-publisher-enhanced"

# 设备列表
devices = [
    "temperature-sensor-001",
    "humidity-sensor-001",
    "pressure-sensor-001",
    "vibration-sensor-001",
    "energy-meter-001"
]

# 资源类型和正常范围
resources = {
    "Temperature": {"min": 15.0, "max": 35.0, "unit": "°C"},
    "Humidity": {"min": 30.0, "max": 80.0, "unit": "%"},
    "Pressure": {"min": 980.0, "max": 1060.0, "unit": "hPa"},
    "Vibration": {"min": 0.0, "max": 10.0, "unit": "mm/s"},
    "Energy": {"min": 50.0, "max": 600.0, "unit": "kWh"}
}

# 异常数据类型
exception_types = [
    "out_of_range_high",  # 超出上限
    "out_of_range_low",   # 超出下限
    "sudden_jump",        # 突然跳变
    "missing_value",      # 缺失值
    "invalid_format"      # 无效格式
]

def create_edgex_event(device_name, resource_name, value, is_exception=False, exception_type=None):
    """创建 EdgeX Foundry 格式的事件"""
    timestamp = int(time.time() * 1000000000)  # 纳秒时间戳
    
    # 处理异常数据
    if is_exception:
        if exception_type == "out_of_range_high":
            # 生成超出上限的值
            range_info = resources[resource_name]
            value = range_info["max"] * 2
        elif exception_type == "out_of_range_low":
            # 生成低于下限的值
            range_info = resources[resource_name]
            value = range_info["min"] * 0.5
        elif exception_type == "sudden_jump":
            # 生成突然跳变的值
            value = random.uniform(1000, 10000)  # 大幅跳变
        elif exception_type == "missing_value":
            # 缺失值
            value = ""
        elif exception_type == "invalid_format":
            # 无效格式
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

def on_connect(client, userdata, flags, rc):
    print(f"已连接到 MQTT Broker (返回码: {rc})")

def on_publish(client, userdata, mid):
    pass  # 静默模式，减少日志输出

def main():
    print("启动增强版 MQTT 测试发布器...")
    print(f"Broker: {BROKER}:{PORT}")
    print("发送主题: edgex/events/core/{device}")
    print("包含正常数据和异常数据")
    print("")
    
    # 创建 MQTT 客户端
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION1, CLIENT_ID)
    client.on_connect = on_connect
    client.on_publish = on_publish
    
    try:
        # 连接到 broker
        client.connect(BROKER, PORT, 60)
        client.loop_start()
        
        # 等待连接
        time.sleep(1)
        
        print("开始发送测试数据... (按 Ctrl+C 停止)")
        print("-" * 80)
        
        message_count = 0
        exception_count = 0
        
        while True:
            # 随机选择设备
            device = random.choice(devices)
            
            # 确定设备对应的资源类型
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
            
            # 决定是否发送异常数据 (10% 概率)
            is_exception = random.random() < 0.1
            exception_type = None
            
            if is_exception:
                exception_type = random.choice(exception_types)
                exception_count += 1
            
            # 生成数据值
            if not is_exception:
                # 生成正常范围内的值，添加一些小的波动
                range_info = resources[resource]
                value = round(random.uniform(range_info["min"], range_info["max"]), 2)
            else:
                # 异常数据将在 create_edgex_event 中处理
                value = 0
            
            # 创建 EdgeX 事件
            event = create_edgex_event(device, resource, value, is_exception, exception_type)
            
            # 发布到 MQTT
            topic = f"edgex/events/core/{device}"
            wrapped_message = {
                "messageType": "event",
                "payload": event
            }
            payload = json.dumps(wrapped_message)
            
            result = client.publish(topic, payload, qos=1)
            result.wait_for_publish()
            
            message_count += 1
            
            # 打印日志
            if is_exception:
                print(f"[{message_count}] {device} - {resource}: {event['readings'][0]['value']} (异常: {exception_type})")
            else:
                print(f"[{message_count}] {device} - {resource}: {value} {resources[resource]['unit']}")
            
            # 随机间隔，模拟真实设备的数据发送频率
            time.sleep(random.uniform(0.5, 2.0))
            
            # 每发送 50 条消息，打印统计信息
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
