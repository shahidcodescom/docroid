package network

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Store handles persistent storage of Docroid virtual networks
type Store struct {
	dir string
	mu  sync.RWMutex
}

// NewStore initializes network store directory
func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(home, ".docroid")
	}

	netDir := filepath.Join(baseDir, "networks")
	if err := os.MkdirAll(netDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create networks dir: %w", err)
	}

	s := &Store{dir: netDir}
	if err := s.initDefaults(); err != nil {
		return nil, fmt.Errorf("failed to initialize default networks: %w", err)
	}

	return s, nil
}

// initDefaults ensures standard 'bridge' and 'host' networks exist
func (s *Store) initDefaults() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bridgePath := filepath.Join(s.dir, "bridge.json")
	if _, err := os.Stat(bridgePath); os.IsNotExist(err) {
		bridge := &Network{
			ID:         GenerateNetworkID(),
			Name:       "bridge",
			Driver:     DriverBridge,
			Scope:      "local",
			Internal:   false,
			Subnet:     "10.89.0.0/16",
			Gateway:    "10.89.0.1",
			Containers: make(map[string]*Endpoint),
			Labels:     map[string]string{"type": "default"},
			CreatedAt:  time.Now(),
		}
		if err := s.saveLocked(bridge); err != nil {
			return err
		}
	}

	hostPath := filepath.Join(s.dir, "host.json")
	if _, err := os.Stat(hostPath); os.IsNotExist(err) {
		host := &Network{
			ID:         GenerateNetworkID(),
			Name:       "host",
			Driver:     DriverHost,
			Scope:      "local",
			Internal:   false,
			Subnet:     "127.0.0.0/8",
			Gateway:    "127.0.0.1",
			Containers: make(map[string]*Endpoint),
			Labels:     map[string]string{"type": "default"},
			CreatedAt:  time.Now(),
		}
		if err := s.saveLocked(host); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) networkFilePath(nameOrID string) string {
	return filepath.Join(s.dir, CleanName(nameOrID)+".json")
}

// Save persists a network definition
func (s *Store) Save(n *Network) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(n)
}

func (s *Store) saveLocked(n *Network) error {
	if n.Name == "" {
		return fmt.Errorf("network name cannot be empty")
	}
	if n.Containers == nil {
		n.Containers = make(map[string]*Endpoint)
	}

	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal network: %w", err)
	}

	path := s.networkFilePath(n.Name)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write network file: %w", err)
	}

	return os.Rename(tmpPath, path)
}

// Get retrieves a network by name or ID
func (s *Store) Get(nameOrID string) (*Network, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	target := CleanName(nameOrID)
	// Try direct by name
	path := s.networkFilePath(target)
	if data, err := os.ReadFile(path); err == nil {
		var n Network
		if err := json.Unmarshal(data, &n); err == nil {
			return &n, nil
		}
	}

	// Try scanning by ID or prefix
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read networks dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
			if err != nil {
				continue
			}
			var n Network
			if err := json.Unmarshal(data, &n); err == nil {
				if n.Name == target || n.ID == target || strings.HasPrefix(n.ID, target) {
					return &n, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("network '%s' not found", nameOrID)
}

// List returns all persisted networks
func (s *Store) List() ([]*Network, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read networks dir: %w", err)
	}

	var networks []*Network
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
			if err != nil {
				continue
			}
			var n Network
			if err := json.Unmarshal(data, &n); err == nil {
				networks = append(networks, &n)
			}
		}
	}

	return networks, nil
}

// Delete removes a user-defined network (default bridge/host cannot be deleted)
func (s *Store) Delete(nameOrID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	net, err := s.getInternal(nameOrID)
	if err != nil {
		return err
	}

	if net.Name == "bridge" || net.Name == "host" {
		return fmt.Errorf("cannot remove default system network '%s'", net.Name)
	}

	if len(net.Containers) > 0 {
		return fmt.Errorf("network '%s' has %d active connected containers; disconnect them first", net.Name, len(net.Containers))
	}

	path := s.networkFilePath(net.Name)
	return os.Remove(path)
}

func (s *Store) getInternal(nameOrID string) (*Network, error) {
	target := CleanName(nameOrID)
	path := s.networkFilePath(target)
	if data, err := os.ReadFile(path); err == nil {
		var n Network
		if err := json.Unmarshal(data, &n); err == nil {
			return &n, nil
		}
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read networks dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
			if err != nil {
				continue
			}
			var n Network
			if err := json.Unmarshal(data, &n); err == nil {
				if n.Name == target || n.ID == target || strings.HasPrefix(n.ID, target) {
					return &n, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("network '%s' not found", nameOrID)
}
