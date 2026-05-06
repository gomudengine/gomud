package pets

import (
	"maps"
	"slices"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/statmods"
)

type PetAbility struct {
	LevelGranted int               `yaml:"levelgranted,omitempty"`
	CombatChance int               `yaml:"combatchance,omitempty"` // odds (out of 100) that it will join in this round of combat
	Damage       items.Damage      `yaml:"damage,omitempty"`
	StatMods     statmods.StatMods `yaml:"statmods,omitempty"`
	BuffIds      []int             `yaml:"buffids,omitempty"`
	Capacity     int               `yaml:"capacity,omitempty"`
}

func (pa PetAbility) Clone() PetAbility {
	pa.Damage = pa.Damage.Clone()
	pa.StatMods = maps.Clone(pa.StatMods)
	pa.BuffIds = slices.Clone(pa.BuffIds)
	return pa
}
