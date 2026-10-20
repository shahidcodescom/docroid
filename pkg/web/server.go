package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/bundle"
	"docroid/pkg/cluster"
	"docroid/pkg/compose"
	"docroid/pkg/doctor"
	"docroid/pkg/engine"
	"docroid/pkg/metrics"
	"docroid/pkg/netproxy"
	"docroid/pkg/network"
	"docroid/pkg/rolling"
	"docroid/pkg/rootfs"
	"docroid/pkg/secrets"
	"docroid/pkg/staticsite"
	"docroid/pkg/supervisor"
	"docroid/pkg/tunnel"
	"docroid/pkg/utils"
)

type Server struct {
	manager       *app.AppManager
	rootfsMgr     *rootfs.Manager
	siteStore     *staticsite.Store
	clusterMgr    *cluster.Manager
	netMgr        *network.NetworkManager
	tunnelMgr     *tunnel.Manager
	dashboardPort int
	mux           *http.ServeMux
}

func NewServer(manager *app.AppManager, rootfsMgr *rootfs.Manager) *Server {
	siteStore, _ := staticsite.NewStore(manager.Store().BaseDir())
	clusterMgr, _ := cluster.NewManager(manager.Store(), manager)
	netMgr, _ := network.NewNetworkManager(manager.Store().BaseDir())
	tunnelStore, _ := tunnel.NewStore(manager.Store().BaseDir())
	tunnelMgr := tunnel.NewManager(tunnelStore)
	s := &Server{
		manager:       manager,
		rootfsMgr:     rootfsMgr,
		siteStore:     siteStore,
		clusterMgr:    clusterMgr,
		netMgr:        netMgr,
		tunnelMgr:     tunnelMgr,
		dashboardPort: 9090,
		mux:           http.NewServeMux(),
	}

	app.RegisterStopHook(func(appID string) {
		supervisor.GetSupervisor().Cancel(appID)
		netproxy.GetProxyManager().StopProxies(appID)
	})

	s.registerRoutes()
	return s
}

func (s *Server) ClusterManager() *cluster.Manager {
	return s.clusterMgr
}

func (s *Server) SetClusterManager(cm *cluster.Manager) {
	s.clusterMgr = cm
}

func (s *Server) TunnelManager() *tunnel.Manager {
	return s.tunnelMgr
}

func (s *Server) SetDashboardPort(port int) {
	s.dashboardPort = port
}

func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("/api/apps", s.handleApps)
	s.mux.HandleFunc("/api/apps/", s.handleAppDetail)
	s.mux.HandleFunc("/api/images", s.handleImages)
	s.mux.HandleFunc("/api/images/pull", s.handleImagePull)
	s.mux.HandleFunc("/api/images/push", s.handleImagePush)
	s.mux.HandleFunc("/api/images/build", s.handleImageBuild)
	s.mux.HandleFunc("/api/images/", s.handleImageDetail)
	s.mux.HandleFunc("/api/bundles/import", s.handleImportBundle)
	s.mux.HandleFunc("/api/secrets", s.handleSecrets)
	s.mux.HandleFunc("/api/secrets/", s.handleSecretDetail)
	s.mux.HandleFunc("/api/system", s.handleSystem)
	s.mux.HandleFunc("/api/rootfs", s.handleRootfs)
	s.mux.HandleFunc("/api/compose/parse", s.handleComposeParse)
	s.mux.HandleFunc("/api/compose/generate", s.handleComposeGenerate)
	s.mux.HandleFunc("/api/compose/up", s.handleComposeUp)
	s.mux.HandleFunc("/api/sites", s.handleSites)
	s.mux.HandleFunc("/api/sites/", s.handleSiteDetail)

	// Virtual Networks API routes
	s.mux.HandleFunc("/api/networks", s.handleNetworks)
	s.mux.HandleFunc("/api/networks/", s.handleNetworkDetail)

	// Cloudflare Tunnel API routes
	s.mux.HandleFunc("/api/tunnels/status", s.handleTunnelStatus)
	s.mux.HandleFunc("/api/tunnels/targets", s.handleTunnelTargets)
	s.mux.HandleFunc("/api/tunnels", s.handleTunnels)
	s.mux.HandleFunc("/api/tunnels/", s.handleTunnelDetail)

	// Cluster & Pods API routes
	s.mux.HandleFunc("/api/cluster/status", s.handleClusterStatus)
	s.mux.HandleFunc("/api/cluster/init", s.handleClusterInit)
	s.mux.HandleFunc("/api/cluster/join", s.handleClusterJoin)
	s.mux.HandleFunc("/api/cluster/leave", s.handleClusterLeave)
	s.mux.HandleFunc("/api/cluster/heartbeat", s.handleClusterHeartbeat)
	s.mux.HandleFunc("/api/cluster/nodes", s.handleClusterNodes)
	s.mux.HandleFunc("/api/cluster/nodes/", s.handleClusterNodeDetail)
	s.mux.HandleFunc("/api/cluster/deployments", s.handleClusterDeployments)
	s.mux.HandleFunc("/api/cluster/deployments/", s.handleClusterDeploymentDetail)
	s.mux.HandleFunc("/api/cluster/scale", s.handleClusterScale)
	s.mux.HandleFunc("/api/cluster/pods", s.handleClusterPods)
	s.mux.HandleFunc("/api/cluster/pods/", s.handleClusterPodDetail)
	// Inter-device worker execution routes
	s.mux.HandleFunc("/api/cluster/node/run-pod", s.handleNodeRunPod)
	s.mux.HandleFunc("/api/cluster/node/stop-pod", s.handleNodeStopPod)
	s.mux.HandleFunc("/api/cluster/node/pod-logs", s.handleNodePodLogs)

	// Embedded Static File Server
	fileServer := http.FileServer(GetFileSystem())
	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Disable caching for dynamic dev updates
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		fileServer.ServeHTTP(w, r)
	}))
}

// AppWithStatsJSON formats the App and Stats for JSON output
type AppWithStatsJSON struct {
	App   *app.App              `json:"app"`
	Stats *metrics.ProcessStats `json:"stats"`
}

