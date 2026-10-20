package bundle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"docroid/pkg/app"
)

const CurrentFormatVersion = "1.0"

// Manifest defines the portable specification for a Docroid package (.dpack)
type Manifest struct {
	FormatVersion string            `json:"format_version"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Architecture  string            `json:"architecture"` // "arm64", "amd64", "arm", "any"
	Engine        app.EngineType    `json:"engine"`       // "native" or "rootfs"
	Command       string            `json:"command"`      // Executable command
	Args          []string          `json:"args,omitempty"`
	Port          int               `json:"port,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	WorkDir       string            `json:"work_dir,omitempty"`
	Rootfs        string            `json:"rootfs,omitempty"`
	Description   string            `json:"description,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	Checksum      string            `json:"checksum,omitempty"`
}

// TaggedName returns name:version
func (m *Manifest) TaggedName() string {
	if m.Version == "" {
		return m.Name + ":latest"
	}
	return fmt.Sprintf("%s:%s", m.Name, m.Version)
}

// Save writes the manifest to a file
func (m *Manifest) Save(filePath string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// LoadManifest reads a manifest from a file
func LoadManifest(filePath string) (*Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	if m.FormatVersion == "" {
		m.FormatVersion = CurrentFormatVersion
	}
	if m.Env == nil {
		m.Env = make(map[string]string)
	}

	return &m, nil
}

// EnsureDir creates directory path if not existing
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

var _ = filepath.Join
