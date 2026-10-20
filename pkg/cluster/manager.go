package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/engine"
	"docroid/pkg/names"
)

// ClusterStatus contains summary information about the cluster
type ClusterStatus struct {
	Initialized     bool      `json:"initialized"`
	ClusterName     string    `json:"cluster_name"`
	Role            NodeRole  `json:"role"`
	LocalNode       *Node     `json:"local_node"`
	JoinToken       string    `json:"join_token,omitempty"`
	LeaderAddr      string    `json:"leader_addr,omitempty"`
	TotalNodes      int       `json:"total_nodes"`
	ReadyNodes      int       `json:"ready_nodes"`
	TotalPods       int       `json:"total_pods"`
	RunningPods     int       `json:"running_pods"`
	TotalReplicas   int       `json:"total_replicas"`
	InitializedAt   time.Time `json:"initialized_at"`
}

// Manager coordinates cluster membership, pod scheduling, and horizontal scaling
type Manager struct {
	store       *Store
	appStore    *app.Store
	appManager  *app.AppManager
	scheduler   *Scheduler
	client      *Client
	proxyMgr    *ProxyManager
	config      *ClusterConfig
	localNode   *Node
	nodes       map[string]*Node       // nodeID -> Node
	deployments map[string]*Deployment // name -> Deployment
	pods        map[string]*Pod        // podID -> Pod
	stopCh      chan struct{}
	mu          sync.RWMutex
}

// NewManager creates a new cluster manager
func NewManager(appStore *app.Store, appMgr *app.AppManager) (*Manager, error) {
	clusterStore, err := NewStore(appStore.BaseDir())
	if err != nil {
		return nil, err
	}

	m := &Manager{
		store:       clusterStore,
		appStore:    appStore,
		appManager:  appMgr,
		scheduler:   NewScheduler(StrategyLeastAllocated, 30*time.Second),
		client:      NewClient(10 * time.Second),
		proxyMgr:    NewProxyManager(),
		nodes:       make(map[string]*Node),
		deployments: make(map[string]*Deployment),
		pods:        make(map[string]*Pod),
		stopCh:      make(chan struct{}),
	}

	// Load existing cluster configuration if available
	cfg, err := clusterStore.LoadConfig()
	if err == nil && cfg != nil {
		m.config = cfg
		m.localNode = DetectLocalNode(cfg.Role, cfg.ControlPlaneAddress, "1.0.0")
		m.localNode.ID = cfg.LocalNodeID

		// Load state
		nodes, _ := clusterStore.LoadNodes()
		for _, n := range nodes {
			if n.ID == m.localNode.ID {
				n.IsLocal = true
				m.localNode = n
			}
			m.nodes[n.ID] = n
		}
		if _, exists := m.nodes[m.localNode.ID]; !exists {
			m.nodes[m.localNode.ID] = m.localNode
		}

		deps, _ := clusterStore.LoadDeployments()
		for _, d := range deps {
			m.deployments[d.Name] = d
		}

		pods, _ := clusterStore.LoadPods()
		for _, p := range pods {
			m.pods[p.ID] = p
		}

		m.startBackgroundLoops()
	}

	return m, nil
}

// InitCluster configures this node as the Control Plane
func (m *Manager) InitCluster(clusterName, listenAddr string) (*ClusterConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clusterName == "" {
		clusterName = "docroid-cluster"
	}

	if listenAddr == "" {
		listenAddr = GetLocalOutboundIP() + ":9090"
	} else {
		listenAddr = CleanAddress(listenAddr, 9090)
	}

	token := GenerateJoinToken()
	localNode := DetectLocalNode(RoleControlPlane, listenAddr, "1.0.0")
	localNode.IsLocal = true

	cfg := &ClusterConfig{
		ClusterName:         clusterName,
		Role:                RoleControlPlane,
		JoinToken:           token,
		ControlPlaneAddress: listenAddr,
		LocalNodeID:         localNode.ID,
		InitializedAt:       time.Now(),
	}

	if err := m.store.SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save cluster config: %w", err)
	}

	m.config = cfg
	m.localNode = localNode
	m.nodes = map[string]*Node{localNode.ID: localNode}
	_ = m.store.SaveNodes([]*Node{localNode})

	m.startBackgroundLoops()
	return cfg, nil
}

