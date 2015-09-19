package filter

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

// AffinityFilter selects only nodes based on other containers on the node.
type AffinityFilter struct {
}

// Name returns the name of the filter
func (f *AffinityFilter) Name() string {
	return "affinity"
}

// Filter is exported
func (f *AffinityFilter) Match(config *cluster.ContainerConfig, node *node.Node) (bool, error) {
	// TODO: Consider how parsing expressions can be done only once
	//			for the whole batch of nodes, instead of compiling for every single call
	affinities, err := parseExprs(config.Affinities())
	if err != nil {
		return false, err
	}

	match := true

	// TODO: Add support for other logical operations.
	for _, affinity := range affinities {
		log.Debugf("matching affinity: %s%s%s", affinity.key, OPERATORS[affinity.operator], affinity.value)

		switch affinity.key {
		case "container":
			containers := []string{}
			for _, container := range node.Containers {
				containers = append(containers, container.Id, strings.TrimPrefix(container.Names[0], "/"))
			}
			match = affinity.Match(containers...)
		case "image":
			images := []string{}
			for _, image := range node.Images {
				images = append(images, image.Id)
				images = append(images, image.RepoTags...)
				for _, tag := range image.RepoTags {
					images = append(images, strings.Split(tag, ":")[0])
				}
			}
			match = affinity.Match(images...)
		default:
			labels := []string{}
			for _, container := range node.Containers {
				labels = append(labels, container.Labels[affinity.key])
			}
			match = affinity.Match(labels...)
		}

		// Since the filtering policy represent logical AND, we break up as soon as a first condition fails
		if !match && !affinity.isSoft {
			break
		}
	}
	return match, nil
}

// String is exported
func (f *AffinityFilter) String(config *cluster.ContainerConfig) string {
	affinities, err := parseExprs(config.Affinities())
	if err != nil {
		log.Errorf("unable to parse affinity expression due to %v", err)
		return ""
	}

	expressions := make([]string, len(affinities))
	for n, affinity := range affinities {
		expressions[n] = fmt.Sprintf("-e affinity:%s%s%s", affinity.key, OPERATORS[affinity.operator], affinity.value)
	}
	return strings.Join(expressions, " ")
}
