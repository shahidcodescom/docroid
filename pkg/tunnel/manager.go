package tunnel

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"docroid/pkg/app"
)

type Manager struct {
	store      *Store
	activeCmds map[string]*exec.Cmd
	mu         sync.RWMutex
}

func NewManager(store *Store) *Manager {
	m := &Manager{
		store:      store,
		activeCmds: make(map[string]*exec.Cmd),
	}
	m.SyncProcesses()
	return m
}

func (m *Manager) Store() *Store {
	return m.store
}

// FindCloudflaredPath searches for the cloudflared binary across standard system and Termux paths
func FindCloudflaredPath() (string, error) {
	// 1. Check PATH
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p, nil
	}

	// 2. Termux PREFIX
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		termuxPath := filepath.Join(prefix, "bin", "cloudflared")
		if _, err := os.Stat(termuxPath); err == nil {
			return termuxPath, nil
		}
	}
	termuxHardcoded := "/data/data/com.termux/files/usr/bin/cloudflared"
	if _, err := os.Stat(termuxHardcoded); err == nil {
		return termuxHardcoded, nil
	}

	// 3. User local and Docroid storage
	defaultBase := app.GetDefaultBaseDir()
	docroidBin := filepath.Join(defaultBase, "bin", "cloudflared")
	if _, err := os.Stat(docroidBin); err == nil {
		return docroidBin, nil
	}

	// 4. Common macOS and Linux locations
	candidates := []string{
		"/opt/homebrew/bin/cloudflared",
		"/usr/local/bin/cloudflared",
		"/usr/bin/cloudflared",
		"/bin/cloudflared",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("cloudflared binary not found in PATH or standard installation locations")
}

// CloudflaredInfo contains details about local cloudflared installation
type CloudflaredInfo struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

// GetCloudflaredInfo returns binary installation status and version
func GetCloudflaredInfo() CloudflaredInfo {
	p, err := FindCloudflaredPath()
	if err != nil {
		return CloudflaredInfo{
			Installed: false,
			Error:     "cloudflared is not installed. Install via 'brew install cloudflared' (macOS), 'pkg install cloudflared' (Termux), or download from Cloudflare releases.",
		}
	}

	cmd := exec.Command(p, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return CloudflaredInfo{
			Installed: true,
			Path:      p,
			Version:   "unknown",
		}
	}

	raw := strings.TrimSpace(string(out))
	version := raw
	// Extract version e.g. "cloudflared version 2026.6.0 (...)" -> "2026.6.0"
	re := regexp.MustCompile(`version\s+([0-9\.]+[a-zA-Z0-9\-]*)`)
	if match := re.FindStringSubmatch(raw); len(match) > 1 {
		version = match[1]
	}

	return CloudflaredInfo{
		Installed: true,
		Path:      p,
		Version:   version,
	}
}

