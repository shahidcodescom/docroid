package staticsite

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/utils"
)

type Store struct {
	baseDir string
	mu      sync.RWMutex
}

func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		baseDir = app.GetDefaultBaseDir()
	}

	sitesDir := filepath.Join(baseDir, "sites")
	if err := os.MkdirAll(sitesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sites directory %s: %w", sitesDir, err)
	}

	return &Store{baseDir: sitesDir}, nil
}

func (s *Store) SitesDir() string {
	return s.baseDir
}

func (s *Store) SiteDir(siteID string) string {
	return filepath.Join(s.baseDir, siteID)
}

func (s *Store) PublicDir(siteID string) string {
	return filepath.Join(s.baseDir, siteID, "public")
}

func (s *Store) ConfigPath(siteID string) string {
	return filepath.Join(s.baseDir, siteID, "site.json")
}

func (s *Store) PidPath(siteID string) string {
	return filepath.Join(s.baseDir, siteID, "server.pid")
}

func (s *Store) CreateSite(name string, port int, domain string, spa bool, dirListing bool) (*Site, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		name = "site-" + GenerateSiteID()[:6]
	}

	// Sanitize name
	name = strings.ToLower(name)
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return nil, fmt.Errorf("invalid site name '%s': only alphanumeric, hyphen, underscore, and dot are allowed", name)
		}
	}

	// Check for duplicates
	existing, _ := s.findSiteUnlocked(name)
	if existing != nil {
		return nil, fmt.Errorf("a static site with name '%s' already exists", name)
	}

	if port <= 0 {
		freePort, err := utils.GetFreePort()
		if err != nil {
			port = 8080
		} else {
			port = freePort
		}
	}

	id := GenerateSiteID()
	siteDir := filepath.Join(s.baseDir, id)
	pubDir := filepath.Join(siteDir, "public")

	if err := os.MkdirAll(pubDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create site directories: %w", err)
	}

	domain = strings.TrimSpace(domain)
	if domain == "*" {
		domain = ""
	}

	site := &Site{
		ID:               id,
		Name:             name,
		Port:             port,
		Domain:           domain,
		RootDir:          pubDir,
		IndexDocument:    "index.html",
		SPAMode:          spa,
		DirectoryListing: dirListing,
		Status:           StatusCreated,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.saveSiteUnlocked(site); err != nil {
		_ = os.RemoveAll(siteDir)
		return nil, err
	}

	return site, nil
}

func (s *Store) SaveSite(site *Site) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveSiteUnlocked(site)
}

func (s *Store) saveSiteUnlocked(site *Site) error {
	site.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(site, "", "  ")
	if err != nil {
		return err
	}
	cfgFile := s.ConfigPath(site.ID)
	return os.WriteFile(cfgFile, data, 0644)
}

func (s *Store) LoadSite(id string) (*Site, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfgFile := s.ConfigPath(id)
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("site '%s' not found: %w", id, err)
	}

	var site Site
	if err := json.Unmarshal(data, &site); err != nil {
		return nil, fmt.Errorf("failed to parse site configuration: %w", err)
	}

	return &site, nil
}

func (s *Store) FindSite(idOrName string) (*Site, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findSiteUnlocked(idOrName)
}

func (s *Store) findSiteUnlocked(idOrName string) (*Site, error) {
	if s, err := s.loadSiteUnlocked(idOrName); err == nil {
		return s, nil
	}

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			site, err := s.loadSiteUnlocked(entry.Name())
			if err == nil && (site.Name == idOrName || site.ID == idOrName) {
				return site, nil
			}
		}
	}

	return nil, fmt.Errorf("site '%s' not found", idOrName)
}

func (s *Store) loadSiteUnlocked(id string) (*Site, error) {
	cfgFile := s.ConfigPath(id)
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	var site Site
	if err := json.Unmarshal(data, &site); err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *Store) ListSites() ([]*Site, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	sites := make([]*Site, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			site, err := s.loadSiteUnlocked(entry.Name())
			if err == nil {
				sites = append(sites, site)
			}
		}
	}

	sort.Slice(sites, func(i, j int) bool {
		return sites[i].CreatedAt.After(sites[j].CreatedAt)
	})

	return sites, nil
}

func (s *Store) DeleteSite(idOrName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	site, err := s.findSiteUnlocked(idOrName)
	if err != nil {
		return err
	}

	siteDir := s.SiteDir(site.ID)
	return os.RemoveAll(siteDir)
}

