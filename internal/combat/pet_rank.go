package combat

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/pets"
)

// PetRank holds the computed metrics for a single pet type across all levels.
type PetRank struct {
	Type string
	Name string

	// Per-level ability snapshots (index 0 = level 1, etc.)
	Levels []PetLevelSnapshot

	// Peak values across all levels
	PeakDPS       float64
	PeakMaxDmg    int
	PeakStatTotal int
	PeakCapacity  int
	PeakBuffCount int
	PeakBuffValue int

	// Aggregate scores
	CombatScore  float64 // weighted combat effectiveness across levels
	UtilityScore float64 // stat mods + buffs + capacity across levels
	OverallScore float64 // combined
}

// PetLevelSnapshot captures a pet's effective ability at a given level.
type PetLevelSnapshot struct {
	Level        int
	CombatChance int
	DiceRoll     string
	AvgDmg       float64
	MaxDmg       int
	DPS          float64 // avgDmg * combatChance/100
	StatMods     map[string]int
	StatTotal    int
	BuffCount    int
	BuffValue    int
	BuffNames    []string
	Capacity     int
}

// RankPets computes ranking metrics for every loaded pet type and returns
// slices sorted by three criteria:
//   - byCombat   – weighted combat effectiveness across all levels
//   - byUtility  – stat mods + buffs + capacity value across all levels
//   - byOverall  – combined score
func RankPets() (byCombat, byUtility, byOverall []PetRank) {
	allSpecs := pets.GetAllPetSpecs()

	ranks := make([]PetRank, 0, len(allSpecs))

	for _, spec := range allSpecs {
		p := spec
		p.Validate()

		rank := PetRank{
			Type:   p.Type,
			Name:   p.Name,
			Levels: make([]PetLevelSnapshot, 10),
		}

		var combatSum, utilitySum float64

		for lvl := 1; lvl <= 10; lvl++ {
			p.Level = lvl
			p.Validate()

			snap := PetLevelSnapshot{Level: lvl}

			combatChance, dmg := p.GetEffectiveDamage()
			snap.CombatChance = combatChance

			if dmg.DiceCount > 0 && dmg.SideCount > 0 {
				attacks := dmg.Attacks
				if attacks < 1 {
					attacks = 1
				}
				snap.DiceRoll = dmg.DiceRoll
				snap.AvgDmg = float64(attacks) * (float64(dmg.DiceCount)*float64(dmg.SideCount+1)/2.0 + float64(dmg.BonusDamage))
				snap.MaxDmg = attacks * (dmg.DiceCount*dmg.SideCount + dmg.BonusDamage)
				snap.DPS = snap.AvgDmg * float64(combatChance) / 100.0
			}

			sm := p.GetEffectiveStatMods()
			snap.StatMods = map[string]int(sm)
			for _, v := range sm {
				snap.StatTotal += int(math.Abs(float64(v)))
			}

			for _, bId := range p.GetEffectiveBuffs() {
				if bs := buffs.GetBuffSpec(bId); bs != nil {
					snap.BuffCount++
					snap.BuffValue += bs.GetValue()
					name, _ := bs.VisibleNameDesc()
					snap.BuffNames = append(snap.BuffNames, name)
				}
			}

			snap.Capacity = p.GetEffectiveCapacity()

			rank.Levels[lvl-1] = snap

			// Track peaks
			if snap.DPS > rank.PeakDPS {
				rank.PeakDPS = snap.DPS
			}
			if snap.MaxDmg > rank.PeakMaxDmg {
				rank.PeakMaxDmg = snap.MaxDmg
			}
			if snap.StatTotal > rank.PeakStatTotal {
				rank.PeakStatTotal = snap.StatTotal
			}
			if snap.Capacity > rank.PeakCapacity {
				rank.PeakCapacity = snap.Capacity
			}
			if snap.BuffCount > rank.PeakBuffCount {
				rank.PeakBuffCount = snap.BuffCount
			}
			if snap.BuffValue > rank.PeakBuffValue {
				rank.PeakBuffValue = snap.BuffValue
			}

			// Higher levels matter more — weight by level
			weight := float64(lvl) / 5.5 // normalizes so avg weight ≈ 1.0
			combatSum += snap.DPS * weight
			utilitySum += (float64(snap.StatTotal) + float64(snap.BuffValue)*0.5 + float64(snap.Capacity)*2.0) * weight
		}

		rank.CombatScore = combatSum / 10.0
		rank.UtilityScore = utilitySum / 10.0
		rank.OverallScore = rank.CombatScore*3.0 + rank.UtilityScore

		ranks = append(ranks, rank)
	}

	byCombat = make([]PetRank, len(ranks))
	copy(byCombat, ranks)
	sort.Slice(byCombat, func(i, j int) bool {
		return byCombat[i].CombatScore > byCombat[j].CombatScore
	})

	byUtility = make([]PetRank, len(ranks))
	copy(byUtility, ranks)
	sort.Slice(byUtility, func(i, j int) bool {
		return byUtility[i].UtilityScore > byUtility[j].UtilityScore
	})

	byOverall = make([]PetRank, len(ranks))
	copy(byOverall, ranks)
	sort.Slice(byOverall, func(i, j int) bool {
		return byOverall[i].OverallScore > byOverall[j].OverallScore
	})

	return byCombat, byUtility, byOverall
}

// FormatPetRankings returns a human-readable table of all three ranking
// views, suitable for logging or printing to a terminal.
func FormatPetRankings() string {
	byCombat, byUtility, byOverall := RankPets()

	var sb strings.Builder

	writeTable := func(title string, rows []PetRank, scoreLabel string, score func(PetRank) string) {
		fmt.Fprintf(&sb, "\n=== %s ===\n", title)
		fmt.Fprintf(&sb, "%-4s %-16s %-10s %-10s %-8s %-8s %-8s %s\n",
			"Rank", "Type", "PeakDPS", "PeakStats", "PeakCap", "Buffs", "BuffVal", scoreLabel)
		fmt.Fprintln(&sb, strings.Repeat("-", 90))
		for i, r := range rows {
			fmt.Fprintf(&sb, "%-4d %-16s %-10.2f %-10d %-8d %-8d %-8d %s\n",
				i+1, r.Type, r.PeakDPS, r.PeakStatTotal, r.PeakCapacity, r.PeakBuffCount, r.PeakBuffValue, score(r))
		}
	}

	writeTable("Ranked by Combat Score", byCombat, "Combat", func(r PetRank) string {
		return fmt.Sprintf("%.3f", r.CombatScore)
	})
	writeTable("Ranked by Utility Score", byUtility, "Utility", func(r PetRank) string {
		return fmt.Sprintf("%.3f", r.UtilityScore)
	})
	writeTable("Ranked by Overall Score", byOverall, "Overall", func(r PetRank) string {
		return fmt.Sprintf("%.3f", r.OverallScore)
	})

	return sb.String()
}
