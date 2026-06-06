package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAICommandAllowedWithinLimit(t *testing.T) {
	cd := NewConnectionDetails(10, nil, nil, nil)
	assert.True(t, cd.AICommandAllowed(1, 2), "1st command in round 1 should be allowed")
	assert.True(t, cd.AICommandAllowed(1, 2), "2nd command in round 1 should be allowed")
	assert.False(t, cd.AICommandAllowed(1, 2), "3rd command in round 1 should be denied (max 2)")
}

func TestAICommandAllowedResetsNextRound(t *testing.T) {
	cd := NewConnectionDetails(11, nil, nil, nil)
	cd.AICommandAllowed(1, 1) // uses the round-1 budget
	assert.False(t, cd.AICommandAllowed(1, 1), "2nd command in round 1 should be denied (max 1)")
	assert.True(t, cd.AICommandAllowed(2, 1), "1st command in round 2 should be allowed after reset")
}