// CreateTunnel configures and initializes a new tunnel entry in the store
func (m *Manager) CreateTunnel(opts TunnelOptions) (*Tunnel, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		if opts.AppName != "" {
			name = fmt.Sprintf("%s-tunnel", opts.AppName)
		} else if opts.TargetPort > 0 {
			name = fmt.Sprintf("port-%d", opts.TargetPort)
		} else {
			name = fmt.Sprintf("tunnel-%s", GenerateTunnelID()[:6])
		}
	}

	// Sanitize name
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9\-_.]`).ReplaceAllString(name, "-")

	mode := opts.Mode
	if mode != ModeToken {
		mode = ModeQuick
	}

	if mode == ModeToken && strings.TrimSpace(opts.Token) == "" {
		return nil, fmt.Errorf("tunnel token is required for token-based named tunnels")
	}

	normalizedTarget, targetPort, err := NormalizeTarget(opts.Target, opts.TargetPort)
	if err != nil && mode != ModeToken {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	id := GenerateTunnelID()
	t := &Tunnel{
		ID:          id,
		Name:        name,
		AppID:       opts.AppID,
		AppName:     opts.AppName,
		Target:      normalizedTarget,
		TargetPort:  targetPort,
		Mode:        mode,
		Token:       strings.TrimSpace(opts.Token),
		Hostname:    strings.TrimSpace(opts.Hostname),
		Status:      StatusStopped,
		PID:         0,
		CreatedAt:   time.Now(),
		LogFile:     m.store.LogPath(id),
		AutoRestart: opts.AutoRestart,
	}

	if err := m.store.SaveTunnel(t); err != nil {
		return nil, fmt.Errorf("failed to save tunnel: %w", err)
	}

	return t, nil
}

// Start begins execution of cloudflared for the given tunnel ID
func (m *Manager) Start(tunnelID string) (*Tunnel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, err := m.store.GetTunnel(tunnelID)
	if err != nil {
		return nil, err
	}

	// If already running and alive, return
	if t.Status == StatusRunning || t.Status == StatusStarting {
		if t.PID > 0 && isPIDAlive(t.PID) {
			return t, nil
		}
	}

	binPath, err := FindCloudflaredPath()
	if err != nil {
		t.Status = StatusError
		t.LastError = err.Error()
		_ = m.store.SaveTunnel(t)
		return nil, err
	}

	var args []string
	if t.Mode == ModeToken {
		args = []string{"tunnel", "run", "--token", t.Token, "--no-autoupdate"}
	} else {
		// Quick tunnel: provide an empty config file so cloudflared doesn't read ~/.cloudflared/config.yml
		emptyCfg := filepath.Join(m.store.BaseDir(), "empty-config.yml")
		if _, err := os.Stat(emptyCfg); os.IsNotExist(err) {
			_ = os.WriteFile(emptyCfg, []byte("# docroid isolated tunnel config\n"), 0644)
		}
		args = []string{"tunnel", "--config", emptyCfg, "--url", t.Target, "--no-autoupdate"}

		// Forward target host header so origin local servers / vhosts accept requests
		if u, err := url.Parse(t.Target); err == nil && u.Host != "" {
			args = append(args, "--http-host-header", u.Host)
		}
		if t.Hostname != "" {
			args = append(args, "--hostname", t.Hostname)
		}
	}

	cmd := exec.Command(binPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Create log file
	logPath := m.store.LogPath(t.ID)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Status = StatusError
		t.LastError = fmt.Sprintf("failed to open log file: %v", err)
		_ = m.store.SaveTunnel(t)
		return nil, err
	}

	// Add timestamp marker to log
	_, _ = fmt.Fprintf(logFile, "\n--- Tunnel %s starting at %s ---\n", t.ID, time.Now().Format(time.RFC3339))

	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		t.Status = StatusError
		t.LastError = fmt.Sprintf("failed to start cloudflared: %v", err)
		_ = m.store.SaveTunnel(t)
		return nil, err
	}

	t.PID = cmd.Process.Pid
	t.Status = StatusStarting
	t.StartedAt = time.Now()
	t.LastError = ""
	if t.Mode == ModeQuick {
		t.PublicURL = ""
	}
	_ = m.store.SaveTunnel(t)

	m.activeCmds[t.ID] = cmd

	// Background watcher for URL extraction and process completion
	go func(tunID string, logF *os.File, c *exec.Cmd) {
		defer logF.Close()

		// Background polling for URL from log file
		pollStop := make(chan struct{})
		go func() {
			ticker := time.NewTicker(300 * time.Millisecond)
			defer ticker.Stop()
			for i := 0; i < 80; i++ { // poll for up to 24s
				select {
				case <-pollStop:
					return
				case <-ticker.C:
					data, err := os.ReadFile(m.store.LogPath(tunID))
					if err == nil {
						content := extractLatestSessionLog(tunID, string(data))
						if t.Mode == ModeQuick {
							if u := ExtractPublicURL(content); u != "" {
								m.mu.Lock()
								if cur, err := m.store.GetTunnel(tunID); err == nil {
									cur.PublicURL = u
									cur.Status = StatusRunning
									_ = m.store.SaveTunnel(cur)
								}
								m.mu.Unlock()
								return
							}
						} else if t.Mode == ModeToken {
							if strings.Contains(content, "Registered tunnel connection") || (strings.Contains(content, "Connection") && strings.Contains(content, "registered")) {
								m.mu.Lock()
								if cur, err := m.store.GetTunnel(tunID); err == nil {
									cur.Status = StatusRunning
									_ = m.store.SaveTunnel(cur)
								}
								m.mu.Unlock()
								return
							}
						}
					}
				}
			}
		}()

		// Wait for command completion
		waitErr := c.Wait()
		close(pollStop)

		m.mu.Lock()
		delete(m.activeCmds, tunID)

		cur, err := m.store.GetTunnel(tunID)
		if err == nil {
			cur.PID = 0
			cur.StoppedAt = time.Now()
			if cur.Status != StatusStopped {
				if waitErr != nil {
					cur.Status = StatusError
					cur.LastError = fmt.Sprintf("cloudflared exited: %v", waitErr)
				} else {
					cur.Status = StatusStopped
				}
			}
			_ = m.store.SaveTunnel(cur)

			// Auto-restart handling if enabled and not stopped manually
			if cur.AutoRestart && cur.Status != StatusStopped {
				m.mu.Unlock()
				time.Sleep(3 * time.Second)
				_, _ = m.Start(tunID)
				return
			}
		}
		m.mu.Unlock()
	}(t.ID, logFile, cmd)

	return t, nil
}

// Stop terminates a running tunnel
func (m *Manager) Stop(tunnelID string) (*Tunnel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, err := m.store.GetTunnel(tunnelID)
	if err != nil {
		return nil, err
	}

	if cmd, exists := m.activeCmds[tunnelID]; exists && cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
		}
		delete(m.activeCmds, tunnelID)
	} else if t.PID > 0 {
		// Process might have been started by a previous Docroid run
		if proc, err := os.FindProcess(t.PID); err == nil {
			_ = proc.Signal(syscall.SIGTERM)
			time.Sleep(300 * time.Millisecond)
			if isPIDAlive(t.PID) {
				_ = proc.Kill()
			}
		}
	}

	t.Status = StatusStopped
	t.PID = 0
	t.StoppedAt = time.Now()
	_ = m.store.SaveTunnel(t)

	return t, nil
}

// Restart stops and restarts a tunnel
func (m *Manager) Restart(tunnelID string) (*Tunnel, error) {
	_, _ = m.Stop(tunnelID)
	time.Sleep(500 * time.Millisecond)
	return m.Start(tunnelID)
}

// Delete stops and permanently removes a tunnel
func (m *Manager) Delete(tunnelID string) error {
	_, _ = m.Stop(tunnelID)
	return m.store.DeleteTunnel(tunnelID)
}

// GetLogs returns the last N lines from the tunnel log file
func (m *Manager) GetLogs(tunnelID string, maxLines int) (string, error) {
	t, err := m.store.GetTunnel(tunnelID)
	if err != nil {
		return "", err
	}

	logPath := m.store.LogPath(t.ID)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return "(no logs generated yet)", nil
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if maxLines <= 0 || len(lines) <= maxLines {
		return string(data), nil
	}

	tail := lines[len(lines)-maxLines:]
	return strings.Join(tail, "\n"), nil
}

// SyncProcesses reconciles persistent store with alive PIDs on startup
func (m *Manager) SyncProcesses() {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnels, err := m.store.ListTunnels()
	if err != nil {
		return
	}

	for _, t := range tunnels {
		if t.PID > 0 && isPIDAlive(t.PID) {
			if t.Mode == ModeQuick {
				if data, err := os.ReadFile(m.store.LogPath(t.ID)); err == nil {
					content := extractLatestSessionLog(t.ID, string(data))
					if u := ExtractPublicURL(content); u != "" {
						t.PublicURL = u
						t.Status = StatusRunning
						_ = m.store.SaveTunnel(t)
					}
				}
			} else if t.Mode == ModeToken {
				t.Status = StatusRunning
				_ = m.store.SaveTunnel(t)
			}
		} else if t.Status == StatusRunning || t.Status == StatusStarting {
			t.Status = StatusStopped
			t.PID = 0
			t.StoppedAt = time.Now()
			_ = m.store.SaveTunnel(t)
		}
	}
}

// extractLatestSessionLog isolates the log output from the most recent tunnel invocation
func extractLatestSessionLog(tunID, fullLog string) string {
	markerPrefix := fmt.Sprintf("--- Tunnel %s starting at", tunID)
	if idx := strings.LastIndex(fullLog, markerPrefix); idx != -1 {
		return fullLog[idx:]
	}
	return fullLog
}

// StopAll terminates all running tunnels (called on server shutdown)
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, cmd := range m.activeCmds {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		delete(m.activeCmds, id)
	}

	tunnels, _ := m.store.ListTunnels()
	for _, t := range tunnels {
		if t.Status == StatusRunning || t.Status == StatusStarting {
			t.Status = StatusStopped
			t.PID = 0
			t.StoppedAt = time.Now()
			_ = m.store.SaveTunnel(t)
		}
	}
}

// isPIDAlive checks if a process with the given PID is currently active
func isPIDAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// DiscoverableTarget represents an available local port/service to tunnel
type DiscoverableTarget struct {
	Type     string `json:"type"`      // "app", "site", "dashboard", "custom"
	ID       string `json:"id"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	Selected bool   `json:"selected,omitempty"`
}

