package connections

import "testing"

func TestAICommandAllowedWithinLimit(t *testing.T) {
	cd := NewConnectionDetails(10, nil, nil, nil)
	if !cd.AICommandAllowed(1, 2) {
		t.Error("1st command in round 1 should be allowed")
	}
	if !cd.AICommandAllowed(1, 2) {
		t.Error("2nd command in round 1 should be allowed")
	}
	if cd.AICommandAllowed(1, 2) {
		t.Error("3rd command in round 1 should be denied (max 2)")
	}
}

func TestAICommandAllowedResetsNextRound(t *testing.T) {
	cd := NewConnectionDetails(11, nil, nil, nil)
	cd.AICommandAllowed(1, 1) // uses the round-1 budget
	if cd.AICommandAllowed(1, 1) {
		t.Error("2nd command in round 1 should be denied (max 1)")
	}
	if !cd.AICommandAllowed(2, 1) {
		t.Error("1st command in round 2 should be allowed after reset")
	}
}
