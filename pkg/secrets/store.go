package secrets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Store manages secure key-value secrets with restricted file permissions (0600)
type Store struct {
	baseDir string
	mu      sync.RWMutex
}

// NewStore initializes the secrets directory in ~/.docroid/secrets/
func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		if env := os.Getenv("DOCROID_HOME"); env != "" {
			baseDir = env
		} else if home := os.Getenv("HOME"); home != "" {
			baseDir = filepath.Join(home, ".docroid")
		} else {
			baseDir = filepath.Join(os.TempDir(), "docroid")
		}
	}

	secretsDir := filepath.Join(baseDir, "secrets")
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create secrets directory: %w", err)
	}

	return &Store{baseDir: secretsDir}, nil
}

// Set saves a secret key-value pair with 0600 permissions
func (s *Store) Set(key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("secret key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.baseDir, sanitizeKey(key))
	// Write with 0600 (owner read/write only)
	return os.WriteFile(filePath, []byte(value), 0600)
}

// Get retrieves the value for a given key
func (s *Store) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.baseDir, sanitizeKey(key))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("secret '%s' not found: %w", key, err)
	}
	return string(data), nil
}

// List returns all secret keys and masked values (or full values)
func (s *Store) List() (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.baseDir, entry.Name()))
		if err == nil {
			result[entry.Name()] = string(data)
		}
	}
	return result, nil
}

// Delete removes a stored secret
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.baseDir, sanitizeKey(key))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("secret '%s' does not exist", key)
	}
	return os.Remove(filePath)
}

// ParseEnvFile parses a .env file into key-value pairs
func ParseEnvFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file %s: %w", filePath, err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Support "export KEY=VAL"
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip surrounding single or double quotes
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			if len(val) >= 2 {
				val = val[1 : len(val)-1]
			}
		}

		if key != "" {
			result[key] = val
		}
	}

	return result, scanner.Err()
}

func sanitizeKey(key string) string {
	return strings.ReplaceAll(filepath.Clean(key), string(os.PathSeparator), "_")
}
