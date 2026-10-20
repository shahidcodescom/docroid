package netproxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"docroid/pkg/utils"
)

// PortMapping defines a host-to-container port forward specification
type PortMapping struct {
	HostIP        string `json:"host_ip,omitempty"`
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol,omitempty"` // "tcp"
}

// String returns formatted mapping e.g. "8080:3000" or "127.0.0.1:8080:3000"
func (p *PortMapping) String() string {
	if p.HostIP != "" {
		return fmt.Sprintf("%s:%d:%d", p.HostIP, p.HostPort, p.ContainerPort)
	}
	if p.HostPort == p.ContainerPort {
		return fmt.Sprintf("%d", p.HostPort)
	}
	return fmt.Sprintf("%d:%d", p.HostPort, p.ContainerPort)
}

// ParsePortMapping parses specs like "8080:3000", "127.0.0.1:8080:3000", or "8080"
func ParsePortMapping(spec string) (*PortMapping, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("empty port spec")
	}

	parts := strings.Split(spec, ":")
	switch len(parts) {
	case 1:
		// Just "8080" (HostPort == ContainerPort)
		p, err := strconv.Atoi(parts[0])
		if err != nil || p <= 0 || p > 65535 {
			return nil, fmt.Errorf("invalid port number: %s", parts[0])
		}
		return &PortMapping{HostPort: p, ContainerPort: p, Protocol: "tcp"}, nil

	case 2:
		// "hostPort:containerPort" e.g. "8080:3000"
		hp, err1 := strconv.Atoi(parts[0])
		cp, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || hp <= 0 || hp > 65535 || cp <= 0 || cp > 65535 {
			return nil, fmt.Errorf("invalid port mapping '%s': ports must be between 1 and 65535", spec)
		}
		return &PortMapping{HostPort: hp, ContainerPort: cp, Protocol: "tcp"}, nil

	case 3:
		// "hostIP:hostPort:containerPort" e.g. "127.0.0.1:8080:3000"
		ip := parts[0]
		hp, err1 := strconv.Atoi(parts[1])
		cp, err2 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || hp <= 0 || hp > 65535 || cp <= 0 || cp > 65535 {
			return nil, fmt.Errorf("invalid port mapping '%s'", spec)
		}
		return &PortMapping{HostIP: ip, HostPort: hp, ContainerPort: cp, Protocol: "tcp"}, nil

	default:
		return nil, fmt.Errorf("invalid port mapping format: '%s'. Expected [host_ip:]host_port:container_port", spec)
	}
}

// ProxySession represents an active forwarder listener with a dynamically switchable target port
type ProxySession struct {
	AppID        string
	Mapping      *PortMapping
	Listener     net.Listener
	ctrlListener net.Listener
	ctrlPath     string
	targetPort   int32
	activeConns  int64
	closed       chan struct{}
}

func (s *ProxySession) TargetPort() int {
	return int(atomic.LoadInt32(&s.targetPort))
}

func (s *ProxySession) SetTargetPort(port int) {
	atomic.StoreInt32(&s.targetPort, int32(port))
}

// ProxyManager manages active TCP forwarder listeners and zero-downtime cutovers
type ProxyManager struct {
	mu       sync.RWMutex
	sessions map[string][]*ProxySession // keyed by appID
}

var globalProxyManager = &ProxyManager{
	sessions: make(map[string][]*ProxySession),
}

func GetProxyManager() *ProxyManager {
	return globalProxyManager
}

// StartProxy starts a bidirectional TCP forwarder
func (m *ProxyManager) StartProxy(appID string, mapping *PortMapping) error {
	return m.StartProxyWithControl(appID, mapping, "")
}

// StartProxyWithControl starts a bidirectional TCP forwarder and an IPC control socket
func (m *ProxyManager) StartProxyWithControl(appID string, mapping *PortMapping, ctrlDir string) error {
	if mapping.HostPort == mapping.ContainerPort {
		// No forwarding needed, check if available
		return utils.CheckPortInUse(mapping.HostPort)
	}

	// Verify host port availability
	if err := utils.CheckPortInUse(mapping.HostPort); err != nil {
		return err
	}

	hostAddr := fmt.Sprintf("%s:%d", mapping.HostIP, mapping.HostPort)
	if mapping.HostIP == "" {
		hostAddr = fmt.Sprintf("0.0.0.0:%d", mapping.HostPort)
	}

	ln, err := net.Listen("tcp", hostAddr)
	if err != nil {
		return fmt.Errorf("failed to bind host port %d: %w", mapping.HostPort, err)
	}

	sess := &ProxySession{
		AppID:      appID,
		Mapping:    mapping,
		Listener:   ln,
		targetPort: int32(mapping.ContainerPort),
		closed:     make(chan struct{}),
	}

	if ctrlDir != "" {
		_ = os.MkdirAll(ctrlDir, 0755)
		ctrlPath := filepath.Join(ctrlDir, "proxy.sock")
		_ = os.Remove(ctrlPath)
		if ctrlLn, err := net.Listen("unix", ctrlPath); err == nil {
			sess.ctrlListener = ctrlLn
			sess.ctrlPath = ctrlPath
			go m.serveControl(sess, ctrlLn)
		}
	}

	m.mu.Lock()
	m.sessions[appID] = append(m.sessions[appID], sess)
	m.mu.Unlock()

	// Launch forwarder goroutine
	go func() {
		for {
			clientConn, err := ln.Accept()
			if err != nil {
				// Listener closed
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				atomic.AddInt64(&sess.activeConns, 1)
				defer atomic.AddInt64(&sess.activeConns, -1)

				targetPort := sess.TargetPort()
				targetAddr := fmt.Sprintf("127.0.0.1:%d", targetPort)

				// Connect to container port with timeout
				targetConn, err := net.DialTimeout("tcp", targetAddr, 2*time.Second)
				if err != nil {
					return
				}
				defer targetConn.Close()

				// Bidirectional copy
				errChan := make(chan error, 2)
				go func() {
					_, err := io.Copy(targetConn, c)
					errChan <- err
				}()
				go func() {
					_, err := io.Copy(c, targetConn)
					errChan <- err
				}()

				<-errChan
			}(clientConn)
		}
	}()

	return nil
}

