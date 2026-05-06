package items

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/statmods"
)

func TestItemCloneOwnsOverrideSpec(t *testing.T) {
	original := Item{
		ItemId:     3001,
		Adjectives: []string{"old"},
		Spec: &ItemSpec{
			BuffIds:     []int{1},
			WornBuffIds: []int{2},
			Damage:      Damage{CritBuffIds: []int{3}},
			StatMods:    statmods.StatMods{"strength": 4},
		},
	}

	cloned := original.Clone()
	cloned.Adjectives[0] = "new"
	cloned.Spec.BuffIds[0] = 10
	cloned.Spec.WornBuffIds[0] = 20
	cloned.Spec.Damage.CritBuffIds[0] = 30
	cloned.Spec.StatMods["strength"] = 40

	if original.Adjectives[0] != "old" {
		t.Fatalf("original adjective = %q, want old", original.Adjectives[0])
	}
	if original.Spec.BuffIds[0] != 1 {
		t.Fatalf("original buff id = %d, want 1", original.Spec.BuffIds[0])
	}
	if original.Spec.WornBuffIds[0] != 2 {
		t.Fatalf("original worn buff id = %d, want 2", original.Spec.WornBuffIds[0])
	}
	if original.Spec.Damage.CritBuffIds[0] != 3 {
		t.Fatalf("original crit buff id = %d, want 3", original.Spec.Damage.CritBuffIds[0])
	}
	if original.Spec.StatMods["strength"] != 4 {
		t.Fatalf("original strength stat mod = %d, want 4", original.Spec.StatMods["strength"])
	}
}
