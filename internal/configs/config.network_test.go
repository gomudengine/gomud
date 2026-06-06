package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkValidateAIDefaults(t *testing.T) {
	n := &Network{
		AIPort:             -1,
		MaxAIConnections:   0,
		AICommandsPerRound: 0,
	}
	n.Validate()

	assert.Equal(t, 0, int(n.AIPort), "negative AIPort should clamp to 0 (disabled)")
	assert.Equal(t, 20, int(n.MaxAIConnections), "MaxAIConnections <1 should default to 20")
	assert.Equal(t, 2, int(n.AICommandsPerRound), "AICommandsPerRound <1 should default to 2")
}

func TestNetworkValidateAIPreservesValidValues(t *testing.T) {
	n := &Network{
		AIPort:             55555,
		MaxAIConnections:   10,
		AICommandsPerRound: 5,
	}
	n.Validate()

	assert.Equal(t, 55555, int(n.AIPort))
	assert.Equal(t, 10, int(n.MaxAIConnections))
	assert.Equal(t, 5, int(n.AICommandsPerRound))
}
