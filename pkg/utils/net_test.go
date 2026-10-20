package utils

import (
	"net"
	"testing"
)

func TestPortAvailability(t *testing.T) {
	// Dynamically find a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on port 0: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	// Currently port is occupied
	if IsPortAvailable(port) {
		t.Errorf("expected port %d to be in use, but IsPortAvailable returned true", port)
	}

	// Close listener
	_ = ln.Close()

	// Port should now be available
	if !IsPortAvailable(port) {
		t.Errorf("expected port %d to be available, but IsPortAvailable returned false", port)
	}
}
