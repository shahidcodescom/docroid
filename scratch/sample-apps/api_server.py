import os
import sys
import json
import time
from http.server import HTTPServer, BaseHTTPRequestHandler

port = int(os.environ.get("PORT", "8081"))

class HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        payload = {
            "service": "api-gateway",
            "status": "online",
            "version": "1.4.2",
            "runtime": "docroid-native",
            "timestamp": time.time(),
            "uptime_seconds": round(time.time() - start_time, 1)
        }
        self.wfile.write(json.dumps(payload, indent=2).encode("utf-8"))

    def log_message(self, format, *args):
        print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] HTTP - {self.address_string()} - {format%args}", flush=True)

start_time = time.time()
print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] Starting API Gateway on port {port}...", flush=True)
server = HTTPServer(("0.0.0.0", port), HealthHandler)
print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] API Gateway is ready to handle requests.", flush=True)
server.serve_forever()