func (s *Server) handleApps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		items, err := s.manager.ListAppsWithStats()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}

		resp := make([]AppWithStatsJSON, 0, len(items))
		for _, item := range items {
			resp = append(resp, AppWithStatsJSON{
				App:   item.App,
				Stats: item.Stats,
			})
		}
		_ = json.NewEncoder(w).Encode(resp)

	case http.MethodPost:
		var req struct {
			Name          string            `json:"name"`
			Engine        string            `json:"engine"`
			Command       string            `json:"command"`
			Args          []string          `json:"args"`
			Port          int               `json:"port"`
			PortMapping   string            `json:"port_mapping"`
			WorkDir       string            `json:"work_dir"`
			Env           map[string]string `json:"env"`
			Rootfs        string            `json:"rootfs"`
			Volumes       []app.VolumeMount `json:"volumes"`
			RestartPolicy string            `json:"restart_policy"`
			Detached      bool              `json:"detached"`
			Network       string            `json:"network"`
			NetworkAliases []string         `json:"network_aliases"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON request"}`, http.StatusBadRequest)
			return
		}

		if req.Command == "" && req.Engine != string(app.EngineRootfs) {
			http.Error(w, `{"error":"command is required"}`, http.StatusBadRequest)
			return
		}

		engType := app.EngineNative
		if req.Engine == string(app.EngineRootfs) {
			engType = app.EngineRootfs
		}

		var pMap *netproxy.PortMapping
		targetPort := req.Port
		if req.PortMapping != "" {
			if pm, err := netproxy.ParsePortMapping(req.PortMapping); err == nil {
				pMap = pm
				if pMap.HostPort == pMap.ContainerPort {
					if freePort, err := utils.GetFreePort(); err == nil {
						oldPortStr := strconv.Itoa(pMap.HostPort)
						pMap.ContainerPort = freePort
						newPortStr := strconv.Itoa(freePort)
						for i, a := range req.Args {
							if a == oldPortStr {
								req.Args[i] = newPortStr
							}
						}
					}
				}
				targetPort = pMap.ContainerPort
			}
		}

		restartPolicy := "no"
		if req.RestartPolicy != "" {
			restartPolicy = req.RestartPolicy
		}

		createdApp, err := s.manager.CreateApp(
			req.Name,
			engType,
			req.Command,
			req.Args,
			req.Env,
			req.WorkDir,
			targetPort,
			req.Volumes,
			req.Rootfs,
			restartPolicy,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		if pMap != nil {
			createdApp.PortMapping = pMap.String()
			_ = s.manager.Store().SaveApp(createdApp)
		}

		if req.Network != "" {
			createdApp.Network = req.Network
		}
		if len(req.NetworkAliases) > 0 {
			createdApp.NetworkAliases = req.NetworkAliases
		}
		_ = s.manager.Store().SaveApp(createdApp)

		// Launch application
		var eng engine.Engine
		if engType == app.EngineRootfs {
			eng = engine.NewRootfsEngine()
		} else {
			eng = engine.NewNativeEngine()
		}

		cfg := &engine.RunConfig{
			App:      createdApp,
			Store:    s.manager.Store(),
			Detached: true,
		}

		proc, err := eng.Run(context.Background(), cfg)
		if err != nil {
			createdApp.Status = app.StatusFailed
			_ = s.manager.Store().SaveApp(createdApp)
			http.Error(w, fmt.Sprintf(`{"error":"failed to start: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		createdApp.PID = proc.Pid
		createdApp.Status = app.StatusRunning
		createdApp.StartedAt = time.Now()
		_ = s.manager.Store().SaveApp(createdApp)

		if pMap != nil && pMap.HostPort != pMap.ContainerPort {
			_ = netproxy.GetProxyManager().StartProxyWithControl(createdApp.ID, pMap, s.manager.Store().AppDir(createdApp.ID))
		}
		if createdApp.RestartPolicy != "" && createdApp.RestartPolicy != "no" {
			supervisor.GetSupervisor().Watch(createdApp, s.manager.Store(), eng)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createdApp)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAppDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/apps/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, `{"error":"missing app id"}`, http.StatusBadRequest)
		return
	}

	targetApp, err := s.manager.Store().FindApp(id)
	if err != nil {
		http.Error(w, `{"error":"app not found"}`, http.StatusNotFound)
		return
	}

	// Action dispatch: /api/apps/{id}/[stop|start|restart|logs]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "stop":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := s.manager.StopApp(targetApp.ID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopped", "id": targetApp.ID})

	case "start":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		if targetApp.Status == app.StatusRunning {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "already running", "id": targetApp.ID})
			return
		}

		var eng engine.Engine
		if targetApp.Engine == app.EngineRootfs {
			eng = engine.NewRootfsEngine()
		} else {
			eng = engine.NewNativeEngine()
		}

		cfg := &engine.RunConfig{
			App:      targetApp,
			Store:    s.manager.Store(),
			Detached: true,
		}

		proc, err := eng.Run(context.Background(), cfg)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		targetApp.PID = proc.Pid
		targetApp.Status = app.StatusRunning
		targetApp.StartedAt = time.Now()
		_ = s.manager.Store().SaveApp(targetApp)

		if targetApp.PortMapping != "" {
			if pm, err := netproxy.ParsePortMapping(targetApp.PortMapping); err == nil && pm.HostPort != pm.ContainerPort {
				_ = netproxy.GetProxyManager().StartProxyWithControl(targetApp.ID, pm, s.manager.Store().AppDir(targetApp.ID))
			}
		}
		if targetApp.RestartPolicy != "" && targetApp.RestartPolicy != "no" {
			supervisor.GetSupervisor().Watch(targetApp, s.manager.Store(), eng)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "started", "pid": proc.Pid})

	case "restart":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		_ = s.manager.StopApp(targetApp.ID)
		time.Sleep(150 * time.Millisecond)

		var eng engine.Engine
		if targetApp.Engine == app.EngineRootfs {
			eng = engine.NewRootfsEngine()
		} else {
			eng = engine.NewNativeEngine()
		}

		cfg := &engine.RunConfig{
			App:      targetApp,
			Store:    s.manager.Store(),
			Detached: true,
		}

		proc, err := eng.Run(context.Background(), cfg)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		targetApp.PID = proc.Pid
		targetApp.Status = app.StatusRunning
		targetApp.StartedAt = time.Now()
		_ = s.manager.Store().SaveApp(targetApp)

		if targetApp.PortMapping != "" {
			if pm, err := netproxy.ParsePortMapping(targetApp.PortMapping); err == nil && pm.HostPort != pm.ContainerPort {
				_ = netproxy.GetProxyManager().StartProxyWithControl(targetApp.ID, pm, s.manager.Store().AppDir(targetApp.ID))
			}
		}
		if targetApp.RestartPolicy != "" && targetApp.RestartPolicy != "no" {
			supervisor.GetSupervisor().Watch(targetApp, s.manager.Store(), eng)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "restarted", "pid": proc.Pid})

	case "rolling-restart":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		res, err := rolling.ExecuteZeroDowntimeRestart(r.Context(), s.manager.Store(), s.manager, targetApp.ID, rolling.RolloutOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(res)

	case "update":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Command      string            `json:"command"`
			Args         []string          `json:"args"`
			Env          map[string]string `json:"env"`
			Image        string            `json:"image"`
			DrainTimeout float64           `json:"drain_timeout"`
			ProbeTimeout float64           `json:"probe_timeout"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		opts := rolling.RolloutOptions{
			NewCommand:   req.Command,
			NewArgs:      req.Args,
			NewEnv:       req.Env,
			NewImage:     req.Image,
			DrainTimeout: time.Duration(req.DrainTimeout * float64(time.Second)),
			ProbeTimeout: time.Duration(req.ProbeTimeout * float64(time.Second)),
		}
		res, err := rolling.ExecuteZeroDowntimeRestart(r.Context(), s.manager.Store(), s.manager, targetApp.ID, opts)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(res)

	case "logs":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		lines := 50
		if lStr := r.URL.Query().Get("lines"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
				lines = l
			}
		}
		_ = s.manager.StreamLogs(targetApp.ID, false, lines, w)

	case "export":
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.dpack"`, targetApp.Name))
		w.Header().Set("Content-Type", "application/gzip")
		tempFile, err := os.CreateTemp("", "dpack_export_*")
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		tempName := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempName)

		if err := bundle.ExportApp(s.manager.Store(), targetApp.ID, tempName, io.Discard); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, tempName)

	case "exec":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			Env     []string `json:"env"`
			WorkDir string   `json:"workdir"`
			User    string   `json:"user"`
			Timeout int      `json:"timeout"`
			Force   bool     `json:"force"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}

		if !req.Force && targetApp.Status != app.StatusRunning {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  fmt.Sprintf("container '%s' is not running (status: %s)", targetApp.ID, targetApp.Status),
				"status": targetApp.Status,
			})
			return
		}

		cmdName := req.Command
		cmdArgs := req.Args
		if cmdName == "" {
			cmdName = "/bin/sh"
			if len(cmdArgs) == 0 {
				cmdArgs = []string{"-c", "id"}
			}
		}

		workDir := req.WorkDir
		if workDir == "" {
			if targetApp.WorkDir != "" {
				workDir = targetApp.WorkDir
			} else if targetApp.Engine == app.EngineRootfs {
				workDir = "/workspace"
			} else {
				workDir = s.manager.Store().WorkspaceDir(targetApp.ID)
			}
		}

		execEnvs := make(map[string]string)
		for k, v := range targetApp.Env {
			execEnvs[k] = v
		}
		for _, e := range req.Env {
			if strings.Contains(e, "=") {
				kv := strings.SplitN(e, "=", 2)
				execEnvs[kv[0]] = kv[1]
			}
		}

		tempApp := &app.App{
			ID:             targetApp.ID,
			Name:           targetApp.Name + "-exec",
			Engine:         targetApp.Engine,
			Command:        cmdName,
			Args:           cmdArgs,
			Env:            execEnvs,
			WorkDir:        workDir,
			Rootfs:         targetApp.Rootfs,
			Volumes:        targetApp.Volumes,
			Network:        targetApp.Network,
			NetworkAliases: targetApp.NetworkAliases,
			Port:           targetApp.Port,
			PortMapping:    targetApp.PortMapping,
		}

		var eng engine.Engine
		if targetApp.Engine == app.EngineRootfs {
			eng = engine.NewRootfsEngine()
		} else {
			eng = engine.NewNativeEngine()
		}

		var stdoutBuf, stderrBuf bytes.Buffer
		cfg := &engine.RunConfig{
			App:         tempApp,
			Store:       s.manager.Store(),
			Stdin:       nil,
			Stdout:      &stdoutBuf,
			Stderr:      &stderrBuf,
			Interactive: false,
			Detached:    false,
			User:        req.User,
		}

		timeout := 30 * time.Second
		if req.Timeout > 0 {
			timeout = time.Duration(req.Timeout) * time.Second
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		cmd, err := eng.Prepare(cfg)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"prepare failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		if err := cmd.Start(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"start failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		var exitCode int
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			w.WriteHeader(http.StatusGatewayTimeout)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "command timed out",
				"exit_code": -1,
				"stdout":    stdoutBuf.String(),
				"stderr":    stderrBuf.String(),
			})
			return
		case err := <-done:
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					exitCode = 1
				}
			} else {
				exitCode = 0
			}
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"exit_code": exitCode,
			"stdout":    stdoutBuf.String(),
			"stderr":    stderrBuf.String(),
		})

	case "":
		// Direct operations on /api/apps/{id}
		if r.Method == http.MethodDelete {
			if err := s.manager.RemoveApp(targetApp.ID, true); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"deleted": targetApp.ID})
		} else if r.Method == http.MethodGet {
			st, _ := metrics.GetProcessStats(targetApp.PID)
			_ = json.NewEncoder(w).Encode(AppWithStatsJSON{
				App:   targetApp,
				Stats: st,
			})
		} else {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}

	default:
		http.Error(w, `{"error":"unknown action"}`, http.StatusNotFound)
	}
}

func (s *Server) handleImages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	images, err := bundle.ListImages(s.manager.Store())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(images)
}

func (s *Server) handleImageDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tag := strings.TrimPrefix(r.URL.Path, "/api/images/")
	if tag == "" {
		http.Error(w, `{"error":"missing image tag"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		img, err := bundle.GetImage(s.manager.Store(), tag)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(img)
	case http.MethodDelete:
		err := bundle.RemoveImage(s.manager.Store(), tag)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "message": fmt.Sprintf("Image %s removed", tag)})
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleImagePull(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Image) == "" {
		http.Error(w, `{"error":"missing 'image' field"}`, http.StatusBadRequest)
		return
	}

	var out bytes.Buffer
	res, err := bundle.PullImage(s.manager.Store(), strings.TrimSpace(req.Image), &out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": fmt.Sprintf("pull failed: %s", err.Error()),
			"logs":  out.String(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"result":  res,
		"logs":    out.String(),
	})
}