// JoinCluster joins an existing cluster as a Worker node
func (m *Manager) JoinCluster(leaderAddr, token, customName, myAddr string) (*ClusterConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("cluster join token is required")
	}
	if leaderAddr == "" {
		return nil, fmt.Errorf("cluster control-plane address is required")
	}

	if myAddr == "" {
		myAddr = GetLocalOutboundIP() + ":9090"
	} else {
		myAddr = CleanAddress(myAddr, 9090)
	}

	localNode := DetectLocalNode(RoleWorker, myAddr, "1.0.0")
	localNode.IsLocal = true
	if customName != "" {
		localNode.Name = customName
	}

	req := &JoinRequest{
		Token: token,
		Node:  *localNode,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	joinResp, err := m.client.JoinCluster(ctx, leaderAddr, req)
	if err != nil {
		return nil, err
	}

	if joinResp.AssignedID != "" {
		localNode.ID = joinResp.AssignedID
	}

	cfg := &ClusterConfig{
		ClusterName:         joinResp.ClusterName,
		Role:                RoleWorker,
		JoinToken:           token,
		ControlPlaneAddress: leaderAddr,
		LocalNodeID:         localNode.ID,
		InitializedAt:       time.Now(),
	}

	if err := m.store.SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to persist join config: %w", err)
	}

	m.config = cfg
	m.localNode = localNode
	m.nodes = map[string]*Node{localNode.ID: localNode}
	for _, n := range joinResp.Nodes {
		if n.ID != localNode.ID {
			m.nodes[n.ID] = n
		}
	}
	_ = m.store.SaveNodes(m.getNodeListUnsafe())

	m.startBackgroundLoops()
	return cfg, nil
}

// LeaveCluster disconnects the node from the cluster
func (m *Manager) LeaveCluster() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop loops
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
		m.stopCh = make(chan struct{})
	}

	// Stop proxies
	m.proxyMgr.StopAll()

	// Terminate local pods
	for _, p := range m.pods {
		if p.NodeID == m.localNode.ID && p.LocalAppID != "" {
			_ = m.appManager.StopApp(p.LocalAppID)
		}
	}

	m.config = nil
	m.localNode = nil
	m.nodes = make(map[string]*Node)
	m.deployments = make(map[string]*Deployment)
	m.pods = make(map[string]*Pod)

	return m.store.ClearConfig()
}

// GetClusterStatus returns summary of cluster state
func (m *Manager) GetClusterStatus() (*ClusterStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil || m.localNode == nil {
		return &ClusterStatus{Initialized: false}, nil
	}

	readyNodes := 0
	for _, n := range m.nodes {
		if n.Status == NodeStatusReady && n.IsAlive(30*time.Second) {
			readyNodes++
		}
	}

	runningPods := 0
	for _, p := range m.pods {
		if p.Status == PodRunning {
			runningPods++
		}
	}

	totalReplicas := 0
	for _, d := range m.deployments {
		totalReplicas += d.Replicas
	}

	return &ClusterStatus{
		Initialized:   true,
		ClusterName:   m.config.ClusterName,
		Role:          m.config.Role,
		LocalNode:     m.localNode,
		JoinToken:     m.config.JoinToken,
		LeaderAddr:    m.config.ControlPlaneAddress,
		TotalNodes:    len(m.nodes),
		ReadyNodes:    readyNodes,
		TotalPods:     len(m.pods),
		RunningPods:   runningPods,
		TotalReplicas: totalReplicas,
		InitializedAt: m.config.InitializedAt,
	}, nil
}

// Config returns active cluster config
func (m *Manager) Config() *ClusterConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// LocalNode returns the local node instance
func (m *Manager) LocalNode() *Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.localNode
}

// ListNodes returns all known cluster nodes
func (m *Manager) ListNodes() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getNodeListUnsafe()
}

