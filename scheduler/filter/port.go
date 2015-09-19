// +build ignore
package filter

import (
	"fmt"
	"strings"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"github.com/samalba/dockerclient"
)

// PortFilter guarantees that, when scheduling a container binding a public
// port, only nodes that have not already allocated that same port will be
// considered.
type PortFilter struct {
}

// Name returns the name of the filter
func (p *PortFilter) Name() string {
	return "port"
}

// Filter is exported
func (p *PortFilter) Match(config *cluster.ContainerConfig, node *node.Node) (bool, error) {
	if config.HostConfig.NetworkMode == "host" {
		return p.filterHost(config, node)
	}

	return p.filterBridge(config, node)
}

// String is exported
func (p *PortFilter) String(config *cluster.ContainerConfig) string {
	if config.HostConfig.NetworkMode == "host" {
		return p.stringHost(config)
	}

	return p.stringBridge(config)
}

func (p *PortFilter) filterHost(config *cluster.ContainerConfig, node *node.Node) (bool, error) {
	for port := range config.ExposedPorts {
		if p.portAlreadyExposed(node, port) {
			return false, nil
		}
	}
	return true, nil
}

func (p *PortFilter) filterBridge(config *cluster.ContainerConfig, node *node.Node) (bool, error) {
	for _, port := range config.HostConfig.PortBindings {
		for _, binding := range port {
			if p.portAlreadyInUse(node, binding) {
				return false, nil
			}
		}
	}
	return true, nil
}

func (p *PortFilter) portAlreadyExposed(node *node.Node, requestedPort string) bool {
	for _, c := range node.Containers {
		if c.Info.HostConfig.NetworkMode == "host" {
			for port := range c.Info.Config.ExposedPorts {
				if port == requestedPort {
					return true
				}
			}
		}
	}
	return false
}

func (p *PortFilter) portAlreadyInUse(node *node.Node, requested dockerclient.PortBinding) bool {
	for _, c := range node.Containers {
		// HostConfig.PortBindings contains the requested ports.
		// NetworkSettings.Ports contains the actual ports.
		//
		// We have to check both because:
		// 1/ If the port was not specifically bound (e.g. -p 80), then
		//    HostConfig.PortBindings.HostPort will be empty and we have to check
		//    NetworkSettings.Port.HostPort to find out which port got dynamically
		//    allocated.
		// 2/ If the port was bound (e.g. -p 80:80) but the container is stopped,
		//    NetworkSettings.Port will be null and we have to check
		//    HostConfig.PortBindings to find out the mapping.

		if p.compare(requested, c.Info.HostConfig.PortBindings) || p.compare(requested, c.Info.NetworkSettings.Ports) {
			return true
		}
	}
	return false
}

func (p *PortFilter) compare(requested dockerclient.PortBinding, bindings map[string][]dockerclient.PortBinding) bool {
	for _, binding := range bindings {
		for _, b := range binding {
			if b.HostPort == "" {
				// Skip undefined HostPorts. This happens in bindings that
				// didn't explicitly specify an external port.
				continue
			}

			if b.HostPort == requested.HostPort {
				// Another container on the same host is binding on the same
				// port/protocol.  Verify if they are requesting the same
				// binding IP, or if the other container is already binding on
				// every interface.
				if requested.HostIp == b.HostIp || bindsAllInterfaces(requested) || bindsAllInterfaces(b) {
					return true
				}
			}
		}
	}
	return false
}

func (p *PortFilter) stringHost(config *cluster.ContainerConfig) string {
	if len(config.ExposedPorts) == 0 {
		return ""
	}

	options := []string{}
	for port := range config.ExposedPorts {
		options = append(options, "--expose="+port)
	}
	return "--net=host " + strings.Join(options, " ")
}

func (p *PortFilter) stringBridge(config *cluster.ContainerConfig) string {
	if len(config.HostConfig.PortBindings) == 0 {
		return ""
	}

	options := []string{}
	for _, port := range config.HostConfig.PortBindings {
		for _, binding := range port {
			options = append(options, fmt.Sprintf("-p %s:%s", port, binding.HostPort))
		}
	}
	return strings.Join(options, " ")
}

func bindsAllInterfaces(binding dockerclient.PortBinding) bool {
	return binding.HostIp == "0.0.0.0" || binding.HostIp == ""
}