func (s *Server) handleImagePush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Image) == "" {
		http.Error(w, `{"error":"missing 'image' field"}`, http.StatusBadRequest)
		return
	}

	var out bytes.Buffer
	digest, err := bundle.PushImage(s.manager.Store(), strings.TrimSpace(req.Image), &out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": fmt.Sprintf("push failed: %s", err.Error()),
			"logs":  out.String(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"digest":  digest,
		"logs":    out.String(),
	})
}

func (s *Server) handleImageBuild(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tag         string `json:"tag"`
		Dockerfile  string `json:"dockerfile"`
		Docroidfile string `json:"docroidfile"`
		ContextDir  string `json:"context_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	tag := strings.TrimSpace(req.Tag)
	if tag == "" {
		tag = "custom-build:latest"
	}

	buildContext := strings.TrimSpace(req.ContextDir)
	dockerfileContent := req.Dockerfile
	if dockerfileContent == "" {
		dockerfileContent = req.Docroidfile
	}

	if buildContext == "" || dockerfileContent != "" {
		tempDir, err := os.MkdirTemp("", "docroid-build-*")
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to create temp build context: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(tempDir)

		if buildContext != "" {
			_ = copyServerDir(buildContext, tempDir)
		}

		if dockerfileContent != "" {
			dfilePath := filepath.Join(tempDir, "Dockerfile")
			if err := os.WriteFile(dfilePath, []byte(dockerfileContent), 0644); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"failed to write build file: %s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
		}
		buildContext = tempDir
	}

	var out bytes.Buffer
	builder := bundle.NewBuilder(s.manager.Store())
	manifest, err := builder.BuildImage("", buildContext, tag, &out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": fmt.Sprintf("build failed: %s", err.Error()),
			"logs":  out.String(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"manifest": manifest,
		"logs":     out.String(),
	})
}

func copyServerDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			_ = os.MkdirAll(d, 0755)
			_ = copyServerDir(s, d)
		} else {
			data, err := os.ReadFile(s)
			if err == nil {
				_ = os.WriteFile(d, data, 0644)
			}
		}
	}
	return nil
}

func (s *Server) handleImportBundle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with max 100MB
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to parse form: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("bundle")
	if err != nil {
		http.Error(w, `{"error":"missing 'bundle' file field"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	manifest, err := bundle.ImportPackageReader(s.manager.Store(), file, io.Discard)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to import bundle: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(manifest)
}

func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	checks := []doctor.CheckResult{}

	// Host
	checks = append(checks, doctor.CheckResult{
		Name:    "Host Platform",
		Status:  "OK",
		Message: fmt.Sprintf("%s/%s (Mobile & Linux Optimized)", runtime.GOOS, runtime.GOARCH),
	})

	// Kernel
	kernel := "unknown"
	if data, err := os.ReadFile("/proc/version"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 3 {
			kernel = parts[2]
		}
	} else if out, err := exec.Command("uname", "-r").Output(); err == nil {
		kernel = strings.TrimSpace(string(out))
	}
	checks = append(checks, doctor.CheckResult{
		Name:    "Kernel Version",
		Status:  "OK",
		Message: kernel,
	})

	// Termux check
	termuxPrefix := os.Getenv("PREFIX")
	if strings.Contains(termuxPrefix, "com.termux") || termuxPrefix != "" {
		checks = append(checks, doctor.CheckResult{
			Name:    "Termux Environment",
			Status:  "OK",
			Message: fmt.Sprintf("Active (%s)", termuxPrefix),
		})
	} else {
		checks = append(checks, doctor.CheckResult{
			Name:    "Termux Environment",
			Status:  "WARN",
			Message: "Standard environment detected",
		})
	}

	// PRoot check
	prootPath, _ := exec.LookPath("proot")
	if prootPath == "" {
		if _, err := os.Stat("/data/data/com.termux/files/usr/bin/proot"); err == nil {
			prootPath = "/data/data/com.termux/files/usr/bin/proot"
		}
	}
	if prootPath != "" {
		checks = append(checks, doctor.CheckResult{
			Name:    "PRoot Sandbox Engine",
			Status:  "OK",
			Message: fmt.Sprintf("Ready (%s)", prootPath),
		})
	} else {
		checks = append(checks, doctor.CheckResult{
			Name:    "PRoot Sandbox Engine",
			Status:  "WARN",
			Message: "Run 'pkg install proot' to enable Linux rootfs containers",
		})
	}

	checks = append(checks, doctor.CheckResult{
		Name:    "Storage Base Directory",
		Status:  "OK",
		Message: s.manager.Store().BaseDir(),
	})

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"checks": checks,
		"os":     runtime.GOOS,
		"arch":   runtime.GOARCH,
	})
}