func (m *Manager) getNodeListUnsafe() []*Node {
	list := make([]*Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		// Clone node to avoid race conditions
		clone := *n
		if clone.ID == m.localNode.ID {
			clone.IsLocal = true
			clone.Status = NodeStatusReady
			clone.LastHeartbeat = time.Now()
		} else if !clone.IsAlive(30 * time.Second) {
			clone.Status = NodeStatusNotReady
		}
		list = append(list, &clone)
	}
	return list
}

// RegisterWorker is called on Control Plane when a worker joins
func (m *Manager) RegisterWorker(req *JoinRequest) (*JoinResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil || m.config.Role != RoleControlPlane {
		return nil, fmt.Errorf("this node is not a cluster control-plane")
	}

	if req.Token != m.config.JoinToken {
		return nil, fmt.Errorf("invalid join token")
	}

	workerNode := req.Node
	if workerNode.ID == "" {
		workerNode.ID = GenerateNodeID()
	}
	workerNode.Role = RoleWorker
	workerNode.Status = NodeStatusReady
	workerNode.LastHeartbeat = time.Now()
	workerNode.JoinedAt = time.Now()
	workerNode.IsLocal = false

	m.nodes[workerNode.ID] = &workerNode
	_ = m.store.SaveNodes(m.getNodeListUnsafe())

	return &JoinResponse{
		ClusterName: m.config.ClusterName,
		LeaderAddr:  m.config.ControlPlaneAddress,
		AssignedID:  workerNode.ID,
		Nodes:       m.getNodeListUnsafe(),
	}, nil
}

// ProcessHeartbeat is called on Control Plane when a worker pings
func (m *Manager) ProcessHeartbeat(req *HeartbeatRequest) (*HeartbeatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil || m.config.Role != RoleControlPlane {
		return nil, fmt.Errorf("not a cluster control-plane")
	}

	if req.Token != m.config.JoinToken {
		return nil, fmt.Errorf("unauthorized heartbeat token")
	}

	node, exists := m.nodes[req.NodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not registered", req.NodeID)
	}

	node.LastHeartbeat = time.Now()
	node.Status = NodeStatusReady
	node.AllocatedPods = req.AllocatedPods

	// Update reported pod statuses
	for podID, statusStr := range req.PodStatuses {
		if pod, exists := m.pods[podID]; exists {
			pod.Status = PodStatus(statusStr)
		}
	}

	return &HeartbeatResponse{Acknowledged: true}, nil
}

// RemoveNode removes a node from the cluster
func (m *Manager) RemoveNode(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node, exists := m.nodes[nodeID]; exists {
		if node.IsLocal {
			return fmt.Errorf("cannot remove local control-plane node; use 'docroid cluster leave' instead")
		}
		delete(m.nodes, nodeID)
		_ = m.store.SaveNodes(m.getNodeListUnsafe())

		// Mark its pods as Failed so reconciler reschedules them
		for _, p := range m.pods {
			if p.NodeID == nodeID {
				p.Status = PodFailed
				p.ErrorMessage = "Node removed from cluster"
			}
		}
		_ = m.store.SavePods(m.getPodListUnsafe())
	}
	return nil
}

// --- Deployments & Pods Management ---

// CreateDeployment registers a new deployment and scales to target replicas
func (m *Manager) CreateDeployment(dep *Deployment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if dep.Name == "" {
		dep.Name = names.GenerateRandomName()
	}
	if dep.Namespace == "" {
		dep.Namespace = "default"
	}
	if dep.Replicas <= 0 {
		dep.Replicas = 1
	}
	dep.CreatedAt = time.Now()
	dep.UpdatedAt = time.Now()

	m.deployments[dep.Name] = dep
	_ = m.store.SaveDeployments(m.getDeploymentListUnsafe())

	// Trigger scaling to meet replicas
	go func(name string, count int) {
		_ = m.ScaleDeployment(name, count)
	}(dep.Name, dep.Replicas)

	return nil
}

