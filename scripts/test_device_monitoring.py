import paho.mqtt.client as mqtt
import json
import time
import random
from datetime import datetime

# MQTT 配置
BROKER = "localhost"
PORT = 1883
CLIENT_ID = "sfsedgestore-device-monitoring-test"

# 设备列表
devices = [
    "temperature-sensor-001",
    "humidity-sensor-001"
]

# 资源类型
resources = ["Temperature", "Humidity"]

# 阈值设置
THRESHOLDS = {
    "Temperature": 50.0,  # 温度阈值
    "Humidity": 90.0      # 湿度阈值
}

def create_edgex_event(device_name, resource_name, value):
    """创建 EdgeX Foundry 格式的事件"""
    timestamp = int(time.time() * 1000000000)  # 纳秒时间戳
    
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
                "valueType": "Float64",
                "value": str(value)
            }
        ]
    }
    return event

def on_connect(client, userdata, flags, rc):
    print(f"[CONNECT] 已连接到 MQTT Broker (返回码: {rc})")

def on_publish(client, userdata, mid):
    pass

def publish_data(client, device, resource, value):
    """发布数据到 MQTT"""
    event = create_edgex_event(device, resource, value)
    topic = f"edgex/events/core/{device}"
    payload = json.dumps(event)
    
    result = client.publish(topic, payload)
    result.wait_for_publish()
    print(f"[PUBLISH] {device} - {resource}: {value}")

def test_normal_data(client):
    """测试正常数据"""
    print("\n[TEST] 测试正常数据...")
    for i in range(5):
        for j, device in enumerate(devices):
            resource = resources[j]
            if resource == "Temperature":
                value = round(random.uniform(20.0, 30.0), 2)
            else:  # Humidity
                value = round(random.uniform(40.0, 70.0), 2)
            publish_data(client, device, resource, value)
            time.sleep(1)

def test_threshold_violation(client):
    """测试阈值违反"""
    print("\n[TEST] 测试阈值违反...")
    for j, device in enumerate(devices):
        resource = resources[j]
        threshold = THRESHOLDS[resource]
        # 发送超过阈值的数据
        value = threshold + 10.0
        publish_data(client, device, resource, value)
        time.sleep(1)

def test_data_anomaly(client):
    """测试数据突变"""
    print("\n[TEST] 测试数据突变...")
    for j, device in enumerate(devices):
        resource = resources[j]
        # 先发送一个正常值
        normal_value = round(random.uniform(20.0, 30.0), 2) if resource == "Temperature" else round(random.uniform(40.0, 70.0), 2)
        publish_data(client, device, resource, normal_value)
        time.sleep(1)
        # 发送一个突变值（变化率超过50%）
        anomaly_value = normal_value * 1.6  # 增加60%
        publish_data(client, device, resource, anomaly_value)
        time.sleep(1)

def test_data_trend(client):
    """测试数据趋势"""
    print("\n[TEST] 测试数据趋势...")
    for j, device in enumerate(devices):
        resource = resources[j]
        # 发送连续上升的数据
        value = 20.0 if resource == "Temperature" else 40.0
        for i in range(5):
            value += 2.0
            publish_data(client, device, resource, value)
            time.sleep(1)

def test_device_offline(client):
    """测试设备离线"""
    print("\n[TEST] 测试设备离线...")
    print("[INFO] 停止发送数据 60 秒，模拟设备离线...")
    time.sleep(60)

def main():
    print("[START] 启动设备监控测试...")
    print(f"[INFO] Broker: {BROKER}:{PORT}")
    print("")
    
    # 创建 MQTT 客户端
    client = mqtt.Client()
    client.on_connect = on_connect
    client.on_publish = on_publish
    
    try:
        # 连接到 broker
        client.connect(BROKER, PORT, 60)
        client.loop_start()
        
        # 等待连接
        time.sleep(2)
        
        print("""
测试计划:
1. 发送正常数据
2. 发送超过阈值的数据
3. 发送数据突变
4. 发送连续上升趋势数据
5. 停止发送数据 60 秒（模拟离线）
6. 重新发送数据（模拟恢复在线）
""")
        
        # 1. 测试正常数据
        test_normal_data(client)
        
        # 2. 测试阈值违反
        test_threshold_violation(client)
        
        # 3. 测试数据突变
        test_data_anomaly(client)
        
        # 4. 测试数据趋势
        test_data_trend(client)
        
        # 5. 测试设备离线
        test_device_offline(client)
        
        # 6. 测试设备恢复在线
        print("\n[TEST] 测试设备恢复在线...")
        test_normal_data(client)
        
        print("\n[DONE] 测试完成！")
        
    except KeyboardInterrupt:
        print("\n[STOP] 收到停止信号")
    except Exception as e:
        print(f"\n[ERROR] 错误: {e}")
    finally:
        client.loop_stop()
        client.disconnect()
        print("[INFO] 已断开连接")

if __name__ == "__main__":
    main()
