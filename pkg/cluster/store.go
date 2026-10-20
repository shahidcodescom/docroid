package cluster

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"docroid/pkg/app"
)

// ClusterConfig represents the cluster configuration for the local node
type ClusterConfig struct {
	ClusterName         string    `json:"cluster_name"`
	Role                NodeRole  `json:"role"`
	JoinToken           string    `json:"join_token"`
	ControlPlaneAddress string    `json:"control_plane_address,omitempty"`
	LocalNodeID         string    `json:"local_node_id"`
	InitializedAt       time.Time `json:"initialized_at"`
}

// Store handles persistent storage of cluster metadata, nodes, deployments, and pods
type Store struct {
	baseDir string
	mu      sync.RWMutex
}

// NewStore initializes cluster store directories
func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		baseDir = app.GetDefaultBaseDir()
	}

	clusterDir := filepath.Join(baseDir, "cluster")
	if err := os.MkdirAll(clusterDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cluster directory %s: %w", clusterDir, err)
	}

	return &Store{baseDir: clusterDir}, nil
}

func (s *Store) ConfigPath() string {
	return filepath.Join(s.baseDir, "config.json")
}

func (s *Store) NodesPath() string {
	return filepath.Join(s.baseDir, "nodes.json")
}

func (s *Store) DeploymentsPath() string {
	return filepath.Join(s.baseDir, "deployments.json")
}

func (s *Store) PodsPath() string {
	return filepath.Join(s.baseDir, "pods.json")
}

// GenerateJoinToken creates a secure cluster join token
func GenerateJoinToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("dcr-tok-%s", hex.EncodeToString(b))
}

// LoadConfig reads the cluster configuration
func (s *Store) LoadConfig() (*ClusterConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not in a cluster yet
		}
		return nil, err
	}

	var cfg ClusterConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes the cluster configuration
func (s *Store) SaveConfig(cfg *ClusterConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.ConfigPath(), data, 0644)
}

// ClearConfig removes the cluster configuration (leaving cluster)
func (s *Store) ClearConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = os.Remove(s.ConfigPath())
	_ = os.Remove(s.NodesPath())
	_ = os.Remove(s.DeploymentsPath())
	_ = os.Remove(s.PodsPath())
	return nil
}

// LoadNodes reads the list of registered nodes
func (s *Store) LoadNodes() ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.NodesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []*Node{}, nil
		}
		return nil, err
	}

	var nodes []*Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// SaveNodes writes the list of registered nodes
func (s *Store) SaveNodes(nodes []*Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.NodesPath(), data, 0644)
}

// LoadDeployments reads stored deployments
func (s *Store) LoadDeployments() ([]*Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.DeploymentsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []*Deployment{}, nil
		}
		return nil, err
	}

	var deps []*Deployment
	if err := json.Unmarshal(data, &deps); err != nil {
		return nil, err
	}
	return deps, nil
}

// SaveDeployments writes deployments to disk
func (s *Store) SaveDeployments(deps []*Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(deps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.DeploymentsPath(), data, 0644)
}

// LoadPods reads stored pods
func (s *Store) LoadPods() ([]*Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.PodsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []*Pod{}, nil
		}
		return nil, err
	}

	var pods []*Pod
	if err := json.Unmarshal(data, &pods); err != nil {
		return nil, err
	}
	return pods, nil
}

// SavePods writes pods to disk
func (s *Store) SavePods(pods []*Pod) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(pods, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.PodsPath(), data, 0644)
}