// GetDeployment returns deployment by name
func (m *Manager) GetDeployment(name string) (*Deployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dep, exists := m.deployments[name]
	if !exists {
		return nil, fmt.Errorf("deployment %s not found", name)
	}
	return dep, nil
}

// ListDeployments returns all deployments
func (m *Manager) ListDeployments() []*Deployment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getDeploymentListUnsafe()
}

func (m *Manager) getDeploymentListUnsafe() []*Deployment {
	list := make([]*Deployment, 0, len(m.deployments))
	for _, d := range m.deployments {
		clone := *d
		// Calculate available replicas dynamically
		runningCount := 0
		for _, p := range m.pods {
			if p.Deployment == d.Name && p.Status == PodRunning {
				runningCount++
			}
		}
		clone.AvailableReplicas = runningCount
		list = append(list, &clone)
	}
	return list
}

// DeleteDeployment removes a deployment and its pods
func (m *Manager) DeleteDeployment(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.deployments[name]; !exists {
		return fmt.Errorf("deployment %s not found", name)
	}

	delete(m.deployments, name)
	_ = m.store.SaveDeployments(m.getDeploymentListUnsafe())

	// Stop service proxy
	m.proxyMgr.StopProxy(name)

	// Terminate all pods for this deployment
	var podsToStop []*Pod
	for _, p := range m.pods {
		if p.Deployment == name {
			podsToStop = append(podsToStop, p)
			delete(m.pods, p.ID)
		}
	}
	_ = m.store.SavePods(m.getPodListUnsafe())

	// Stop pods asynchronously
	go func(pods []*Pod) {
		for _, p := range pods {
			m.stopPodInstance(p)
		}
	}(podsToStop)

	return nil
}

// ScaleDeployment scales the deployment to the requested number of replicas
func (m *Manager) ScaleDeployment(name string, targetReplicas int) error {
	m.mu.Lock()

	dep, exists := m.deployments[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("deployment %s not found", name)
	}

	if targetReplicas < 0 {
		targetReplicas = 0
	}
	dep.Replicas = targetReplicas
	dep.UpdatedAt = time.Now()
	_ = m.store.SaveDeployments(m.getDeploymentListUnsafe())

	// Find current active pods for this deployment
	var activePods []*Pod
	for _, p := range m.pods {
		if p.Deployment == name && (p.Status == PodRunning || p.Status == PodPending) {
			activePods = append(activePods, p)
		}
	}

	currentCount := len(activePods)
	m.mu.Unlock()

	if currentCount < targetReplicas {
		// Scale UP: schedule additional pods
		toAdd := targetReplicas - currentCount
		for i := 0; i < toAdd; i++ {
			if err := m.launchPodReplica(dep); err != nil {
				return fmt.Errorf("failed to scale up pod: %w", err)
			}
		}
	} else if currentCount > targetReplicas {
		// Scale DOWN: select and stop excess pods
		toRemove := currentCount - targetReplicas
		for i := 0; i < toRemove && i < len(activePods); i++ {
			p := activePods[i]
			m.mu.Lock()
			delete(m.pods, p.ID)
			if node, ok := m.nodes[p.NodeID]; ok && node.AllocatedPods > 0 {
				node.AllocatedPods--
			}
			_ = m.store.SavePods(m.getPodListUnsafe())
			m.mu.Unlock()

			go m.stopPodInstance(p)
		}
	}

	m.refreshServiceProxy(name)
	return nil
}

