package cluster

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

type NodeRole string

const (
	RoleControlPlane NodeRole = "control-plane"
	RoleWorker       NodeRole = "worker"
)

type NodeStatus string

const (
	NodeStatusReady    NodeStatus = "Ready"
	NodeStatusNotReady NodeStatus = "NotReady"
	NodeStatusOffline  NodeStatus = "Offline"
)

// NodeCapacity describes the compute and pod capacity of a node
type NodeCapacity struct {
	CPUCores int   `json:"cpu_cores"`
	MemoryMB int64 `json:"memory_mb"`
	MaxPods  int   `json:"max_pods"`
}

// Node represents a device running Docroid interconnected in the cluster
type Node struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Role           NodeRole          `json:"role"`
	Address        string            `json:"address"` // e.g. "192.168.1.50:9090"
	Status         NodeStatus        `json:"status"`
	OS             string            `json:"os"`
	Arch           string            `json:"arch"`
	Version        string            `json:"version"`
	Labels         map[string]string `json:"labels"`
	Capacity       NodeCapacity      `json:"capacity"`
	AllocatedPods  int               `json:"allocated_pods"`
	LastHeartbeat  time.Time         `json:"last_heartbeat"`
	JoinedAt       time.Time         `json:"joined_at"`
	IsLocal        bool              `json:"is_local,omitempty"`
}

// GenerateNodeID creates a unique node identifier
func GenerateNodeID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("node-%s", hex.EncodeToString(b))
}

// DetectLocalNode creates a Node definition representing the local machine
func DetectLocalNode(role NodeRole, address string, docroidVersion string) *Node {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "docroid-node"
	}

	// Detect if running on Android/Termux
	detectedOS := runtime.GOOS
	if _, err := os.Stat("/data/data/com.termux"); err == nil {
		detectedOS = "android"
	}

	cores := runtime.NumCPU()
	if cores <= 0 {
		cores = 2
	}

	// Default address if empty
	if address == "" {
		address = GetLocalOutboundIP() + ":9090"
	}

	labels := map[string]string{
		"os":                  detectedOS,
		"arch":                runtime.GOARCH,
		"kubernetes.io/os":    detectedOS,
		"device.docroid/role": string(role),
	}
	if detectedOS == "android" {
		labels["device.docroid/type"] = "mobile"
	} else {
		labels["device.docroid/type"] = "host"
	}

	return &Node{
		ID:            GenerateNodeID(),
		Name:          hostname,
		Role:          role,
		Address:       address,
		Status:        NodeStatusReady,
		OS:            detectedOS,
		Arch:          runtime.GOARCH,
		Version:       docroidVersion,
		Labels:        labels,
		Capacity: NodeCapacity{
			CPUCores: cores,
			MemoryMB: 2048, // reasonable mobile/edge baseline
			MaxPods:  20,
		},
		AllocatedPods: 0,
		LastHeartbeat: time.Now(),
		JoinedAt:      time.Now(),
		IsLocal:       true,
	}
}

// IsAlive checks if the node has checked in recently
func (n *Node) IsAlive(timeout time.Duration) bool {
	if n.IsLocal {
		return true
	}
	return time.Since(n.LastHeartbeat) <= timeout
}

// MatchesSelector checks if the node satisfies all key-value labels in the selector
func (n *Node) MatchesSelector(selector map[string]string) bool {
	if len(selector) == 0 {
		return true
	}
	for k, v := range selector {
		if val, exists := n.Labels[k]; !exists || val != v {
			return false
		}
	}
	return true
}

// GetLocalOutboundIP returns the preferred outbound IP address of this machine
func GetLocalOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	// Fallback to interface scan
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return "127.0.0.1"
}

// CleanAddress ensures the address has a valid host and port
func CleanAddress(addr string, defaultPort int) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fmt.Sprintf("%s:%d", GetLocalOutboundIP(), defaultPort)
	}
	if strings.HasPrefix(addr, ":") {
		return fmt.Sprintf("%s%s", GetLocalOutboundIP(), addr)
	}
	if !strings.Contains(addr, ":") {
		return fmt.Sprintf("%s:%d", addr, defaultPort)
	}
	return addr
}
