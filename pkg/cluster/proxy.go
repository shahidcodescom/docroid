package cluster

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ServiceProxy manages round-robin load balancing for a cluster service
type ServiceProxy struct {
	deploymentName string
	listenPort     int
	listener       net.Listener
	endpoints      []string
	counter        uint64
	mu             sync.RWMutex
	stopChan       chan struct{}
}

// NewServiceProxy creates a proxy listener for a deployment's ServicePort
func NewServiceProxy(deploymentName string, port int) (*ServiceProxy, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to bind service port %d: %w", port, err)
	}

	p := &ServiceProxy{
		deploymentName: deploymentName,
		listenPort:     port,
		listener:       ln,
		endpoints:      []string{},
		stopChan:       make(chan struct{}),
	}

	go p.serve()
	return p, nil
}

// UpdateEndpoints sets the current active pod endpoints (host:port)
func (p *ServiceProxy) UpdateEndpoints(endpoints []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.endpoints = append([]string{}, endpoints...)
}

func (p *ServiceProxy) nextEndpoint() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.endpoints) == 0 {
		return ""
	}
	idx := atomic.AddUint64(&p.counter, 1) % uint64(len(p.endpoints))
	return p.endpoints[idx]
}

func (p *ServiceProxy) serve() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.stopChan:
				return
			default:
				time.Sleep(50 * time.Millisecond)
				continue
			}
		}

		go p.forward(conn)
	}
}

func (p *ServiceProxy) forward(src net.Conn) {
	defer src.Close()

	endpoint := p.nextEndpoint()
	if endpoint == "" {
		_ = src.SetDeadline(time.Now().Add(5 * time.Second))
		_, _ = src.Write([]byte("HTTP/1.1 503 Service Unavailable\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nNo healthy pods available in cluster\n"))
		return
	}

	dest, err := net.DialTimeout("tcp", endpoint, 3*time.Second)
	if err != nil {
		_, _ = src.Write([]byte(fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nFailed to connect to pod %s: %v\n", endpoint, err)))
		return
	}
	defer dest.Close()

	// Bi-directional pipe
	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(dest, src)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(src, dest)
		done <- struct{}{}
	}()

	<-done
}

// Stop closes the listener and halts load balancing
func (p *ServiceProxy) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.stopChan:
		return
	default:
		close(p.stopChan)
		if p.listener != nil {
			_ = p.listener.Close()
		}
	}
}

// ProxyManager tracks active service proxies for all deployments
type ProxyManager struct {
	proxies map[string]*ServiceProxy // deploymentName -> ServiceProxy
	mu      sync.Mutex
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		proxies: make(map[string]*ServiceProxy),
	}
}

// EnsureProxy creates or updates a service proxy for a deployment
func (pm *ProxyManager) EnsureProxy(deploymentName string, servicePort int, endpoints []string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if servicePort <= 0 {
		// If proxy existed previously, shut it down
		if existing, ok := pm.proxies[deploymentName]; ok {
			existing.Stop()
			delete(pm.proxies, deploymentName)
		}
		return nil
	}

	if existing, ok := pm.proxies[deploymentName]; ok {
		if existing.listenPort == servicePort {
			existing.UpdateEndpoints(endpoints)
			return nil
		}
		existing.Stop()
		delete(pm.proxies, deploymentName)
	}

	proxy, err := NewServiceProxy(deploymentName, servicePort)
	if err != nil {
		return err
	}
	proxy.UpdateEndpoints(endpoints)
	pm.proxies[deploymentName] = proxy
	return nil
}

// StopProxy removes a service proxy
func (pm *ProxyManager) StopProxy(deploymentName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p, ok := pm.proxies[deploymentName]; ok {
		p.Stop()
		delete(pm.proxies, deploymentName)
	}
}

// StopAll stops all active proxies
func (pm *ProxyManager) StopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, p := range pm.proxies {
		p.Stop()
		delete(pm.proxies, name)
	}
}
