package network

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// NetworkManager manages virtual container networks and inter-container service discovery
type NetworkManager struct {
	store *Store
	mu    sync.Mutex
}

// NewNetworkManager creates a new network manager with persistence
func NewNetworkManager(baseDir string) (*NetworkManager, error) {
	store, err := NewStore(baseDir)
	if err != nil {
		return nil, err
	}
	return &NetworkManager{store: store}, nil
}

// Store returns the underlying store
func (m *NetworkManager) Store() *Store {
	return m.store
}

// CreateNetwork creates a user-defined virtual network
func (m *NetworkManager) CreateNetwork(name string, driver DriverType, subnet string, labels map[string]string) (*Network, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := CleanName(name)
	if cleaned == "" {
		return nil, fmt.Errorf("network name cannot be empty")
	}

	if existing, _ := m.store.Get(cleaned); existing != nil {
		return nil, fmt.Errorf("network '%s' already exists", cleaned)
	}

	if driver == "" {
		driver = DriverBridge
	}

	if subnet == "" {
		// Pick next free subnet space
		all, _ := m.store.List()
		subnetIndex := len(all) + 89
		subnet = fmt.Sprintf("10.%d.0.0/16", subnetIndex)
	}

	gateway := "10.89.0.1"
	parts := strings.Split(subnet, "/")
	if len(parts) > 0 {
		ipParts := strings.Split(parts[0], ".")
		if len(ipParts) >= 3 {
			gateway = fmt.Sprintf("%s.%s.%s.1", ipParts[0], ipParts[1], ipParts[2])
		}
	}

	net := &Network{
		ID:         GenerateNetworkID(),
		Name:       cleaned,
		Driver:     driver,
		Scope:      "local",
		Internal:   false,
		Subnet:     subnet,
		Gateway:    gateway,
		Containers: make(map[string]*Endpoint),
		Labels:     labels,
		CreatedAt:  time.Now(),
	}

	if err := m.store.Save(net); err != nil {
		return nil, err
	}

	return net, nil
}

// RemoveNetwork removes a user-defined network
func (m *NetworkManager) RemoveNetwork(nameOrID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store.Delete(nameOrID)
}

// GetNetwork retrieves a network by name or ID
func (m *NetworkManager) GetNetwork(nameOrID string) (*Network, error) {
	return m.store.Get(nameOrID)
}

// ListNetworks lists all networks
func (m *NetworkManager) ListNetworks() ([]*Network, error) {
	return m.store.List()
}

// Connect attaches a container to a network with optional aliases
func (m *NetworkManager) Connect(nameOrID string, containerID, containerName string, port int, aliases []string) (*Endpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	net, err := m.store.Get(nameOrID)
	if err != nil {
		return nil, err
	}

	// Check if already connected
	if ep, exists := net.Containers[containerID]; exists {
		// Update port & aliases
		if port > 0 {
			ep.Port = port
		}
		for _, a := range aliases {
			if a != "" && !contains(ep.Aliases, a) {
				ep.Aliases = append(ep.Aliases, a)
			}
		}
		_ = m.store.Save(net)
		return ep, nil
	}

	// Allocate next IP
	ip, err := net.NextAvailableIP()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP on network '%s': %w", net.Name, err)
	}

	// Always add container name as primary alias
	var cleanedAliases []string
	if containerName != "" {
		cleanedAliases = append(cleanedAliases, containerName)
	}
	for _, a := range aliases {
		a = strings.TrimSpace(a)
		if a != "" && !contains(cleanedAliases, a) {
			cleanedAliases = append(cleanedAliases, a)
		}
	}

	ep := &Endpoint{
		ContainerID:   containerID,
		ContainerName: containerName,
		IPv4Address:   ip,
		Port:          port,
		Aliases:       cleanedAliases,
		ConnectedAt:   time.Now(),
	}

	net.Containers[containerID] = ep
	if err := m.store.Save(net); err != nil {
		return nil, fmt.Errorf("failed to save network connection: %w", err)
	}

	return ep, nil
}

// Disconnect removes a container from a network
func (m *NetworkManager) Disconnect(nameOrID string, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	net, err := m.store.Get(nameOrID)
	if err != nil {
		return err
	}

	// Find by ID or Name
	foundKey := ""
	for id, ep := range net.Containers {
		if id == containerID || ep.ContainerName == containerID {
			foundKey = id
			break
		}
	}

	if foundKey == "" {
		return fmt.Errorf("container '%s' is not connected to network '%s'", containerID, net.Name)
	}

	delete(net.Containers, foundKey)
	return m.store.Save(net)
}

