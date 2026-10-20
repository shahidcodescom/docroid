package tunnel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"docroid/pkg/app"
)

type Store struct {
	baseDir string
	mu      sync.RWMutex
}

func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		baseDir = app.GetDefaultBaseDir()
	}

	tunnelsDir := filepath.Join(baseDir, "tunnels")
	if err := os.MkdirAll(tunnelsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tunnels directory %s: %w", tunnelsDir, err)
	}

	logsDir := filepath.Join(tunnelsDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tunnel logs directory %s: %w", logsDir, err)
	}

	return &Store{baseDir: tunnelsDir}, nil
}

func (s *Store) BaseDir() string {
	return s.baseDir
}

func (s *Store) jsonPath() string {
	return filepath.Join(s.baseDir, "tunnels.json")
}

func (s *Store) LogPath(tunnelID string) string {
	return filepath.Join(s.baseDir, "logs", tunnelID+".log")
}

func (s *Store) loadUnlocked() (map[string]*Tunnel, error) {
	path := s.jsonPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string]*Tunnel), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tunnels file: %w", err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return make(map[string]*Tunnel), nil
	}

	var tunnels map[string]*Tunnel
	if err := json.Unmarshal(data, &tunnels); err != nil {
		return nil, fmt.Errorf("failed to parse tunnels json: %w", err)
	}

	if tunnels == nil {
		tunnels = make(map[string]*Tunnel)
	}
	return tunnels, nil
}

func (s *Store) saveUnlocked(tunnels map[string]*Tunnel) error {
	data, err := json.MarshalIndent(tunnels, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tunnels: %w", err)
	}

	tmpFile := s.jsonPath() + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tmp tunnels file: %w", err)
	}

	if err := os.Rename(tmpFile, s.jsonPath()); err != nil {
		return fmt.Errorf("failed to commit tunnels file: %w", err)
	}
	return nil
}

// SaveTunnel persists or updates a tunnel entry
func (s *Store) SaveTunnel(t *Tunnel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tunnels, err := s.loadUnlocked()
	if err != nil {
		return err
	}

	tunnels[t.ID] = t
	return s.saveUnlocked(tunnels)
}

// GetTunnel retrieves a tunnel by exact ID
func (s *Store) GetTunnel(id string) (*Tunnel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tunnels, err := s.loadUnlocked()
	if err != nil {
		return nil, err
	}

	t, exists := tunnels[id]
	if !exists {
		return nil, fmt.Errorf("tunnel '%s' not found", id)
	}
	return t, nil
}

// FindTunnel matches a tunnel by exact ID, ID prefix, or case-insensitive Name
func (s *Store) FindTunnel(query string) (*Tunnel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tunnels, err := s.loadUnlocked()
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	// 1. Exact ID
	if t, exists := tunnels[query]; exists {
		return t, nil
	}

	// 2. Exact Name
	for _, t := range tunnels {
		if strings.EqualFold(t.Name, query) {
			return t, nil
		}
	}

	// 3. ID Prefix
	var prefixMatches []*Tunnel
	for id, t := range tunnels {
		if strings.HasPrefix(id, query) {
			prefixMatches = append(prefixMatches, t)
		}
	}
	if len(prefixMatches) == 1 {
		return prefixMatches[0], nil
	}
	if len(prefixMatches) > 1 {
		return nil, fmt.Errorf("ambiguous tunnel identifier '%s', matches multiple tunnels", query)
	}

	return nil, fmt.Errorf("tunnel '%s' not found", query)
}

// ListTunnels returns all stored tunnels sorted by creation date descending
func (s *Store) ListTunnels() ([]*Tunnel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tunnels, err := s.loadUnlocked()
	if err != nil {
		return nil, err
	}

	list := make([]*Tunnel, 0, len(tunnels))
	for _, t := range tunnels {
		list = append(list, t)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	return list, nil
}

// DeleteTunnel removes a tunnel from store and deletes its log file
func (s *Store) DeleteTunnel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tunnels, err := s.loadUnlocked()
	if err != nil {
		return err
	}

	delete(tunnels, id)
	if err := s.saveUnlocked(tunnels); err != nil {
		return err
	}

	_ = os.Remove(s.LogPath(id))
	return nil
}
