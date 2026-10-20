# Docroid

> **Ultra-Lightweight, High-Performance Application & Container Engine for Android (Termux)**  
> *Zero daemon. Zero root required. 0% CPU loss.*

---

## Why Docroid?

Running Docker on Android mobile fails for two reasons:
1. **Resource Bloat**: The `dockerd` + `containerd` daemons consume 200MB–400MB of RAM idle, quickly draining your battery and causing Android's Out-Of-Memory (OOM) killer to terminate background apps.
2. **Kernel Constraints**: Standard Docker requires root, cgroups v2, network bridges (`veth`), and `overlayfs` kernel modules—none of which exist on standard non-rooted Android devices.

**Docroid** is purpose-built in pure Go for Android and Termux:
- **Zero Daemon (No Battery Drain)**: State is managed via atomic metadata files in `~/.docroid/`. Zero idle memory footprint.
- **Works 100% on Non-Rooted Devices**: Runs out-of-the-box on any standard Android smartphone.
- **Native Fast-Path Engine (0% CPU Loss)**: Runs Go binaries, static executables, and scripts directly on the Android kernel with 100% native CPU speed, isolated workspaces, and port checking.
- **Rootfs Sandbox Engine (Ultra-Lean Linux)**: Runs applications requiring standard Linux root structures (`/lib`, `/bin/sh`, glibc/musl, `apk` package manager) using a ~3.2 MB Alpine micro-rootfs and non-root PRoot virtualization.
- **Container-Like Supervision**: Detached background execution (`-d`), live resource stats (`docroid top`), automatic stdout/stderr log capture (`docroid logs`), port binding, and graceful signal stops (`docroid stop`).

---

## Architecture Overview

```
                      ┌──────────────────────────────────────┐
                      │             docroid CLI              │
                      │  run, ps, stop, start, logs, top...  │
                      └──────────────────┬───────────────────┘
                                         │
                   ┌─────────────────────┴─────────────────────┐
                   ▼                                           ▼
       ┌───────────────────────┐                   ┌───────────────────────┐
       │   Native Engine       │                   │ Rootfs Sandbox Engine │
       │  (Zero-Overhead)      │                   │   (Non-Root PRoot)    │
       ├───────────────────────┤                   ├───────────────────────┤
       │ • Go binaries         │                   │ • Linux binaries      │
       │ • Static executables  │                   │ • Alpine rootfs (3MB) │
       │ • 0% CPU loss         │                   │ • Fake root (UID 0)   │
       │ • <5MB RAM footprint  │                   │ • Virtualized /dev/etc│
       └───────────┬───────────┘                   └───────────┬───────────┘
                   │                                           │
                   └─────────────────────┬─────────────────────┘
                                         │
                                         ▼
                     ┌───────────────────────────────────────┐
                     │    Supervisor & Process Manager       │
                     │  • Process Groups (PGID)              │
                     │  • Circular Log Buffers               │
                     │  • Real-time CPU & RSS RAM metrics    │
                     │  • Port & Workspace Sandboxing        │
                     └───────────────────┬───────────────────┘
                                         │
                                         ▼
                     ┌───────────────────────────────────────┐
                     │     Storage & State (~/.docroid/)     │
                     │  • apps/<id>/ (config, logs, pid)     │
                     │  • workspaces/<id>/ (isolated disk)   │
                     │  • rootfs/<name>/ (cached base OS)    │
                     └───────────────────────────────────────┘
```

---

## Quick Installation on Termux (Android)

### Option 1: Direct Binary Copy
Download the pre-compiled static `docroid-linux-arm64` binary to Termux:

```bash
# Inside Termux:
curl -L -o $PREFIX/bin/docroid https://github.com/your-repo/docroid/releases/latest/download/docroid-linux-arm64
chmod +x $PREFIX/bin/docroid

# Run doctor to verify setup
docroid doctor
```

### Option 2: Build from Source in Termux
```bash
# Install Go in Termux
pkg install -y golang git

# Clone and install
git clone https://github.com/your-repo/docroid.git
cd docroid
make install
```

---

## Getting Started & Usage Examples

### 1. Run a Go Binary in Background (Native Engine)

```bash
# Compile or download any Go binary
go build -o my-server main.go

# Run it detached in the background on port 8080
docroid run -d -n api-server -p 8080 ./my-server

# Output:
# Started api-server (ID: 3a9f0e1d82, PID: 14210)
# Run 'docroid logs api-server' or 'docroid top api-server' to monitor.
```

### 2. View Running Applications (`docroid ps`)

```bash
docroid ps
```
```
APP ID       NAME         ENGINE   STATUS             CPU    MEM        PORT   COMMAND
3a9f0e1d82   api-server   native   Up 14 minutes ago  0.3%   11.4 MB    8080   ./my-server
```

To see all apps (including stopped ones):
```bash
docroid ps -a
```

### 3. Monitor Live Resource Usage (`docroid top`)

Shows real-time CPU %, RSS memory, threads, and uptime:

```bash
docroid top api-server
```
```
Live Metrics for api-server (PID 14210)
-----------------------------------------
State:       S (sleeping/event-loop)
CPU Usage:   0.30%
Memory RSS:  11.40 MB
Threads:     4
Uptime:      14 minutes ago
Port:        8080
Workspace:   /data/data/com.termux/files/home/.docroid/workspaces/3a9f0e1d82
```

### 4. View and Follow Logs (`docroid logs`)

