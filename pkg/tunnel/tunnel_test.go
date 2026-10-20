package tunnel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		target      string
		port        int
		expectedURL string
		expectedP   int
		expectErr   bool
	}{
		{"", 8080, "http://127.0.0.1:8080", 8080, false},
		{"8080", 0, "http://127.0.0.1:8080", 8080, false},
		{":3000", 0, "http://127.0.0.1:3000", 3000, false},
		{"localhost:9090", 0, "http://127.0.0.1:9090", 9090, false},
		{"127.0.0.1:5000", 0, "http://127.0.0.1:5000", 5000, false},
		{"http://localhost:4000", 0, "http://127.0.0.1:4000", 4000, false},
		{"http://127.0.0.1:8080/api", 0, "http://127.0.0.1:8080/api", 8080, false},
		{"https://localhost:8443", 0, "https://127.0.0.1:8443", 8443, false},
		{"invalid-port", 0, "", 0, true},
		{"", 0, "", 0, true},
		{"", 70000, "", 0, true},
	}

	for _, tt := range tests {
		u, p, err := NormalizeTarget(tt.target, tt.port)
		if tt.expectErr {
			if err == nil {
				t.Errorf("NormalizeTarget(%q, %d) expected error, got nil", tt.target, tt.port)
			}
		} else {
			if err != nil {
				t.Errorf("NormalizeTarget(%q, %d) unexpected error: %v", tt.target, tt.port, err)
			}
			if u != tt.expectedURL {
				t.Errorf("NormalizeTarget(%q, %d) expected URL %q, got %q", tt.target, tt.port, tt.expectedURL, u)
			}
			if p != tt.expectedP {
				t.Errorf("NormalizeTarget(%q, %d) expected Port %d, got %d", tt.target, tt.port, tt.expectedP, p)
			}
		}
	}
}

func TestExtractPublicURL(t *testing.T) {
	logContent := `
2026-10-20T18:09:05Z INF Requesting new quick Tunnel on trycloudflare.com...
2026-10-20T18:09:10Z INF +--------------------------------------------------------------------------------------------+
2026-10-20T18:09:10Z INF |  Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):  |
2026-10-20T18:09:10Z INF |  https://funeral-instrumental-connecticut-held.trycloudflare.com                           |
2026-10-20T18:09:10Z INF +--------------------------------------------------------------------------------------------+
`
	url := ExtractPublicURL(logContent)
	expected := "https://funeral-instrumental-connecticut-held.trycloudflare.com"
	if url != expected {
		t.Fatalf("ExtractPublicURL failed: expected %s, got %s", expected, url)
	}

	empty := ExtractPublicURL("No url in here yet")
	if empty != "" {
		t.Fatalf("ExtractPublicURL expected empty, got %s", empty)
	}
}

func TestStoreCRUD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tunnel-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	mgr := NewManager(store)

	// Create Tunnel
	tun, err := mgr.CreateTunnel(TunnelOptions{
		Name:       "Test-Service",
		Target:     "8080",
		TargetPort: 8080,
		Mode:       ModeQuick,
	})
	if err != nil {
		t.Fatalf("CreateTunnel failed: %v", err)
	}

	if tun.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if tun.Name != "test-service" {
		t.Fatalf("expected lowercase name test-service, got %s", tun.Name)
	}
	if tun.TargetPort != 8080 {
		t.Fatalf("expected target port 8080, got %d", tun.TargetPort)
	}

	// Get Tunnel
	fetched, err := store.GetTunnel(tun.ID)
	if err != nil {
		t.Fatalf("GetTunnel failed: %v", err)
	}
	if fetched.ID != tun.ID {
		t.Fatalf("fetched ID mismatch")
	}

	// Find by Name
	found, err := store.FindTunnel("test-service")
	if err != nil {
		t.Fatalf("FindTunnel by name failed: %v", err)
	}
	if found.ID != tun.ID {
		t.Fatalf("found ID mismatch")
	}

	// Find by Prefix
	foundPrefix, err := store.FindTunnel(tun.ID[:4])
	if err != nil {
		t.Fatalf("FindTunnel by prefix failed: %v", err)
	}
	if foundPrefix.ID != tun.ID {
		t.Fatalf("found prefix mismatch")
	}

	// List
	list, err := store.ListTunnels()
	if err != nil {
		t.Fatalf("ListTunnels failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(list))
	}

	// Delete
	if err := mgr.Delete(tun.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	afterList, err := store.ListTunnels()
	if err != nil {
		t.Fatalf("ListTunnels failed: %v", err)
	}
	if len(afterList) != 0 {
		t.Fatalf("expected 0 tunnels after delete, got %d", len(afterList))
	}
}

func TestDiscoverTargets(t *testing.T) {
	store, _ := NewStore(filepath.Join(os.TempDir(), "discover-test"))
	mgr := NewManager(store)

	apps := []struct {
		ID     string
		Name   string
		Port   int
		Status string
	}{
		{"app1", "web-api", 3000, "running"},
		{"app2", "worker", 0, "running"},
	}

	sites := []struct {
		ID     string
		Name   string
		Port   int
		Status string
	}{
		{"site1", "docs-site", 8080, "running"},
	}

	targets := mgr.DiscoverTargets(apps, sites, 9090)
	if len(targets) != 3 {
		t.Fatalf("expected 3 discoverable targets (dashboard + web-api + docs-site), got %d", len(targets))
	}
}
