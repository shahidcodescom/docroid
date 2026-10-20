package rolling

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/engine"
	"docroid/pkg/netproxy"
	"docroid/pkg/utils"
)

// RolloutOptions specifies settings for a zero-downtime rolling update or restart
type RolloutOptions struct {
	NewCommand   string            `json:"new_command,omitempty"`
	NewArgs      []string          `json:"new_args,omitempty"`
	NewEnv       map[string]string `json:"new_env,omitempty"`
	NewImage     string            `json:"new_image,omitempty"`
	ProbeTimeout time.Duration     `json:"probe_timeout,omitempty"` // default: 10s
	DrainTimeout time.Duration     `json:"drain_timeout,omitempty"` // default: 1.5s
	ReadinessURL string            `json:"readiness_url,omitempty"` // optional HTTP path e.g. "/healthz"
}

// RolloutResult summarizes the completed zero-downtime switchover
type RolloutResult struct {
	Status        string        `json:"status"`
	AppID         string        `json:"app_id"`
	OldPID        int           `json:"old_pid"`
	NewPID        int           `json:"new_pid"`
	TargetPort    int           `json:"target_port,omitempty"`
	DrainDuration time.Duration `json:"drain_duration"`
	TotalDuration time.Duration `json:"total_duration"`
	RolledBack    bool          `json:"rolled_back"`
	Message       string        `json:"message"`
}