```bash
# View last 50 log lines
docroid logs api-server

# Follow live log stream
docroid logs -f api-server
```

### 5. Stop and Restart Apps

```bash
# Gracefully stop the application (SIGTERM -> SIGKILL)
docroid stop api-server

# Restart it
docroid start api-server

# Restart with one command
docroid restart api-server

# Remove application and its workspace
docroid rm -f api-server
```

---

## Built-in Static Website Hosting & VHost Routing

Docroid includes a built-in static web server and virtual host router. Deploy websites by uploading HTML/static files or `.zip` bundles, binding any port, and routing by domain (`Host` header).

### 1. Host a Directory, File, or ZIP via CLI

```bash
# Serve a build directory in background on port 8080 with a custom domain
docroid static serve ./dist -p 8080 -d mysite.local --name portfolio -b

# Serve a ZIP bundle directly with SPA fallback routing
docroid static serve myapp.zip -p 3000 --spa -b

# Serve a single HTML file in foreground
docroid static serve index.html -p 8080
```

### 2. Manage Hosted Static Sites

```bash
# List all hosted websites with status, port, domain, and URLs
docroid static ps

# Inspect files inside a hosted website
docroid static files portfolio

# Export a hosted website as a .zip bundle
docroid static export portfolio -o backup.zip

# Stop and start servers
docroid static stop portfolio
docroid static start portfolio

# Remove website and its assets
docroid static rm portfolio
```

### 3. Deploy via Web Management Dashboard (`docroid ui`)

Launch the web dashboard (`docroid ui`) and click **Websites** in the sidebar:
- **ZIP Bundle Upload**: Drag-and-drop any `.zip` archive. Docroid automatically detects `index.html` inside `dist/`, `build/`, or root with Zip-Slip protection.
- **HTML & Static Assets**: Drag-and-drop multiple HTML, CSS, JS, images, or entire folders (`webkitdirectory`).
- **Direct HTML Editor**: Write or paste HTML code directly in the browser.
- **Virtual Host Domain Routing**: Configure domains like `docs.local` or `mysite.com`. Multiple sites can share a port if their domains differ.
- **SPA Fallback Routing**: Seamlessly route deep client-side URLs to `index.html` for React, Vue, and Angular apps.
- **In-Browser File Explorer**: Browse file trees, preview code/images, and download `.zip` exports directly in the dashboard.

---

## Running Full Linux Micro-Containers (Rootfs Engine)

If your app requires standard Linux paths (`/bin/sh`, `/lib`, `apt`, `apk`, `python3`, etc.):

### Step 1: Install PRoot (Only needed for Rootfs mode)
```bash
pkg install -y proot
```

### Step 2: Pull the Alpine Linux Micro-Rootfs (~3.2 MB)
```bash
docroid rootfs pull alpine
```

### Step 3: Launch Containers Inside the Rootfs
```bash
# Interactive shell with fake root (UID 0)
docroid run --rootfs alpine /bin/sh

# Inside Alpine rootfs, install packages with apk!
/ # apk update && apk add curl python3
/ # python3 -c "print('Hello from Alpine inside Android!')"
/ # exit
```

Run background services inside Alpine rootfs:
```bash
docroid run -d --rootfs alpine -n py-service -p 5000 python3 app.py
```

---

## Container Execution (`docroid exec`)

Docroid supports executing arbitrary commands or interactive terminal sessions inside running containers, similar to `docker exec`. It inherits the container's workspace, volumes, network aliases, and environment variables with full signal forwarding and exit code propagation.

```bash
# Enter an interactive shell inside a running container
docroid exec -it myapp sh
docroid exec -it myapp   # defaults to /bin/sh or $SHELL

# Execute a one-off command and capture output
docroid exec myapp ls -la /workspace

# Run with custom environment variable overrides
docroid exec -e DEBUG=1 -e API_KEY=secret123 myapp ./check_status.sh

# Run in a specific working directory
docroid exec -w /tmp myapp pwd

# Run in detached mode (background process)
docroid exec -d myapp python3 worker.py

# Force execution even if the container is currently stopped
docroid exec -f myapp cat config.json
```

### Flags & Options

| Flag | Long Flag | Description |
|------|-----------|-------------|
| `-i` | `--interactive` | Keep STDIN open even if not attached |
| `-t` | `--tty` | Allocate a pseudo-TTY |
| `-d` | `--detach` | Run command in the background |
| `-e` | `--env` | Set environment variables (`-e KEY=VAL` or `-e KEY` to inherit) |
| `-w` | `--workdir` | Working directory inside the container |
| `-u` | `--user` | Username or UID (`<name\|uid>[:<group\|gid>]`) |
| `-f` | `--force` | Execute even if container is stopped |
| `--` | `--privileged` | Extended privileges inside the execution sandbox |

### REST API Execution

You can also execute commands programmatically via the built-in HTTP management API:

```bash
curl -X POST http://localhost:9190/api/apps/<id_or_name>/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "sh",
    "args": ["-c", "echo Connected to container; env"],
    "env": ["DEBUG=1"],
    "workdir": "/workspace",
    "timeout": 30
  }'
```

Response:
```json
{
  "exit_code": 0,
  "stdout": "Connected to container\n...",
  "stderr": ""
}
```

---

## Image Lifecycle: `build`, `pull`, `push` & `.dpack` Portability

