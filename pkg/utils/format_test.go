package utils

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.02 kB"},
		{1048576, "1.05 MB"},
		{1073741824, "1.07 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %s; want %s", tt.bytes, got, tt.expected)
		}
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()
	if res := FormatTimeAgo(now.Add(-2 * time.Second)); res != "2 seconds ago" {
		t.Errorf("expected '2 seconds ago', got %s", res)
	}
	if res := FormatTimeAgo(now.Add(-5 * time.Minute)); res != "5 minutes ago" {
		t.Errorf("expected '5 minutes ago', got %s", res)
	}
}

func TestTruncateString(t *testing.T) {
	if got := TruncateString("hello world", 5); got != "he..." {
		t.Errorf("expected 'he...', got '%s'", got)
	}
	if got := TruncateString("hi", 5); got != "hi" {
		t.Errorf("expected 'hi', got '%s'", got)
	}
}
