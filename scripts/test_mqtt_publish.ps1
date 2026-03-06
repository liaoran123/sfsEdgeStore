# 测试 MQTT 数据发布脚本
# 需要安装 paho-mqtt (pip install paho-mqtt)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MQTT 测试数据发布工具" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Python 是否可用
try {
    $pythonVersion = python --version 2>&1
    Write-Host "Python 已安装: $pythonVersion" -ForegroundColor Green
} catch {
    Write-Host "Python 未安装，请先安装 Python" -ForegroundColor Red
    exit 1
}

# 创建测试 Python 脚本
$pythonScript = @'
import paho.mqtt.client as mqtt
import json
import time
import random
from datetime import datetime

# MQTT 配置
BROKER = "localhost"
PORT = 1883
TOPIC = "edgex/events/core/#"
CLIENT_ID = "sfsedgestore-test-publisher"

# 设备列表
devices = [
    "temperature-sensor-001",
    "humidity-sensor-001",
    "pressure-sensor-001",
    "vibration-sensor-001",
    "energy-meter-001"
]

# 资源类型
resources = ["Temperature", "Humidity", "Pressure", "Vibration", "Energy"]

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
    print(f"✅ 已连接到 MQTT Broker (返回码: {rc})")

def on_publish(client, userdata, mid):
    print(f"📤 消息已发布 (ID: {mid})")

def main():
    print("🚀 启动 MQTT 测试发布器...")
    print(f"📡 Broker: {BROKER}:{PORT}")
    print(f"📋 Topic: {TOPIC}")
    print("")
    
    # 创建 MQTT 客户端
    client = mqtt.Client(CLIENT_ID)
    client.on_connect = on_connect
    client.on_publish = on_publish
    
    try:
        # 连接到 broker
        client.connect(BROKER, PORT, 60)
        client.loop_start()
        
        # 等待连接
        time.sleep(1)
        
        print("")
        print("🔄 开始发送测试数据... (按 Ctrl+C 停止)")
        print("")
        
        message_count = 0
        while True:
            # 随机选择设备和资源
            device_idx = random.randint(0, len(devices) - 1)
            device = devices[device_idx]
            resource = resources[device_idx]
            
            # 生成随机值
            if resource == "Temperature":
                value = round(random.uniform(20.0, 30.0), 2)
            elif resource == "Humidity":
                value = round(random.uniform(40.0, 70.0), 2)
            elif resource == "Pressure":
                value = round(random.uniform(1000.0, 1050.0), 2)
            elif resource == "Vibration":
                value = round(random.uniform(0.0, 5.0), 2)
            elif resource == "Energy":
                value = round(random.uniform(100.0, 500.0), 2)
            else:
                value = round(random.uniform(0.0, 100.0), 2)
            
            # 创建 EdgeX 事件
            event = create_edgex_event(device, resource, value)
            
            # 发布到 MQTT
            topic = f"edgex/events/core/{device}"
            payload = json.dumps(event)
            
            result = client.publish(topic, payload)
            result.wait_for_publish()
            
            message_count += 1
            print(f"[{message_count}] {device} - {resource}: {value}")
            
            # 随机间隔
            time.sleep(random.uniform(0.5, 2.0))
            
    except KeyboardInterrupt:
        print("")
        print("🛑 收到停止信号")
    except Exception as e:
        print(f"❌ 错误: {e}")
    finally:
        client.loop_stop()
        client.disconnect()
        print("✅ 已断开连接")

if __name__ == "__main__":
    main()
'@

# 保存 Python 脚本
$pythonScriptPath = "d:\MyGo\src\sfsEdgeStore\scripts\test_mqtt_publisher.py"
Set-Content -Path $pythonScriptPath -Value $pythonScript -Encoding UTF8

Write-Host "Python 测试脚本已创建: $pythonScriptPath" -ForegroundColor Green
Write-Host ""
Write-Host "使用方法:" -ForegroundColor Yellow
Write-Host "  1. 安装 paho-mqtt: pip install paho-mqtt" -ForegroundColor Cyan
Write-Host "  2. 运行脚本: python scripts\test_mqtt_publisher.py" -ForegroundColor Cyan
Write-Host ""
Write-Host "脚本将持续向 MQTT broker 发送 EdgeX 格式的测试数据" -ForegroundColor Cyan
