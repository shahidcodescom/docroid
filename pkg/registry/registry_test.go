package registry

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		input       string
		expectedReg string
		expectedRep string
		expectedTag string
		expectedDig string
	}{
		{
			input:       "alpine",
			expectedReg: "registry-1.docker.io",
			expectedRep: "library/alpine",
			expectedTag: "latest",
		},
		{
			input:       "alpine:3.19",
			expectedReg: "registry-1.docker.io",
			expectedRep: "library/alpine",
			expectedTag: "3.19",
		},
		{
			input:       "myuser/app:v1.2",
			expectedReg: "registry-1.docker.io",
			expectedRep: "myuser/app",
			expectedTag: "v1.2",
		},
		{
			input:       "ghcr.io/org/repo:edge",
			expectedReg: "ghcr.io",
			expectedRep: "org/repo",
			expectedTag: "edge",
		},
		{
			input:       "localhost:5000/test-app:latest",
			expectedReg: "localhost:5000",
			expectedRep: "test-app",
			expectedTag: "latest",
		},
		{
			input:       "alpine@sha256:e1c082e3d3c45cccac829840a25941e6e",
			expectedReg: "registry-1.docker.io",
			expectedRep: "library/alpine",
			expectedTag: "",
			expectedDig: "sha256:e1c082e3d3c45cccac829840a25941e6e",
		},
	}

	for _, tt := range tests {
		ref, err := ParseReference(tt.input)
		if err != nil {
			t.Fatalf("ParseReference(%s) returned error: %v", tt.input, err)
		}
		if ref.Registry != tt.expectedReg {
			t.Errorf("[%s] Registry = %s, expected %s", tt.input, ref.Registry, tt.expectedReg)
		}
		if ref.Repository != tt.expectedRep {
			t.Errorf("[%s] Repository = %s, expected %s", tt.input, ref.Repository, tt.expectedRep)
		}
		if ref.Tag != tt.expectedTag {
			t.Errorf("[%s] Tag = %s, expected %s", tt.input, ref.Tag, tt.expectedTag)
		}
		if ref.Digest != tt.expectedDig {
			t.Errorf("[%s] Digest = %s, expected %s", tt.input, ref.Digest, tt.expectedDig)
		}
	}
}

// TestRegistryPushAndPull spins up a local mock OCI V2 registry and tests full push and pull lifecycle
func TestRegistryPushAndPull(t *testing.T) {
	// In-memory mock registry state
	var mu sync.Mutex
	blobs := make(map[string][]byte)
	manifests := make(map[string][]byte)

	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Token auth endpoint
		if path == "/v2/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": "mock-auth-token-xyz",
			})
			return
		}

		// Auth check for /v2/
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer mock-auth-token-xyz" {
			w.Header().Set("Www-Authenticate", fmt.Sprintf(`Bearer realm="%s/v2/token",service="mock-registry"`, serverURL))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Blob upload initiation: POST /v2/<repo>/blobs/uploads/
		if strings.HasSuffix(path, "/blobs/uploads/") && r.Method == http.MethodPost {
			uploadID := "upload-12345"
			w.Header().Set("Location", fmt.Sprintf("%s/v2/myrepo/blobs/uploads/%s", serverURL, uploadID))
			w.WriteHeader(http.StatusAccepted)
			return
		}

		// Blob upload complete: PUT /v2/<repo>/blobs/uploads/<id>?digest=...
		if strings.Contains(path, "/blobs/uploads/") && r.Method == http.MethodPut {
			digest := r.URL.Query().Get("digest")
			data, _ := io.ReadAll(r.Body)
			mu.Lock()
			blobs[digest] = data
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			return
		}

		// Blob HEAD/GET: /v2/<repo>/blobs/<digest>
		if strings.Contains(path, "/blobs/sha256:") {
			parts := strings.Split(path, "/blobs/")
			digest := parts[len(parts)-1]
			mu.Lock()
			data, exists := blobs[digest]
			mu.Unlock()
			if !exists {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}

		// Manifest PUT / GET: /v2/<repo>/manifests/<tag>
		if strings.Contains(path, "/manifests/") {
			parts := strings.Split(path, "/manifests/")
			tag := parts[len(parts)-1]

			if r.Method == http.MethodPut {
				data, _ := io.ReadAll(r.Body)
				mu.Lock()
				manifests[tag] = data
				mu.Unlock()
				w.WriteHeader(http.StatusCreated)
				return
			}

			if r.Method == http.MethodGet {
				mu.Lock()
				data, exists := manifests[tag]
				mu.Unlock()
				if !exists {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", MediaTypeOCIV1Manifest)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(data)
				return
			}
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()
	serverURL = ts.URL

	u, _ := url.Parse(ts.URL)
	registryHost := u.Host

	ref, err := ParseReference(fmt.Sprintf("%s/myrepo:v1.0", registryHost))
	if err != nil {
		t.Fatalf("ParseReference error: %v", err)
	}

	// Create dummy local image to push
	tempDir := t.TempDir()
	imageDir := filepath.Join(tempDir, "push-image")
	payloadDir := filepath.Join(imageDir, "payload")
	_ = os.MkdirAll(payloadDir, 0755)

	testFile := filepath.Join(payloadDir, "hello.txt")
	_ = os.WriteFile(testFile, []byte("Hello from Docroid Push!"), 0644)

	manifestJSON := []byte(`{
		"format_version": "1.0",
		"name": "myrepo",
		"version": "v1.0",
		"architecture": "arm64",
		"engine": "rootfs",
		"command": "/bin/echo",
		"args": ["hello"]
	}`)
	_ = os.WriteFile(filepath.Join(imageDir, "manifest.json"), manifestJSON, 0644)

	client := NewClient()

	// 1. Test Push
	var pushOut bytes.Buffer
	_, err = client.Push(ref, imageDir, &pushOut)
	if err != nil {
		t.Fatalf("client.Push failed: %v, output: %s", err, pushOut.String())
	}

	if !strings.Contains(pushOut.String(), "Successfully pushed") {
		t.Errorf("Expected 'Successfully pushed', got: %s", pushOut.String())
	}

	// 2. Test Pull into fresh directory
	pullDir := filepath.Join(tempDir, "pulled-image")
	var pullOut bytes.Buffer
	pulledManifest, err := client.Pull(ref, pullDir, &pullOut)
	if err != nil {
		t.Fatalf("client.Pull failed: %v, output: %s", err, pullOut.String())
	}

	if pulledManifest == nil || pulledManifest.Name != ref.FullName() {
		t.Errorf("Unexpected pulled manifest: %+v, expected name: %s", pulledManifest, ref.FullName())
	}

	// Verify extracted file content
	extractedFile := filepath.Join(pullDir, "payload", "hello.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Extracted file missing: %v", err)
	}
	if string(content) != "Hello from Docroid Push!" {
		t.Errorf("File content mismatch: expected 'Hello from Docroid Push!', got: '%s'", string(content))
	}
}

// Helper to create a gzip tar buffer for tests
func makeTarGz(files map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