// launchPodReplica selects a node and runs a new pod instance
func (m *Manager) launchPodReplica(dep *Deployment) error {
	m.mu.Lock()
	nodes := m.getNodeListUnsafe()
	node, err := m.scheduler.SelectNode(&dep.Template, nodes)
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("scheduling failed: %w", err)
	}

	podID := GeneratePodID(dep.Name)
	pod := &Pod{
		ID:           podID,
		Name:         fmt.Sprintf("%s-%s", dep.Name, podID[len(podID)-5:]),
		Namespace:    dep.Namespace,
		Deployment:   dep.Name,
		NodeID:       node.ID,
		NodeName:     node.Name,
		NodeAddress:  node.Address,
		Status:       PodPending,
		Spec:         dep.Template,
		RestartCount: 0,
		CreatedAt:    time.Now(),
	}

	// Update node allocation
	if actualNode, exists := m.nodes[node.ID]; exists {
		actualNode.AllocatedPods++
	}
	m.pods[pod.ID] = pod
	_ = m.store.SavePods(m.getPodListUnsafe())
	_ = m.store.SaveNodes(m.getNodeListUnsafe())

	isLocal := node.IsLocal || (m.localNode != nil && node.ID == m.localNode.ID)
	token := ""
	if m.config != nil {
		token = m.config.JoinToken
	}
	m.mu.Unlock()

	// Execute pod on target node
	if isLocal {
		resp, err := m.ExecuteLocalPod(pod)
		m.mu.Lock()
		if err != nil {
			pod.Status = PodFailed
			pod.ErrorMessage = err.Error()
		} else {
			pod.Status = PodRunning
			pod.LocalAppID = resp.LocalAppID
			pod.HostPort = resp.HostPort
			pod.PID = resp.PID
			pod.StartedAt = time.Now()
		}
		_ = m.store.SavePods(m.getPodListUnsafe())
		m.mu.Unlock()
	} else {
		// Remote node dispatch over HTTP
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resp, err := m.client.DispatchPod(ctx, node.Address, token, pod)
		m.mu.Lock()
		if err != nil {
			pod.Status = PodFailed
			pod.ErrorMessage = err.Error()
		} else {
			pod.Status = PodRunning
			pod.LocalAppID = resp.LocalAppID
			pod.HostPort = resp.HostPort
			pod.PID = resp.PID
			pod.StartedAt = time.Now()
		}
		_ = m.store.SavePods(m.getPodListUnsafe())
		m.mu.Unlock()
	}

	m.refreshServiceProxy(dep.Name)
	return nil
}

// stopPodInstance terminates pod either locally or via remote node
func (m *Manager) stopPodInstance(pod *Pod) {
	if pod == nil {
		return
	}

	isLocal := (m.localNode != nil && pod.NodeID == m.localNode.ID)
	if isLocal {
		_ = m.StopLocalPod(pod.ID)
	} else {
		token := ""
		m.mu.RLock()
		if m.config != nil {
			token = m.config.JoinToken
		}
		m.mu.RUnlock()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = m.client.StopPod(ctx, pod.NodeAddress, token, pod.ID)
	}
}

// refreshServiceProxy updates active endpoints for a deployment's service port
func (m *Manager) refreshServiceProxy(deploymentName string) {
	m.mu.RLock()
	dep, exists := m.deployments[deploymentName]
	if !exists || dep.ServicePort <= 0 {
		m.mu.RUnlock()
		return
	}

	var endpoints []string
	for _, p := range m.pods {
		if p.Deployment == deploymentName && p.Status == PodRunning && p.HostPort > 0 {
			ep := p.Endpoint()
			if ep != "" {
				endpoints = append(endpoints, ep)
			}
		}
	}
	m.mu.RUnlock()

	_ = m.proxyMgr.EnsureProxy(deploymentName, dep.ServicePort, endpoints)
}

// ListPods returns all known pods
func (m *Manager) ListPods() []*Pod {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getPodListUnsafe()
}

func (m *Manager) getPodListUnsafe() []*Pod {
	list := make([]*Pod, 0, len(m.pods))
	for _, p := range m.pods {
		clone := *p
		list = append(list, &clone)
	}
	return list
}

// GetPod retrieves a single pod by ID
func (m *Manager) GetPod(podID string) (*Pod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pod, exists := m.pods[podID]
	if !exists {
		return nil, fmt.Errorf("pod %s not found", podID)
	}
	return pod, nil
}

