package filter

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

// Filter is exported
type Filter interface {
	// Returns a name of the filtering policy.
	Name() string

	// Returns true if node was accepted by the filtering policy.
	// Might return error e.g. in case of expression parsing error.
	Match(*cluster.ContainerConfig, *node.Node) (bool, error)

	// Returns a cli representation of the config, or an empty string.
	String(*cluster.ContainerConfig) string
}

var (
	filters []Filter
	// ErrNotSupported is exported
	ErrNotSupported = errors.New("filter not supported")
)

func init() {
	filters = []Filter{
		// &AffinityFilter{},
		&HealthFilter{},
		&ConstraintFilter{},
		// &PortFilter{},
		// &DependencyFilter{},
	}
}

// New is exported
func New(names []string) ([]Filter, error) {
	var selectedFilters []Filter

	for _, name := range names {
		found := false
		for _, filter := range filters {
			if filter.Name() == name {
				log.WithField("name", name).Debug("Initializing filter")
				selectedFilters = append(selectedFilters, filter)
				found = true
				break
			}
		}
		if !found {
			return nil, ErrNotSupported
		}
	}
	return selectedFilters, nil
}

// ApplyFilters applies a set of filters in batch and performs the logical AND.
func ApplyFilters(filters []Filter, config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error) {
	var err error

	for _, filter := range filters {
		nodes, err = ApplyFilter(filter, config, nodes)
		if err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// ApplyFilter applies a single filter to a set of nodes
func ApplyFilter(filter Filter, config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error) {
	candidates := []*node.Node{}
	for _, node := range nodes {
		if ok, err := filter.Match(config, node); err != nil {
			return nil, err
		} else if ok {
			candidates = append(candidates, node)
		}
	}
	return candidates, nil
}

// List returns the names of all the available filters
func List() []string {
	names := []string{}

	for _, filter := range filters {
		names = append(names, filter.Name())
	}

	return names
}
