package cluster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"docroid/pkg/app"
)

func setupTestStore(t *testing.T) (*app.Store, *app.AppManager, func()) {
	tmpDir, err := os.MkdirTemp("", "docroid-cluster-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	appStore, err := app.NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create app store: %v", err)
	}

	appMgr := app.NewAppManager(appStore)

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return appStore, appMgr, cleanup
}

func TestNodeDetectionAndMatching(t *testing.T) {
	node := DetectLocalNode(RoleControlPlane, "192.168.1.50:9090", "1.0.0")
	if node.ID == "" {
		t.Fatalf("expected non-empty node ID")
	}
	if node.Role != RoleControlPlane {
		t.Fatalf("expected role control-plane, got %s", node.Role)
	}
	if node.Status != NodeStatusReady {
		t.Fatalf("expected ready status")
	}

	// Selector matching
	if !node.MatchesSelector(map[string]string{"os": node.OS}) {
		t.Fatalf("expected node to match its own OS")
	}
	if node.MatchesSelector(map[string]string{"nonexistent": "val"}) {
		t.Fatalf("expected node to fail matching nonexistent label")
	}
}

func TestSchedulerLeastAllocated(t *testing.T) {
	scheduler := NewScheduler(StrategyLeastAllocated, 30*time.Second)

	node1 := &Node{
		ID:            "node-1",
		Name:          "device-1",
		Status:        NodeStatusReady,
		AllocatedPods: 4,
		LastHeartbeat: time.Now(),
		Capacity:      NodeCapacity{MaxPods: 10},
	}
	node2 := &Node{
		ID:            "node-2",
		Name:          "device-2",
		Status:        NodeStatusReady,
		AllocatedPods: 1, // Fewest pods
		LastHeartbeat: time.Now(),
		Capacity:      NodeCapacity{MaxPods: 10},
	}
	node3 := &Node{
		ID:            "node-3",
		Name:          "device-3",
		Status:        NodeStatusNotReady, // Not ready
		AllocatedPods: 0,
		LastHeartbeat: time.Now(),
		Capacity:      NodeCapacity{MaxPods: 10},
	}

	nodes := []*Node{node1, node2, node3}
	spec := &PodSpec{Name: "test-app"}

	selected, err := scheduler.SelectNode(spec, nodes)
	if err != nil {
		t.Fatalf("SelectNode error: %v", err)
	}
	if selected.ID != "node-2" {
		t.Fatalf("expected node-2 (least allocated), got %s", selected.ID)
	}
}

func TestSchedulerNodeFailureDetection(t *testing.T) {
	scheduler := NewScheduler(StrategyLeastAllocated, 1*time.Second)

	healthyNode := &Node{
		ID:            "node-healthy",
		Status:        NodeStatusReady,
		LastHeartbeat: time.Now(),
	}
	deadNode := &Node{
		ID:            "node-dead",
		Status:        NodeStatusReady,
		LastHeartbeat: time.Now().Add(-5 * time.Second), // Expired
	}

	nodes := []*Node{healthyNode, deadNode}
	pods := []*Pod{
		{ID: "pod-1", NodeID: "node-healthy", Status: PodRunning},
		{ID: "pod-2", NodeID: "node-dead", Status: PodRunning},
	}

	evictable := scheduler.FindEvictablePods(nodes, pods)
	if len(evictable) != 1 || evictable[0].ID != "pod-2" {
		t.Fatalf("expected pod-2 to be evictable, got %v", evictable)
	}
}

