package tunnel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TunnelMode string

const (
	ModeQuick TunnelMode = "quick" // Ephemeral trycloudflare.com tunnel (no account required)
	ModeToken TunnelMode = "token" // Cloudflare Zero Trust named tunnel (via token)
)

type TunnelStatus string

const (
	StatusStarting TunnelStatus = "starting"
	StatusRunning  TunnelStatus = "running"
	StatusStopped  TunnelStatus = "stopped"
	StatusError    TunnelStatus = "error"
)

// Tunnel holds the runtime and persistent state for a Cloudflare Tunnel
type Tunnel struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AppID       string       `json:"app_id,omitempty"`
	AppName     string       `json:"app_name,omitempty"`
	Target      string       `json:"target"` // normalized target e.g. "http://127.0.0.1:8080"
	TargetPort  int          `json:"target_port"`
	Mode        TunnelMode   `json:"mode"`
	Token       string       `json:"token,omitempty"`
	Hostname    string       `json:"hostname,omitempty"`
	PublicURL   string       `json:"public_url,omitempty"`
	Status      TunnelStatus `json:"status"`
	PID         int          `json:"pid,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	StartedAt   time.Time    `json:"started_at,omitempty"`
	StoppedAt   time.Time    `json:"stopped_at,omitempty"`
	LastError   string       `json:"last_error,omitempty"`
	LogFile     string       `json:"log_file,omitempty"`
	AutoRestart bool         `json:"auto_restart"`
}

// TunnelOptions defines creation options for a tunnel
type TunnelOptions struct {
	Name        string     `json:"name"`
	AppID       string     `json:"app_id,omitempty"`
	AppName     string     `json:"app_name,omitempty"`
	Target      string     `json:"target"`
	TargetPort  int        `json:"target_port"`
	Mode        TunnelMode `json:"mode"`
	Token       string     `json:"token,omitempty"`
	Hostname    string     `json:"hostname,omitempty"`
	AutoRestart bool       `json:"auto_restart"`
}

// GenerateTunnelID generates a unique 10-character hex identifier
func GenerateTunnelID() string {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())[:10]
	}
	return hex.EncodeToString(bytes)
}

// NormalizeTarget parses and normalizes target specifications (e.g. "8080", ":8080", "127.0.0.1:8080", "http://localhost:3000")
func NormalizeTarget(target string, port int) (string, int, error) {
	target = strings.TrimSpace(target)

	// If target is empty, use port
	if target == "" {
		if port <= 0 || port > 65535 {
			return "", 0, fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
		}
		return fmt.Sprintf("http://127.0.0.1:%d", port), port, nil
	}

	// If target is just digits e.g. "8080"
	if p, err := strconv.Atoi(target); err == nil {
		if p <= 0 || p > 65535 {
			return "", 0, fmt.Errorf("invalid port %d: must be between 1 and 65535", p)
		}
		return fmt.Sprintf("http://127.0.0.1:%d", p), p, nil
	}

	// If target starts with colon e.g. ":8080"
	if strings.HasPrefix(target, ":") {
		p, err := strconv.Atoi(strings.TrimPrefix(target, ":"))
		if err != nil || p <= 0 || p > 65535 {
			return "", 0, fmt.Errorf("invalid port format '%s'", target)
		}
		return fmt.Sprintf("http://127.0.0.1:%d", p), p, nil
	}

	// If target does not specify protocol, assume http://
	rawTarget := target
	if !strings.Contains(target, "://") {
		rawTarget = "http://" + target
	}

	parsed, err := url.Parse(rawTarget)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse target URL '%s': %w", target, err)
	}

	var resolvedPort int
	if parsed.Port() != "" {
		p, err := strconv.Atoi(parsed.Port())
		if err != nil || p <= 0 || p > 65535 {
			return "", 0, fmt.Errorf("invalid port in target '%s'", target)
		}
		resolvedPort = p
	} else if port > 0 && port <= 65535 {
		resolvedPort = port
	} else if strings.Contains(target, "://") {
		// Only infer default scheme port if user explicitly supplied full scheme URL
		switch strings.ToLower(parsed.Scheme) {
		case "https":
			resolvedPort = 443
		case "http":
			resolvedPort = 80
		default:
			return "", 0, fmt.Errorf("target '%s' must include a valid port number", target)
		}
	} else {
		return "", 0, fmt.Errorf("target '%s' must include a port number (e.g. 8080 or localhost:8080)", target)
	}

	// Normalize hostname
	host := parsed.Hostname()
	if host == "localhost" || host == "0.0.0.0" || host == "" {
		host = "127.0.0.1"
	}

	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "http"
	}

	normalized := fmt.Sprintf("%s://%s:%d", scheme, host, resolvedPort)
	if parsed.Path != "" && parsed.Path != "/" {
		normalized += parsed.Path
	}

	return normalized, resolvedPort, nil
}

var (
	tryCloudflareRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
)

// ExtractPublicURL searches for a trycloudflare.com URL in cloudflared log text
func ExtractPublicURL(content string) string {
	matches := tryCloudflareRegex.FindAllString(content, -1)
	if len(matches) > 0 {
		return matches[len(matches)-1]
	}
	return ""
}
