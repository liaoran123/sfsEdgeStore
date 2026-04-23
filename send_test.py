import json
import time
import random
import urllib.request
import urllib.error
import sys

BASE_URL = "http://localhost:8081"

def send_test_data():
    devices = [
        "temperature-sensor-001", "temperature-sensor-002", "temperature-sensor-003",
        "humidity-sensor-001", "humidity-sensor-002",
        "pressure-sensor-001", "pressure-sensor-002",
        "flow-meter-001", "flow-meter-002",
        "power-meter-001"
    ]

    readings = ["Temperature", "Humidity", "Pressure", "Flow", "Power"]
    data = {
        "deviceName": random.choice(devices),
        "reading": random.choice(readings),
        "value": round(random.uniform(20, 100), 2),
        "valueType": "Float32",
        "baseType": "Float",
        "timestamp": int(time.time() * 1000000000),
        "metadata": json.dumps({"location": "Factory", "unit": "C"})
    }
    return data

def send_data(data):
    try:
        url = f"{BASE_URL}/api/test-edgex"
        headers = {'Content-Type': 'application/json'}
        data_json = json.dumps(data).encode('utf-8')
        req = urllib.request.Request(url, data=data_json, headers=headers, method='POST')
        with urllib.request.urlopen(req) as response:
            if response.status == 200:
                print(f"[OK] Data sent: {data['deviceName']} - {data['reading']}: {data['value']}")
            else:
                print(f"[FAIL] Status: {response.status}")
    except urllib.error.HTTPError as e:
        print(f"[ERROR] HTTP {e.code}: {e.read().decode('utf-8')}")
    except Exception as e:
        print(f"[ERROR] {e}")

def main():
    print("Starting test data generator...")
    print(f"Server: {BASE_URL}")
    print("Press Ctrl+C to stop")
    count = 0
    try:
        while True:
            send_test_data()
            send_data(send_test_data())
            count += 1
            if count % 10 == 0:
                print(f"[INFO] Sent {count} messages")
            time.sleep(random.uniform(0.5, 1.5))
    except KeyboardInterrupt:
        print(f"\n[STOPPED] Total messages sent: {count}")

if __name__ == "__main__":
    main()
