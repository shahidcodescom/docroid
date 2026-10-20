package cluster

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type PodStatus string

const (
	PodPending    PodStatus = "Pending"
	PodRunning    PodStatus = "Running"
	PodFailed     PodStatus = "Failed"
	PodSucceeded  PodStatus = "Succeeded"
	PodTerminated PodStatus = "Terminated"
)

// PodSpec defines the desired execution specification for a pod
type PodSpec struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace,omitempty"`
	Image         string            `json:"image,omitempty"`
	Command       string            `json:"command"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Port          int               `json:"port,omitempty"` // container port
	Engine        string            `json:"engine,omitempty"` // "native" or "rootfs"
	Rootfs        string            `json:"rootfs,omitempty"` // e.g. "alpine"
	RestartPolicy string            `json:"restart_policy,omitempty"` // "no", "always", "on-failure"
	WorkingDir    string            `json:"working_dir,omitempty"`
	Volumes       []string          `json:"volumes,omitempty"`
	NodeSelector  map[string]string `json:"node_selector,omitempty"`
}

// Pod represents a single runtime instance of a workload in the cluster
type Pod struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Deployment   string            `json:"deployment,omitempty"`
	NodeID       string            `json:"node_id"`
	NodeName     string            `json:"node_name"`
	NodeAddress  string            `json:"node_address"`
	Status       PodStatus         `json:"status"`
	Spec         PodSpec           `json:"spec"`
	HostPort     int               `json:"host_port,omitempty"`
	LocalAppID   string            `json:"local_app_id,omitempty"`
	PID          int               `json:"pid,omitempty"`
	RestartCount int               `json:"restart_count"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    time.Time         `json:"started_at,omitempty"`
	FinishedAt   time.Time         `json:"finished_at,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// GeneratePodID generates a pod ID with an optional prefix
func GeneratePodID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	hash := hex.EncodeToString(b)
	if prefix != "" {
		return fmt.Sprintf("pod-%s-%s", prefix, hash)
	}
	return fmt.Sprintf("pod-%s", hash)
}

// Endpoint returns the host:port reachable address of this pod
func (p *Pod) Endpoint() string {
	if p.HostPort <= 0 || p.NodeAddress == "" {
		return ""
	}
	// Extract host from NodeAddress
	host := p.NodeAddress
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			host = host[:i]
			break
		}
	}
	return fmt.Sprintf("%s:%d", host, p.HostPort)
}