func (s *Server) handleRootfs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.rootfsMgr == nil {
		s.rootfsMgr = rootfs.NewManager(s.manager.Store())
	}

	list, err := s.rootfsMgr.List()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"installed": list,
	})
}

func (s *Server) handleSecrets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	secStore, err := secrets.NewStore(s.manager.Store().BaseDir())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := secStore.List()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		type SecretItem struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		res := make([]SecretItem, 0, len(items))
		for k, v := range items {
			res = append(res, SecretItem{Key: k, Value: v})
		}
		_ = json.NewEncoder(w).Encode(res)

	case http.MethodPost:
		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Key) == "" {
			http.Error(w, `{"error":"key cannot be empty"}`, http.StatusBadRequest)
			return
		}
		if err := secStore.Set(req.Key, req.Value); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "saved", "key": req.Key})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSecretDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	secStore, err := secrets.NewStore(s.manager.Store().BaseDir())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/api/secrets/")
	if key == "" {
		http.Error(w, `{"error":"missing secret key"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		val, err := secStore.Get(key)
		if err != nil {
			http.Error(w, `{"error":"secret not found"}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})

	case http.MethodDelete:
		if err := secStore.Delete(key); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "key": key})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleComposeParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var yamlBytes []byte
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to parse multipart form: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("compose")
		if err != nil {
			http.Error(w, `{"error":"missing 'compose' file field"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()
		b, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to read uploaded file: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		yamlBytes = b
	} else {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to read request body: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		var jsonReq struct {
			YAML string `json:"yaml"`
		}
		if err := json.Unmarshal(b, &jsonReq); err == nil && jsonReq.YAML != "" {
			yamlBytes = []byte(jsonReq.YAML)
		} else {
			yamlBytes = b
		}
	}

	if len(strings.TrimSpace(string(yamlBytes))) == 0 {
		http.Error(w, `{"error":"empty YAML content"}`, http.StatusBadRequest)
		return
	}

	cfg, err := compose.ParseYAML(yamlBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to parse compose YAML: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	var services []compose.ResolvedService
	for key, spec := range cfg.Services {
		res, err := compose.NormalizeService(key, &spec)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"service '%s': %s"}`, key, err.Error()), http.StatusBadRequest)
			return
		}
		services = append(services, *res)
	}

	resp := map[string]interface{}{
		"version":  cfg.Version,
		"services": services,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleComposeGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name          string            `json:"name"`
		Engine        string            `json:"engine"`
		Rootfs        string            `json:"rootfs"`
		Command       string            `json:"command"`
		Args          []string          `json:"args"`
		PortMapping   string            `json:"port_mapping"`
		RestartPolicy string            `json:"restart_policy"`
		WorkDir       string            `json:"work_dir"`
		Env           map[string]string `json:"env"`
		Volumes       []string          `json:"volumes"`
		Secrets       []string          `json:"secrets"`
		Detached      bool              `json:"detached"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	yamlBytes, err := compose.GenerateComposeYAML(
		req.Name,
		req.Engine,
		req.Rootfs,
		req.Command,
		req.Args,
		req.PortMapping,
		req.RestartPolicy,
		req.WorkDir,
		req.Env,
		req.Volumes,
		req.Secrets,
		req.Detached,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to generate YAML: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("download") == "false" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"yaml": string(yamlBytes)})
		return
	}

	filename := "docroid-compose.yml"
	if req.Name != "" {
		filename = fmt.Sprintf("docroid-compose-%s.yml", req.Name)
	}

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(yamlBytes)
}

func (s *Server) handleComposeUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var yamlBytes []byte
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read request"}`, http.StatusBadRequest)
		return
	}
	var jsonReq struct {
		YAML string `json:"yaml"`
	}
	if err := json.Unmarshal(b, &jsonReq); err == nil && jsonReq.YAML != "" {
		yamlBytes = []byte(jsonReq.YAML)
	} else {
		yamlBytes = b
	}

	cfg, err := compose.ParseYAML(yamlBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to parse compose YAML: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	launched, err := compose.Up(r.Context(), s.manager.Store(), s.manager, cfg, true)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to launch services: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "launched",
		"count":    len(launched),
		"services": launched,
	})
}

