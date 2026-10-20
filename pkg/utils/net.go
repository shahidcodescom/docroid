package utils

import (
	"fmt"
	"net"
)

// IsPortAvailable checks if a TCP port is currently available on localhost
func IsPortAvailable(port int) bool {
	if port <= 0 || port > 65535 {
		return false
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// CheckPortInUse returns an error if the port is already bound
func CheckPortInUse(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}
	if !IsPortAvailable(port) {
		return fmt.Errorf("port %d is already in use by another application", port)
	}
	return nil
}

// GetFreePort returns an available ephemeral port on 127.0.0.1
func GetFreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to allocate free port: %w", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

