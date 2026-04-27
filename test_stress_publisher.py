import paho.mqtt.client as mqtt
import json
import time
import random
import threading
import uuid
import sys

MQTT_BROKER = "localhost"
MQTT_PORT = 1883
TOTAL_DEVICES = 500
SEND_INTERVAL = (0.05, 0.2)

stats = {"sent": 0, "errors": 0}
stats_lock = threading.Lock()

def device_thread(device_id, dev_type):
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id=f"stress-{dev_type}-{device_id}")
    try:
        client.connect(MQTT_BROKER, MQTT_PORT, 60)
        client.loop_start()
        topic = f"edgex/events/device/{dev_type}-sensor-{device_id:03d}"
        while True:
            try:
                ns = int(time.time() * 1e9)
                payload = json.dumps({
                    "apiVersion": "v2",
                    "correlationId": str(uuid.uuid4()),
                    "messageType": "event",
                    "origin": ns,
                    "payload": {
                        "id": str(uuid.uuid4()),
                        "deviceName": f"{dev_type}-sensor-{device_id:03d}",
                        "profileName": f"{dev_type}-profile",
                        "sourceName": f"{dev_type}-source",
                        "origin": ns,
                        "readings": [{
                            "id": str(uuid.uuid4()),
                            "origin": ns,
                            "deviceName": f"{dev_type}-sensor-{device_id:03d}",
                            "resourceName": dev_type,
                            "profileName": f"{dev_type}-profile",
                            "valueType": "Float64",
                            "value": str(round(random.uniform(0, 100), 2)),
                            "baseType": "Simple"
                        }]
                    }
                })
                client.publish(topic, payload, qos=1)
                with stats_lock:
                    stats["sent"] += 1
            except Exception:
                with stats_lock:
                    stats["errors"] += 1
            time.sleep(random.uniform(*SEND_INTERVAL))
    except Exception as e:
        print(f"Error: {device_id} {dev_type}: {e}")

def stats_printer():
    start = time.time()
    while True:
        time.sleep(30)
        elapsed = time.time() - start
        with stats_lock:
            sent = stats["sent"]
            errors = stats["errors"]
        rate = sent / elapsed if elapsed > 0 else 0
        print(f"\n{'='*50}\n  Stress Test ({elapsed/60:.1f}min)\n  Sent: {sent}\n  Errors: {errors}\n  Rate: {rate:.0f} msg/sec\n{'='*50}\n")

def main():
    print("=" * 60)
    print("  sfsEdgeStore STRESS TEST")
    print("=" * 60)
    print(f"  Devices: {TOTAL_DEVICES}")
    print(f"  Interval: {SEND_INTERVAL[0]}-{SEND_INTERVAL[1]}s")
    print(f"  Expected rate: ~{TOTAL_DEVICES * (1/((SEND_INTERVAL[0]+SEND_INTERVAL[1])/2)):.0f} msg/sec")
    print("=" * 60)
    
    dev_types = ["temp", "humid", "press", "vib", "energy"]
    devices_per_type = TOTAL_DEVICES // len(dev_types)
    
    threads = []
    for dt in dev_types:
        for i in range(devices_per_type):
            t = threading.Thread(target=device_thread, args=(i, dt))
            threads.append(t)
            t.start()
            time.sleep(0.01)
    
    threading.Thread(target=stats_printer, daemon=True).start()
    print(f"All {TOTAL_DEVICES} stress devices started!")
    print("Press Ctrl+C to stop.\n")
    
    try:
        for t in threads:
            t.join()
    except KeyboardInterrupt:
        print("\nStopping stress test...")

if __name__ == "__main__":
    main()