// StartServer launches the HTTP dashboard server
func StartServer(manager *app.AppManager, rootfsMgr *rootfs.Manager, host string, port int) error {
	if host == "" {
		host = "0.0.0.0"
	}
	if port <= 0 {
		port = 9090
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	server := NewServer(manager, rootfsMgr)
	server.SetDashboardPort(port)

	fmt.Printf("\nDocroid Web Dashboard & Management Server\n")
	fmt.Printf("==========================================\n")
	fmt.Printf("• Local access:   http://localhost:%d\n", port)

	// Print LAN IP addresses for easy connection from other devices (e.g. laptop)
	if ips := getLocalIPs(); len(ips) > 0 {
		for _, ip := range ips {
			fmt.Printf("• Network access: http://%s:%d\n", ip, port)
		}
	}
	fmt.Printf("\nPress Ctrl+C to stop the dashboard server.\n\n")

	return http.ListenAndServe(addr, server.mux)
}

func getLocalIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	return ips
}

// Static Site API Handlers

func (s *Server) handleSites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		sites, err := s.siteStore.ListSites()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// Refresh liveness for each site
		for _, site := range sites {
			staticsite.IsSiteRunning(site, s.siteStore)
		}

		_ = json.NewEncoder(w).Encode(sites)

	case http.MethodPost:
		// Supports multipart/form-data for file/zip uploads, or JSON
		ct := r.Header.Get("Content-Type")

		var name string
		var port int
		var domain string
		var spa bool
		var dirListing bool
		var autoStart = true
		var rawHTML string

		if strings.Contains(ct, "multipart/form-data") {
			// 100MB limit for zip bundles and assets
			if err := r.ParseMultipartForm(100 << 20); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"failed to parse multipart form: %s"}`, err.Error()), http.StatusBadRequest)
				return
			}

			name = strings.TrimSpace(r.FormValue("name"))
			if pStr := strings.TrimSpace(r.FormValue("port")); pStr != "" {
				port, _ = strconv.Atoi(pStr)
			}
			domain = strings.TrimSpace(r.FormValue("domain"))
			spa = r.FormValue("spa") == "true" || r.FormValue("spa") == "1"
			dirListing = r.FormValue("dir_listing") == "true" || r.FormValue("dir_listing") == "1"
			if asStr := r.FormValue("auto_start"); asStr != "" {
				autoStart = asStr == "true" || asStr == "1"
			}
			rawHTML = r.FormValue("html")
		} else {
			var jsonReq struct {
				Name             string `json:"name"`
				Port             int    `json:"port"`
				Domain           string `json:"domain"`
				SPAMode          bool   `json:"spa"`
				DirectoryListing bool   `json:"dir_listing"`
				AutoStart        *bool  `json:"auto_start"`
				HTML             string `json:"html"`
			}
			if err := json.NewDecoder(r.Body).Decode(&jsonReq); err != nil {
				http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
				return
			}
			name = jsonReq.Name
			port = jsonReq.Port
			domain = jsonReq.Domain
			spa = jsonReq.SPAMode
			dirListing = jsonReq.DirectoryListing
			if jsonReq.AutoStart != nil {
				autoStart = *jsonReq.AutoStart
			}
			rawHTML = jsonReq.HTML
		}

		createdSite, err := s.siteStore.CreateSite(name, port, domain, spa, dirListing)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		// Handle file sources
		hasAssets := false

		// 1. Check for uploaded ZIP bundle
		if r.MultipartForm != nil && len(r.MultipartForm.File["bundle"]) > 0 {
			fileHeader := r.MultipartForm.File["bundle"][0]
			f, err := fileHeader.Open()
			if err == nil {
				defer f.Close()
				if seeker, ok := f.(io.ReaderAt); ok {
					if err := s.siteStore.ExtractZipBundle(createdSite.ID, seeker, fileHeader.Size); err == nil {
						hasAssets = true
					} else {
						http.Error(w, fmt.Sprintf(`{"error":"zip extraction failed: %s"}`, err.Error()), http.StatusBadRequest)
						return
					}
				}
			}
		}

		// 2. Check for uploaded files (multi-file or folder upload)
		if !hasAssets && r.MultipartForm != nil && len(r.MultipartForm.File["files"]) > 0 {
			fileMap := make(map[string][]byte)
			for _, fh := range r.MultipartForm.File["files"] {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err == nil {
					// Use original filename which may include relative path
					relPath := fh.Filename
					fileMap[relPath] = data
				}
			}
			if len(fileMap) > 0 {
				if err := s.siteStore.SaveUploadedFiles(createdSite.ID, fileMap); err == nil {
					hasAssets = true
				}
			}
		}

		// 3. Check for raw HTML
		if !hasAssets && strings.TrimSpace(rawHTML) != "" {
			if err := s.siteStore.SaveRawHTML(createdSite.ID, rawHTML); err == nil {
				hasAssets = true
			}
		}

		// 4. If still no assets, generate a beautiful starter HTML page
		if !hasAssets {
			starterHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s - Hosted on Docroid</title>
  <style>
    :root {
      --bg: #0f172a;
      --card: #1e293b;
      --text: #f8fafc;
      --accent: #38bdf8;
      --muted: #94a3b8;
    }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: var(--bg);
      color: var(--text);
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      margin: 0;
      padding: 1.5rem;
      box-sizing: border-box;
    }
    .card {
      background: var(--card);
      border: 1px solid #334155;
      padding: 2.5rem;
      border-radius: 16px;
      max-width: 580px;
      text-align: center;
      box-shadow: 0 20px 40px rgba(0,0,0,0.4);
    }
    .badge {
      display: inline-block;
      background: rgba(56, 189, 248, 0.15);
      color: var(--accent);
      padding: 4px 14px;
      border-radius: 9999px;
      font-size: 0.85rem;
      font-weight: 600;
      margin-bottom: 1rem;
    }
    h1 { margin: 0 0 0.5rem; font-size: 2rem; color: #fff; }
    p { color: var(--muted); line-height: 1.6; margin-bottom: 1.5rem; }
    .specs {
      background: #0b1120;
      padding: 1rem;
      border-radius: 8px;
      display: flex;
      justify-content: space-around;
      font-size: 0.9rem;
      margin-bottom: 1.5rem;
    }
    .spec-item span { display: block; color: var(--muted); font-size: 0.75rem; text-transform: uppercase; }
    .spec-item strong { color: var(--accent); font-family: monospace; font-size: 1.05rem; }
    .footer { font-size: 0.8rem; color: #64748b; }
  </style>
</head>
<body>
  <div class="card">
    <div class="badge">Docroid Static Web Server</div>
    <h1>%s</h1>
    <p>Your static website is live and served directly from Docroid's zero-daemon mobile container engine!</p>
    <div class="specs">
      <div class="spec-item"><span>Port</span><strong>%d</strong></div>
      <div class="spec-item"><span>Domain</span><strong>%s</strong></div>
      <div class="spec-item"><span>Engine</span><strong>Native Go</strong></div>
    </div>
    <div class="footer">Replace this file by uploading HTML, static assets, or a ZIP bundle in the Docroid Dashboard.</div>
  </div>
</body>
</html>`, html.EscapeString(createdSite.Name), html.EscapeString(createdSite.Name), createdSite.Port, html.EscapeString(createdSite.Domain))
			_ = s.siteStore.SaveRawHTML(createdSite.ID, starterHTML)
		}

		// Auto-start if requested
		if autoStart {
			if err := staticsite.StartInProcessServer(createdSite.Port, s.siteStore); err == nil {
				createdSite.Status = staticsite.StatusRunning
				createdSite.PID = os.Getpid()
				createdSite.StartedAt = time.Now()
				_ = s.siteStore.SaveSite(createdSite)
			}
		}

		// Reload fresh site data
		refreshedSite, _ := s.siteStore.LoadSite(createdSite.ID)
		if refreshedSite != nil {
			createdSite = refreshedSite
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createdSite)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSiteDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if id == "" {
		http.Error(w, `{"error":"missing site id"}`, http.StatusBadRequest)
		return
	}

	site, err := s.siteStore.FindSite(id)
	if err != nil {
		http.Error(w, `{"error":"site not found"}`, http.StatusNotFound)
		return
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "start":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := staticsite.StartInProcessServer(site.Port, s.siteStore); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to start server on port %d: %s"}`, site.Port, err.Error()), http.StatusInternalServerError)
			return
		}
		site.Status = staticsite.StatusRunning
		site.PID = os.Getpid()
		site.StartedAt = time.Now()
		_ = s.siteStore.SaveSite(site)
		_ = json.NewEncoder(w).Encode(site)

	case "stop":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		_ = staticsite.StopInProcessServer(site.Port)
		_ = staticsite.StopSiteProcess(site.ID, s.siteStore)
		site.Status = staticsite.StatusStopped
		site.PID = 0
		_ = s.siteStore.SaveSite(site)
		_ = json.NewEncoder(w).Encode(site)

	case "upload":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(100 << 20); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to parse form: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		// ZIP bundle upload
		if len(r.MultipartForm.File["bundle"]) > 0 {
			fileHeader := r.MultipartForm.File["bundle"][0]
			f, err := fileHeader.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			defer f.Close()
			if seeker, ok := f.(io.ReaderAt); ok {
				if err := s.siteStore.ExtractZipBundle(site.ID, seeker, fileHeader.Size); err != nil {
					http.Error(w, fmt.Sprintf(`{"error":"zip extraction failed: %s"}`, err.Error()), http.StatusBadRequest)
					return
				}
			}
		} else if len(r.MultipartForm.File["files"]) > 0 {
			fileMap := make(map[string][]byte)
			for _, fh := range r.MultipartForm.File["files"] {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err == nil {
					fileMap[fh.Filename] = data
				}
			}
			if err := s.siteStore.SaveUploadedFiles(site.ID, fileMap); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"failed to save files: %s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
		} else if rawHTML := r.FormValue("html"); rawHTML != "" {
			if err := s.siteStore.SaveRawHTML(site.ID, rawHTML); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"failed to save HTML: %s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
		}

		updated, _ := s.siteStore.LoadSite(site.ID)
		_ = json.NewEncoder(w).Encode(updated)

	case "export":
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, site.Name))
		w.Header().Set("Content-Type", "application/zip")
		if err := s.siteStore.ExportSiteZip(site.ID, w); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

	case "files":
		// File listing or file content preview
		subpath := r.URL.Query().Get("subpath")
		filePath := r.URL.Query().Get("path")
		rawContent := r.URL.Query().Get("content") == "true"

		if filePath != "" {
			data, mimeType, err := s.siteStore.ReadSiteFile(site.ID, filePath)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
				return
			}

			if rawContent {
				w.Header().Set("Content-Type", mimeType)
				_, _ = w.Write(data)
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"path":      filePath,
				"mime_type": mimeType,
				"size":      len(data),
				"content":   string(data),
			})
			return
		}

		files, err := s.siteStore.ListSiteFiles(site.ID, subpath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(files)

	case "":
		if r.Method == http.MethodGet {
			staticsite.IsSiteRunning(site, s.siteStore)
			_ = json.NewEncoder(w).Encode(site)
		} else if r.Method == http.MethodDelete {
			_ = staticsite.StopInProcessServer(site.Port)
			_ = staticsite.StopSiteProcess(site.ID, s.siteStore)
			if err := s.siteStore.DeleteSite(site.ID); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"deleted": site.ID})
		} else {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}

	default:
		http.Error(w, `{"error":"unknown action"}`, http.StatusNotFound)
	}
}

