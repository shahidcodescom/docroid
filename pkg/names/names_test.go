package names

import (
	"strings"
	"testing"
)

func TestGenerateRandomName(t *testing.T) {
	name := GenerateRandomName()
	if name == "" {
		t.Fatalf("expected non-empty name")
	}
	parts := strings.Split(name, "_")
	if len(parts) != 2 {
		t.Fatalf("expected format 'adjective_surname', got %s", name)
	}
}
