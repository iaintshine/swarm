package filter

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

// ConstraintFilter selects only nodes that match certain labels.
type ConstraintFilter struct {
}

// Name returns the name of the filter
func (f *ConstraintFilter) Name() string {
	return "constraint"
}

// Match is exported
func (f *ConstraintFilter) Match(config *cluster.ContainerConfig, node *node.Node) (bool, error) {
	// TODO: Consider how parsing expressions can be done only once
	//			for the whole batch of nodes, instead of compiling for every single call
	constraints, err := parseExprs(config.Constraints())
	if err != nil {
		return false, err
	}

	match := true

	// TODO: Add support for other logical operations.
	for _, constraint := range constraints {
		log.Debugf("matching constraint: %s %s %s", constraint.key, OPERATORS[constraint.operator], constraint.value)

		switch constraint.key {
		case "node":
			// "node" label is a special case pinning a container to a specific node.
			match = constraint.Match(node.ID, node.Name) || constraint.isSoft
		default:
			match = constraint.Match(node.Labels[constraint.key]) || constraint.isSoft
		}

		// Since the filtering policy represent logical AND, we break up as soon as a first condition fails
		if !match {
			break
		}
	}

	return match, nil
}

// String is exported
func (f *ConstraintFilter) String(config *cluster.ContainerConfig) string {
	constraints, err := parseExprs(config.Constraints())
	if err != nil {
		log.Errorf("unable to parse constraint expression due to %v", err)
		return ""
	}

	expressions := make([]string, len(constraints))
	for n, constraint := range constraints {
		expressions[n] = fmt.Sprintf("-e %s%s%s", constraint.key, OPERATORS[constraint.operator], constraint.value)
	}
	return strings.Join(expressions, " ")
}
