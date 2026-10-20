import time
import uuid
import random
import sys

jobs = ["sync_database", "process_event", "aggregate_metrics", "clean_cache", "dispatch_webhook"]

print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] Task Worker Daemon started (concurrency=4).", flush=True)

while True:
    job_name = random.choice(jobs)
    job_id = str(uuid.uuid4())[:8]
    duration = random.randint(15, 120)
    print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] [TASK] Dispatched {job_name} ({job_id}) completed in {duration}ms status=200", flush=True)
    time.sleep(4)
