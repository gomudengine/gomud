package connections

import "testing"

func TestConnTypeDefaultsToHuman(t *testing.T) {
	cd := NewConnectionDetails(1, nil, nil, nil)
	if cd.ConnType() != ConnHuman {
		t.Errorf("default ConnType should be ConnHuman, got %d", cd.ConnType())
	}
}

func TestConnTypeSetGet(t *testing.T) {
	cd := NewConnectionDetails(2, nil, nil, nil)
	cd.SetConnType(ConnAI)
	if cd.ConnType() != ConnAI {
		t.Errorf("ConnType should be ConnAI after SetConnType, got %d", cd.ConnType())
	}
}
