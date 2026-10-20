package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store struct {
	baseDir string
	mu      sync.RWMutex
}

// GetDefaultBaseDir resolves the root storage directory for Docroid
func GetDefaultBaseDir() string {
	if env := os.Getenv("DOCROID_HOME"); env != "" {
		return env
	}
	// In Termux, $HOME is /data/data/com.termux/files/home
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".docroid")
	}
	return filepath.Join(os.TempDir(), "docroid")
}

// NewStore initializes the store directory structure
func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		baseDir = GetDefaultBaseDir()
	}

	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "apps"),
		filepath.Join(baseDir, "workspaces"),
		filepath.Join(baseDir, "rootfs"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	return &Store{baseDir: baseDir}, nil
}

// BaseDir returns the root storage path
func (s *Store) BaseDir() string {
	return s.baseDir
}

// AppDir returns the directory for a specific app
func (s *Store) AppDir(appID string) string {
	return filepath.Join(s.baseDir, "apps", appID)
}

// WorkspaceDir returns the isolated workspace directory for an app
func (s *Store) WorkspaceDir(appID string) string {
	return filepath.Join(s.baseDir, "workspaces", appID)
}

// RootfsDir returns the directory containing unpacked rootfs images
func (s *Store) RootfsDir(name string) string {
	return filepath.Join(s.baseDir, "rootfs", name)
}

// LogPaths returns the stdout and stderr log file paths for an app
func (s *Store) LogPaths(appID string) (stdout string, stderr string) {
	dir := s.AppDir(appID)
	return filepath.Join(dir, "stdout.log"), filepath.Join(dir, "stderr.log")
}

// SaveApp writes the App metadata to apps/<id>/app.json atomically
func (s *Store) SaveApp(a *App) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := s.AppDir(a.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Also ensure workspace directory exists
	wsDir := s.WorkspaceDir(a.ID)
	_ = os.MkdirAll(wsDir, 0755)

	filePath := filepath.Join(dir, "app.json")
	tempPath := filePath + ".tmp"

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal app JSON: %w", err)
	}

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write app temp JSON: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("failed to rename app JSON: %w", err)
	}

	return nil
}

// LoadApp reads app metadata for a given app ID
func (s *Store) LoadApp(appID string) (*App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.AppDir(appID), "app.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("app %s not found: %w", appID, err)
	}

	var a App
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("failed to parse app JSON for %s: %w", appID, err)
	}

	return &a, nil
}

// FindApp locates an app by ID (exact or prefix) or Name
func (s *Store) FindApp(idOrName string) (*App, error) {
	apps, err := s.ListApps()
	if err != nil {
		return nil, err
	}

	for _, a := range apps {
		if a.ID == idOrName || strings.HasPrefix(a.ID, idOrName) || a.Name == idOrName {
			return a, nil
		}
	}

	return nil, fmt.Errorf("no application found matching '%s'", idOrName)
}

// ListApps returns all registered applications
func (s *Store) ListApps() ([]*App, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	appsDir := filepath.Join(s.baseDir, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, err
	}

	var result []*App
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		jsonPath := filepath.Join(appsDir, entry.Name(), "app.json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			continue
		}
		var a App
		if err := json.Unmarshal(data, &a); err == nil {
			result = append(result, &a)
		}
	}

	return result, nil
}

// DeleteApp removes an app and optionally its workspace
func (s *Store) DeleteApp(appID string, removeWorkspace bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = os.RemoveAll(s.AppDir(appID))
	if removeWorkspace {
		_ = os.RemoveAll(s.WorkspaceDir(appID))
	}
	return nil
}