// ExecuteZeroDowntimeRestart conducts an atomic, zero-downtime cutover to a new candidate process
func ExecuteZeroDowntimeRestart(
	ctx context.Context,
	store *app.Store,
	manager *app.AppManager,
	appID string,
	opts RolloutOptions,
) (*RolloutResult, error) {
	startTime := time.Now()

	// 1. Validate target application
	currentApp, err := store.FindApp(appID)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	if currentApp.Status != app.StatusRunning || currentApp.PID <= 0 {
		return nil, fmt.Errorf("app '%s' is not running (zero-downtime update requires an active instance)", currentApp.Name)
	}

	oldPID := currentApp.PID

	// Defaults
	probeTimeout := opts.ProbeTimeout
	if probeTimeout <= 0 {
		probeTimeout = 10 * time.Second
	}
	drainTimeout := opts.DrainTimeout
	if drainTimeout <= 0 {
		drainTimeout = 1500 * time.Millisecond
	}

	// 2. Determine engine
	var eng engine.Engine
	if currentApp.Engine == app.EngineRootfs {
		eng = engine.NewRootfsEngine()
	} else {
		eng = engine.NewNativeEngine()
	}

	// 3. Allocate candidate internal port if application is network-facing
	candidatePort := 0
	if currentApp.Port > 0 || currentApp.PortMapping != "" {
		freePort, err := utils.GetFreePort()
		if err == nil {
			candidatePort = freePort
		} else {
			candidatePort = currentApp.Port + 1000
		}
	}

	// 4. Prepare candidate app configuration
	candidateApp := *currentApp
	candidateApp.ID = currentApp.ID // share identity
	if opts.NewCommand != "" {
		candidateApp.Command = opts.NewCommand
	}
	if opts.NewArgs != nil {
		candidateApp.Args = opts.NewArgs
	} else {
		clonedArgs := make([]string, len(currentApp.Args))
		copy(clonedArgs, currentApp.Args)
		candidateApp.Args = clonedArgs
	}

	// Copy environment
	candidateEnv := make(map[string]string)
	for k, v := range currentApp.Env {
		candidateEnv[k] = v
	}
	for k, v := range opts.NewEnv {
		candidateEnv[k] = v
	}
	candidateApp.Env = candidateEnv

	// If candidate has an internal port, update arguments and env
	if candidatePort > 0 {
		candidateApp.Port = candidatePort
		if currentApp.Port > 0 {
			oldPortStr := fmt.Sprintf("%d", currentApp.Port)
			newPortStr := fmt.Sprintf("%d", candidatePort)
			for i, a := range candidateApp.Args {
				if a == oldPortStr {
					candidateApp.Args[i] = newPortStr
				}
			}
		}
		candidateApp.Env["PORT"] = fmt.Sprintf("%d", candidatePort)
		candidateApp.Env["DOCROID_PORT"] = fmt.Sprintf("%d", candidatePort)
	}

	// 5. Spawn Candidate Process (Instance B) in background
	runCfg := &engine.RunConfig{
		App:      &candidateApp,
		Store:    store,
		Detached: true,
	}

	candidateProc, err := eng.Run(ctx, runCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to start candidate process: %w", err)
	}
	candidatePID := candidateProc.Pid

	// Helper: Rollback cleanup on failure
	rollback := func(reason string) (*RolloutResult, error) {
		_ = syscall.Kill(-candidatePID, syscall.SIGKILL)
		_ = candidateProc.Kill()
		return &RolloutResult{
			Status:        "failed",
			AppID:         currentApp.ID,
			OldPID:        oldPID,
			NewPID:        candidatePID,
			RolledBack:    true,
			TotalDuration: time.Since(startTime),
			Message:       fmt.Sprintf("Rollback triggered: %s. Old instance remains running.", reason),
		}, fmt.Errorf("zero-downtime update failed: %s (old instance PID %d preserved)", reason, oldPID)
	}

	// 6. Readiness Probe / Health Check
	if candidatePort > 0 {
		targetAddr := fmt.Sprintf("127.0.0.1:%d", candidatePort)
		deadline := time.Now().Add(probeTimeout)
		ready := false

		for time.Now().Before(deadline) {
			// Check if candidate process terminated prematurely
			if err := candidateProc.Signal(syscall.Signal(0)); err != nil {
				return rollback("candidate process exited unexpectedly before becoming ready")
			}

			// Try TCP connection probe
			conn, err := net.DialTimeout("tcp", targetAddr, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				// Optional HTTP readiness check if requested
				if opts.ReadinessURL != "" {
					httpURL := fmt.Sprintf("http://127.0.0.1:%d%s", candidatePort, opts.ReadinessURL)
					client := http.Client{Timeout: 500 * time.Millisecond}
					resp, err := client.Get(httpURL)
					if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
						resp.Body.Close()
						ready = true
						break
					}
					if resp != nil {
						resp.Body.Close()
					}
				} else {
					ready = true
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
		}

		if !ready {
			return rollback(fmt.Sprintf("readiness probe timed out after %v on port %d", probeTimeout, candidatePort))
		}
	} else {
		// Non-network worker: wait 350ms to ensure stability
		time.Sleep(350 * time.Millisecond)
		if err := candidateProc.Signal(syscall.Signal(0)); err != nil {
			return rollback("worker process failed immediately after launch")
		}
	}

	// 7. Atomic Cutover
	appDir := store.AppDir(currentApp.ID)
	if candidatePort > 0 {
		if netproxy.HasProxyAcrossProcesses(currentApp.ID, appDir) {
			if err := netproxy.SwitchTargetAcrossProcesses(currentApp.ID, appDir, candidatePort); err != nil {
				return rollback(fmt.Sprintf("failed to atomically switch proxy target: %v", err))
			}
		} else if currentApp.PortMapping != "" {
			// Start proxy if it wasn't running
			pm, err := netproxy.ParsePortMapping(currentApp.PortMapping)
			if err == nil {
				pm.ContainerPort = candidatePort
				_ = netproxy.GetProxyManager().StartProxyWithControl(currentApp.ID, pm, appDir)
			}
		}
	}

	// 8. Graceful Connection Draining for Old Instance
	drainStart := time.Now()
	if candidatePort > 0 && netproxy.HasProxyAcrossProcesses(currentApp.ID, appDir) {
		netproxy.DrainAcrossProcesses(currentApp.ID, appDir, drainTimeout)
	} else {
		time.Sleep(200 * time.Millisecond)
	}
	drainDuration := time.Since(drainStart)

	// 9. Graceful Retirement of Old Instance (Instance A)
	if oldProc, err := os.FindProcess(oldPID); err == nil {
		_ = syscall.Kill(-oldPID, syscall.SIGTERM)
		_ = oldProc.Signal(syscall.SIGTERM)

		// Wait up to 3s for old instance to exit
		stopped := false
		for i := 0; i < 30; i++ {
			time.Sleep(100 * time.Millisecond)
			if err := oldProc.Signal(syscall.Signal(0)); err != nil {
				stopped = true
				break
			}
		}
		if !stopped {
			_ = syscall.Kill(-oldPID, syscall.SIGKILL)
			_ = oldProc.Kill()
		}
	}

	// 10. Atomic State Persistence
	currentApp.PID = candidatePID
	currentApp.StartedAt = time.Now()
	if candidatePort > 0 {
		currentApp.Port = candidatePort
		if currentApp.PortMapping != "" {
			if pm, err := netproxy.ParsePortMapping(currentApp.PortMapping); err == nil {
				pm.ContainerPort = candidatePort
				currentApp.PortMapping = pm.String()
			}
		}
	}
	if opts.NewCommand != "" {
		currentApp.Command = opts.NewCommand
	}
	if opts.NewArgs != nil {
		currentApp.Args = opts.NewArgs
	} else {
		currentApp.Args = candidateApp.Args
	}
	if opts.NewEnv != nil {
		currentApp.Env = candidateApp.Env
	}
	currentApp.Status = app.StatusRunning
	_ = store.SaveApp(currentApp)

	return &RolloutResult{
		Status:        "completed",
		AppID:         currentApp.ID,
		OldPID:        oldPID,
		NewPID:        candidatePID,
		TargetPort:    candidatePort,
		DrainDuration: drainDuration,
		TotalDuration: time.Since(startTime),
		RolledBack:    false,
		Message:       fmt.Sprintf("Zero-downtime cutover complete! Swapped PID %d -> %d seamlessly.", oldPID, candidatePID),
	}, nil
}

// StartApp starts a stopped application using the appropriate engine and updates state
func StartApp(ctx context.Context, store *app.Store, a *app.App) (*os.Process, error) {
	var eng engine.Engine
	if a.Engine == app.EngineRootfs {
		eng = engine.NewRootfsEngine()
	} else {
		eng = engine.NewNativeEngine()
	}

	cfg := &engine.RunConfig{
		App:      a,
		Store:    store,
		Detached: true,
	}

	proc, err := eng.Run(ctx, cfg)
	if err != nil {
		a.Status = app.StatusFailed
		_ = store.SaveApp(a)
		return nil, err
	}

	a.PID = proc.Pid
	a.Status = app.StatusRunning
	a.StartedAt = time.Now()
	if err := store.SaveApp(a); err != nil {
		return nil, err
	}

	return proc, nil
}


var _ = filepath.Join