Docroid provides full image lifecycle and registry support matching modern container tooling (`docker build`, `docker pull`, `docker push`), optimized specifically for non-root mobile environments:

```
         ┌─────────────────────────┐           ┌────────────────────────┐
         │ Dockerfile / Docroidfile│           │  OCI / Docker Registry │
         │ FROM, RUN, COPY, CMD... │           │  (Docker Hub, GHCR...) │
         └────────────┬────────────┘           └───────────┬────────────┘
                      │ docroid build -t myapp             │ docroid pull
                      ▼                                    ▼
         ┌──────────────────────────────────────────────────────────────┐
         │             Docroid Local Image Repository                   │
         │                  (~/.docroid/images/)                        │
         └─────────────┬───────────────────────────┬────────────────────┘
                       │ docroid push              │ docroid export -o .dpack
                       ▼                           ▼
         ┌─────────────────────────┐  ┌─────────────────────────────────┐
         │ OCI V2 Registry (Push)  │  │  Portable .dpack Bundle         │
         │ (Layer blobs + Manifest)│  │  (Air-gapped Wi-Fi/Bluetooth)   │
         └─────────────────────────┘  └─────────────────────────────────┘
```

### 1. Build Images (`docroid build`)

Docroid supports both standard **`Dockerfile`** and **`Docroidfile`** syntax with step-by-step layer generation:

```dockerfile
# Dockerfile
FROM alpine:latest
WORKDIR /app
COPY . /app
RUN apk add --no-cache curl python3
ENV APP_ENV=production
EXPOSE 8080
CMD ["python3", "server.py"]
```

Build and tag your image from the current context:
```bash
docroid build -t myapp:1.0 .

# Or specify a custom Dockerfile path:
docroid build -t myapp:1.0 -f ./deploy/Dockerfile .
```

Supported build instructions:
* `FROM <image|rootfs|local>` — Pulls base image or uses local rootfs/native sandbox
* `WORKDIR <path>` — Sets working directory for subsequent steps
* `COPY <src> <dest>` — Copies files from context directory into image
* `ADD <src|url|tar> <dest>` — Copies files, downloads remote URLs, or unpacks archives
* `RUN <command>` — Executes build commands in context
* `ENV KEY=VALUE` — Injects persistent environment variables
* `EXPOSE <port>` — Configures default exposed network port
* `ENTRYPOINT` & `CMD` — Sets executable command and default arguments

---

### 2. Pull Images (`docroid pull`)

Pull OCI v2 and Docker Hub container images directly to your Android device without Docker daemon or root privileges:

```bash
# Pull from Docker Hub
docroid pull alpine:latest

# Pull specific tags or organizations
docroid pull python:3.11-alpine

# Pull from GitHub Container Registry (GHCR) or private registries
docroid pull ghcr.io/username/app:v1.2.0
docroid pull registry.example.com:5000/microservice:latest

# Pull from direct .dpack bundle URL
docroid pull https://example.com/downloads/myapp.dpack
```

> **Auto-Pull on Run**: Running `docroid run -d <image>` will automatically pull the image from the registry if not found locally!

---

### 3. Push Images (`docroid push`)

Push locally built or customized Docroid images to any OCI V2 compliant container registry:

```bash
# Push to Docker Hub
docroid push myusername/myapp:1.0

# Push to private or local registry
docroid push 192.168.1.50:5000/myapp:latest
docroid push ghcr.io/myorg/myapp:v1.0
```

Docroid automatically compresses the filesystem into tar.gz layer blobs, calculates SHA-256 digests, generates an OCI Image Config and Manifest, handles Bearer token authentication challenges, and uploads to the registry.

---

### 4. Inspect & Manage Images (`docroid images`, `docroid rmi`)

```bash
# List installed images
docroid images

# REPOSITORY    TAG     ENGINE   ARCH    CREATED         COMMAND
# alpine        latest  rootfs   arm64   2 minutes ago   /bin/sh
# myapp         1.0     rootfs   arm64   10 seconds ago  python3 server.py

# Remove an image
docroid rmi myapp:1.0
```

---

### 5. Export & Import Air-Gapped Bundles (`.dpack`)

For environments without internet access or registry accounts, export to a single `.dpack` file:

```bash
# Export built image to .dpack
docroid export -o myapp.dpack myapp:1.0

# Export running app directly
docroid export -o live-app.dpack my-active-app

# Import on another device
docroid import -i myapp.dpack

# Run immediately!
docroid run -d myapp:1.0

# 3. Verify it's running
docroid ps
docroid logs api-service
```

### 5. Web Dashboard Export & Import

- **One-Click Export**: Click the **Export** button on any application card to download `<app>.dpack` directly in your phone's browser.
- **Drag-and-Drop Import**: Click the **Import .dpack** button in the dashboard navigation bar to upload and install any bundle instantly.

---

## Multi-Service Orchestration: Docroid Compose (`docroid-compose.yml`)

Just like **Docker Compose** orchestrates multi-container stacks via `docker-compose.yml`, Docroid provides built-in support for **`docroid-compose.yml`** to declaratively define, deploy, and manage multi-service microservice architectures on Android without any background daemon.

```yaml
# docroid-compose.yml
version: "1.0"
services:
  web:
    name: web-frontend
    engine: native
    command: python3 -m http.server 8080
    ports:
      - "8080:8080"
    environment:
      APP_ENV: production
      PORT: "8080"
    restart: always
    volumes:
      - "/sdcard/Download:/downloads"
    secrets:
      - DB_PASSWORD

  api:
    name: api-backend
    engine: native
    command: ["./api-server", "--verbose"]
    ports:
      - "9000:9000"
    restart: on-failure
    environment:
      - DB_HOST=127.0.0.1
      - DB_PORT=5432
```

