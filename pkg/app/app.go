package app

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusFailed  Status = "failed"
)

type EngineType string

const (
	EngineNative EngineType = "native"
	EngineRootfs EngineType = "rootfs"
)

// VolumeMount defines a directory binding
type VolumeMount struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
}

// App represents an application or container instance managed by Docroid
type App struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Engine        EngineType        `json:"engine"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	WorkDir       string            `json:"work_dir"`
	Port          int               `json:"port,omitempty"`
	PortMapping   string            `json:"port_mapping,omitempty"`
	Volumes       []VolumeMount     `json:"volumes,omitempty"`
	Rootfs        string            `json:"rootfs,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty"` // "no", "always", "on-failure"
	RestartCount  int               `json:"restart_count,omitempty"`
	Network       string            `json:"network,omitempty"`         // e.g. "bridge", "my-net"
	NetworkAliases []string         `json:"network_aliases,omitempty"` // e.g. ["db", "database"]

	// Runtime state
	Status    Status    `json:"status"`
	PID       int       `json:"pid"`
	ExitCode  int       `json:"exit_code"`
	CreatedAt time.Time `json:"created_at"`
	StartedAt time.Time `json:"started_at"`
}

// GenerateID returns a unique 10-character hexadecimal ID
func GenerateID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// FullCommandString returns the formatted command and arguments
func (a *App) FullCommandString() string {
	res := a.Command
	for _, arg := range a.Args {
		res += " " + arg
	}
	return res
}
