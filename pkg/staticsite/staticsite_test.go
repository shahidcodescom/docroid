package staticsite

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func createTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "docroid_staticsite_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	return store, cleanup
}

func TestCreateAndListSite(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("my-blog", 8080, "blog.local", true, false)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}

	if site.Name != "my-blog" || site.Port != 8080 || site.Domain != "blog.local" {
		t.Errorf("unexpected site values: %+v", site)
	}
	if !site.SPAMode {
		t.Errorf("expected SPAMode=true")
	}

	loaded, err := store.FindSite("my-blog")
	if err != nil {
		t.Fatalf("failed to find site: %v", err)
	}
	if loaded.ID != site.ID {
		t.Errorf("expected id %s, got %s", site.ID, loaded.ID)
	}

	sites, err := store.ListSites()
	if err != nil || len(sites) != 1 {
		t.Fatalf("expected 1 site, got %d, err: %v", len(sites), err)
	}
}

func TestSaveRawHTML(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("quick-site", 8081, "", false, false)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}

	html := "<h1>Welcome to Docroid Static Site</h1>"
	if err := store.SaveRawHTML(site.ID, html); err != nil {
		t.Fatalf("failed to save raw HTML: %v", err)
	}

	files, err := store.ListSiteFiles(site.ID, "")
	if err != nil || len(files) != 1 {
		t.Fatalf("expected 1 file, got %d, err: %v", len(files), err)
	}
	if files[0].Name != "index.html" {
		t.Errorf("expected index.html, got %s", files[0].Name)
	}

	data, mimeType, err := store.ReadSiteFile(site.ID, "index.html")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != html {
		t.Errorf("content mismatch: got %s", string(data))
	}
	if !strings.Contains(mimeType, "text/html") {
		t.Errorf("expected text/html mime, got %s", mimeType)
	}
}

func TestSaveUploadedFiles(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("multi-file-site", 8082, "test.local", false, false)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}

	uploaded := map[string][]byte{
		"index.html":       []byte("<html><body>Hello</body></html>"),
		"css/style.css":    []byte("body { background: #000; }"),
		"js/main.js":       []byte("console.log('hi');"),
		"assets/logo.svg":  []byte("<svg></svg>"),
	}

	if err := store.SaveUploadedFiles(site.ID, uploaded); err != nil {
		t.Fatalf("failed to save uploaded files: %v", err)
	}

	updatedSite, err := store.FindSite("multi-file-site")
	if err != nil {
		t.Fatalf("failed to load site: %v", err)
	}
	if updatedSite.FileCount != 4 {
		t.Errorf("expected 4 files, got %d", updatedSite.FileCount)
	}
}

func TestExtractZipBundle(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("zip-site", 8083, "", false, false)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}

	// Create test zip archive in memory (with nested "dist/" folder to test smart root unwrapping)
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	f1, _ := zw.Create("dist/index.html")
	_, _ = f1.Write([]byte("<h1>From Dist</h1>"))

	f2, _ := zw.Create("dist/assets/app.js")
	_, _ = f2.Write([]byte("console.log('from dist');"))

	_ = zw.Close()

	reader := bytes.NewReader(zipBuf.Bytes())
	if err := store.ExtractZipBundle(site.ID, reader, int64(zipBuf.Len())); err != nil {
		t.Fatalf("failed to extract zip: %v", err)
	}

	// Because of smart prefix detection, index.html should be directly at root
	files, err := store.ListSiteFiles(site.ID, "")
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}

	hasIndex := false
	for _, f := range files {
		if f.Name == "index.html" {
			hasIndex = true
		}
	}
	if !hasIndex {
		t.Errorf("expected unwrapped index.html at root, got: %+v", files)
	}
}

func TestExtractZipSlipPrevention(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("slip-site", 8084, "", false, false)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}

	// Create malicious zip with ../ traversal
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, _ := zw.Create("../../../malicious.txt")
	_, _ = f.Write([]byte("evil"))
	_ = zw.Close()

	reader := bytes.NewReader(zipBuf.Bytes())
	err = store.ExtractZipBundle(site.ID, reader, int64(zipBuf.Len()))
	if err == nil {
		t.Fatalf("expected error for Zip Slip attempt, but got nil")
	}
}

func TestVHostDomainRouting(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	// Setup two sites on the SAME port: 8085 with different domains
	siteA, err := store.CreateSite("site-a", 8085, "alpha.example.com", false, false)
	if err != nil {
		t.Fatalf("failed siteA: %v", err)
	}
	_ = store.SaveRawHTML(siteA.ID, "<h1>Site Alpha</h1>")

	siteB, err := store.CreateSite("site-b", 8085, "beta.example.com", false, false)
	if err != nil {
		t.Fatalf("failed siteB: %v", err)
	}
	_ = store.SaveRawHTML(siteB.ID, "<h1>Site Beta</h1>")

	vh := NewVHostHandler(8085, store)

	// Test request for alpha.example.com
	reqA := httptest.NewRequest(http.MethodGet, "http://alpha.example.com:8085/", nil)
	recA := httptest.NewRecorder()
	vh.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recA.Code)
	}
	if !strings.Contains(recA.Body.String(), "Site Alpha") {
		t.Errorf("expected Site Alpha, got: %s", recA.Body.String())
	}

	// Test request for beta.example.com
	reqB := httptest.NewRequest(http.MethodGet, "http://beta.example.com:8085/", nil)
	recB := httptest.NewRecorder()
	vh.ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recB.Code)
	}
	if !strings.Contains(recB.Body.String(), "Site Beta") {
		t.Errorf("expected Site Beta, got: %s", recB.Body.String())
	}

	// Test request for unknown domain on this port
	reqUnknown := httptest.NewRequest(http.MethodGet, "http://unknown.example.com:8085/", nil)
	recUnknown := httptest.NewRecorder()
	vh.ServeHTTP(recUnknown, reqUnknown)
	if recUnknown.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown vhost, got %d", recUnknown.Code)
	}
	if !strings.Contains(recUnknown.Body.String(), "Virtual Host Routing") {
		t.Errorf("expected vhost mismatch explanation page, got: %s", recUnknown.Body.String())
	}
}

func TestSPAModeRouting(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("spa-app", 8086, "", true, false)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	_ = store.SaveRawHTML(site.ID, "<div id='root'>SPA Root</div>")

	vh := NewVHostHandler(8086, store)

	// Direct route to deep URL (not an actual file)
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8086/users/profile/settings", nil)
	rec := httptest.NewRecorder()
	vh.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for SPA route, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "SPA Root") {
		t.Errorf("expected SPA index content, got: %s", rec.Body.String())
	}
}

func TestExportSiteZip(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	site, err := store.CreateSite("export-test", 8087, "", false, false)
	if err != nil {
		t.Fatalf("failed site: %v", err)
	}
	_ = store.SaveRawHTML(site.ID, "<h1>Exportable Site</h1>")

	var buf bytes.Buffer
	if err := store.ExportSiteZip(site.ID, &buf); err != nil {
		t.Fatalf("failed to export zip: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("invalid exported zip: %v", err)
	}
	if len(zr.File) != 1 || zr.File[0].Name != "index.html" {
		t.Errorf("unexpected exported files: %+v", zr.File)
	}
}
