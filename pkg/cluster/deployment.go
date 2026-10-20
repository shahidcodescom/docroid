package cluster

import (
	"time"
)

// Deployment manages desired replica count and template specification for pods
type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int               `json:"replicas"`
	AvailableReplicas int               `json:"available_replicas"`
	Selector          map[string]string `json:"selector,omitempty"`
	Template          PodSpec           `json:"template"`
	ServicePort       int               `json:"service_port,omitempty"` // Port for cluster load balancer
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// DeploymentScaleRequest represents a request to scale a deployment
type DeploymentScaleRequest struct {
	Deployment string `json:"deployment"`
	Replicas   int    `json:"replicas"`
}
