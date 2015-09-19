package filter

import (
	"errors"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

var (
	// ErrNoHealthyNodeAvailable is exported
	ErrNoHealthyNodeAvailable = errors.New("No healthy node available in the cluster")
)

// HealthFilter only schedules containers on healthy nodes.
type HealthFilter struct {
}

// Name returns the name of the filter
func (f *HealthFilter) Name() string {
	return "health"
}

// Match is exported
func (f *HealthFilter) Match(_ *cluster.ContainerConfig, node *node.Node) (bool, error) {
	return node.IsHealthy, nil
}

// String is exported
func (f *HealthFilter) String(_ *cluster.ContainerConfig) string {
	return ""
}
