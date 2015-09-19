package filter

import (
	"testing"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	"github.com/stretchr/testify/assert"
)

func TestHealthyFilter(t *testing.T) {
	var (
		f = HealthFilter{}
	)

	cases := []struct {
		node     *node.Node
		expected bool
	}{
		{
			&node.Node{
				ID:        "node-0-id",
				Name:      "node-0-name",
				IsHealthy: false,
			},
			false,
		},
		{
			&node.Node{
				ID:        "node-1-id",
				Name:      "node-1-name",
				IsHealthy: true,
			},
			true,
		},
	}

	for _, test := range cases {
		match, err := f.Match(&cluster.ContainerConfig{}, test.node)

		assert.NoError(t, err)
		assert.Equal(t, match, test.expected)
	}
}
