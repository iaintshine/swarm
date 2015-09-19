// +build ignore
package filter

import (
	"fmt"
	"strings"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

// DependencyFilter co-schedules dependent containers on the same node.
type DependencyFilter struct {
}

// Name returns the name of the filter
func (f *DependencyFilter) Name() string {
	return "dependency"
}

// Filter is exported
func (f *DependencyFilter) Match(config *cluster.ContainerConfig, node *node.Node) (bool, error) {
	// Volumes
	volumes := []string{}
	for _, volume := range config.HostConfig.VolumesFrom {
		volumes = append(volumes, strings.SplitN(volume, ":", 2)[0])
	}

	// Extract containers from links.
	links := []string{}
	for _, link := range config.HostConfig.Links {
		links = append(links, strings.SplitN(link, ":", 2)[0])
	}

	// Check if --net points to a container.
	net := []string{}
	if strings.HasPrefix(config.HostConfig.NetworkMode, "container:") {
		net = append(net, strings.TrimPrefix(config.HostConfig.NetworkMode, "container:"))
	}

	return f.check(volumes, node) &&
		f.check(links, node) &&
		f.check(net, node), nil
}

// Get a string representation of the dependencies found in the container config.
func (f *DependencyFilter) String(config *cluster.ContainerConfig) string {
	dependencies := []string{}
	for _, volume := range config.HostConfig.VolumesFrom {
		dependencies = append(dependencies, fmt.Sprintf("--volumes-from=%s", volume))
	}
	for _, link := range config.HostConfig.Links {
		dependencies = append(dependencies, fmt.Sprintf("--link=%s", link))
	}
	if strings.HasPrefix(config.HostConfig.NetworkMode, "container:") {
		dependencies = append(dependencies, fmt.Sprintf("--net=%s", config.HostConfig.NetworkMode))
	}
	return strings.Join(dependencies, " ")
}

// Ensure that the node contains all dependent containers.
func (f *DependencyFilter) check(dependencies []string, node *node.Node) bool {
	for _, dependency := range dependencies {
		if node.Container(dependency) == nil {
			return false
		}
	}
	return true
}
