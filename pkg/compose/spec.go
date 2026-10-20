package compose

// ComposeConfig represents the root docroid-compose.yml structure
type ComposeConfig struct {
	Version  string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Name     string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Services map[string]ServiceSpec `yaml:"services" json:"services"`
	Networks map[string]NetworkSpec `yaml:"networks,omitempty" json:"networks,omitempty"`
}

// NetworkSpec represents an individual network definition in compose
type NetworkSpec struct {
	Driver   string            `yaml:"driver,omitempty" json:"driver,omitempty"`
	External bool              `yaml:"external,omitempty" json:"external,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ServiceSpec represents an individual service within docroid-compose.yml
type ServiceSpec struct {
	Name        string      `yaml:"name,omitempty" json:"name,omitempty"`
	Image       string      `yaml:"image,omitempty" json:"image,omitempty"`
	Engine      string      `yaml:"engine,omitempty" json:"engine,omitempty"` // "native" (default) or "rootfs"
	Rootfs      string      `yaml:"rootfs,omitempty" json:"rootfs,omitempty"` // e.g. "alpine"
	Command     interface{} `yaml:"command,omitempty" json:"command,omitempty"` // string or []interface{} / []string
	Args        []string    `yaml:"args,omitempty" json:"args,omitempty"`
	Ports       []string    `yaml:"ports,omitempty" json:"ports,omitempty"` // e.g. ["8080:8080", "9000"]
	Environment interface{} `yaml:"environment,omitempty" json:"environment,omitempty"` // map[string]interface{} or []interface{}
	Volumes     []string    `yaml:"volumes,omitempty" json:"volumes,omitempty"` // e.g. ["/sdcard/Download:/downloads"]
	Restart     string      `yaml:"restart,omitempty" json:"restart,omitempty"` // "no", "always", "on-failure", "unless-stopped"
	WorkingDir  string      `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Secrets     []string    `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Networks    interface{} `yaml:"networks,omitempty" json:"networks,omitempty"` // []string or map
	Detached    *bool       `yaml:"detached,omitempty" json:"detached,omitempty"`
}

// ResolvedService is a normalized, strongly-typed service representation ready for execution
type ResolvedService struct {
	ServiceName    string            `json:"service_name"`
	Name           string            `json:"name"`
	Image          string            `json:"image,omitempty"`
	Engine         string            `json:"engine"`
	Rootfs         string            `json:"rootfs,omitempty"`
	Command        string            `json:"command"`
	Args           []string          `json:"args"`
	PortMapping    string            `json:"port_mapping,omitempty"`
	Port           int               `json:"port,omitempty"`
	Environment    map[string]string `json:"environment"`
	Volumes        []string          `json:"volumes"`
	RestartPolicy  string            `json:"restart_policy"`
	WorkingDir     string            `json:"working_dir"`
	Secrets        []string          `json:"secrets,omitempty"`
	Network        string            `json:"network,omitempty"`
	NetworkAliases []string          `json:"network_aliases,omitempty"`
	Detached       bool              `json:"detached"`
}
