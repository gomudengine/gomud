package configs

import "testing"

func TestNetworkValidateAIDefaults(t *testing.T) {
	n := &Network{
		AIPort:             -1,
		MaxAIConnections:   0,
		AICommandsPerRound: 0,
	}
	n.Validate()

	if int(n.AIPort) != 0 {
		t.Errorf("AIPort: negative should clamp to 0 (disabled), got %d", int(n.AIPort))
	}
	if int(n.MaxAIConnections) != 20 {
		t.Errorf("MaxAIConnections: <1 should default to 20, got %d", int(n.MaxAIConnections))
	}
	if int(n.AICommandsPerRound) != 2 {
		t.Errorf("AICommandsPerRound: <1 should default to 2, got %d", int(n.AICommandsPerRound))
	}
}

func TestNetworkValidateAIPreservesValidValues(t *testing.T) {
	n := &Network{
		AIPort:             55555,
		MaxAIConnections:   10,
		AICommandsPerRound: 5,
	}
	n.Validate()

	if int(n.AIPort) != 55555 || int(n.MaxAIConnections) != 10 || int(n.AICommandsPerRound) != 5 {
		t.Errorf("Validate must not overwrite valid values: got AIPort=%d Max=%d Cmds=%d",
			int(n.AIPort), int(n.MaxAIConnections), int(n.AICommandsPerRound))
	}
}
