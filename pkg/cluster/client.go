package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// JoinRequest is sent by a worker node to join the cluster
type JoinRequest struct {
	Token string `json:"token"`
	Node  Node   `json:"node"`
}

// JoinResponse is returned by the control plane upon successful join
type JoinResponse struct {
	ClusterName string  `json:"cluster_name"`
	LeaderAddr  string  `json:"leader_addr"`
	AssignedID  string  `json:"assigned_id"`
	Nodes       []*Node `json:"nodes"`
}

// HeartbeatRequest is sent periodically by worker nodes to the control plane
type HeartbeatRequest struct {
	NodeID        string            `json:"node_id"`
	Token         string            `json:"token"`
	AllocatedPods int               `json:"allocated_pods"`
	PodStatuses   map[string]string `json:"pod_statuses,omitempty"` // podID -> PodStatus
}

// HeartbeatResponse is returned by the control plane
type HeartbeatResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// RunPodRequest instructs a node to run a pod locally
type RunPodRequest struct {
	Token string  `json:"token"`
	Pod   Pod     `json:"pod"`
}

// RunPodResponse is returned by a node after running a pod
type RunPodResponse struct {
	PodID      string `json:"pod_id"`
	LocalAppID string `json:"local_app_id"`
	HostPort   int    `json:"host_port"`
	PID        int    `json:"pid"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

// StopPodRequest instructs a node to stop a pod
type StopPodRequest struct {
	Token string `json:"token"`
	PodID string `json:"pod_id"`
}

// Client facilitates inter-device cluster communications over HTTP
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new cluster HTTP client with sane timeouts
func NewClient(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func normalizeURL(addr, path string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	addr = strings.TrimRight(addr, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return addr + path
}

// JoinCluster registers this worker device with the cluster control plane
func (c *Client) JoinCluster(ctx context.Context, leaderAddr string, req *JoinRequest) (*JoinResponse, error) {
	endpoint := normalizeURL(leaderAddr, "/api/cluster/join")
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.Token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to contact cluster leader at %s: %w", leaderAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cluster join rejected (status %d): %s", resp.StatusCode, string(respBody))
	}

	var joinResp JoinResponse
	if err := json.NewDecoder(resp.Body).Decode(&joinResp); err != nil {
		return nil, fmt.Errorf("failed to decode join response: %w", err)
	}
	return &joinResp, nil
}

// SendHeartbeat sends a heartbeat from a worker to the control plane
func (c *Client) SendHeartbeat(ctx context.Context, leaderAddr string, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	endpoint := normalizeURL(leaderAddr, "/api/cluster/heartbeat")
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.Token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("heartbeat error status: %d", resp.StatusCode)
	}

	var hbResp HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&hbResp); err != nil {
		return nil, err
	}
	return &hbResp, nil
}

// DispatchPod asks a remote node to launch a pod
func (c *Client) DispatchPod(ctx context.Context, nodeAddr string, token string, pod *Pod) (*RunPodResponse, error) {
	endpoint := normalizeURL(nodeAddr, "/api/cluster/node/run-pod")
	req := RunPodRequest{
		Token: token,
		Pod:   *pod,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("dispatch to node %s failed: %w", nodeAddr, err)
	}
	defer resp.Body.Close()

	var runResp RunPodResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		return nil, fmt.Errorf("invalid response from node: %w", err)
	}

	if resp.StatusCode != http.StatusOK || runResp.Error != "" {
		if runResp.Error != "" {
			return nil, fmt.Errorf("remote node error: %s", runResp.Error)
		}
		return nil, fmt.Errorf("node responded with status %d", resp.StatusCode)
	}

	return &runResp, nil
}

// StopPod asks a remote node to stop a pod
func (c *Client) StopPod(ctx context.Context, nodeAddr string, token string, podID string) error {
	endpoint := normalizeURL(nodeAddr, "/api/cluster/node/stop-pod")
	req := StopPodRequest{
		Token: token,
		PodID: podID,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("stop pod on node %s failed: %w", nodeAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("node stop-pod failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// FetchPodLogs retrieves logs for a pod running on a remote node
func (c *Client) FetchPodLogs(ctx context.Context, nodeAddr string, token string, podID string, lines int) (string, error) {
	endpoint := normalizeURL(nodeAddr, fmt.Sprintf("/api/cluster/node/pod-logs?pod_id=%s&lines=%d", podID, lines))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("fetch logs from node %s failed: %w", nodeAddr, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
