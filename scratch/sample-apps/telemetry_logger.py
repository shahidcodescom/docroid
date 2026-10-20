import time
import random
import sys

print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] Telemetry Agent v2.3.0 initialized.", flush=True)
print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] Connected to Docroid runtime telemetry bus.", flush=True)

counter = 0
while True:
    counter += 1
    cpu = random.randint(12, 42)
    mem = random.randint(52, 94)
    conns = random.randint(140, 480)
    latency = round(random.uniform(1.2, 8.5), 2)
    print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] [METRIC] tick={counter} cpu={cpu}% mem={mem}MB conns={conns} avg_latency={latency}ms", flush=True)
    time.sleep(3)
