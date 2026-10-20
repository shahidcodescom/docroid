package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"docroid/pkg/metrics"
	"docroid/pkg/names"
	"docroid/pkg/utils"
)

var (
	stopHooks   []func(appID string)
	stopHooksMu sync.Mutex
)

// RegisterStopHook registers a callback to be invoked when an application is stopped
func RegisterStopHook(fn func(appID string)) {
	stopHooksMu.Lock()
	defer stopHooksMu.Unlock()
	stopHooks = append(stopHooks, fn)
}

type AppManager struct {
	store *Store
}

func NewAppManager(store *Store) *AppManager {
	return &AppManager{store: store}
}

func (m *AppManager) Store() *Store {
	return m.store
}

// AppWithStats combines App metadata with real-time process metrics
type AppWithStats struct {
	App   *App
	Stats *metrics.ProcessStats
}

// ListAppsWithStats returns all apps with live status & metrics, auto-reaping dead processes
func (m *AppManager) ListAppsWithStats() ([]*AppWithStats, error) {
	apps, err := m.store.ListApps()
	if err != nil {
		return nil, err
	}

	var result []*AppWithStats
	for _, a := range apps {
		stats := &metrics.ProcessStats{PID: a.PID}
		if a.Status == StatusRunning && a.PID > 0 {
			st, err := metrics.GetProcessStats(a.PID)
			if err == nil && st.Alive {
				stats = st
			} else {
				// Process is no longer running, update state
				a.Status = StatusStopped
				_ = m.store.SaveApp(a)
				stats.Alive = false
			}
		}
		result = append(result, &AppWithStats{
			App:   a,
			Stats: stats,
		})
	}

	return result, nil
}

// CreateApp creates and persists a new application instance
func (m *AppManager) CreateApp(
	name string,
	engine EngineType,
	command string,
	args []string,
	env map[string]string,
	workDir string,
	port int,
	volumes []VolumeMount,
	rootfs string,
	restartPolicy string,
) (*App, error) {
	if command == "" && engine == EngineNative {
		return nil, fmt.Errorf("command or binary path is required")
	}

	if name == "" {
		name = names.GenerateRandomName()
	}

	// Check if name already exists
	if existing, _ := m.store.FindApp(name); existing != nil {
		return nil, fmt.Errorf("an application with name '%s' already exists", name)
	}

	// Validate port availability if requested
	if port > 0 {
		if err := utils.CheckPortInUse(port); err != nil {
			return nil, err
		}
	}

	appID := GenerateID()
	a := &App{
		ID:            appID,
		Name:          name,
		Engine:        engine,
		Command:       command,
		Args:          args,
		Env:           env,
		WorkDir:       workDir,
		Port:          port,
		Volumes:       volumes,
		Rootfs:        rootfs,
		RestartPolicy: restartPolicy,
		Status:        StatusCreated,
		CreatedAt:     time.Now(),
	}

	if err := m.store.SaveApp(a); err != nil {
		return nil, err
	}

	return a, nil
}

// StopApp gracefully stops a running application using SIGTERM then SIGKILL
func (m *AppManager) StopApp(idOrName string) error {
	a, err := m.store.FindApp(idOrName)
	if err != nil {
		return err
	}

	stopHooksMu.Lock()
	for _, fn := range stopHooks {
		fn(a.ID)
	}
	stopHooksMu.Unlock()

	// Terminate supervisor process if one exists
	supFile := filepath.Join(m.store.AppDir(a.ID), "supervisor.pid")
	if data, err := os.ReadFile(supFile); err == nil {
		if supPid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && supPid > 0 {
			if proc, err := os.FindProcess(supPid); err == nil {
				_ = syscall.Kill(-supPid, syscall.SIGTERM)
				_ = proc.Signal(syscall.SIGTERM)
			}
		}
		_ = os.Remove(supFile)
	}

	if a.Status != StatusRunning || a.PID <= 0 {
		a.Status = StatusStopped
		_ = m.store.SaveApp(a)
		return fmt.Errorf("app '%s' is not running", idOrName)
	}

	// Attempt graceful termination
	process, err := os.FindProcess(a.PID)
	if err != nil {
		a.Status = StatusStopped
		_ = m.store.SaveApp(a)
		return nil
	}

	// Try sending SIGTERM to process group (negative PID)
	_ = syscall.Kill(-a.PID, syscall.SIGTERM)
	_ = process.Signal(syscall.SIGTERM)

	// Wait up to 3 seconds for graceful exit
	stopped := false
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if err := process.Signal(syscall.Signal(0)); err != nil {
			stopped = true
			break
		}
	}

	// If still alive, force kill
	if !stopped {
		_ = syscall.Kill(-a.PID, syscall.SIGKILL)
		_ = process.Kill()
	}

	a.Status = StatusStopped
	return m.store.SaveApp(a)
}

// RemoveApp deletes an app and its logs/state
func (m *AppManager) RemoveApp(idOrName string, force bool) error {
	a, err := m.store.FindApp(idOrName)
	if err != nil {
		return err
	}

	if a.Status == StatusRunning {
		if !force {
			return fmt.Errorf("app '%s' is running. Stop it first or use --force", idOrName)
		}
		_ = m.StopApp(idOrName)
	}

	return m.store.DeleteApp(a.ID, true)
}

// StreamLogs writes app logs to the given writer and optionally follows output
func (m *AppManager) StreamLogs(idOrName string, follow bool, lines int, out io.Writer) error {
	a, err := m.store.FindApp(idOrName)
	if err != nil {
		return err
	}

	stdoutPath, stderrPath := m.store.LogPaths(a.ID)

	// Read existing content
	printFileTail := func(path string, prefix string) int64 {
		f, err := os.Open(path)
		if err != nil {
			return 0
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return 0
		}

		strLines := strings.Split(string(data), "\n")
		start := 0
		if lines > 0 && len(strLines) > lines {
			start = len(strLines) - lines
		}
		for _, l := range strLines[start:] {
			if l != "" {
				fmt.Fprintln(out, l)
			}
		}

		info, _ := f.Stat()
		if info != nil {
			return info.Size()
		}
		return 0
	}

	outOffset := printFileTail(stdoutPath, "")
	errOffset := printFileTail(stderrPath, "[stderr] ")

	if !follow {
		return nil
	}

	// Follow mode: poll file for newly written bytes
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Follow stdout
		if f, err := os.Open(stdoutPath); err == nil {
			info, _ := f.Stat()
			if info != nil && info.Size() > outOffset {
				_, _ = f.Seek(outOffset, io.SeekStart)
				n, _ := io.Copy(out, f)
				outOffset += n
			}
			f.Close()
		}

		// Follow stderr
		if f, err := os.Open(stderrPath); err == nil {
			info, _ := f.Stat()
			if info != nil && info.Size() > errOffset {
				_, _ = f.Seek(errOffset, io.SeekStart)
				n, _ := io.Copy(out, f)
				errOffset += n
			}
			f.Close()
		}
	}

	return nil
}