// DeletePod deletes a single pod
func (m *Manager) DeletePod(podID string) error {
	m.mu.Lock()
	pod, exists := m.pods[podID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("pod %s not found", podID)
	}

	delete(m.pods, podID)
	if node, ok := m.nodes[pod.NodeID]; ok && node.AllocatedPods > 0 {
		node.AllocatedPods--
	}
	_ = m.store.SavePods(m.getPodListUnsafe())
	_ = m.store.SaveNodes(m.getNodeListUnsafe())
	depName := pod.Deployment
	m.mu.Unlock()

	go m.stopPodInstance(pod)
	if depName != "" {
		m.refreshServiceProxy(depName)
	}
	return nil
}

// --- Local Pod Execution (Runner) ---

// ExecuteLocalPod launches a pod on the local machine using Docroid's AppManager
func (m *Manager) ExecuteLocalPod(pod *Pod) (*RunPodResponse, error) {
	spec := pod.Spec

	// Resolve command & args
	cmdStr := spec.Command
	cmdArgs := spec.Args
	if cmdStr == "" && len(cmdArgs) > 0 {
		cmdStr = cmdArgs[0]
		cmdArgs = cmdArgs[1:]
	}
	if cmdStr == "" {
		return nil, fmt.Errorf("pod command cannot be empty")
	}

	// Engine
	engineType := app.EngineNative
	if strings.ToLower(spec.Engine) == "rootfs" {
		engineType = app.EngineRootfs
	}

	var volumes []app.VolumeMount
	for _, v := range spec.Volumes {
		parts := strings.Split(v, ":")
		if len(parts) >= 2 {
			volumes = append(volumes, app.VolumeMount{
				HostPath:      parts[0],
				ContainerPath: parts[1],
			})
		}
	}

	restartPolicy := spec.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = "no"
	}

	localApp, err := m.appManager.CreateApp(
		pod.Name,
		engineType,
		cmdStr,
		cmdArgs,
		spec.Env,
		spec.WorkingDir,
		spec.Port,
		volumes,
		spec.Rootfs,
		restartPolicy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create local pod app: %w", err)
	}

	var eng engine.Engine
	if engineType == app.EngineRootfs {
		eng = engine.NewRootfsEngine()
	} else {
		eng = engine.NewNativeEngine()
	}

	cfg := &engine.RunConfig{
		App:      localApp,
		Store:    m.appStore,
		Detached: true,
	}

	proc, err := eng.Run(context.Background(), cfg)
	if err != nil {
		localApp.Status = app.StatusFailed
		_ = m.appStore.SaveApp(localApp)
		return nil, fmt.Errorf("failed to start local pod app: %w", err)
	}

	localApp.PID = proc.Pid
	localApp.Status = app.StatusRunning
	localApp.StartedAt = time.Now()
	_ = m.appStore.SaveApp(localApp)

	return &RunPodResponse{
		PodID:      pod.ID,
		LocalAppID: localApp.ID,
		HostPort:   localApp.Port,
		PID:        localApp.PID,
		Status:     string(PodRunning),
	}, nil
}

// StopLocalPod stops a pod running on this local machine
func (m *Manager) StopLocalPod(podID string) error {
	m.mu.Lock()
	pod, exists := m.pods[podID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("pod %s not found", podID)
	}

	localAppID := pod.LocalAppID
	pod.Status = PodTerminated
	pod.FinishedAt = time.Now()
	_ = m.store.SavePods(m.getPodListUnsafe())
	m.mu.Unlock()

	if localAppID != "" {
		_ = m.appManager.StopApp(localAppID)
	}
	return nil
}

