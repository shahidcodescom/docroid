package bundle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docroid/pkg/app"
)

func TestBuildExportImportLifecycle(t *testing.T) {
	// 1. Create temporary store and build context
	tempDir, err := os.MkdirTemp("", "docroid_bundle_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storeDir := filepath.Join(tempDir, "store")
	contextDir := filepath.Join(tempDir, "context")
	_ = os.MkdirAll(contextDir, 0755)

	store, err := app.NewStore(storeDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// 2. Create sample payload file and Docroidfile
	serverBinPath := filepath.Join(contextDir, "server")
	if err := os.WriteFile(serverBinPath, []byte("#!/bin/sh\necho 'Hello from bundle'"), 0755); err != nil {
		t.Fatalf("failed to write dummy server: %v", err)
	}

	docroidfileContent := `
FROM native
NAME test-api
VERSION 1.0.0
PORT 8080
ENV KEY=VAL
COPY server /bin/server
CMD ["/bin/server"]
`
	dfPath := filepath.Join(contextDir, "Docroidfile")
	if err := os.WriteFile(dfPath, []byte(docroidfileContent), 0644); err != nil {
		t.Fatalf("failed to write Docroidfile: %v", err)
	}

	// 3. Build image
	builder := NewBuilder(store)
	var buildBuf bytes.Buffer
	manifest, err := builder.BuildImage(dfPath, contextDir, "test-api:1.0.0", &buildBuf)
	if err != nil {
		t.Fatalf("BuildImage failed: %v", err)
	}

	if manifest.Name != "test-api" || manifest.Version != "1.0.0" {
		t.Fatalf("unexpected manifest name/version: %s:%s", manifest.Name, manifest.Version)
	}

	// 4. List Images
	images, err := ListImages(store)
	if err != nil || len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}

	// 5. Export Image to .dpack
	dpackPath := filepath.Join(tempDir, "test-api.dpack")
	var exportBuf bytes.Buffer
	if err := ExportImage(store, "test-api:1.0.0", dpackPath, &exportBuf); err != nil {
		t.Fatalf("ExportImage failed: %v", err)
	}

	if _, err := os.Stat(dpackPath); os.IsNotExist(err) {
		t.Fatalf("dpack file was not created: %s", dpackPath)
	}

	// 6. Delete image from store
	if err := RemoveImage(store, "test-api:1.0.0"); err != nil {
		t.Fatalf("RemoveImage failed: %v", err)
	}
	imagesAfterDel, _ := ListImages(store)
	if len(imagesAfterDel) != 0 {
		t.Fatalf("expected 0 images after deletion")
	}

	// 7. Import .dpack into store
	var importBuf bytes.Buffer
	importedManifest, err := ImportPackage(store, dpackPath, &importBuf)
	if err != nil {
		t.Fatalf("ImportPackage failed: %v", err)
	}

	if importedManifest.Name != "test-api" {
		t.Errorf("expected imported name 'test-api', got '%s'", importedManifest.Name)
	}

	// 8. Verify image exists again
	reloaded, err := GetImage(store, "test-api:1.0.0")
	if err != nil || reloaded == nil {
		t.Fatalf("GetImage failed after import: %v", err)
	}
}

func TestDockerfileBuild(t *testing.T) {
	tempDir := t.TempDir()
	storeDir := filepath.Join(tempDir, "store")
	contextDir := filepath.Join(tempDir, "context")
	_ = os.MkdirAll(contextDir, 0755)

	store, err := app.NewStore(storeDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	appScript := filepath.Join(contextDir, "main.sh")
	_ = os.WriteFile(appScript, []byte("#!/bin/sh\necho 'Running app'"), 0755)

	dockerfileContent := `
FROM native
WORKDIR /app
COPY main.sh ./
ENV APP_ENV=production
RUN echo "build_step_completed" > built.txt
EXPOSE 9000
CMD ["sh", "main.sh"]
`
	_ = os.WriteFile(filepath.Join(contextDir, "Dockerfile"), []byte(dockerfileContent), 0644)

	builder := NewBuilder(store)
	var buildBuf bytes.Buffer
	manifest, err := builder.BuildImage("", contextDir, "my-docker-app:v2", &buildBuf)
	if err != nil {
		t.Fatalf("BuildImage from Dockerfile failed: %v, output: %s", err, buildBuf.String())
	}

	if manifest.Name != "my-docker-app" || manifest.Version != "v2" {
		t.Errorf("Unexpected manifest name/version: %s:%s", manifest.Name, manifest.Version)
	}
	if manifest.WorkDir != "/app" {
		t.Errorf("Expected WorkDir /app, got %s", manifest.WorkDir)
	}
	if manifest.Env["APP_ENV"] != "production" {
		t.Errorf("Expected APP_ENV=production, got %s", manifest.Env["APP_ENV"])
	}
	if manifest.Port != 9000 {
		t.Errorf("Expected Port 9000, got %d", manifest.Port)
	}

	// Verify RUN step executed and created built.txt inside payload
	builtFile := filepath.Join(store.BaseDir(), "images", "my-docker-app:v2", "payload", "app", "built.txt")
	content, err := os.ReadFile(builtFile)
	if err != nil {
		t.Fatalf("built.txt not created by RUN step: %v", err)
	}
	if !strings.Contains(string(content), "build_step_completed") {
		t.Errorf("Unexpected content in built.txt: %s", string(content))
	}
}

