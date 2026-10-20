package network

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNetworkCreationAndDefaults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "docroid-net-test-*")
	if err != nil {
		t.Fatalf("failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Verify default networks 'bridge' and 'host'
	nets, err := mgr.ListNetworks()
	if err != nil {
		t.Fatalf("failed to list networks: %v", err)
	}
	if len(nets) < 2 {
		t.Fatalf("expected at least 2 default networks, got %d", len(nets))
	}

	bridge, err := mgr.GetNetwork("bridge")
	if err != nil || bridge == nil {
		t.Fatalf("expected bridge network to exist, got err: %v", err)
	}
	if bridge.Driver != DriverBridge {
		t.Errorf("expected bridge driver, got %s", bridge.Driver)
	}

	host, err := mgr.GetNetwork("host")
	if err != nil || host == nil {
		t.Fatalf("expected host network to exist, got err: %v", err)
	}

	// Create user-defined network
	custom, err := mgr.CreateNetwork("my-app-net", DriverBridge, "10.90.0.0/16", map[string]string{"env": "test"})
	if err != nil {
		t.Fatalf("failed to create custom network: %v", err)
	}
	if custom.Name != "my-app-net" {
		t.Errorf("expected name my-app-net, got %s", custom.Name)
	}
	if custom.Gateway != "10.90.0.1" {
		t.Errorf("expected gateway 10.90.0.1, got %s", custom.Gateway)
	}

	// Duplicate creation should fail
	_, err = mgr.CreateNetwork("my-app-net", DriverBridge, "", nil)
	if err == nil {
		t.Errorf("expected duplicate creation to fail")
	}
}

func TestContainerConnectionAndResolution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "docroid-net-test-*")
	if err != nil {
		t.Fatalf("failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewNetworkManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	netName := "test-intercom"
	_, err = mgr.CreateNetwork(netName, DriverBridge, "10.95.0.0/16", nil)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

	// Connect container 1 (database)
	ep1, err := mgr.Connect(netName, "c1-db", "postgres-db", 5432, []string{"db", "database"})
	if err != nil {
		t.Fatalf("failed to connect c1: %v", err)
	}
	if ep1.IPv4Address != "10.95.0.2" {
		t.Errorf("expected IP 10.95.0.2, got %s", ep1.IPv4Address)
	}

	// Connect container 2 (web api)
	ep2, err := mgr.Connect(netName, "c2-web", "web-api", 8080, []string{"web", "api"})
	if err != nil {
		t.Fatalf("failed to connect c2: %v", err)
	}
	if ep2.IPv4Address != "10.95.0.3" {
		t.Errorf("expected IP 10.95.0.3, got %s", ep2.IPv4Address)
	}

	// Test hosts content generation for c2 (web api)
	hostsContent, err := mgr.GenerateHostsContent(netName, "c2-web")
	if err != nil {
		t.Fatalf("failed to generate hosts: %v", err)
	}

	// Should contain peer entries: postgres-db, db, database
	if !strings.Contains(hostsContent, "postgres-db") || !strings.Contains(hostsContent, "db") {
		t.Errorf("expected hosts content to contain 'postgres-db' and 'db', got:\n%s", hostsContent)
	}
	// Should resolve to 127.0.0.1
	if !strings.Contains(hostsContent, "127.0.0.1\tpostgres-db db database") {
		t.Errorf("expected mapping '127.0.0.1\\tpostgres-db db database', got:\n%s", hostsContent)
	}

	// Test Docker-style Link Environment Variables for c2 (web api)
	envs, err := mgr.GenerateLinkEnvs(netName, "c2-web")
	if err != nil {
		t.Fatalf("failed to generate link envs: %v", err)
	}

	if envs["DB_HOST"] != "127.0.0.1" || envs["DB_PORT"] != "5432" {
		t.Errorf("expected DB_HOST=127.0.0.1 and DB_PORT=5432, got DB_HOST=%s DB_PORT=%s", envs["DB_HOST"], envs["DB_PORT"])
	}
	if envs["POSTGRES_DB_PORT"] != "5432" {
		t.Errorf("expected POSTGRES_DB_PORT=5432, got %s", envs["POSTGRES_DB_PORT"])
	}

	// Test writing hosts file to workspace
	wsDir := filepath.Join(tmpDir, "ws-web")
	hostsPath, err := mgr.WriteContainerHosts(wsDir, netName, "c2-web")
	if err != nil {
		t.Fatalf("failed to write container hosts: %v", err)
	}

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatalf("failed to read hosts file: %v", err)
	}
	if !strings.Contains(string(data), "postgres-db") {
		t.Errorf("expected written hosts file to contain postgres-db")
	}

	// Test Disconnect
	err = mgr.Disconnect(netName, "c1-db")
	if err != nil {
		t.Fatalf("failed to disconnect c1: %v", err)
	}

	// Hosts content should no longer have c1
	newHosts, _ := mgr.GenerateHostsContent(netName, "c2-web")
	if strings.Contains(newHosts, "postgres-db") {
		t.Errorf("disconnected container postgres-db should not appear in hosts")
	}

	// Remove network
	err = mgr.RemoveNetwork(netName)
	// Should fail because c2 is still connected
	if err == nil {
		t.Errorf("expected deletion to fail while container connected")
	}

	_ = mgr.Disconnect(netName, "c2-web")
	err = mgr.RemoveNetwork(netName)
	if err != nil {
		t.Fatalf("failed to remove empty network: %v", err)
	}
}
