package pets

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/statmods"
)

func TestPetCloneOwnsInventoryAndAbilities(t *testing.T) {
	original := Pet{
		Level: 1,
		Items: []items.Item{
			{ItemId: 3001, Adjectives: []string{"old"}},
		},
		Abilities: []PetAbility{
			{
				Damage:   items.Damage{CritBuffIds: []int{1}},
				StatMods: statmods.StatMods{"speed": 2},
				BuffIds:  []int{3},
			},
		},
	}

	cloned := original.Clone()
	cloned.Items[0].Adjectives[0] = "new"
	cloned.Abilities[0].Damage.CritBuffIds[0] = 10
	cloned.Abilities[0].StatMods["speed"] = 20
	cloned.Abilities[0].BuffIds[0] = 30

	if original.Items[0].Adjectives[0] != "old" {
		t.Fatalf("original item adjective = %q, want old", original.Items[0].Adjectives[0])
	}
	if original.Abilities[0].Damage.CritBuffIds[0] != 1 {
		t.Fatalf("original crit buff id = %d, want 1", original.Abilities[0].Damage.CritBuffIds[0])
	}
	if original.Abilities[0].StatMods["speed"] != 2 {
		t.Fatalf("original speed stat mod = %d, want 2", original.Abilities[0].StatMods["speed"])
	}
	if original.Abilities[0].BuffIds[0] != 3 {
		t.Fatalf("original buff id = %d, want 3", original.Abilities[0].BuffIds[0])
	}
}
