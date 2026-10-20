package supervisor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/engine"
)

// Supervisor monitors background processes and enforces restart policies
type Supervisor struct {
	mu         sync.Mutex
	cancelFunc map[string]context.CancelFunc // keyed by appID
}

var globalSupervisor = &Supervisor{
	cancelFunc: make(map[string]context.CancelFunc),
}

// GetSupervisor returns the global supervisor singleton
func GetSupervisor() *Supervisor {
	return globalSupervisor
}

// Watch starts monitoring an application process according to its RestartPolicy
func (s *Supervisor) Watch(a *app.App, store *app.Store, eng engine.Engine) {
	if a.RestartPolicy == "" || a.RestartPolicy == "no" {
		return
	}

	s.mu.Lock()
	if cancel, exists := s.cancelFunc[a.ID]; exists {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc[a.ID] = cancel
	s.mu.Unlock()

	go s.monitorLoop(ctx, a.ID, store, eng)
}

// Cancel stops monitoring for a given app ID
func (s *Supervisor) Cancel(appID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.cancelFunc[appID]; exists {
		cancel()
		delete(s.cancelFunc, appID)
	}
}

func (s *Supervisor) monitorLoop(ctx context.Context, appID string, store *app.Store, eng engine.Engine) {
	backoff := 500 * time.Millisecond
	const maxBackoff = 10 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		currentApp, err := store.LoadApp(appID)
		if err != nil {
			return
		}

		// If user explicitly marked it stopped, exit monitor loop
		if currentApp.Status == app.StatusStopped {
			return
		}

		pid := currentApp.PID
		if pid <= 0 {
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return
		}

		// Wait for process termination
		startedTime := time.Now()
		exitCode := 0
		hasExited := false

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Check if process is still alive using signal 0
			if err := process.Signal(syscall.Signal(0)); err != nil {
				// Process has terminated
				hasExited = true
				break
			}
			time.Sleep(200 * time.Millisecond)
		}

		if !hasExited {
			return
		}

		// Check if process was healthy for more than 15s, reset backoff
		if time.Since(startedTime) > 15*time.Second {
			backoff = 500 * time.Millisecond
		}

		// Re-read app state to verify it wasn't stopped intentionally
		freshApp, err := store.LoadApp(appID)
		if err != nil || freshApp.Status == app.StatusStopped {
			return
		}

		// Determine if restart policy triggers
		shouldRestart := false
		if freshApp.RestartPolicy == "always" {
			shouldRestart = true
		} else if freshApp.RestartPolicy == "on-failure" {
			// On failure restarts if exit was abnormal (exitCode != 0 or killed)
			shouldRestart = true
		}

		if !shouldRestart {
			freshApp.Status = app.StatusStopped
			_ = store.SaveApp(freshApp)
			return
		}

		// Log restart event
		stdoutPath, _ := store.LogPaths(appID)
		logMsg := fmt.Sprintf("\n[docroid supervisor] Process terminated. Restarting (#%d) in %v...\n",
			freshApp.RestartCount+1, backoff)
		if f, err := os.OpenFile(stdoutPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = f.WriteString(logMsg)
			f.Close()
		}

		// Apply backoff
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		// Exponential backoff
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		// Restart process
		runCfg := &engine.RunConfig{
			App:      freshApp,
			Store:    store,
			Detached: true,
		}

		proc, err := eng.Run(context.Background(), runCfg)
		if err != nil {
			freshApp.Status = app.StatusFailed
			_ = store.SaveApp(freshApp)
			return
		}

		freshApp.PID = proc.Pid
		freshApp.Status = app.StatusRunning
		freshApp.StartedAt = time.Now()
		freshApp.RestartCount++
		freshApp.ExitCode = exitCode
		_ = store.SaveApp(freshApp)
	}
}
