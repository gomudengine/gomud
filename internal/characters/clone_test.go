package characters

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/items"
)

func TestCharacterCloneOwnsNestedMutableState(t *testing.T) {
	original := New()
	original.Shop = Shop{{ItemId: 4001, Quantity: 1}}
	original.Cooldowns = Cooldowns{"cast": 3}
	original.KD = KDStats{
		Kills:        map[int]int{101: 2},
		PlayerKills:  map[string]int{"1:player": 1},
		PlayerDeaths: map[string]int{"2:player": 1},
	}
	original.MobMastery = MobMasteries{Tame: map[int]int{201: 4}}
	original.ZonesVisited = map[string]RoomBitset{"start": {1: 1}}
	original.Equipment.Weapon = items.Item{ItemId: 3001, Adjectives: []string{"old"}}

	cloned := original.Clone()
	cloned.Shop[0].Quantity = 5
	cloned.Cooldowns["cast"] = 9
	cloned.KD.Kills[101] = 7
	cloned.KD.PlayerKills["1:player"] = 8
	cloned.KD.PlayerDeaths["2:player"] = 9
	cloned.MobMastery.Tame[201] = 6
	cloned.ZonesVisited["start"].Set(130)
	cloned.Equipment.Weapon.Adjectives[0] = "new"

	if original.Shop[0].Quantity != 1 {
		t.Fatalf("original shop quantity = %d, want 1", original.Shop[0].Quantity)
	}
	if original.Cooldowns["cast"] != 3 {
		t.Fatalf("original cooldown = %d, want 3", original.Cooldowns["cast"])
	}
	if original.KD.Kills[101] != 2 {
		t.Fatalf("original mob kills = %d, want 2", original.KD.Kills[101])
	}
	if original.KD.PlayerKills["1:player"] != 1 {
		t.Fatalf("original player kills = %d, want 1", original.KD.PlayerKills["1:player"])
	}
	if original.KD.PlayerDeaths["2:player"] != 1 {
		t.Fatalf("original player deaths = %d, want 1", original.KD.PlayerDeaths["2:player"])
	}
	if original.MobMastery.Tame[201] != 4 {
		t.Fatalf("original tame mastery = %d, want 4", original.MobMastery.Tame[201])
	}
	if original.ZonesVisited["start"].Has(130) {
		t.Fatal("original zone visit bitset changed after mutating clone")
	}
	if original.Equipment.Weapon.Adjectives[0] != "old" {
		t.Fatalf("original weapon adjective = %q, want old", original.Equipment.Weapon.Adjectives[0])
	}
}