// GenerateHostsContent builds an /etc/hosts formatted string for a specific container on a network.
// It maps localhost, loopback, and all peer containers/services to 127.0.0.1 (so connections to "http://db:5432" resolve to the local port).
func (m *NetworkManager) GenerateHostsContent(networkName string, currentContainerID string) (string, error) {
	if networkName == "" {
		networkName = "bridge"
	}

	net, err := m.store.Get(networkName)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("# Docroid Virtual Network Hosts\n")
	b.WriteString("# Network: " + net.Name + " (" + net.Subnet + ")\n")
	b.WriteString("127.0.0.1\tlocalhost\n")
	b.WriteString("::1\tlocalhost ip6-localhost ip6-loopback\n")

	// Map current container's name & aliases
	if currentContainerID != "" {
		if curEp, ok := net.Containers[currentContainerID]; ok {
			var hostnames []string
			if curEp.ContainerName != "" {
				hostnames = append(hostnames, curEp.ContainerName)
			}
			for _, alias := range curEp.Aliases {
				if !contains(hostnames, alias) {
					hostnames = append(hostnames, alias)
				}
			}
			if len(hostnames) > 0 {
				b.WriteString(fmt.Sprintf("127.0.0.1\t%s\n", strings.Join(hostnames, " ")))
			}
		}
	}

	// Map all peer containers on this network to 127.0.0.1 (or loopback) so inter-container calls work
	for cid, ep := range net.Containers {
		if cid == currentContainerID {
			continue
		}
		var peerNames []string
		if ep.ContainerName != "" {
			peerNames = append(peerNames, ep.ContainerName)
		}
		for _, alias := range ep.Aliases {
			if !contains(peerNames, alias) {
				peerNames = append(peerNames, alias)
			}
		}
		if len(peerNames) > 0 {
			// In Docroid non-root shared network stack, all peer processes listen on localhost ports
			b.WriteString(fmt.Sprintf("127.0.0.1\t%s\n", strings.Join(peerNames, " ")))
		}
	}

	return b.String(), nil
}

// GenerateLinkEnvs generates Docker link / Compose environment variables for peer containers
func (m *NetworkManager) GenerateLinkEnvs(networkName string, currentContainerID string) (map[string]string, error) {
	if networkName == "" {
		networkName = "bridge"
	}

	envs := make(map[string]string)
	net, err := m.store.Get(networkName)
	if err != nil {
		return envs, err
	}

	envs["DOCROID_NETWORK"] = net.Name
	envs["DOCROID_NETWORK_DRIVER"] = string(net.Driver)
	envs["DOCROID_NETWORK_SUBNET"] = net.Subnet

	for cid, ep := range net.Containers {
		if cid == currentContainerID {
			continue
		}

		namesToExpose := []string{ep.ContainerName}
		namesToExpose = append(namesToExpose, ep.Aliases...)

		for _, name := range namesToExpose {
			if name == "" {
				continue
			}
			sanitized := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(name, "-", "_"), ".", "_"))
			envs[sanitized+"_HOST"] = "127.0.0.1"
			if ep.Port > 0 {
				envs[sanitized+"_PORT"] = fmt.Sprintf("%d", ep.Port)
				envs[sanitized+"_PORT_"+fmt.Sprintf("%d", ep.Port)+"_TCP"] = fmt.Sprintf("tcp://127.0.0.1:%d", ep.Port)
				envs[sanitized+"_PORT_"+fmt.Sprintf("%d", ep.Port)+"_TCP_ADDR"] = "127.0.0.1"
				envs[sanitized+"_PORT_"+fmt.Sprintf("%d", ep.Port)+"_TCP_PORT"] = fmt.Sprintf("%d", ep.Port)
				envs[sanitized+"_PORT_"+fmt.Sprintf("%d", ep.Port)+"_TCP_PROTO"] = "tcp"
			}
		}
	}

	return envs, nil
}

// WriteContainerHosts writes the generated /etc/hosts file into workspace directory
// and writes an auxiliary hosts file for HOSTALIASES resolution.
func (m *NetworkManager) WriteContainerHosts(wsDir string, networkName string, containerID string) (string, error) {
	hostsContent, err := m.GenerateHostsContent(networkName, containerID)
	if err != nil {
		return "", err
	}

	// 1. Write <wsDir>/etc/hosts (for PRoot mounting)
	etcDir := filepath.Join(wsDir, "etc")
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create etc dir: %w", err)
	}

	hostsPath := filepath.Join(etcDir, "hosts")
	if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write etc/hosts: %w", err)
	}

	// 2. Also write <wsDir>/hosts (for HOSTALIASES in native processes)
	auxHostsPath := filepath.Join(wsDir, "hosts")
	_ = os.WriteFile(auxHostsPath, []byte(hostsContent), 0644)

	return hostsPath, nil
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
