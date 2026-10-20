package network

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

// DriverType defines the network driver
type DriverType string

const (
	DriverBridge DriverType = "bridge"
	DriverHost   DriverType = "host"
)

// Endpoint represents a container connected to a network
type Endpoint struct {
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	IPv4Address   string    `json:"ipv4_address"`
	Port          int       `json:"port,omitempty"`
	Aliases       []string  `json:"aliases,omitempty"`
	ConnectedAt   time.Time `json:"connected_at"`
}

// Network represents a virtual container network (Docker-like)
type Network struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Driver     DriverType           `json:"driver"`
	Scope      string               `json:"scope"`
	Internal   bool                 `json:"internal"`
	Subnet     string               `json:"subnet"`
	Gateway    string               `json:"gateway"`
	Containers map[string]*Endpoint `json:"containers"`
	Labels     map[string]string    `json:"labels,omitempty"`
	CreatedAt  time.Time            `json:"created_at"`
}

// GenerateNetworkID generates a 12-character hex ID (similar to Docker network ID)
func GenerateNetworkID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NextAvailableIP allocates the next available IP address in the network subnet
func (n *Network) NextAvailableIP() (string, error) {
	if n.Driver == DriverHost {
		return "127.0.0.1", nil
	}

	subnetStr := n.Subnet
	if subnetStr == "" {
		subnetStr = "10.89.0.0/16"
	}

	_, ipNet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return "", fmt.Errorf("invalid subnet '%s': %w", subnetStr, err)
	}

	usedIPs := make(map[string]bool)
	if n.Gateway != "" {
		usedIPs[n.Gateway] = true
	}
	for _, ep := range n.Containers {
		if ep.IPv4Address != "" {
			usedIPs[ep.IPv4Address] = true
		}
	}

	// Start scanning from base IP + 2 (since .1 is gateway)
	currIP := make(net.IP, len(ipNet.IP))
	copy(currIP, ipNet.IP)

	// Increment
	incrementIP(currIP) // .1 (gateway)
	for {
		incrementIP(currIP)
		if !ipNet.Contains(currIP) {
			return "", fmt.Errorf("no more IP addresses available in subnet %s", subnetStr)
		}
		ipStr := currIP.String()
		if !usedIPs[ipStr] {
			return ipStr, nil
		}
	}
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// CleanName sanitizes network names
func CleanName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	return name
}