// ExtractZipBundle extracts a zip bundle into the site's public directory.
// Includes Zip Slip protection and smart root unwrap (e.g. if files are nested in dist/ or build/).
func (s *Store) ExtractZipBundle(siteID string, r io.ReaderAt, size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	site, err := s.loadSiteUnlocked(siteID)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(r, size)
	if err != nil {
		return fmt.Errorf("invalid zip archive: %w", err)
	}

	pubDir := s.PublicDir(site.ID)
	_ = os.RemoveAll(pubDir)
	if err := os.MkdirAll(pubDir, 0755); err != nil {
		return err
	}

	// Detect if all entries share a single common top-level directory (e.g. "dist/" or "my-site/")
	commonPrefix := detectCommonZipPrefix(zr.File)

	for _, f := range zr.File {
		cleanName := f.Name
		if commonPrefix != "" && strings.HasPrefix(cleanName, commonPrefix) {
			cleanName = strings.TrimPrefix(cleanName, commonPrefix)
		}
		cleanName = strings.TrimPrefix(cleanName, "/")
		if cleanName == "" {
			continue
		}

		targetPath := filepath.Join(pubDir, filepath.Clean(cleanName))

		// Zip Slip protection
		if !strings.HasPrefix(targetPath, filepath.Clean(pubDir)+string(filepath.Separator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	s.recalculateStatsUnlocked(site)
	return s.saveSiteUnlocked(site)
}

func detectCommonZipPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}

	var firstDir string
	allShare := true

	for _, f := range files {
		clean := strings.TrimPrefix(filepath.Clean(f.Name), "/")
		parts := strings.Split(clean, string(filepath.Separator))
		if len(parts) <= 1 && !f.FileInfo().IsDir() {
			// There's a file at root, so no common directory prefix
			return ""
		}
		topDir := parts[0]
		if firstDir == "" {
			firstDir = topDir
		} else if firstDir != topDir {
			allShare = false
			break
		}
	}

	if allShare && firstDir != "" && firstDir != "." {
		return firstDir + "/"
	}
	return ""
}

// SaveUploadedFiles saves multiple uploaded files with relative paths into the site's public directory
func (s *Store) SaveUploadedFiles(siteID string, files map[string][]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	site, err := s.loadSiteUnlocked(siteID)
	if err != nil {
		return err
	}

	pubDir := s.PublicDir(site.ID)
	_ = os.MkdirAll(pubDir, 0755)

	for relPath, content := range files {
		relPath = strings.TrimPrefix(filepath.Clean(relPath), "/")
		if relPath == "" || relPath == "." {
			continue
		}

		targetPath := filepath.Join(pubDir, relPath)
		if !strings.HasPrefix(targetPath, filepath.Clean(pubDir)+string(filepath.Separator)) {
			return fmt.Errorf("illegal relative file path: %s", relPath)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return err
		}
	}

	s.recalculateStatsUnlocked(site)
	return s.saveSiteUnlocked(site)
}

// SaveRawHTML writes or updates the index.html file
func (s *Store) SaveRawHTML(siteID string, htmlContent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	site, err := s.loadSiteUnlocked(siteID)
	if err != nil {
		return err
	}

	pubDir := s.PublicDir(site.ID)
	_ = os.MkdirAll(pubDir, 0755)

	indexPath := filepath.Join(pubDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(htmlContent), 0644); err != nil {
		return err
	}

	s.recalculateStatsUnlocked(site)
	return s.saveSiteUnlocked(site)
}

func (s *Store) recalculateStatsUnlocked(site *Site) {
	pubDir := s.PublicDir(site.ID)
	var count int
	var totalSize int64

	_ = filepath.Walk(pubDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
			totalSize += info.Size()
		}
		return nil
	})

	site.FileCount = count
	site.TotalSize = totalSize
}

// ListSiteFiles returns all files within the site or a specific subpath
func (s *Store) ListSiteFiles(siteID string, subPath string) ([]SiteFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	site, err := s.loadSiteUnlocked(siteID)
	if err != nil {
		return nil, err
	}

	pubDir := s.PublicDir(site.ID)
	targetDir := pubDir
	if subPath != "" {
		cleanSub := strings.TrimPrefix(filepath.Clean(subPath), "/")
		targetDir = filepath.Join(pubDir, cleanSub)
		if !strings.HasPrefix(targetDir, filepath.Clean(pubDir)) {
			return nil, fmt.Errorf("invalid path")
		}
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SiteFile{}, nil
		}
		return nil, err
	}

	results := make([]SiteFile, 0)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(targetDir, e.Name())
		relPath, _ := filepath.Rel(pubDir, fullPath)
		ext := filepath.Ext(e.Name())
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		results = append(results, SiteFile{
			Name:      e.Name(),
			Path:      relPath,
			IsDir:     e.IsDir(),
			Size:      info.Size(),
			MimeType:  mimeType,
			ModTime:   info.ModTime(),
			Extension: ext,
		})
	}

	// Sort directories first, then alphabetically
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDir != results[j].IsDir {
			return results[i].IsDir
		}
		return results[i].Name < results[j].Name
	})

	return results, nil
}

// ReadSiteFile reads the contents of a specific site file
func (s *Store) ReadSiteFile(siteID string, filePath string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	site, err := s.loadSiteUnlocked(siteID)
	if err != nil {
		return nil, "", err
	}

	pubDir := s.PublicDir(site.ID)
	cleanPath := strings.TrimPrefix(filepath.Clean(filePath), "/")
	fullPath := filepath.Join(pubDir, cleanPath)

	if !strings.HasPrefix(fullPath, filepath.Clean(pubDir)) {
		return nil, "", fmt.Errorf("illegal file path")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(cleanPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return data, mimeType, nil
}

// ExportSiteZip creates a zip archive of the site's public directory
func (s *Store) ExportSiteZip(siteID string, w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	site, err := s.loadSiteUnlocked(siteID)
	if err != nil {
		return err
	}

	pubDir := s.PublicDir(site.ID)
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.Walk(pubDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == pubDir {
			return nil
		}

		relPath, err := filepath.Rel(pubDir, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(writer, f)
		return err
	})
}

var _ = bytes.NewBuffer
