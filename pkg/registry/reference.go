package registry

import (
	"fmt"
	"strings"
)

// Reference represents a parsed Docker/OCI image reference
type Reference struct {
	Registry   string // e.g. "registry-1.docker.io", "ghcr.io"
	Repository string // e.g. "library/alpine", "myuser/myimage"
	Tag        string // e.g. "latest", "3.19"
	Digest     string // e.g. "sha256:abcd..."
}

// Scheme returns "http" for localhost/insecure registries, otherwise "https"
func (r *Reference) Scheme() string {
	if strings.HasPrefix(r.Registry, "localhost") || strings.HasPrefix(r.Registry, "127.0.0.1") || strings.Contains(r.Registry, ":5000") {
		return "http"
	}
	return "https"
}

// BaseURL returns the protocol and registry host e.g. "https://registry-1.docker.io"
func (r *Reference) BaseURL() string {
	return fmt.Sprintf("%s://%s", r.Scheme(), r.Registry)
}

// ParseReference parses an image name like "alpine", "ubuntu:22.04", or "ghcr.io/foo/bar:v1"
func ParseReference(imageName string) (*Reference, error) {
	if strings.TrimSpace(imageName) == "" {
		return nil, fmt.Errorf("empty image name")
	}

	tag := "latest"
	digest := ""
	repoWithRegistry := imageName

	// Check if there is a digest (@sha256:...)
	if atIdx := strings.Index(imageName, "@"); atIdx != -1 {
		digest = imageName[atIdx+1:]
		repoWithRegistry = imageName[:atIdx]
		tag = ""
	} else {
		// Check if there's a tag
		colonIdx := strings.LastIndex(imageName, ":")
		slashIdx := strings.LastIndex(imageName, "/")
		if colonIdx > slashIdx && colonIdx != -1 {
			tag = imageName[colonIdx+1:]
			repoWithRegistry = imageName[:colonIdx]
		}
	}

	parts := strings.Split(repoWithRegistry, "/")
	var registry, repo string

	// Determine if the first part is a domain name
	if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") || parts[0] == "localhost") {
		registry = parts[0]
		repo = strings.Join(parts[1:], "/")
	} else {
		// Default to Docker Hub
		registry = "registry-1.docker.io"
		if len(parts) == 1 {
			repo = "library/" + parts[0]
		} else {
			repo = strings.Join(parts, "/")
		}
	}

	return &Reference{
		Registry:   registry,
		Repository: repo,
		Tag:        tag,
		Digest:     digest,
	}, nil
}

// FullName returns the canonical repository name e.g. "library/alpine" or "ghcr.io/foo/bar"
func (r *Reference) FullName() string {
	if r.Registry == "registry-1.docker.io" {
		if strings.HasPrefix(r.Repository, "library/") {
			return strings.TrimPrefix(r.Repository, "library/")
		}
		return r.Repository
	}
	return fmt.Sprintf("%s/%s", r.Registry, r.Repository)
}

// TaggedName returns the canonical name with tag e.g. "alpine:latest"
func (r *Reference) TaggedName() string {
	if r.Digest != "" {
		return fmt.Sprintf("%s@%s", r.FullName(), r.Digest)
	}
	tag := r.Tag
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", r.FullName(), tag)
}
