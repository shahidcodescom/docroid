package netproxy

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestParsePortMapping(t *testing.T) {
	tests := []struct {
		input       string
		wantHost    int
		wantCont    int
		wantIP      string
		expectError bool
	}{
		{"8080:3000", 8080, 3000, "", false},
		{"127.0.0.1:9000:80", 9000, 80, "127.0.0.1", false},
		{"5000", 5000, 5000, "", false},
		{"invalid", 0, 0, "", true},
		{"8080:99999", 0, 0, "", true},
		{"-1:80", 0, 0, "", true},
	}

	for _, tt := range tests {
		m, err := ParsePortMapping(tt.input)
		if tt.expectError {
			if err == nil {
				t.Errorf("ParsePortMapping(%s) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePortMapping(%s) unexpected error: %v", tt.input, err)
			continue
		}
		if m.HostPort != tt.wantHost || m.ContainerPort != tt.wantCont || m.HostIP != tt.wantIP {
			t.Errorf("ParsePortMapping(%s) = %+v, want HostPort=%d, ContPort=%d, IP=%s",
				tt.input, m, tt.wantHost, tt.wantCont, tt.wantIP)
		}
	}
}

func TestTCPProxyForwarding(t *testing.T) {
	// 1. Start a mock container TCP server on a free port
	contLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock container listener: %v", err)
	}
	defer contLn.Close()
	contPort := contLn.Addr().(*net.TCPAddr).Port

	// Echo server logic
	go func() {
		for {
			conn, err := contLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.WriteString(c, "HELLO_FROM_CONTAINER")
			}(conn)
		}
	}()

	// 2. Find a free host port
	tempLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free host port: %v", err)
	}
	hostPort := tempLn.Addr().(*net.TCPAddr).Port
	tempLn.Close() // Release so proxy can bind

	// 3. Start proxy
	mapping := &PortMapping{
		HostIP:        "127.0.0.1",
		HostPort:      hostPort,
		ContainerPort: contPort,
	}

	mgr := GetProxyManager()
	appID := "test-proxy-app"
	if err := mgr.StartProxy(appID, mapping); err != nil {
		t.Fatalf("StartProxy failed: %v", err)
	}
	defer mgr.StopProxies(appID)

	time.Sleep(50 * time.Millisecond)

	// 4. Connect to host port and read response
	client, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", hostPort))
	if err != nil {
		t.Fatalf("failed to dial host port %d: %v", hostPort, err)
	}
	defer client.Close()

	buf := make([]byte, 64)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("failed to read from client: %v", err)
	}

	response := string(buf[:n])
	if response != "HELLO_FROM_CONTAINER" {
		t.Errorf("expected 'HELLO_FROM_CONTAINER', got '%s'", response)
	}
}

func TestProxyAtomicTargetSwitch(t *testing.T) {
	// 1. Start Server 1 (Old instance)
	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start ln1: %v", err)
	}
	defer ln1.Close()
	port1 := ln1.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := ln1.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.WriteString(c, "VERSION_1")
			}(conn)
		}
	}()

	// 2. Start Server 2 (New candidate)
	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start ln2: %v", err)
	}
	defer ln2.Close()
	port2 := ln2.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := ln2.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.WriteString(c, "VERSION_2")
			}(conn)
		}
	}()

	// 3. Find free host port
	tempLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free host port: %v", err)
	}
	hostPort := tempLn.Addr().(*net.TCPAddr).Port
	tempLn.Close()

	appID := "atomic-switch-test"
	mgr := GetProxyManager()
	mapping := &PortMapping{HostIP: "127.0.0.1", HostPort: hostPort, ContainerPort: port1}
	if err := mgr.StartProxy(appID, mapping); err != nil {
		t.Fatalf("StartProxy failed: %v", err)
	}
	defer mgr.StopProxies(appID)

	// Verify connecting to hostPort reaches Version 1
	c1, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", hostPort))
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	buf := make([]byte, 32)
	n, _ := c1.Read(buf)
	c1.Close()
	if string(buf[:n]) != "VERSION_1" {
		t.Fatalf("expected VERSION_1, got %s", string(buf[:n]))
	}

	// 4. Perform atomic target switch to Server 2
	if err := mgr.SwitchTarget(appID, port2); err != nil {
		t.Fatalf("SwitchTarget failed: %v", err)
	}

	// Verify connecting to the EXACT same hostPort now immediately reaches Version 2
	c2, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", hostPort))
	if err != nil {
		t.Fatalf("dial after switch failed: %v", err)
	}
	n, _ = c2.Read(buf)
	c2.Close()
	if string(buf[:n]) != "VERSION_2" {
		t.Fatalf("expected VERSION_2, got %s", string(buf[:n]))
	}
}

