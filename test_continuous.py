import json
import time
import random
import urllib.request
import urllib.error

BASE_URL = "http://localhost:8081"

def send_request():
    data = json.dumps({
        "deviceName": random.choice([
            "temperature-sensor-001", "temperature-sensor-002",
            "humidity-sensor-001", "pressure-sensor-001",
            "flow-meter-001", "power-meter-001"
        ]),
        "reading": random.choice(["Temperature", "Humidity", "Pressure", "Flow", "Power"]),
        "value": round(random.uniform(20, 100), 2),
        "valueType": "Float32",
        "baseType": "Float",
        "timestamp": int(time.time() * 1000000000),
        "metadata": "{}"
    }).encode('utf-8')

    try:
        req = urllib.request.Request(
            f"{BASE_URL}/api/test-edgex",
            data=data,
            headers={'Content-Type': 'application/json'},
            method='POST'
        )
        with urllib.request.urlopen(req) as response:
            return response.read().decode()
    except Exception as e:
        return f"Error: {e}"

print("Starting continuous test...")
print(f"Server: {BASE_URL}")
count = 0
try:
    while True:
        result = send_request()
        count += 1
        if "success" in result:
            print(f"[{count}] OK")
        else:
            print(f"[{count}] {result[:100]}")
        if count % 20 == 0:
            print(f"--- Sent {count} requests ---")
        time.sleep(0.3)
except KeyboardInterrupt:
    print(f"\nStopped. Total: {count}")