var _ = bytes.NewBuffer

// --- Cluster & Pods HTTP Handlers ---

func (s *Server) handleClusterStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if s.clusterMgr == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": false})
		return
	}

	status, err := s.clusterMgr.GetClusterStatus()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) handleClusterInit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	addr := ""
	if req.Port > 0 {
		addr = fmt.Sprintf(":%d", req.Port)
	}

	cfg, err := s.clusterMgr.InitCluster(req.Name, addr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"config":  cfg,
	})
}

func (s *Server) handleClusterJoin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Read body bytes to inspect structure
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to read body: %v"}`, err), http.StatusBadRequest)
		return
	}

	// 1. Check if this is a worker node joining this control plane
	var workerReq cluster.JoinRequest
	if err := json.Unmarshal(data, &workerReq); err == nil && workerReq.Node.Name != "" {
		joinResp, err := s.clusterMgr.RegisterWorker(&workerReq)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(joinResp)
		return
	}

	// 2. Check if this is a request to join this node to a remote cluster
	var uiJoinReq struct {
		LeaderAddr string `json:"leader_addr"`
		Token      string `json:"token"`
		NodeName   string `json:"node_name"`
	}
	if err := json.Unmarshal(data, &uiJoinReq); err == nil && uiJoinReq.LeaderAddr != "" {
		cfg, err := s.clusterMgr.JoinCluster(uiJoinReq.LeaderAddr, uiJoinReq.Token, uiJoinReq.NodeName, "")
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"config":  cfg,
		})
		return
	}

	http.Error(w, `{"error":"invalid join payload"}`, http.StatusBadRequest)
}

