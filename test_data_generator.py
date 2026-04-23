import json
import time
import random
import urllib.request
import urllib.error

# 服务器URL
BASE_URL = "http://localhost:8081"

# 模拟设备数据
def generate_test_data():
    """生成模拟测试数据"""
    devices = ["Device001", "Device002", "Device003"]
    readings = ["Temperature", "Humidity", "Pressure"]
    
    data = {
        "deviceName": random.choice(devices),
        "reading": random.choice(readings),
        "value": round(random.uniform(20, 50), 2),
        "valueType": "Float32",
        "baseType": "Float",
        "timestamp": int(time.time() * 1000),
        "metadata": json.dumps({"location": "Factory", "sensor": "DHT22"})
    }
    return data

def send_data(data):
    """发送数据到服务器"""
    try:
        # 准备请求
        url = f"{BASE_URL}/api/data"
        headers = {
            'Content-Type': 'application/json',
        }
        data_json = json.dumps(data).encode('utf-8')
        
        # 发送请求
        req = urllib.request.Request(url, data=data_json, headers=headers)
        with urllib.request.urlopen(req) as response:
            if response.status == 200:
                print(f"✓ 数据发送成功: {data['deviceName']} - {data['reading']}: {data['value']}")
            else:
                print(f"✗ 数据发送失败: {response.status}")
    except urllib.error.HTTPError as e:
        print(f"✗ HTTP错误: {e.code} - {e.read().decode('utf-8')}")
    except Exception as e:
        print(f"✗ 发送数据时出错: {e}")

def main():
    """主函数"""
    print("开始发送测试数据...")
    print(f"服务器地址: {BASE_URL}")
    print("按 Ctrl+C 停止")
    
    try:
        while True:
            # 生成并发送数据
            data = generate_test_data()
            send_data(data)
            
            # 随机间隔
            time.sleep(random.uniform(0.5, 2))
            
    except KeyboardInterrupt:
        print("\n测试数据发送已停止")

if __name__ == "__main__":
    main()
