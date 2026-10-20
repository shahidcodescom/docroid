package cluster

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// SchedulingStrategy defines the algorithm used for pod node placement
type SchedulingStrategy string

const (
	StrategyLeastAllocated SchedulingStrategy = "least-allocated"
	StrategyRoundRobin     SchedulingStrategy = "round-robin"
)

// Scheduler decides which node runs each pod replica
type Scheduler struct {
	strategy         SchedulingStrategy
	heartbeatTimeout time.Duration
	roundRobinIdx    int
	mu               sync.Mutex
}

// NewScheduler creates a new scheduler with the specified strategy and timeout
func NewScheduler(strategy SchedulingStrategy, heartbeatTimeout time.Duration) *Scheduler {
	if strategy == "" {
		strategy = StrategyLeastAllocated
	}
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 30 * time.Second
	}
	return &Scheduler{
		strategy:         strategy,
		heartbeatTimeout: heartbeatTimeout,
	}
}

// SelectNode picks the best candidate node for a given PodSpec
func (s *Scheduler) SelectNode(spec *PodSpec, nodes []*Node) (*Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Filter phase
	var eligible []*Node
	for _, n := range nodes {
		// Must be Ready and alive
		if n.Status != NodeStatusReady || !n.IsAlive(s.heartbeatTimeout) {
			continue
		}
		// Check capacity
		if n.Capacity.MaxPods > 0 && n.AllocatedPods >= n.Capacity.MaxPods {
			continue
		}
		// Check label selector
		if !n.MatchesSelector(spec.NodeSelector) {
			continue
		}
		eligible = append(eligible, n)
	}

	if len(eligible) == 0 {
		return nil, fmt.Errorf("no eligible nodes available (healthy and matching selector)")
	}

	// 2. Prioritize / Score phase
	switch s.strategy {
	case StrategyRoundRobin:
		idx := s.roundRobinIdx % len(eligible)
		s.roundRobinIdx++
		return eligible[idx], nil

	case StrategyLeastAllocated:
		fallthrough
	default:
		// Sort by fewest allocated pods first
		sort.SliceStable(eligible, func(i, j int) bool {
			if eligible[i].AllocatedPods != eligible[j].AllocatedPods {
				return eligible[i].AllocatedPods < eligible[j].AllocatedPods
			}
			// Prefer local node as tie-breaker for lower network latency
			return eligible[i].IsLocal
		})
		return eligible[0], nil
	}
}

// FindEvictablePods identifies pods that need rescheduling due to dead/lost nodes
func (s *Scheduler) FindEvictablePods(nodes []*Node, pods []*Pod) []*Pod {
	nodeHealthMap := make(map[string]bool)
	for _, n := range nodes {
		nodeHealthMap[n.ID] = (n.Status == NodeStatusReady && n.IsAlive(s.heartbeatTimeout))
	}

	var evictable []*Pod
	for _, p := range pods {
		if p.Status == PodRunning || p.Status == PodPending {
			healthy, exists := nodeHealthMap[p.NodeID]
			if !exists || !healthy {
				evictable = append(evictable, p)
			}
		}
	}
	return evictable
}