// SwitchTarget atomically updates the destination container port for all new incoming connections
func (m *ProxyManager) SwitchTarget(appID string, newContainerPort int) error {
	m.mu.RLock()
	sessions, ok := m.sessions[appID]
	m.mu.RUnlock()

	if !ok || len(sessions) == 0 {
		return fmt.Errorf("no active proxy session for app %s", appID)
	}

	for _, sess := range sessions {
		sess.SetTargetPort(newContainerPort)
	}
	return nil
}

// GetTargetPort returns the current target port for an app proxy session
func (m *ProxyManager) GetTargetPort(appID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions, ok := m.sessions[appID]
	if !ok || len(sessions) == 0 {
		return 0, fmt.Errorf("no active proxy session for app %s", appID)
	}
	return sessions[0].TargetPort(), nil
}

// HasProxy returns true if an app has an active proxy session
func (m *ProxyManager) HasProxy(appID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions, ok := m.sessions[appID]
	return ok && len(sessions) > 0
}

// Drain waits for active in-flight connections to complete, up to timeout
func (m *ProxyManager) Drain(appID string, timeout time.Duration) {
	m.mu.RLock()
	sessions := m.sessions[appID]
	m.mu.RUnlock()

	if len(sessions) == 0 {
		return
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		allZero := true
		for _, s := range sessions {
			if atomic.LoadInt64(&s.activeConns) > 0 {
				allZero = false
				break
			}
		}
		if allZero {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// StopProxies stops and closes all active listeners for an application
func (m *ProxyManager) StopProxies(appID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessions, ok := m.sessions[appID]
	if !ok {
		return
	}

	for _, sess := range sessions {
		_ = sess.Listener.Close()
		if sess.ctrlListener != nil {
			_ = sess.ctrlListener.Close()
		}
		if sess.ctrlPath != "" {
			_ = os.Remove(sess.ctrlPath)
		}
	}
	delete(m.sessions, appID)
}

func (m *ProxyManager) serveControl(sess *ProxySession, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			scanner := bufio.NewScanner(c)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "SWITCH ") {
					portStr := strings.TrimSpace(strings.TrimPrefix(line, "SWITCH "))
					if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
						sess.SetTargetPort(p)
						fmt.Fprintf(c, "OK\n")
					} else {
						fmt.Fprintf(c, "ERR invalid port\n")
					}
				} else if strings.HasPrefix(line, "DRAIN ") {
					msStr := strings.TrimSpace(strings.TrimPrefix(line, "DRAIN "))
					ms, _ := strconv.Atoi(msStr)
					if ms <= 0 {
						ms = 1500
					}
					m.Drain(sess.AppID, time.Duration(ms)*time.Millisecond)
					fmt.Fprintf(c, "OK\n")
				} else if line == "TARGET" {
					fmt.Fprintf(c, "%d\n", sess.TargetPort())
				} else if line == "PING" {
					fmt.Fprintf(c, "PONG\n")
				}
			}
		}(conn)
	}
}

// SwitchTargetAcrossProcesses switches target either locally if this process runs the proxy, or via unix socket if held by another daemonless process
func SwitchTargetAcrossProcesses(appID, ctrlDir string, newTargetPort int) error {
	if GetProxyManager().HasProxy(appID) {
		return GetProxyManager().SwitchTarget(appID, newTargetPort)
	}
	if ctrlDir != "" {
		sockPath := filepath.Join(ctrlDir, "proxy.sock")
		conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
		if err == nil {
			defer conn.Close()
			fmt.Fprintf(conn, "SWITCH %d\n", newTargetPort)
			buf := make([]byte, 16)
			n, _ := conn.Read(buf)
			if strings.HasPrefix(string(buf[:n]), "OK") {
				return nil
			}
		}
	}
	return fmt.Errorf("no active proxy found locally or via socket for app %s", appID)
}

// DrainAcrossProcesses drains either locally or via unix socket
func DrainAcrossProcesses(appID, ctrlDir string, timeout time.Duration) {
	if GetProxyManager().HasProxy(appID) {
		GetProxyManager().Drain(appID, timeout)
		return
	}
	if ctrlDir != "" {
		sockPath := filepath.Join(ctrlDir, "proxy.sock")
		conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
		if err == nil {
			defer conn.Close()
			fmt.Fprintf(conn, "DRAIN %d\n", timeout.Milliseconds())
			buf := make([]byte, 16)
			_, _ = conn.Read(buf)
			return
		}
	}
	time.Sleep(timeout)
}

// HasProxyAcrossProcesses returns true if proxy is active either in-process or via socket
func HasProxyAcrossProcesses(appID, ctrlDir string) bool {
	if GetProxyManager().HasProxy(appID) {
		return true
	}
	if ctrlDir != "" {
		sockPath := filepath.Join(ctrlDir, "proxy.sock")
		conn, err := net.DialTimeout("unix", sockPath, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