func (s *Server) handleClusterLeave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := s.clusterMgr.LeaveCluster(); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (s *Server) handleClusterHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req cluster.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid heartbeat payload"}`, http.StatusBadRequest)
		return
	}

	resp, err := s.clusterMgr.ProcessHeartbeat(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleClusterNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	nodes := s.clusterMgr.ListNodes()
	_ = json.NewEncoder(w).Encode(nodes)
}

func (s *Server) handleClusterNodeDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	nodeID := strings.TrimPrefix(r.URL.Path, "/api/cluster/nodes/")
	nodeID = strings.TrimSpace(nodeID)

	if r.Method == http.MethodDelete {
		if err := s.clusterMgr.RemoveNode(nodeID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"deleted": nodeID})
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleClusterDeployments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		deps := s.clusterMgr.ListDeployments()
		_ = json.NewEncoder(w).Encode(deps)

	case http.MethodPost:
		var dep cluster.Deployment
		if err := json.NewDecoder(r.Body).Decode(&dep); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid deployment payload: %v"}`, err), http.StatusBadRequest)
			return
		}

		if err := s.clusterMgr.CreateDeployment(&dep); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dep)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleClusterDeploymentDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.TrimPrefix(r.URL.Path, "/api/cluster/deployments/")
	name = strings.TrimSpace(name)

	switch r.Method {
	case http.MethodGet:
		dep, err := s.clusterMgr.GetDeployment(name)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(dep)

	case http.MethodDelete:
		if err := s.clusterMgr.DeleteDeployment(name); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"deleted": name})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleClusterScale(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req cluster.DeploymentScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid scale payload"}`, http.StatusBadRequest)
		return
	}

	if err := s.clusterMgr.ScaleDeployment(req.Deployment, req.Replicas); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"deployment": req.Deployment,
		"replicas":   req.Replicas,
	})
}

func (s *Server) handleClusterPods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	pods := s.clusterMgr.ListPods()
	_ = json.NewEncoder(w).Encode(pods)
}

func (s *Server) handleClusterPodDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/cluster/pods/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	podID := parts[0]

	if len(parts) > 1 && parts[1] == "logs" {
		lines := 100
		if lStr := r.URL.Query().Get("lines"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
				lines = l
			}
		}
		logs, err := s.clusterMgr.GetPodLogs(podID, lines)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching logs: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(logs))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		pod, err := s.clusterMgr.GetPod(podID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(pod)

	case http.MethodDelete:
		if err := s.clusterMgr.DeletePod(podID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"deleted": podID})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// Inter-device worker node execution endpoints

func (s *Server) handleNodeRunPod(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req cluster.RunPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid run-pod request: %v"}`, err), http.StatusBadRequest)
		return
	}

	resp, err := s.clusterMgr.ExecuteLocalPod(&req.Pod)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleNodeStopPod(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req cluster.StopPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid stop-pod request: %v"}`, err), http.StatusBadRequest)
		return
	}

	if err := s.clusterMgr.StopLocalPod(req.PodID); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"stopped": req.PodID})
}

func (s *Server) handleNodePodLogs(w http.ResponseWriter, r *http.Request) {
	podID := r.URL.Query().Get("pod_id")
	if podID == "" {
		http.Error(w, "pod_id is required", http.StatusBadRequest)
		return
	}

	lines := 100
	if lStr := r.URL.Query().Get("lines"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			lines = l
		}
	}

	logs, err := s.clusterMgr.GetPodLogs(podID, lines)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading logs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(logs))
}

// ----------------------------------------------------
// Virtual Networks API Handlers
// ----------------------------------------------------

