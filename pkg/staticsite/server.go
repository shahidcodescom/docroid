package staticsite

import (
	"context"
	"fmt"
	"html"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"docroid/pkg/metrics"
	"docroid/pkg/utils"
)

// VHostHandler routes incoming HTTP requests to the appropriate Site based on the Host header
type VHostHandler struct {
	port  int
	store *Store
}

func NewVHostHandler(port int, store *Store) *VHostHandler {
	return &VHostHandler{
		port:  port,
		store: store,
	}
}

func (vh *VHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract and normalize hostname
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSpace(host))

	sites, err := vh.store.ListSites()
	if err != nil {
		http.Error(w, "Failed to load sites", http.StatusInternalServerError)
		return
	}

	// Filter sites configured on this port
	var portSites []*Site
	for _, s := range sites {
		if s.Port == vh.port {
			portSites = append(portSites, s)
		}
	}

	if len(portSites) == 0 {
		vh.renderNoSitesFound(w, host)
		return
	}

	// 1. Check for exact domain match
	var matchedSite *Site
	var defaultSite *Site

	for _, s := range portSites {
		siteDomain := strings.ToLower(strings.TrimSpace(s.Domain))
		if siteDomain == "" || siteDomain == "*" {
			defaultSite = s
		} else if siteDomain == host {
			matchedSite = s
			break
		}
	}

	// 2. If no exact match and only one site on port, or default site available
	if matchedSite == nil {
		if defaultSite != nil {
			matchedSite = defaultSite
		} else if len(portSites) == 1 {
			matchedSite = portSites[0]
		}
	}

	if matchedSite == nil {
		vh.renderVHostMismatch(w, host, portSites)
		return
	}

	// Serve the matched site
	handler := NewSiteHandler(matchedSite)
	handler.ServeHTTP(w, r)
}

