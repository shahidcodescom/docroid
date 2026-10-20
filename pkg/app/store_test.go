package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreAppLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_store_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	testApp := &App{
		ID:        "app1234567",
		Name:      "test-service",
		Engine:    EngineNative,
		Command:   "echo",
		Args:      []string{"hello"},
		Status:    StatusCreated,
		CreatedAt: time.Now(),
	}

	// 1. Save App
	if err := store.SaveApp(testApp); err != nil {
		t.Fatalf("failed to save app: %v", err)
	}

	// 2. Load App
	loaded, err := store.LoadApp("app1234567")
	if err != nil {
		t.Fatalf("failed to load app: %v", err)
	}
	if loaded.Name != "test-service" {
		t.Errorf("expected name 'test-service', got '%s'", loaded.Name)
	}

	// 3. Find App by name and prefix
	foundByName, err := store.FindApp("test-service")
	if err != nil || foundByName.ID != testApp.ID {
		t.Fatalf("FindApp by name failed")
	}

	foundByPrefix, err := store.FindApp("app123")
	if err != nil || foundByPrefix.ID != testApp.ID {
		t.Fatalf("FindApp by prefix failed")
	}

	// 4. List Apps
	apps, err := store.ListApps()
	if err != nil || len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}

	// 5. Delete App
	if err := store.DeleteApp(testApp.ID, true); err != nil {
		t.Fatalf("failed to delete app: %v", err)
	}

	// Verify deleted
	_, err = store.LoadApp(testApp.ID)
	if err == nil {
		t.Fatalf("expected error loading deleted app, got nil")
	}
}

func TestStoreWorkspaceAndLogPaths(t *testing.T) {
	store, _ := NewStore("/tmp/test_docroid_paths")
	defer os.RemoveAll("/tmp/test_docroid_paths")

	ws := store.WorkspaceDir("xyz123")
	expectedWs := filepath.Join("/tmp/test_docroid_paths", "workspaces", "xyz123")
	if ws != expectedWs {
		t.Errorf("expected ws '%s', got '%s'", expectedWs, ws)
	}

	out, errOut := store.LogPaths("xyz123")
	if !filepath.IsAbs(out) || !filepath.IsAbs(errOut) {
		t.Errorf("expected absolute log paths")
	}
}
