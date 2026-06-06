package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkValidateAIDefaults(t *testing.T) {
	n := &Network{
		AI: AINetwork{Port: -1, MaxConnections: 0, CommandsPerRound: 0},
	}
	n.Validate()

	assert.Equal(t, 0, int(n.AI.Port), "negative AI.Port should clamp to 0 (disabled)")
	assert.Equal(t, 20, int(n.AI.MaxConnections), "AI.MaxConnections <1 should default to 20")
	assert.Equal(t, 2, int(n.AI.CommandsPerRound), "AI.CommandsPerRound <1 should default to 2")
}

func TestNetworkValidateAIPreservesValidValues(t *testing.T) {
	n := &Network{
		AI: AINetwork{Port: 55555, MaxConnections: 10, CommandsPerRound: 5},
	}
	n.Validate()

	assert.Equal(t, 55555, int(n.AI.Port))
	assert.Equal(t, 10, int(n.AI.MaxConnections))
	assert.Equal(t, 5, int(n.AI.CommandsPerRound))
}