// GetPodLogs fetches stdout/stderr logs for a pod
func (m *Manager) GetPodLogs(podID string, lines int) (string, error) {
	m.mu.RLock()
	pod, exists := m.pods[podID]
	if !exists {
		m.mu.RUnlock()
		return "", fmt.Errorf("pod %s not found", podID)
	}

	isLocal := (m.localNode != nil && pod.NodeID == m.localNode.ID)
	nodeAddr := pod.NodeAddress
	localAppID := pod.LocalAppID
	token := ""
	if m.config != nil {
		token = m.config.JoinToken
	}
	m.mu.RUnlock()

	if isLocal {
		if localAppID == "" {
			return "No logs available for pending pod", nil
		}
		logPath := filepath.Join(m.appStore.AppDir(localAppID), "app.log")
		data, err := os.ReadFile(logPath)
		if err != nil {
			return fmt.Sprintf("No logs available yet: %v", err), nil
		}
		linesArr := strings.Split(string(data), "\n")
		if lines > 0 && len(linesArr) > lines {
			linesArr = linesArr[len(linesArr)-lines:]
		}
		return strings.Join(linesArr, "\n"), nil
	}

	// Remote node logs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.FetchPodLogs(ctx, nodeAddr, token, podID, lines)
}

// --- Background Loops (Control Plane Reconciler & Worker Heartbeat) ---

func (m *Manager) startBackgroundLoops() {
	if m.config == nil {
		return
	}

	if m.config.Role == RoleControlPlane {
		go m.controlPlaneReconciliationLoop()
	} else if m.config.Role == RoleWorker {
		go m.workerHeartbeatLoop()
	}
}

// controlPlaneReconciliationLoop monitors node health and maintains replica counts
func (m *Manager) controlPlaneReconciliationLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.reconcileCluster()
		}
	}
}

func (m *Manager) reconcileCluster() {
	m.mu.Lock()
	if m.config == nil || m.config.Role != RoleControlPlane {
		m.mu.Unlock()
		return
	}

	// 1. Check for dead/unreachable nodes
	nodesChanged := false
	for _, node := range m.nodes {
		if node.IsLocal {
			continue
		}
		if !node.IsAlive(30 * time.Second) {
			if node.Status == NodeStatusReady {
				node.Status = NodeStatusNotReady
				nodesChanged = true
			}
		} else {
			if node.Status == NodeStatusNotReady {
				node.Status = NodeStatusReady
				nodesChanged = true
			}
		}
	}
	if nodesChanged {
		_ = m.store.SaveNodes(m.getNodeListUnsafe())
	}

	// 2. Identify evictable pods on dead nodes
	evictable := m.scheduler.FindEvictablePods(m.getNodeListUnsafe(), m.getPodListUnsafe())
	for _, p := range evictable {
		p.Status = PodFailed
		p.ErrorMessage = "Node heartbeat lost"
	}
	if len(evictable) > 0 {
		_ = m.store.SavePods(m.getPodListUnsafe())
	}

	// 3. Reconcile deployments to meet target replica counts
	for _, dep := range m.deployments {
		running := 0
		for _, p := range m.pods {
			if p.Deployment == dep.Name && (p.Status == PodRunning || p.Status == PodPending) {
				running++
			}
		}

		if running < dep.Replicas {
			depCopy := *dep
			m.mu.Unlock()
			_ = m.launchPodReplica(&depCopy)
			m.mu.Lock()
		}
	}
	m.mu.Unlock()
}

// workerHeartbeatLoop reports local worker health to the control plane
func (m *Manager) workerHeartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.sendWorkerHeartbeat()
		}
	}
}

func (m *Manager) sendWorkerHeartbeat() {
	m.mu.RLock()
	if m.config == nil || m.localNode == nil || m.config.Role != RoleWorker {
		m.mu.RUnlock()
		return
	}

	leaderAddr := m.config.ControlPlaneAddress
	token := m.config.JoinToken
	nodeID := m.localNode.ID

	// Count local allocated pods and map statuses
	podStatuses := make(map[string]string)
	allocated := 0
	for _, p := range m.pods {
		if p.NodeID == nodeID {
			allocated++
			podStatuses[p.ID] = string(p.Status)
		}
	}
	m.mu.RUnlock()

	req := &HeartbeatRequest{
		NodeID:        nodeID,
		Token:         token,
		AllocatedPods: allocated,
		PodStatuses:   podStatuses,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = m.client.SendHeartbeat(ctx, leaderAddr, req)
}

// Stop halts all cluster manager background activities
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
	m.proxyMgr.StopAll()
}

// Helper to parse int flag safely
func ParseInt(s string, defaultVal int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