### 1. Web Dashboard Integration (Upload & Download)
- **Upload & Auto-Parse**: In the dashboard's **Deploy New Application** modal, click **Upload Compose YAML** (or drag & drop any `.yml` / `.yaml` file onto the form). Docroid automatically parses the file and auto-populates all form inputs (Name, Engine, Command, Arguments, Port Mapping, Environment, Volumes, Restart Policy).
- **Multi-Service Switcher**: If the compose file contains multiple services, an interactive dropdown lets you switch between services to inspect or edit their configurations.
- **One-Click Deploy All (`Compose Up`)**: Click **Deploy All Services (Compose Up)** in the modal to provision the entire stack in one click.
- **Generate & Download YAML**: Click **Download Compose YAML** at any time to generate and export a ready-to-use `docroid-compose.yml` based on the current field values in the deployment form.

### 2. CLI Commands (`docroid compose`)
```bash
# 1. Start all services defined in docroid-compose.yml in background
docroid compose up -d

# 2. View status of compose stack services
docroid compose ps

# 3. Validate and view resolved configuration
docroid compose config

# 4. Generate a docroid-compose.yml from an existing running app
docroid compose generate my-running-app -o docroid-compose.yml

# 5. Stop and remove all services defined in docroid-compose.yml
docroid compose down
```

---

## Virtual Networking & Inter-Container Communication (Docker-like)

Docroid features a first-class, **zero-dependency virtual networking layer** designed specifically for Android (Termux) and mobile environments without requiring root, kernel `veth` pairs, or `iptables`:

```
               ┌────────────────────────────────────────────────────────┐
               │         Virtual Bridge Network (10.89.0.0/16)          │
               │                                                        │
               │  ┌──────────────────────┐    ┌──────────────────────┐  │
               │  │     backend-api      │    │    frontend-web      │  │
               │  │  IP: 10.89.0.2:8080  │◄───│  IP: 10.89.0.3:3000  │  │
               │  │  Aliases: api, db    │    │  Resolves: "api:8080"│  │
               │  └──────────────────────┘    └──────────────────────┘  │
               └────────────────────────────────────────────────────────┘
                       ▲                                  ▲
            /etc/hosts Hostname Resolution       Docker Link Env Vars
             ("127.0.0.1  backend-api api")       (API_PORT, API_HOST...)
```

### 1. Zero-Root Service Discovery Architecture
- **Automatic IPAM**: Docroid allocates virtual subnet ranges (`10.89.0.0/16`, `10.90.0.0/16`, etc.) and assigns private container IPs.
- **Isolated `/etc/hosts` Mounting**: For Rootfs containers (`docroid run --rootfs alpine`), Docroid dynamically builds a container-specific `/etc/hosts` mapping all sibling container names and network aliases to `127.0.0.1`, bind-mounted automatically via PRoot.
- **Docker Standard Link Environment Variables**: For all engines (native and rootfs), Docroid injects standard Docker link variables:
  - `<SERVICE>_HOST=127.0.0.1`
  - `<SERVICE>_PORT=<port>`
  - `<SERVICE>_PORT_<port>_TCP_ADDR=127.0.0.1`
  - `<SERVICE>_PORT_<port>_TCP_PORT=<port>`
- **`HOSTALIASES` Support**: Native processes automatically receive `HOSTALIASES` pointing to the network hosts lookup table.

### 2. Network CLI Commands (`docroid network`)
```bash
# 1. List virtual networks
docroid network ls

# 2. Create a custom isolated network
docroid network create my-net
docroid network create --subnet 10.95.0.0/16 backend-net

# 3. Inspect network details, subnets, and connected containers
docroid network inspect my-net

# 4. Run containers attached to a network with aliases
docroid run -d -n redis --network backend-net --network-alias cache redis-server
docroid run -d -n api --network backend-net -p 8080:8080 node server.js

# 5. Connect and disconnect running containers dynamically
docroid network connect backend-net frontend-web --alias web
docroid network disconnect backend-net frontend-web

# 6. Delete a user-defined virtual network
docroid network rm my-net
```

### 3. Compose Multi-Container Network Integration
In `docroid-compose.yml`, define user-defined networks or allow Docroid to automatically provision a project-scoped `<project>_default` bridge network:

```yaml
version: "1.0"
services:
  db:
    name: postgres-db
    command: postgres
    ports: ["5432:5432"]
    networks:
      - app-net

  web:
    name: api-server
    command: node server.js
    ports: ["8080:8080"]
    networks:
      - app-net

networks:
  app-net:
    driver: bridge
```

Inside `web`, you can seamlessly connect to PostgreSQL via `postgres://postgres:password@db:5432/mydb` or `$DB_HOST:$DB_PORT`!

---

## Built-In Web Dashboard & Management Interface

Docroid includes a built-in, zero-dependency web dashboard embedded directly into the binary:

```bash
# Launch dashboard on default port 9090
docroid ui

# Or specify custom port
docroid ui -p 9090
```

```
Docroid Web Dashboard & Management Server
==========================================
• Local access:   http://localhost:9090
• Network access: http://192.168.1.50:9090
```