func (vh *VHostHandler) renderNoSitesFound(w http.ResponseWriter, host string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>404 - No Site Hosted</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
    .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; max-width: 500px; text-align: center; border: 1px solid #334155; box-shadow: 0 10px 25px rgba(0,0,0,0.5); }
    h1 { color: #38bdf8; margin-top: 0; font-size: 1.8rem; }
    p { color: #94a3b8; line-height: 1.6; }
    .badge { display: inline-block; background: #334155; color: #38bdf8; padding: 4px 12px; border-radius: 9999px; font-size: 0.85rem; font-weight: 600; margin-top: 1rem; }
  </style>
</head>
<body>
  <div class="card">
    <h1><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align: -3px; margin-right: 6px;"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>Docroid Static Server</h1>
    <p>No static website is currently configured or active on port <strong>%d</strong> for host <code>%s</code>.</p>
    <div class="badge">Docroid Engine • Mobile Container</div>
  </div>
</body>
</html>`, vh.port, html.EscapeString(host))
}

func (vh *VHostHandler) renderVHostMismatch(w http.ResponseWriter, host string, sites []*Site) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	var listItems strings.Builder
	for _, s := range sites {
		dom := s.Domain
		if dom == "" {
			dom = "(wildcard / default)"
		}
		listItems.WriteString(fmt.Sprintf("<li><strong>%s</strong> &mdash; Domain: <code>%s</code></li>", html.EscapeString(s.Name), html.EscapeString(dom)))
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Virtual Host Mismatch</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; }
    .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; max-width: 540px; text-align: left; border: 1px solid #334155; box-shadow: 0 10px 25px rgba(0,0,0,0.5); }
    h1 { color: #f59e0b; margin-top: 0; font-size: 1.6rem; display: flex; align-items: center; gap: 8px; }
    p { color: #94a3b8; line-height: 1.6; }
    ul { background: #0f172a; border-radius: 8px; padding: 1rem 1.5rem; list-style-type: square; color: #cbd5e1; }
    li { margin-bottom: 0.5rem; }
    code { background: #334155; padding: 2px 6px; border-radius: 4px; color: #38bdf8; font-family: monospace; }
  </style>
</head>
<body>
  <div class="card">
    <h1><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align: -3px; margin-right: 6px;"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>Virtual Host Routing</h1>
    <p>Incoming request host <code>%s</code> does not match any configured domain on port <strong>%d</strong>.</p>
    <p>Available sites on this port:</p>
    <ul>%s</ul>
    <p style="font-size:0.85rem; color:#64748b; margin-bottom:0;">Configure local DNS or /etc/hosts to direct your domain to this server, or set domain to <code>*</code> in Docroid.</p>
  </div>
</body>
</html>`, html.EscapeString(host), vh.port, listItems.String())
}

// SiteHandler serves static files for a specific site with SPA fallback and directory listing
type SiteHandler struct {
	site *Site
}

func NewSiteHandler(site *Site) *SiteHandler {
	return &SiteHandler{site: site}
}

func (sh *SiteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Canonical path cleaning to prevent path traversal attacks
	cleanPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	rootDir := filepath.Clean(sh.site.RootDir)
	targetPath := filepath.Join(rootDir, cleanPath)

	// Directory traversal guard
	if !strings.HasPrefix(targetPath, rootDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(targetPath)
	if err == nil {
		if info.IsDir() {
			// Check if index document exists
			indexDoc := sh.site.IndexDocument
			if indexDoc == "" {
				indexDoc = "index.html"
			}
			indexPath := filepath.Join(targetPath, indexDoc)
			if idxInfo, err := os.Stat(indexPath); err == nil && !idxInfo.IsDir() {
				sh.serveFile(w, r, indexPath)
				return
			}

			// No index file found
			if sh.site.DirectoryListing {
				sh.renderDirectoryListing(w, r, targetPath, r.URL.Path)
				return
			}

			// No directory listing allowed
			http.Error(w, "403 Forbidden - Directory listing disabled", http.StatusForbidden)
			return
		}

		// It's a file, serve it directly
		sh.serveFile(w, r, targetPath)
		return
	}

	// File not found (os.IsNotExist)
	// Check SPA mode: for client-side routing, route non-existent paths to index.html
	if sh.site.SPAMode {
		indexDoc := sh.site.IndexDocument
		if indexDoc == "" {
			indexDoc = "index.html"
		}
		spaIndexPath := filepath.Join(rootDir, indexDoc)
		if spaInfo, err := os.Stat(spaIndexPath); err == nil && !spaInfo.IsDir() {
			sh.serveFile(w, r, spaIndexPath)
			return
		}
	}

	// Check for custom error document e.g. 404.html
	if sh.site.ErrorDocument != "" {
		errPath := filepath.Join(rootDir, sh.site.ErrorDocument)
		if errInfo, err := os.Stat(errPath); err == nil && !errInfo.IsDir() {
			w.WriteHeader(http.StatusNotFound)
			sh.serveFile(w, r, errPath)
			return
		}
	}

	// Default 404 response
	sh.renderDefault404(w, r.URL.Path)
}

func (sh *SiteHandler) serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		switch strings.ToLower(ext) {
		case ".js", ".mjs":
			mimeType = "application/javascript"
		case ".css":
			mimeType = "text/css; charset=utf-8"
		case ".html", ".htm":
			mimeType = "text/html; charset=utf-8"
		case ".json":
			mimeType = "application/json"
		case ".svg":
			mimeType = "image/svg+xml"
		case ".wasm":
			mimeType = "application/wasm"
		case ".webp":
			mimeType = "image/webp"
		case ".woff2":
			mimeType = "font/woff2"
		case ".woff":
			mimeType = "font/woff"
		default:
			mimeType = "application/octet-stream"
		}
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Server", "Docroid-Static/1.0")

	// Check if this is an asset that can be cached (CSS, JS, Images, Fonts)
	if ext != ".html" && ext != ".htm" && ext != "" {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	} else {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	}

	http.ServeFile(w, r, filePath)
}

func (sh *SiteHandler) renderDefault404(w http.ResponseWriter, reqPath string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>404 Not Found</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
    .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; max-width: 480px; text-align: center; border: 1px solid #334155; }
    h1 { color: #f43f5e; margin: 0 0 1rem; font-size: 2.2rem; }
    p { color: #94a3b8; line-height: 1.5; }
    code { background: #334155; padding: 2px 6px; border-radius: 4px; color: #38bdf8; }
  </style>
</head>
<body>
  <div class="card">
    <h1>404 Not Found</h1>
    <p>The requested file <code>%s</code> was not found on this website.</p>
    <p style="font-size:0.85rem; color:#64748b;">Site: %s</p>
  </div>
</body>
</html>`, html.EscapeString(reqPath), html.EscapeString(sh.site.Name))
}

func (sh *SiteHandler) renderDirectoryListing(w http.ResponseWriter, r *http.Request, dirPath string, urlPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var rows strings.Builder
	cleanURL := strings.TrimSuffix(urlPath, "/")
	if cleanURL != "" {
		parentURL := filepath.Dir(cleanURL)
		if parentURL == "." {
			parentURL = "/"
		}
		rows.WriteString(fmt.Sprintf(`<tr><td><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-2px; margin-right:6px; color:#38bdf8;"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg><a href="%s">.. (Parent Directory)</a></td><td>-</td><td>-</td></tr>`, parentURL))
	}

	for _, e := range entries {
		name := e.Name()
		iconSvg := `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-2px; margin-right:6px; color:#94a3b8;"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>`
		if e.IsDir() {
			iconSvg = `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-2px; margin-right:6px; color:#38bdf8;"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>`
			name += "/"
		} else if strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm") {
			iconSvg = `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-2px; margin-right:6px; color:#38bdf8;"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>`
		} else if strings.HasSuffix(name, ".css") || strings.HasSuffix(name, ".js") {
			iconSvg = `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-2px; margin-right:6px; color:#fbbf24;"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>`
		} else if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".svg") {
			iconSvg = `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-2px; margin-right:6px; color:#a855f7;"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>`
		}

		info, _ := e.Info()
		sizeStr := "-"
		modStr := "-"
		if info != nil {
			if !e.IsDir() {
				sizeStr = utils.FormatBytes(info.Size())
			}
			modStr = info.ModTime().Format("2006-01-02 15:04:05")
		}

		itemURL := cleanURL + "/" + e.Name()
		rows.WriteString(fmt.Sprintf(`<tr>
  <td>%s<a href="%s">%s</a></td>
  <td style="color:#94a3b8; font-family:monospace;">%s</td>
  <td style="color:#64748b; font-size:0.85rem;">%s</td>
</tr>`, iconSvg, itemURL, html.EscapeString(name), sizeStr, modStr))
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Index of %s</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; padding: 2rem; margin: 0; }
    .container { max-width: 900px; margin: 0 auto; background: #1e293b; padding: 2rem; border-radius: 12px; border: 1px solid #334155; }
    h1 { font-size: 1.4rem; color: #38bdf8; margin-top: 0; display: flex; align-items: center; gap: 8px; }
    table { width: 100%%; border-collapse: collapse; margin-top: 1.5rem; }
    th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #334155; }
    th { color: #94a3b8; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; }
    a { color: #38bdf8; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .footer { margin-top: 1.5rem; font-size: 0.8rem; color: #64748b; text-align: right; }
  </style>
</head>
<body>
  <div class="container">
    <h1><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>Index of %s</h1>
    <table>
      <thead>
        <tr><th>Name</th><th>Size</th><th>Last Modified</th></tr>
      </thead>
      <tbody>
        %s
      </tbody>
    </table>
    <div class="footer">Docroid Static Web Engine</div>
  </div>
</body>
</html>`, html.EscapeString(urlPath), html.EscapeString(urlPath), rows.String())
}

// Active listener registry to prevent duplicate in-process listeners
var (
	activeServersMu sync.Mutex
	activeServers   = make(map[int]*http.Server)
)

// StartInProcessServer starts or registers a port listener within the running process (e.g. inside docroid ui)
func StartInProcessServer(port int, store *Store) error {
	activeServersMu.Lock()
	defer activeServersMu.Unlock()

	if _, exists := activeServers[port]; exists {
		return nil // Already running and dynamically handling sites on this port
	}

	handler := NewVHostHandler(port, store)
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: handler,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("failed to bind port %d: %w", port, err)
	}

	activeServers[port] = srv

	go func() {
		_ = srv.Serve(ln)
		activeServersMu.Lock()
		delete(activeServers, port)
		activeServersMu.Unlock()
	}()

	return nil
}

// StopInProcessServer shuts down an active in-process listener on a port
func StopInProcessServer(port int) error {
	return StopInProcessServerForSite(port, nil, "")
}

// StopInProcessServerForSite shuts down an active in-process listener on a port if no other running site requires it
func StopInProcessServerForSite(port int, store *Store, excludeSiteID string) error {
	activeServersMu.Lock()
	defer activeServersMu.Unlock()

	if store != nil {
		if sites, err := store.ListSites(); err == nil {
			for _, s := range sites {
				if s.ID != excludeSiteID && s.Port == port && s.Status == StatusRunning {
					// Another site is still running on this port, keep listener active
					return nil
				}
			}
		}
	}

	srv, exists := activeServers[port]
	delete(activeServers, port)

	if exists && srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
	return nil
}

// StartSiteProcess spawns a background detached static server process
func StartSiteProcess(siteID string, store *Store) error {
	site, err := store.LoadSite(siteID)
	if err != nil {
		return err
	}

	// Check if already running
	if IsSiteRunning(site, store) {
		return nil
	}

	// Verify port availability
	if err := utils.CheckPortInUse(site.Port); err != nil {
		// If another site is on this port, check if it belongs to Docroid static vhost
		sites, _ := store.ListSites()
		otherSiteRunning := false
		for _, s := range sites {
			if s.ID != site.ID && s.Port == site.Port && IsSiteRunning(s, store) {
				otherSiteRunning = true
				break
			}
		}
		if !otherSiteRunning {
			return fmt.Errorf("port %d is already in use by another application: %w", site.Port, err)
		}
	}

	selfPath, err := os.Executable()
	if err != nil {
		selfPath = os.Args[0]
	}

	cmd := exec.Command(selfPath, "_static-serve", "--id", site.ID)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		site.Status = StatusFailed
		_ = store.SaveSite(site)
		return fmt.Errorf("failed to start static server process: %w", err)
	}

	pid := cmd.Process.Pid
	site.PID = pid
	site.Status = StatusRunning
	site.StartedAt = time.Now()
	_ = store.SaveSite(site)

	// Write PID file
	_ = os.WriteFile(store.PidPath(site.ID), []byte(strconv.Itoa(pid)), 0644)

	go func() {
		_ = cmd.Process.Release()
	}()

	return nil
}

// StopSiteProcess terminates the running server process for a site
func StopSiteProcess(siteID string, store *Store) error {
	site, err := store.LoadSite(siteID)
	if err != nil {
		return err
	}

	pid := site.PID
	if pid <= 0 {
		if data, err := os.ReadFile(store.PidPath(site.ID)); err == nil {
			pid, _ = strconv.Atoi(strings.TrimSpace(string(data)))
		}
	}

	// CRITICAL GUARD: Never kill or signal our own process (e.g. if site ran in-process inside Web UI)
	if pid > 0 && pid != os.Getpid() {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Signal(syscall.SIGTERM)
			_ = syscall.Kill(-pid, syscall.SIGTERM)

			// Wait up to 2 seconds for graceful exit
			stopped := false
			for i := 0; i < 20; i++ {
				time.Sleep(100 * time.Millisecond)
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					stopped = true
					break
				}
			}
			if !stopped {
				_ = proc.Kill()
				_ = syscall.Kill(-pid, syscall.SIGKILL)
			}
		}
	}

	_ = os.Remove(store.PidPath(site.ID))
	site.Status = StatusStopped
	site.PID = 0
	return store.SaveSite(site)
}

// IsSiteRunning checks whether the site server process is alive
func IsSiteRunning(site *Site, store *Store) bool {
	pid := site.PID
	if pid <= 0 {
		if data, err := os.ReadFile(store.PidPath(site.ID)); err == nil {
			pid, _ = strconv.Atoi(strings.TrimSpace(string(data)))
		}
	}

	if pid <= 0 {
		if site.Status == StatusRunning {
			site.Status = StatusStopped
			_ = store.SaveSite(site)
		}
		return false
	}

	// If marked with the current process PID (in-process web server)
	if pid == os.Getpid() {
		activeServersMu.Lock()
		_, inProcess := activeServers[site.Port]
		activeServersMu.Unlock()
		if inProcess && site.Status == StatusRunning {
			return true
		}
		site.Status = StatusStopped
		site.PID = 0
		_ = store.SaveSite(site)
		_ = os.Remove(store.PidPath(site.ID))
		return false
	}

	st, err := metrics.GetProcessStats(pid)
	if err == nil && st.Alive {
		return true
	}

	// Process dead, update metadata
	if site.Status == StatusRunning {
		site.Status = StatusStopped
		site.PID = 0
		_ = store.SaveSite(site)
		_ = os.Remove(store.PidPath(site.ID))
	}
	return false
}

// ServeStandalone runs the blocking HTTP server loop for a static site
func ServeStandalone(siteID string, store *Store) error {
	site, err := store.LoadSite(siteID)
	if err != nil {
		return err
	}

	pid := os.Getpid()
	_ = os.WriteFile(store.PidPath(site.ID), []byte(strconv.Itoa(pid)), 0644)
	site.PID = pid
	site.Status = StatusRunning
	site.StartedAt = time.Now()
	_ = store.SaveSite(site)

	defer func() {
		_ = os.Remove(store.PidPath(site.ID))
		site.Status = StatusStopped
		site.PID = 0
		_ = store.SaveSite(site)
	}()

	vh := NewVHostHandler(site.Port, store)
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", site.Port),
		Handler: vh,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		site.Status = StatusFailed
		_ = store.SaveSite(site)
		return fmt.Errorf("failed to bind port %d: %w", site.Port, err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Serve(ln)
	}()

	select {
	case <-sigChan:
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	case err := <-errChan:
		return err
	}
}
