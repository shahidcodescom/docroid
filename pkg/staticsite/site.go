package staticsite

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type SiteStatus string

const (
	StatusCreated SiteStatus = "created"
	StatusRunning SiteStatus = "running"
	StatusStopped SiteStatus = "stopped"
	StatusFailed  SiteStatus = "failed"
)

// Site represents a static website hosted and supervised by Docroid
type Site struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Port             int        `json:"port"`
	Domain           string     `json:"domain,omitempty"` // e.g. "mysite.local", "example.com", or "*" / empty
	RootDir          string     `json:"root_dir"`         // absolute path to public files
	IndexDocument    string     `json:"index_document"`   // default "index.html"
	ErrorDocument    string     `json:"error_document,omitempty"`
	SPAMode          bool       `json:"spa_mode"`          // if true, 404s route to index_document
	DirectoryListing bool       `json:"directory_listing"` // allow directory browsing if no index
	Status           SiteStatus `json:"status"`
	PID              int        `json:"pid,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	StartedAt        time.Time  `json:"started_at,omitempty"`
	FileCount        int        `json:"file_count"`
	TotalSize        int64      `json:"total_size"` // in bytes
}

// SiteFile represents a single file or directory inside a site's public directory
type SiteFile struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"` // relative to rootDir
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type,omitempty"`
	ModTime   time.Time `json:"mod_time"`
	Extension string    `json:"extension,omitempty"`
}

// GenerateSiteID generates a unique 10-character hexadecimal ID
func GenerateSiteID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// URL returns the formatted access URL for this site
func (s *Site) URL(preferredHost string) string {
	domain := s.Domain
	if domain == "" || domain == "*" {
		if preferredHost != "" && preferredHost != "0.0.0.0" {
			domain = preferredHost
		} else {
			domain = "localhost"
		}
	}

	if s.Port == 80 {
		return fmt.Sprintf("http://%s", domain)
	}
	if s.Port == 443 {
		return fmt.Sprintf("https://%s", domain)
	}
	return fmt.Sprintf("http://%s:%d", domain, s.Port)
}