### Features:
- **Mobile & Desktop UI**: Responsive dark theme with glassmorphic cards and telemetry meters.
- **Live Monitoring**: Real-time CPU % load, RSS memory, and process states without page reloads.
- **Interactive Control**: Deploy new Go binaries or Alpine containers, Start, Stop, Restart, and Delete apps.
- **Live Terminal**: Integrated logs drawer with auto-scroll and stream refresh.
- **System Doctor**: Live hardware, Android release, kernel, and PRoot health checks.
- **REST APIs**:
  - `GET /api/apps`: All apps with live `/proc` stats
  - `POST /api/apps`: Deploy new app
  - `POST /api/apps/{id}/stop`: Stop app
  - `POST /api/apps/{id}/start`: Start app
  - `POST /api/apps/{id}/restart`: Restart app
  - `GET /api/secrets`: List stored secrets
  - `POST /api/secrets`: Store a new secret key-value
  - `DELETE /api/secrets/{key}`: Delete a secret
  - `GET /api/system`: System doctor report
  - `GET /api/tunnels`: List all configured Cloudflare tunnels with status & URLs
  - `POST /api/tunnels`: Create & auto-start a tunnel (port, host, app, quick or token mode)
  - `GET /api/tunnels/{id}`: Detailed tunnel inspection
  - `POST /api/tunnels/{id}/start|stop|restart`: Lifecycle control
  - `GET /api/tunnels/{id}/logs`: Real-time tail of tunnel process logs
  - `DELETE /api/tunnels/{id}`: Terminate and delete tunnel
  - `GET /api/tunnels/targets`: Auto-discover running apps, static sites, and UI ports
  - `GET /api/tunnels/status`: Check cloudflared binary installation status and version

---

## Cloudflare Tunnel Direct Configuration (Expose Any Port or App Globally)

Docroid includes a built-in **Cloudflare Tunnel manager** that lets you expose any application running on any port (or any Docroid container app, static website, or the Web Dashboard itself) to the public internet securely—without port forwarding, static IPs, router access, or public IP addresses.

Supports both:
1. **Quick Tunnels (Free & Zero-Config)**: Instant ephemeral public `https://*.trycloudflare.com` URL with zero signups or credentials needed.
2. **Zero Trust Named Tunnels (Custom Domains & Auth)**: Production-ready enterprise tunnels authenticated with a Cloudflare remote tunnel token (`eyJh...`).

### 1. CLI Commands (`docroid tunnel`)

```bash
# Check cloudflared binary status and detected binary path
docroid tunnel status

# 1. Expose any arbitrary local port via Quick Tunnel (ephemeral https://*.trycloudflare.com)
docroid tunnel start -p 8080 -n my-api

# 2. Expose running Docroid container app or static site by name
docroid tunnel start --app web-demo -n web-public

# 3. Expose any custom hostname or external service
docroid tunnel start -t http://127.0.0.1:3000 -n nextjs-app

# 4. Run in background (detached daemon mode)
docroid tunnel start -p 9090 -n docroid-ui -d

# 5. Production Zero Trust Named Tunnel using a Cloudflare Tunnel Token
docroid tunnel start --token eyJh... --hostname api.mydomain.com -n prod-api -d

# 6. List all tunnels, running state, target port, and public URLs
docroid tunnel ls

# 7. View live streaming diagnostic logs
docroid tunnel logs <tunnel-id>

# 8. Manage lifecycle (stop, restart, remove)
docroid tunnel stop <tunnel-id>
docroid tunnel restart <tunnel-id>
docroid tunnel rm <tunnel-id>
```

### 2. Direct Configuration in the Web Management Dashboard

Access the dashboard at `http://localhost:9090/#tunnels`:
- **"+ Direct Configure Tunnel" Modal**:
  - **Auto-Discovery Dropdown**: Select directly from any running Docroid container app, static website, or the Docroid Web Dashboard itself.
  - **Custom Port / Host Input**: Target any arbitrary local port (e.g. `8080`, `3000`, `5000`) or custom target URL (`http://localhost:8000`).
  - **Interactive Mode Selector**: One-click toggle between **Quick Tunnel** (instant free URL) and **Zero Trust Token** (named Cloudflare Tunnel).
  - **Advanced Settings**: Configure custom tunnel names, custom hostnames, auto-start, and auto-restart flags.
- **Dedicated Tunnels Dashboard**:
  - Live KPI strip showing Active Tunnels, Quick Tunnels, Zero Trust Tunnels, and Cloudflared engine runtime status.
  - Interactive tunnel cards with one-click **Copy Public URL** and status indicators (Running, Starting, Stopped, Error).
  - Direct Action Controls: Start, Stop, Restart, View Logs, and Delete tunnels in real time.
- **Embedded Dark Terminal Log Viewer**:
  - Inspect live `cloudflared` diagnostic outputs with auto-refresh and log copy.
- **Seamless Container & Website Integration**:
  - Direct **Tunnel** action button on container application cards and static site cards for 1-click exposure.

---

## Container Features & Advanced Runtime Capabilities

### 1. Port Mapping (Pure Non-Root TCP Forwarding)
Standard Docker uses Linux kernel `iptables` and network bridges (`veth`) to forward ports. Because non-root Android apps cannot manipulate `iptables`, Docroid introduces an intelligent, zero-overhead bidirectional TCP proxy:
- **Direct binding**: `-p 8080` binds container port 8080 directly.
- **Port remapping**: `-p 8080:3000` maps external host port 8080 to container port 3000.
- **IP binding**: `-p 127.0.0.1:8080:3000` binds only to localhost for secure internal microservices.