func (s *Server) handleNetworks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.netMgr == nil {
		http.Error(w, `{"error":"network manager not available"}`, http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		nets, err := s.netMgr.ListNetworks()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		if nets == nil {
			nets = []*network.Network{}
		}
		_ = json.NewEncoder(w).Encode(nets)

	case http.MethodPost:
		var req struct {
			Name   string            `json:"name"`
			Driver string            `json:"driver"`
			Subnet string            `json:"subnet"`
			Labels map[string]string `json:"labels"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		driver := network.DriverBridge
		if req.Driver != "" {
			driver = network.DriverType(req.Driver)
		}
		net, err := s.netMgr.CreateNetwork(req.Name, driver, req.Subnet, req.Labels)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(net)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleNetworkDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.netMgr == nil {
		http.Error(w, `{"error":"network manager not available"}`, http.StatusInternalServerError)
		return
	}

	subPath := strings.TrimPrefix(r.URL.Path, "/api/networks/")
	parts := strings.Split(strings.Trim(subPath, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"network name or ID required"}`, http.StatusBadRequest)
		return
	}
	netID := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			net, err := s.netMgr.GetNetwork(netID)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(net)

		case http.MethodDelete:
			if err := s.netMgr.RemoveNetwork(netID); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"deleted": netID})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	action := parts[1]
	switch action {
	case "connect":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Container string   `json:"container"`
			Aliases   []string `json:"aliases"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		appInst, err := s.manager.Store().FindApp(req.Container)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"container '%s' not found"}`, req.Container), http.StatusNotFound)
			return
		}
		ep, err := s.netMgr.Connect(netID, appInst.ID, appInst.Name, appInst.Port, req.Aliases)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		appInst.Network = netID
		for _, a := range req.Aliases {
			if a != "" && !contains(appInst.NetworkAliases, a) {
				appInst.NetworkAliases = append(appInst.NetworkAliases, a)
			}
		}
		_ = s.manager.Store().SaveApp(appInst)
		_ = json.NewEncoder(w).Encode(ep)

	case "disconnect":
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Container string `json:"container"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		appInst, _ := s.manager.Store().FindApp(req.Container)
		targetID := req.Container
		if appInst != nil {
			targetID = appInst.ID
		}
		if err := s.netMgr.Disconnect(netID, targetID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		if appInst != nil && appInst.Network == netID {
			appInst.Network = "bridge"
			_ = s.manager.Store().SaveApp(appInst)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"disconnected": targetID})

	default:
		http.Error(w, `{"error":"unknown action"}`, http.StatusNotFound)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Cloudflare Tunnel API Handlers

func (s *Server) handleTunnelStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	info := tunnel.GetCloudflaredInfo()
	_ = json.NewEncoder(w).Encode(info)
}

func (s *Server) handleTunnelTargets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var appsList []struct {
		ID     string
		Name   string
		Port   int
		Status string
	}
	if apps, err := s.manager.ListAppsWithStats(); err == nil {
		for _, item := range apps {
			p := item.App.Port
			if p == 0 && item.App.PortMapping != "" {
				if pm, err := netproxy.ParsePortMapping(item.App.PortMapping); err == nil {
					p = pm.HostPort
				}
			}
			appsList = append(appsList, struct {
				ID     string
				Name   string
				Port   int
				Status string
			}{
				ID:     item.App.ID,
				Name:   item.App.Name,
				Port:   p,
				Status: string(item.App.Status),
			})
		}
	}

	var sitesList []struct {
		ID     string
		Name   string
		Port   int
		Status string
	}
	if sites, err := s.siteStore.ListSites(); err == nil {
		for _, site := range sites {
			st := "stopped"
			if staticsite.IsSiteRunning(site, s.siteStore) {
				st = "running"
			}
			sitesList = append(sitesList, struct {
				ID     string
				Name   string
				Port   int
				Status string
			}{
				ID:     site.ID,
				Name:   site.Name,
				Port:   site.Port,
				Status: st,
			})
		}
	}

	targets := s.tunnelMgr.DiscoverTargets(appsList, sitesList, s.dashboardPort)
	_ = json.NewEncoder(w).Encode(targets)
}

func (s *Server) handleTunnels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		tunnels, err := s.tunnelMgr.Store().ListTunnels()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(tunnels)

	case http.MethodPost:
		var req struct {
			Name        string            `json:"name"`
			AppID       string            `json:"app_id"`
			AppName     string            `json:"app_name"`
			Target      string            `json:"target"`
			TargetPort  int               `json:"target_port"`
			Mode        tunnel.TunnelMode `json:"mode"`
			Token       string            `json:"token"`
			Hostname    string            `json:"hostname"`
			AutoRestart bool              `json:"auto_restart"`
			AutoStart   *bool             `json:"auto_start"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		opts := tunnel.TunnelOptions{
			Name:        req.Name,
			AppID:       req.AppID,
			AppName:     req.AppName,
			Target:      req.Target,
			TargetPort:  req.TargetPort,
			Mode:        req.Mode,
			Token:       req.Token,
			Hostname:    req.Hostname,
			AutoRestart: req.AutoRestart,
		}

		tun, err := s.tunnelMgr.CreateTunnel(opts)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		autoStart := true
		if req.AutoStart != nil {
			autoStart = *req.AutoStart
		}

		if autoStart {
			if started, err := s.tunnelMgr.Start(tun.ID); err == nil {
				tun = started
			} else {
				tun, _ = s.tunnelMgr.Store().GetTunnel(tun.ID)
			}
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(tun)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTunnelDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/tunnels/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"tunnel id required"}`, http.StatusBadRequest)
		return
	}

	tunnelID := parts[0]

	if len(parts) >= 2 {
		action := parts[1]
		switch action {
		case "start":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			tun, err := s.tunnelMgr.Start(tunnelID)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(tun)
			return

		case "stop":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			tun, err := s.tunnelMgr.Stop(tunnelID)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(tun)
			return

		case "restart":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			tun, err := s.tunnelMgr.Restart(tunnelID)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(tun)
			return

		case "logs":
			if r.Method != http.MethodGet {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			lines := 100
			if lStr := r.URL.Query().Get("lines"); lStr != "" {
				if parsedL, err := strconv.Atoi(lStr); err == nil && parsedL > 0 {
					lines = parsedL
				}
			}
			logs, err := s.tunnelMgr.GetLogs(tunnelID, lines)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id":   tunnelID,
				"logs": logs,
			})
			return

		default:
			http.Error(w, `{"error":"unknown tunnel action"}`, http.StatusNotFound)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		tun, err := s.tunnelMgr.Store().GetTunnel(tunnelID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(tun)

	case http.MethodDelete:
		if err := s.tunnelMgr.Delete(tunnelID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}


