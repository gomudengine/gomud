package combat

import (
	"math"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// expectedDPS computes the expected damage per round that attacker deals to defender,
// mirroring the logic in calculateCombat without any randomness.
func expectedDPS(atkChar characters.Character, defChar characters.Character) float64 {

	// Attack count: speed differential gives extra attacks, floored at 1.
	attackCount := int(math.Ceil(float64(atkChar.Stats.Speed.ValueAdj-defChar.Stats.Speed.ValueAdj) / 25))
	if attackCount < 1 {
		attackCount = 1
	}
	attackCount += atkChar.StatMod(`attacks`)
	if attackCount < 1 {
		attackCount = 1
	}

	statDmgBonus := atkChar.StatMod(`damage`)

	// Determine weapons the attacker uses, mirroring getAttackWeapons logic.
	attackWeapons := []items.Item{}
	if atkChar.Equipment.Weapon.ItemId > 0 {
		attackWeapons = append(attackWeapons, atkChar.Equipment.Weapon)
	}
	if atkChar.Equipment.Offhand.ItemId > 0 && atkChar.Equipment.Offhand.GetSpec().Type == items.Weapon {
		attackWeapons = append(attackWeapons, atkChar.Equipment.Offhand)
	}
	if len(attackWeapons) == 0 {
		attackWeapons = append(attackWeapons, items.Item{ItemId: 0})
	}

	// Resolve dual-wield: use expected number of weapons per attack.
	// At level 1 only one weapon fires; at level 2 50% chance for both; level 3+ always both.
	weaponWeight := make([]float64, len(attackWeapons))
	if len(attackWeapons) == 1 {
		weaponWeight[0] = 1.0
	} else {
		dwLevel := atkChar.GetSkillLevel(skills.DualWield)
		// Check if both are martial (claws) — always dual wield regardless of skill.
		alwaysDual := atkChar.Equipment.Weapon.GetSpec().Subtype == items.Claws &&
			atkChar.Equipment.Offhand.GetSpec().Subtype == items.Claws
		switch {
		case alwaysDual || dwLevel >= 3:
			weaponWeight[0] = 1.0
			weaponWeight[1] = 1.0
		case dwLevel == 2:
			weaponWeight[0] = 1.0
			weaponWeight[1] = 0.5 // 50% chance offhand fires
		default: // level 0 or 1: only main hand
			weaponWeight[0] = 1.0
			weaponWeight[1] = 0.0
		}
	}

	// Hit chance from speed stats (deterministic version of Hits).
	hitPct := float64(hitChance(atkChar.Stats.Speed.ValueAdj, defChar.Stats.Speed.ValueAdj)) / 100.0
	if hitPct < 0.05 {
		hitPct = 0.05
	}
	if hitPct > 0.95 {
		hitPct = 0.95
	}

	// Dual-wield hit penalty applied as a multiplier on hit chance.
	// We compute a weighted average: single-weapon rounds have no penalty,
	// dual-wield rounds have the penalty. For simplicity we apply it per weapon.
	dwLevel := atkChar.GetSkillLevel(skills.DualWield)
	dwPenalty := 0.0
	if len(attackWeapons) > 1 {
		if dwLevel < 4 {
			dwPenalty = 0.35
		} else {
			dwPenalty = 0.25
		}
	}

	// Crit chance (deterministic version of Crits).
	levelDiff := atkChar.Level - defChar.Level
	if levelDiff < 1 {
		levelDiff = 1
	}
	critPct := float64(5+int(math.Round(float64(atkChar.Stats.Strength.ValueAdj+atkChar.Stats.Speed.ValueAdj)/float64(levelDiff)))) / 100.0
	if critPct < 0.05 {
		critPct = 0.05
	}

	// Defense: expected reduction fraction applied to damage.
	// In calculateCombat: defenseAmt = Rand(defense), reduction = round(defenseAmt/100 * damage).
	// Expected defenseAmt = defense/2, so expected reduction fraction = defense/200.
	defenseFraction := float64(defChar.GetDefense()) / 200.0
	if defenseFraction > 0.95 {
		defenseFraction = 0.95
	}

	totalDPS := 0.0

	for roundIdx := 0; roundIdx < attackCount; roundIdx++ {
		for wIdx, weapon := range attackWeapons {
			wWeight := weaponWeight[wIdx]
			if wWeight <= 0 {
				continue
			}

			var attacks, dCount, dSides, dBonus int
			if weapon.ItemId > 0 {
				attacks, dCount, dSides, dBonus, _ = weapon.GetDiceRoll()
			} else {
				attacks, dCount, dSides, dBonus, _ = atkChar.GetDefaultDiceRoll()
			}
			dBonus += statDmgBonus

			// Average roll of NdS = N*(S+1)/2
			avgRoll := float64(dCount) * float64(dSides+1) / 2.0
			avgDmg := avgRoll + float64(dBonus)
			if avgDmg < 0 {
				avgDmg = 0
			}

			// Crit adds another full roll: dCount*dSides + dBonus on top of the base hit.
			critBonus := float64(dCount*dSides+dBonus) * critPct

			// Effective hit chance for this weapon slot.
			effHit := hitPct
			if wIdx > 0 {
				effHit = math.Max(0.05, hitPct-dwPenalty)
			}

			rawDmg := (avgDmg + critBonus) * effHit
			netDmg := rawDmg * (1.0 - defenseFraction)

			for atkIdx := 0; atkIdx < attacks; atkIdx++ {
				totalDPS += netDmg * wWeight
			}
		}
	}

	return totalDPS
}

// CombatOdds returns the ratio of (rounds for attacker to kill defender) to
// (rounds for defender to kill attacker). A value > 1 means the attacker wins
// faster; < 1 means the defender wins faster.
//
// The returned value is suitable for the consider command: values well above 1.0
// are favorable for the attacker; values near or below 1.0 are dangerous.
func CombatOdds(atkChar characters.Character, defChar characters.Character) float64 {
	atkDPS := expectedDPS(atkChar, defChar)
	defDPS := expectedDPS(defChar, atkChar)

	defHP := float64(defChar.Health)
	if defHP < 1 {
		defHP = 1
	}
	atkHP := float64(atkChar.Health)
	if atkHP < 1 {
		atkHP = 1
	}

	if atkDPS < 0.001 {
		return 0
	}

	atkRoundsToKill := defHP / atkDPS

	if defDPS < 0.001 {
		return math.MaxFloat64
	}

	defRoundsToKill := atkHP / defDPS

	// ratio > 1 means attacker kills faster than defender kills attacker.
	return defRoundsToKill / atkRoundsToKill
}

func ChanceToTame(s *users.UserRecord, t *mobs.Mob) int {

	var MOD_SKILL_MIN int = 1   // Minimum base tame ability
	var MOD_SKILL_MAX int = 100 // Maximum base tame ability

	var MOD_SIZE_SMALL int = 0    // Modifier for small creatures
	var MOD_SIZE_MEDIUM int = -10 // Modifier for medium creatures
	var MOD_SIZE_LARGE int = -25  // Modifier for large creatures

	var MOD_LEVELDIFF_MIN int = -25 // Lowest level delta modifier
	var MOD_LEVELDIFF_MAX int = 25  // Highest level delta modifier

	var MOD_HEALTHPERCENT_MAX float64 = 50 // Highest possible bonus for target HP being reduced

	var FACTOR_IS_AGGRO float64 = .50 // Overall reduction of chance if target is aggro

	proficiencyModifier := s.Character.MobMastery.GetTame(int(t.MobId))

	if proficiencyModifier < MOD_SKILL_MIN {
		proficiencyModifier = MOD_SKILL_MIN
	} else if proficiencyModifier > MOD_SKILL_MAX {
		proficiencyModifier = MOD_SKILL_MAX
	}

	raceInfo := races.GetRace(s.Character.RaceId)

	sizeModifier := 0
	switch raceInfo.Size {
	case races.Large:
		sizeModifier = MOD_SIZE_LARGE
	case races.Small:
		sizeModifier = MOD_SIZE_SMALL
	case races.Medium:
	default:
		sizeModifier = MOD_SIZE_MEDIUM
	}

	levelDiff := s.Character.Level - t.Character.Level
	if levelDiff > MOD_LEVELDIFF_MAX {
		levelDiff = MOD_LEVELDIFF_MAX
	} else if levelDiff < MOD_LEVELDIFF_MIN {
		levelDiff = MOD_LEVELDIFF_MIN
	}

	healthModifier := MOD_HEALTHPERCENT_MAX - math.Ceil(float64(s.Character.Health)/float64(s.Character.HealthMax.Value)*MOD_HEALTHPERCENT_MAX)

	var aggroModifier float64 = 1
	if t.Character.IsAggro(s.UserId, 0) {
		aggroModifier = FACTOR_IS_AGGRO
	}

	return int(math.Ceil((float64(proficiencyModifier) + float64(levelDiff) + healthModifier + float64(sizeModifier)) * aggroModifier))
}

func AlignmentChange(killerAlignment int8, killedAlignment int8) int {

	isKillerGood := killerAlignment > characters.AlignmentNeutralHigh
	isKillerEvil := killerAlignment < characters.AlignmentNeutralLow
	isKillerNeutral := killerAlignment >= characters.AlignmentNeutralLow && killerAlignment <= characters.AlignmentNeutralHigh

	isKilledGood := killedAlignment > characters.AlignmentNeutralHigh
	isKilledEvil := killedAlignment < characters.AlignmentNeutralLow
	isKilledNeutral := killedAlignment >= characters.AlignmentNeutralLow && killedAlignment <= characters.AlignmentNeutralHigh

	// Normalize the delta to positive, then half, so 0-100
	deltaAbs := math.Abs(math.Max(float64(killerAlignment), float64(killedAlignment))-math.Min(float64(killerAlignment), float64(killedAlignment))) * 0.5

	changeAmt := 0
	if deltaAbs <= 10 {
		changeAmt = 0
	} else if deltaAbs <= 30 {
		changeAmt = 1
	} else if deltaAbs <= 60 {
		changeAmt = 2
	} else if deltaAbs <= 80 {
		changeAmt = 3
	} else {
		changeAmt = 4
	}

	factor := 0

	if isKillerGood {

		if isKilledGood { // good vs good is especially evil
			factor = -2
			changeAmt = int(math.Max(float64(changeAmt), 1)) // At least 1 when killing own kind
		} else if isKilledEvil { // good vs evil is good
			factor = 1
		} else if isKilledNeutral { // good vs neutral is evil
			factor = -1
		}

	} else if isKillerEvil {

		if isKilledGood { // evil vs good is evil
			factor = -1
		} else if isKilledEvil { // evil vs evil is especially good
			factor = 2
			changeAmt = int(math.Max(float64(changeAmt), 1)) // At least 1 when killing own kind
		} else if isKilledNeutral { // evil vs neutral is evil
			factor = -1
		}

	} else if isKillerNeutral {

		if isKilledGood { // neutral vs good is evil
			factor = -1
		} else if isKilledEvil { // neutral vs evil is good
			factor = 1
		} else if isKilledNeutral { // neutral vs evil is nothing
			factor = 0
		}

	}

	return factor * changeAmt
}