```bash
# Run server on port 3000, accessible externally on port 8080
docroid run -d -n web -p 8080:3000 ./my-server
```

When stopping or deleting an app (`docroid stop` / `docroid rm`), the forwarder listener is automatically terminated and the port is released instantly.

---

### 2. Restart Policy & Process Supervision
Keep critical mobile services and background bots alive through crashes or Android process interruptions:
- `--restart no`: Default. Do not restart automatically upon exit.
- `--restart always`: Continually restart if the process crashes or exits.
- `--restart on-failure`: Restart automatically if the process exits with a non-zero exit code or is killed.

Docroid applies **exponential backoff** (500ms → 1s → 2s ... up to 10s) to protect CPU and battery from runaway restart loops:

```bash
docroid run -d -n telegram-bot --restart always ./bot-binary
```

View restart counters in `docroid ps`:
```
APP ID       NAME           STATUS                         PORT   COMMAND
a1b2c3d4e5   telegram-bot   Up 2h (restarts: 3)            -      ./bot-binary
```

Stopping an app gracefully cancels active supervision and prevents unwanted restarts:
```bash
docroid stop telegram-bot
```

---

### 3. Volume Binding (Host-to-Container Storage)
Mount host Android storage (SD card, shared documents, databases, logs) directly into your application's isolated workspace:
- **Syntax**: `-v <host_path>:<container_path>`
- **Native Engine**: Automatically maps and creates bidirectional workspace symlinks with full read/write permissions.
- **Rootfs Sandbox Engine**: Injects kernel-level PRoot bind mounts directly into Alpine Linux.

```bash
# Bind Termux home data into app
docroid run -d -n db-app -v /data/data/com.termux/files/home/data:/app/data ./db-engine

# Bind phone's shared SD card storage into an Alpine container
docroid run --rootfs alpine -v /sdcard/Download:/downloads /bin/sh
```

---

### 4. Safe Secrets, Key-Values & Environment Store
Store sensitive API tokens, passwords, database URLs, and private keys securely with strict file permissions (`0600` - owner read/write only) under `~/.docroid/secrets/`.

#### Manage Secrets via CLI:
```bash
# Store a secret
docroid secret set DB_PASSWORD "My$uperSecretPass"
docroid secret set OPENAI_API_KEY "sk-antigravity123"

# List stored secrets (values are automatically masked for privacy)
docroid secret list
# SECRET KEY          VALUE (MASKED)
# DB_PASSWORD         ••••••••••••
# OPENAI_API_KEY      ••••••••••••

# Retrieve a secret value
docroid secret get DB_PASSWORD

# Delete a secret
docroid secret rm DB_PASSWORD
```

#### Inject Secrets & Environment Files into Apps:
```bash
# 1. Inject specific secrets directly into container environment
docroid run -d -n backend --secret DB_PASSWORD --secret API_KEY=OPENAI_API_KEY ./server

# 2. Load environment variables from standard .env files
docroid run -d -n backend --env-file ./production.env ./server
```

#### Manage Secrets via Web Dashboard:
Click the **Secrets** button in the dashboard navbar to view masked secrets, reveal values with one click, add new keys, and delete stale configurations.

---

### 5. Built-in Zero-Downtime Updates & Hot Restarts (Zero Dropped Connections)

Production micro-services, APIs, and background servers on mobile cannot afford downtime or dropped TCP connections during routine code deployments, configuration tweaks, or restarts.

Docroid implements **in-process blue-green / shadow rolling cutovers** with automatic rollback protection:

```
[Incoming Traffic on Host Port 8080]
               │
               ▼
   ┌───────────────────────┐
   │ Dynamic Reverse Proxy │ (Holds host port open continuously)
   └───────────┬───────────┘
               │
        ┌──────┴────────────────────────────────────┐
        │ Active Traffic                            │ (Switches atomically in <1ms)
        ▼                                           ▼
┌───────────────────┐                     ┌───────────────────┐
│ Old Instance A    │                     │ Candidate Inst. B │
│ (PID 1201)        │                     │ (PID 1245)        │
│ Internal Port X   │                     │ Internal Port Y   │
├───────────────────┤                     ├───────────────────┤
│ 1. Handling reqs  │                     │ 2. Shadow spawn   │
│                   │                     │ 3. Readiness probe│
│ 5. Drain (1.5s)   │ ◄───────────────────┼ 4. Atomic Cutover │
│ 6. SIGTERM retire │                     │ 7. Primary active │
└───────────────────┘                     └───────────────────┘
```

#### How It Works:
1. **Shadow Candidate Spawn**: Launches the updated application version on an internal ephemeral port while the old process remains 100% active.
2. **Active Readiness Probing**: Actively verifies candidate health via TCP socket check (or HTTP `/healthz` probe) within a configurable timeout.
3. **Automatic Rollback Guard**: If the candidate fails to boot, crashes, or fails the readiness probe, Docroid immediately aborts the cutover, terminates the failed candidate, and keeps the original instance untouched with zero downtime.
4. **Atomic Netproxy Cutover**: Upon readiness confirmation, Docroid's user-space proxy switches incoming socket dispatch to the candidate's port atomically using sync/atomic.
5. **Graceful Connection Draining**: Gives in-flight requests a configurable drain window (default 1.5s) to complete cleanly.
6. **Graceful Retirement**: Sends `SIGTERM` (with 3-second `SIGKILL` escalation) to the old instance and persists new process state atomically in `app.json`.

