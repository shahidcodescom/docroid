package rolling

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/netproxy"
	"docroid/pkg/utils"
)

func TestZeroDowntimeRolloutSuccess(t *testing.T) {
	tempDir := t.TempDir()
	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	manager := app.NewAppManager(store)

	// Create a tiny mock HTTP server script that prints version based on env or arg
	scriptPath := filepath.Join(tempDir, "server.sh")
	scriptContent := `#!/bin/sh
echo "Server starting on port $PORT..."
python3 -c '
import http.server, os, sys
port = int(os.environ.get("PORT", 8080))
ver = os.environ.get("VERSION", "V1")
class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(ver.encode())
    def log_message(self, format, *args):
        pass
http.server.HTTPServer(("127.0.0.1", port), Handler).serve_forever()
'
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	hostPort, err := utils.GetFreePort()
	if err != nil {
		t.Fatalf("failed to get free host port: %v", err)
	}

	initialPort, err := utils.GetFreePort()
	if err != nil {
		t.Fatalf("failed to get free initial port: %v", err)
	}

	// 1. Create and start initial app
	application, err := manager.CreateApp(
		"zero-downtime-test",
		app.EngineNative,
		"/bin/sh",
		[]string{scriptPath},
		map[string]string{"VERSION": "V1", "PORT": fmt.Sprintf("%d", initialPort)},
		tempDir,
		initialPort,
		nil,
		"",
		"no",
	)
	if err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	mapping := &netproxy.PortMapping{
		HostIP:        "127.0.0.1",
		HostPort:      hostPort,
		ContainerPort: initialPort,
	}
	application.PortMapping = mapping.String()
	_ = store.SaveApp(application)

	// Start proxy
	if err := netproxy.GetProxyManager().StartProxy(application.ID, mapping); err != nil {
		t.Fatalf("StartProxy failed: %v", err)
	}
	defer netproxy.GetProxyManager().StopProxies(application.ID)

	// Start app
	if _, err := StartApp(context.Background(), store, application); err != nil {
		t.Fatalf("StartApp failed: %v", err)
	}
	defer manager.StopApp(application.ID)

	// Wait for server V1 to respond
	testURL := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	var resp *http.Response
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		r, err := http.Get(testURL)
		if err == nil && r.StatusCode == 200 {
			resp = r
			break
		}
	}
	if resp == nil {
		t.Fatalf("initial server V1 failed to start on %s", testURL)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "V1" {
		t.Fatalf("expected V1, got %s", string(body))
	}

	// 2. Perform Zero-Downtime Rollout to V2
	opts := RolloutOptions{
		NewEnv:       map[string]string{"VERSION": "V2"},
		ProbeTimeout: 5 * time.Second,
		DrainTimeout: 500 * time.Millisecond,
	}

	result, err := ExecuteZeroDowntimeRestart(context.Background(), store, manager, application.ID, opts)
	if err != nil {
		t.Fatalf("ExecuteZeroDowntimeRestart failed: %v", err)
	}

	if result.RolledBack {
		t.Fatalf("expected successful rollout, but was rolled back: %s", result.Message)
	}

	// 3. Connect to the SAME hostPort and verify we now get V2
	r2, err := http.Get(testURL)
	if err != nil {
		t.Fatalf("request after rollout failed: %v", err)
	}
	body2, _ := io.ReadAll(r2.Body)
	r2.Body.Close()

	if string(body2) != "V2" {
		t.Fatalf("expected V2 after cutover, got %s", string(body2))
	}
}

func TestZeroDowntimeRollbackOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	store, err := app.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	manager := app.NewAppManager(store)

	scriptPath := filepath.Join(tempDir, "server.sh")
	scriptContent := `#!/bin/sh
python3 -c '
import http.server, os
port = int(os.environ.get("PORT", 8080))
class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"STABLE_OLD")
    def log_message(self, *a): pass
http.server.HTTPServer(("127.0.0.1", port), Handler).serve_forever()
'
`
	_ = os.WriteFile(scriptPath, []byte(scriptContent), 0755)

	hostPort, _ := utils.GetFreePort()
	initialPort, _ := utils.GetFreePort()

	application, _ := manager.CreateApp(
		"rollback-test",
		app.EngineNative,
		"/bin/sh",
		[]string{scriptPath},
		map[string]string{"PORT": fmt.Sprintf("%d", initialPort)},
		tempDir,
		initialPort,
		nil,
		"",
		"no",
	)

	mapping := &netproxy.PortMapping{HostIP: "127.0.0.1", HostPort: hostPort, ContainerPort: initialPort}
	application.PortMapping = mapping.String()
	_ = store.SaveApp(application)
	_ = netproxy.GetProxyManager().StartProxy(application.ID, mapping)
	defer netproxy.GetProxyManager().StopProxies(application.ID)

	_, _ = StartApp(context.Background(), store, application)
	defer manager.StopApp(application.ID)

	testURL := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if r, err := http.Get(testURL); err == nil && r.StatusCode == 200 {
			r.Body.Close()
			break
		}
	}

	oldPID := application.PID

	// Attempt rollout with a command that crashes immediately
	brokenOpts := RolloutOptions{
		NewCommand:   "/bin/sh",
		NewArgs:      []string{"-c", "exit 1"},
		ProbeTimeout: 1 * time.Second,
	}

	_, err = ExecuteZeroDowntimeRestart(context.Background(), store, manager, application.ID, brokenOpts)
	if err == nil {
		t.Fatal("expected rollout to fail and trigger rollback, got nil error")
	}

	// Verify old process is STILL running and serving requests with ZERO downtime!
	r, err := http.Get(testURL)
	if err != nil {
		t.Fatalf("old server unreachable after rollback: %v", err)
	}
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()

	if string(body) != "STABLE_OLD" {
		t.Fatalf("expected STABLE_OLD preserved, got %s", string(body))
	}

	freshApp, _ := store.FindApp(application.ID)
	if freshApp.PID != oldPID {
		t.Fatalf("expected old PID %d to be preserved, got %d", oldPID, freshApp.PID)
	}
}
