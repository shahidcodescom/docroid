package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretsLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docroid_secrets_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create secrets store: %v", err)
	}

	// 1. Set secret
	if err := store.Set("DB_PASSWORD", "super_secret_123"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 2. Verify file permissions (0600)
	secretPath := filepath.Join(store.baseDir, "DB_PASSWORD")
	info, err := os.Stat(secretPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected 0600 permissions, got %o", perm)
	}

	// 3. Get secret
	val, err := store.Get("DB_PASSWORD")
	if err != nil || val != "super_secret_123" {
		t.Fatalf("Get failed: got '%s', want 'super_secret_123'", val)
	}

	// 4. List secrets
	list, err := store.List()
	if err != nil || len(list) != 1 || list["DB_PASSWORD"] != "super_secret_123" {
		t.Fatalf("List failed: %+v", list)
	}

	// 5. Delete secret
	if err := store.Delete("DB_PASSWORD"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify not found
	_, err = store.Get("DB_PASSWORD")
	if err == nil {
		t.Fatalf("expected error getting deleted secret, got nil")
	}
}

func TestParseEnvFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_env_*.env")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `
# Configuration file
export PORT=8080
API_KEY="secret-jwt-key"
DATABASE_URL='postgres://user:pass@localhost/db'
DEBUG=true
EMPTY_VAL=
`
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("failed to write env: %v", err)
	}
	tempFile.Close()

	envMap, err := ParseEnvFile(tempFile.Name())
	if err != nil {
		t.Fatalf("ParseEnvFile failed: %v", err)
	}

	if envMap["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %s", envMap["PORT"])
	}
	if envMap["API_KEY"] != "secret-jwt-key" {
		t.Errorf("expected API_KEY=secret-jwt-key, got %s", envMap["API_KEY"])
	}
	if envMap["DATABASE_URL"] != "postgres://user:pass@localhost/db" {
		t.Errorf("expected DATABASE_URL=postgres://user:pass@localhost/db, got %s", envMap["DATABASE_URL"])
	}
	if envMap["DEBUG"] != "true" {
		t.Errorf("expected DEBUG=true, got %s", envMap["DEBUG"])
	}
}