#### CLI Commands:
```bash
# 1. Hot restart existing process with zero downtime
docroid restart -z my-web-app

# 2. Live update binary executable with zero downtime
docroid update --cmd ./my-server-v2 my-web-app

# 3. Live update environment variables and connection drain window
docroid update -e APP_ENV=production -e LOG_LEVEL=debug --drain 2.0 my-web-app

# 4. Live update from a new portable image
docroid update --image myapp:2.0.0 my-web-app
```

#### Web Dashboard & REST API:
- **One-Click Hot Restart**: Click the **Hot Restart** button on any active application card.
- **Update Dialog**: Click **Update** to open the live update dialog, adjust command, arguments, environment flags, drain window, and probe timeout, and execute a live cutover.
- **REST Endpoints**:
  - `POST /api/apps/{id}/rolling-restart`
  - `POST /api/apps/{id}/update` (JSON body: `command`, `args`, `env`, `image`, `drain_timeout`, `probe_timeout`)

---

## Kubernetes-like Multi-Node Cluster & Pods Scaling

Docroid introduces a lightweight, distributed clustering and pod scaling architecture inspired by Kubernetes, tailored specifically for mobile devices and edge computing.

Multiple devices running Docroid (such as Android smartphones via Termux, laptops, Raspberry Pis, or VPS instances) can interconnect across local Wi-Fi, LAN, or mesh networks (e.g. Tailscale). Workloads run as atomic **Pods**, managed by multi-replica **Deployments**, distributed via an intelligent **Scheduler**, horizontally scalable via CLI and Web UI, and reverse-proxied with an integrated **Cluster Service Load Balancer**.

```
                        ┌────────────────────────────────────────────────────────┐
                        │             Docroid Cluster Control Plane              │
                        │  (Device 1: e.g. Primary Phone / Laptop / Server)      │
                        │                                                        │
                        │  • Cluster Manager & Node Registry                     │
                        │  • Deployment & Pod Controller                         │
                        │  • Intelligent Scheduler (Least-Allocated / Affinity)  │
                        │  • Service Load Balancer (Round-Robin Proxy)           │
                        └───────────────▲────────────────────────▲───────────────┘
                                        │                        │
                    Heartbeats / Remote │ Pod Dispatch           │ Heartbeats / Remote Pod Dispatch
                                        │                        │
         ┌──────────────────────────────┴───────┐        ┌───────┴──────────────────────────────┐
         │         Worker Node: Device 2        │        │         Worker Node: Device 3        │
         │     (e.g. Android Termux Phone)      │        │        (e.g. Raspberry Pi / PC)      │
         │                                      │        │                                      │
         │  • Local AppManager                  │        │  • Local AppManager                  │
         │  • Pod Agent & Heartbeat Daemon      │        │  • Pod Agent & Heartbeat Daemon      │
         │  • Pod Replica #1 (web-api-1)        │        │  • Pod Replica #2 (web-api-2)        │
         └──────────────────────────────────────┘        └──────────────────────────────────────┘
```

### Core Concepts:
- **Node**: Any physical or virtual device running Docroid. Can act as `control-plane` (cluster leader) or `worker`. Real-time heartbeats track device capacity, CPU cores, memory, and health status (`Ready`, `NotReady`).
- **Pod**: The atomic deployable execution unit containing one or more containers, ports, environment variables, and execution parameters.
- **Deployment**: Declarative specification for a desired number of Pod replicas (e.g., 3 instances of `web-api`).
- **Scheduler**: Automatically places pods on healthy nodes using `least-allocated` capacity balancing or round-robin, respecting label affinities (e.g. `os=android`, `arch=arm64`).
- **Horizontal Pod Scaling**: Scale replicas up (1 → 5) or down (5 → 2) instantly via CLI (`docroid scale <app> <N>`) or from the Web Dashboard's interactive slider.
- **Self-Healing**: If a device disconnects or goes offline for >30 seconds, the control-plane automatically evicts its pods and reschedules replacements on surviving healthy nodes.
- **Cluster Service Proxy (Load Balancer)**: When a deployment specifies a `service_port` (e.g. `:8080`), Docroid runs a high-performance TCP/HTTP load balancer distributing incoming traffic across pods on all connected devices.

### 1. Initializing a Cluster (Device 1 - Control Plane)
On your primary phone or machine, initialize the cluster:
```bash
docroid cluster init --name "office-mesh" --port 9090
```
Output:
```
✓ Docroid Cluster 'office-mesh' initialized successfully!

  Role:          control-plane
  Control Plane: 192.168.1.100:9090
  Join Token:    dcr-tok-a8f93bc102...

To connect another device to this cluster, run:
  docroid cluster join 192.168.1.100:9090 --token dcr-tok-a8f93bc102...
```

### 2. Interconnecting Additional Devices (Device 2 - Worker Node)
On any other Android device (or PC/laptop) running Docroid:
```bash
docroid cluster join 192.168.1.100:9090 --token dcr-tok-a8f93bc102... --name "Pixel-7-Termux"
```