func TestClusterManagerInitAndScaling(t *testing.T) {
	appStore, appMgr, cleanup := setupTestStore(t)
	defer cleanup()

	mgr, err := NewManager(appStore, appMgr)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Stop()

	// 1. Initialize cluster
	cfg, err := mgr.InitCluster("test-mesh", "127.0.0.1:9090")
	if err != nil {
		t.Fatalf("InitCluster failed: %v", err)
	}
	if cfg.ClusterName != "test-mesh" {
		t.Fatalf("expected cluster name test-mesh, got %s", cfg.ClusterName)
	}

	status, err := mgr.GetClusterStatus()
	if err != nil || !status.Initialized {
		t.Fatalf("expected cluster to be initialized: %v", err)
	}
	if status.TotalNodes != 1 {
		t.Fatalf("expected 1 initial node, got %d", status.TotalNodes)
	}

	// 2. Register a simulated worker device
	workerJoinReq := &JoinRequest{
		Token: cfg.JoinToken,
		Node: Node{
			Name:    "Phone-Termux-Worker",
			Address: "192.168.1.105:9090",
			OS:      "android",
			Arch:    "arm64",
			Capacity: NodeCapacity{
				CPUCores: 8,
				MemoryMB: 4096,
				MaxPods:  20,
			},
		},
	}
	joinResp, err := mgr.RegisterWorker(workerJoinReq)
	if err != nil {
		t.Fatalf("RegisterWorker failed: %v", err)
	}
	if joinResp.AssignedID == "" {
		t.Fatalf("expected assigned ID for worker")
	}

	nodes := mgr.ListNodes()
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	// 3. Worker sends heartbeat
	hbReq := &HeartbeatRequest{
		NodeID:        joinResp.AssignedID,
		Token:         cfg.JoinToken,
		AllocatedPods: 0,
	}
	hbResp, err := mgr.ProcessHeartbeat(hbReq)
	if err != nil || !hbResp.Acknowledged {
		t.Fatalf("ProcessHeartbeat failed: %v", err)
	}

	// 4. Create a Deployment and scale
	dep := &Deployment{
		Name:     "echo-service",
		Replicas: 2,
		Template: PodSpec{
			Name:    "echo",
			Command: "echo",
			Args:    []string{"hello cluster"},
		},
	}

	err = mgr.CreateDeployment(dep)
	if err != nil {
		t.Fatalf("CreateDeployment failed: %v", err)
	}

	// Allow brief time for scale goroutine
	time.Sleep(200 * time.Millisecond)

	deployments := mgr.ListDeployments()
	if len(deployments) != 1 || deployments[0].Name != "echo-service" {
		t.Fatalf("expected 1 deployment named echo-service, got %v", deployments)
	}

	// Scale down to 1
	err = mgr.ScaleDeployment("echo-service", 1)
	if err != nil {
		t.Fatalf("ScaleDeployment to 1 failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	activePods := mgr.ListPods()
	if len(activePods) > 1 {
		t.Fatalf("expected 1 pod after scale down, got %d", len(activePods))
	}

	// Clean up deployment
	err = mgr.DeleteDeployment("echo-service")
	if err != nil {
		t.Fatalf("DeleteDeployment failed: %v", err)
	}
}

func TestInterDeviceClient(t *testing.T) {
	// Mock HTTP server acting as remote node
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/cluster/join":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"cluster_name":"mock-mesh","assigned_id":"node-xyz"}`))
		case "/api/cluster/heartbeat":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		case "/api/cluster/node/run-pod":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"pod_id":"pod-123","local_app_id":"app-abc","host_port":8080,"status":"Running"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := NewClient(2 * time.Second)
	ctx := context.Background()

	// Test join
	joinResp, err := client.JoinCluster(ctx, ts.URL, &JoinRequest{Token: "test-token"})
	if err != nil {
		t.Fatalf("client.JoinCluster failed: %v", err)
	}
	if joinResp.AssignedID != "node-xyz" {
		t.Fatalf("expected node-xyz, got %s", joinResp.AssignedID)
	}

	// Test heartbeat
	hbResp, err := client.SendHeartbeat(ctx, ts.URL, &HeartbeatRequest{NodeID: "node-xyz", Token: "test-token"})
	if err != nil || !hbResp.Acknowledged {
		t.Fatalf("client.SendHeartbeat failed: %v", err)
	}

	// Test pod dispatch
	pod := &Pod{ID: "pod-123", Name: "mock-pod"}
	runResp, err := client.DispatchPod(ctx, ts.URL, "test-token", pod)
	if err != nil {
		t.Fatalf("client.DispatchPod failed: %v", err)
	}
	if runResp.HostPort != 8080 {
		t.Fatalf("expected host port 8080, got %d", runResp.HostPort)
	}
}

func TestServiceProxy(t *testing.T) {
	// Create mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello from pod"))
	}))
	defer backend.Close()

	// Strip http://
	backendAddr := backend.URL[7:]

	pm := NewProxyManager()
	defer pm.StopAll()

	// Ensure proxy on a free port
	err := pm.EnsureProxy("web-svc", 18888, []string{backendAddr})
	if err != nil {
		t.Fatalf("EnsureProxy failed: %v", err)
	}

	// Test proxy forwarding
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:18888/")
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from proxy, got %d", resp.StatusCode)
	}
}