// DiscoverTargets finds all active Docroid applications, static sites, and services that can be tunneled
func (m *Manager) DiscoverTargets(appsList []struct {
	ID     string
	Name   string
	Port   int
	Status string
}, sitesList []struct {
	ID     string
	Name   string
	Port   int
	Status string
}, dashboardPort int) []DiscoverableTarget {
	var targets []DiscoverableTarget

	// 1. Docroid Dashboard itself
	if dashboardPort > 0 {
		targets = append(targets, DiscoverableTarget{
			Type:   "dashboard",
			ID:     "docroid-dashboard",
			Name:   "Docroid Dashboard",
			Port:   dashboardPort,
			URL:    fmt.Sprintf("http://127.0.0.1:%d", dashboardPort),
			Status: "running",
		})
	}

	// 2. Apps
	for _, a := range appsList {
		if a.Port > 0 {
			targets = append(targets, DiscoverableTarget{
				Type:   "app",
				ID:     a.ID,
				Name:   a.Name,
				Port:   a.Port,
				URL:    fmt.Sprintf("http://127.0.0.1:%d", a.Port),
				Status: a.Status,
			})
		}
	}

	// 3. Static Sites
	for _, s := range sitesList {
		if s.Port > 0 {
			targets = append(targets, DiscoverableTarget{
				Type:   "site",
				ID:     s.ID,
				Name:   s.Name,
				Port:   s.Port,
				URL:    fmt.Sprintf("http://127.0.0.1:%d", s.Port),
				Status: s.Status,
			})
		}
	}

	return targets
}