Check cluster status and list interconnected nodes:
```bash
docroid cluster status
docroid node ls
```
Output:
```
NAME               STATUS   ROLE                    AGE   VERSION   OS/ARCH        PODS   ADDRESS
MacBook-Host       Ready    control-plane (local)   15m   1.0.0     darwin/arm64   2      192.168.1.100:9090
Pixel-7-Termux     Ready    worker                  5m    1.0.0     android/arm64  2      192.168.1.105:9090
```

### 3. Deploying Scalable Pods
Deploy a service with multiple replicas and cluster load balancing:
```bash
docroid pod run "python3 -m http.server 8000" --name web-api --replicas 4 --service-port 8888
```
The scheduler immediately distributes the 4 replicas across your connected devices (e.g., 2 pods on `MacBook-Host` and 2 pods on `Pixel-7-Termux`).

List running pods across the cluster:
```bash
docroid pod ls
```
Output:
```
NAME            POD ID             DEPLOYMENT   DEVICE (NODE)    STATUS    PORT    RESTARTS   AGE
web-api-91a0f   pod-web-api-91a0f  web-api      MacBook-Host     Running   :8000   0          10s
web-api-81b4c   pod-web-api-81b4c  web-api      Pixel-7-Termux   Running   :8000   0          10s
web-api-72c3d   pod-web-api-72c3d  web-api      MacBook-Host     Running   :8001   0          10s
web-api-63d2e   pod-web-api-63d2e  web-api      Pixel-7-Termux   Running   :8001   0          10s
```

Test cluster load balancer:
```bash
curl http://localhost:8888
```
Requests arriving on port `8888` are automatically load balanced across all 4 pods running on both devices!

### 4. Horizontally Scaling Replicas
Scale your application up or down dynamically:
```bash
# Scale up to 6 replicas
docroid scale web-api 6

# Scale down to 2 replicas
docroid scale web-api 2
```

### 5. Inspecting Pod Logs & Terminating
```bash
# Stream stdout/stderr from any pod (local or remote device)
docroid pod logs pod-web-api-81b4c -n 50

# Terminate a single pod
docroid pod rm pod-web-api-81b4c

# Delete entire deployment and all its pods across all devices
docroid deployment rm web-api
```

---

## CLI Command Reference

| Command | Subcommand / Flags | Description |
|---------|--------------------|-------------|
| `docroid ui` | `-p <port>`, `--host <ip>` | Launch built-in web dashboard (default :9090) |
| `docroid cluster` | `init`, `join`, `status`, `leave` | Manage multi-device cluster mesh, tokens, and roles |
| `docroid node` | `ls`, `inspect <id>`, `rm <id>` | Inspect or manage interconnected device nodes |
| `docroid pod` | `run`, `ls`, `logs <id>`, `rm <id>` | Deploy, inspect, or manage atomic cluster pods |
| `docroid scale` | `<deployment> <replicas>` | Horizontally scale pod replicas across cluster devices |
| `docroid deployment`| `ls`, `rm <name>` | Manage multi-replica deployments |
| `docroid run` | `-d`, `-n <name>`, `-p [h:]p[:c]`, `-v h:c`, `--restart <pol>`, `--env-file <file>`, `--secret <k>` | Launch container / binary |
| `docroid ps` | `-a` | List running or all apps with ports and restart counters |
| `docroid secret` | `set`, `get`, `list`, `rm` | Manage secure key-values and encrypted tokens |
| `docroid stop` | `<name_or_id>` | Gracefully terminate app, supervisor, and proxy |
| `docroid start` | `<name_or_id>` | Start a stopped app and restore proxies/supervisors |
| `docroid restart` | `-z, --zero-downtime`, `<name_or_id>` | Restart app (standard or zero-downtime hot cutover) |
| `docroid update` | `--cmd`, `--image`, `-e`, `--env-file`, `--drain`, `--timeout` | Execute atomic zero-downtime rolling update |
| `docroid rm` | `-f` | Delete app, logs, and workspace |
| `docroid logs` | `-f`, `-n <lines>` | Stream stdout/stderr logs |
| `docroid top` | `<name_or_id>` | Display live CPU %, RSS RAM, and threads |
| `docroid exec` | `[OPTIONS] <name_or_id> [cmd]` | Execute command or interactive shell inside container (`-it`, `-d`, `-e`, `-w`, `-u`, `-f`) |
| `docroid build` | `-t <name:tag> [dir]` | Build portable image bundle from Docroidfile |
| `docroid export` | `-o <file.dpack> <target>` | Export image or app to portable `.dpack` package |
| `docroid import` | `-i <file.dpack>` | Import and register `.dpack` package into image store |
| `docroid images` | - | List local image repository |
| `docroid rmi` | `<name:tag>` | Remove installed image |
| `docroid rootfs` | `list`, `pull <name>`, `rm <name>` | Manage rootfs environments (Alpine ~3MB) |
| `docroid doctor` | - | System, kernel, and Termux diagnostics |
| `docroid version` | - | Show version and architecture |


---

## Android & Termux Battery Optimization Tips

1. **Keep Background Apps Alive with Screen Off**:
   Termux can be suspended by Android's battery saver. Run:
   ```bash
   termux-wake-lock
   ```
2. **Android 12+ Phantom Process Killer**:
   Android 12 introduced a limit of 32 child processes per app. To disable it via ADB:
   ```bash
   adb shell "/system/bin/device_config set_sync_disabled_for_tests persistent"
   adb shell "/system/bin/device_config put activity_manager max_phantom_processes 2147483647"
   ```

---

## License

MIT License. Designed for the open-source Android and Go communities.
