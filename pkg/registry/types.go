package registry

// OCI and Docker Registry V2 Media Types
const (
	MediaTypeDockerV2ManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	MediaTypeDockerV2Manifest     = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeDockerV2Config       = "application/vnd.docker.container.image.v1+json"
	MediaTypeDockerV2Layer        = "application/vnd.docker.image.rootfs.diff.tar.gzip"

	MediaTypeOCIV1Index    = "application/vnd.oci.image.index.v1+json"
	MediaTypeOCIV1Manifest = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeOCIV1Config   = "application/vnd.oci.image.config.v1+json"
	MediaTypeOCIV1Layer    = "application/vnd.oci.image.layer.v1.tar+gzip"
)

// Descriptor represents an OCI/Docker content descriptor
type Descriptor struct {
	MediaType string    `json:"mediaType"`
	Size      int64     `json:"size"`
	Digest    string    `json:"digest"`
	Platform  *Platform `json:"platform,omitempty"`
}

// Platform represents target OS and architecture
type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

// IndexManifest represents a multi-architecture manifest list or OCI index
type IndexManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType,omitempty"`
	Manifests     []Descriptor `json:"manifests"`
}

// ImageManifest represents a single platform image manifest
type ImageManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType,omitempty"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
}

// ImageConfig represents the runtime container configuration JSON blob
type ImageConfig struct {
	Architecture string          `json:"architecture"`
	OS           string          `json:"os"`
	Config       ContainerConfig `json:"config"`
}

// ContainerConfig holds environment variables, commands, and runtime options
type ContainerConfig struct {
	User         string                 `json:"User,omitempty"`
	ExposedPorts map[string]struct{}    `json:"ExposedPorts,omitempty"`
	Env          []string               `json:"Env,omitempty"`
	Entrypoint   []string               `json:"Entrypoint,omitempty"`
	Cmd          []string               `json:"Cmd,omitempty"`
	WorkingDir   string                 `json:"WorkingDir,omitempty"`
	Labels       map[string]string      `json:"Labels,omitempty"`
}

// PullResult represents summary information of a successfully pulled image
type PullResult struct {
	Name         string            `json:"name"`
	Tag          string            `json:"tag"`
	Architecture string            `json:"architecture"`
	Engine       string            `json:"engine"`
	Command      string            `json:"command"`
	Args         []string          `json:"args,omitempty"`
	Port         int               `json:"port,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	WorkDir      string            `json:"work_dir,omitempty"`
	Digest       string            `json:"digest,omitempty"`
}

