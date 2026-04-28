package combat

import (
	"math"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// resolveAttackWeapons returns the candidate weapon list for a character,
// applying the same selection logic used by calculateCombat.
// It does not trim for dual-wield skill — callers handle that themselves.
func resolveAttackWeapons(char characters.Character) []items.Item {
	attackWeapons := []items.Item{}
	if char.Equipment.Weapon.ItemId > 0 {
		attackWeapons = append(attackWeapons, char.Equipment.Weapon)
	}
	if char.Equipment.Offhand.ItemId > 0 && char.Equipment.Offhand.GetSpec().Type == items.Weapon {
		attackWeapons = append(attackWeapons, char.Equipment.Offhand)
	}
	if len(attackWeapons) == 0 {
		attackWeapons = append(attackWeapons, items.Item{ItemId: 0})
	}
	return attackWeapons
}

// attackCount returns the number of attack iterations given raw stat values.
func attackCount(atkSpd, defSpd, attacksMod int) int {
	count := int(math.Ceil(float64(atkSpd-defSpd) / 25))
	if count < 1 {
		count = 1
	}
	count += attacksMod
	if count < 1 {
		count = 1
	}
	return count
}

// combatAttackCount is a thin wrapper around attackCount that extracts the
// required fields from full Character values, preserving existing call sites.
func combatAttackCount(sourceChar characters.Character, targetChar characters.Character) int {
	return attackCount(
		sourceChar.Stats.Speed.ValueAdj,
		targetChar.Stats.Speed.ValueAdj,
		sourceChar.StatMod(`attacks`),
	)
}

// hitChance returns a hit probability in the range [30, 100] based on speeds.
func hitChance(attackSpd, defendSpd int) int {
	atkPlusDef := float64(attackSpd + defendSpd)
	if atkPlusDef < 1 {
		atkPlusDef = 1
	}
	return 30 + int(float64(attackSpd)/atkPlusDef*70)
}

// Hits returns whether an attack connects, incorporating an optional modifier.
func Hits(attackSpd, defendSpd, hitModifier int) bool {
	toHit := hitChance(attackSpd, defendSpd)
	if hitModifier != 0 {
		toHit += hitModifier
	}

	if toHit < 5 {
		toHit = 5
	}
	if toHit > 95 {
		toHit = 95
	}

	hitRoll := util.Rand(100)
	util.LogRoll(`Hits`, hitRoll, toHit)
	return hitRoll < toHit
}

// critChance returns the integer crit probability (0-100) given attacker stats
// and optional buff flags. It does not perform a roll.
func critChance(atkStr, atkSpd, levelDiff int, hasAccuracy, targetHasBlink bool) int {
	if levelDiff < 1 {
		levelDiff = 1
	}
	chance := 5 + int(math.Round(float64(atkStr+atkSpd)/float64(levelDiff)))
	if hasAccuracy {
		chance *= 2
	}
	if targetHasBlink {
		chance /= 2
	}
	if chance < 5 {
		chance = 5
	}
	if chance > 75 {
		chance = 75
	}
	return chance
}

// Crits rolls whether an attack is a critical hit.
func Crits(sourceChar characters.Character, targetChar characters.Character) bool {
	chance := critChance(
		sourceChar.Stats.Strength.ValueAdj,
		sourceChar.Stats.Speed.ValueAdj,
		sourceChar.Level-targetChar.Level,
		sourceChar.HasBuffFlag(buffs.Accuracy),
		targetChar.HasBuffFlag(buffs.Blink),
	)
	critRoll := util.Rand(100)
	util.LogRoll(`Crits`, critRoll, chance)
	return critRoll < chance
}

// dualWieldHitPenalty returns the negative hit modifier applied to the offhand
// weapon based on the attacker's dual-wield skill level.
func dualWieldHitPenalty(dwLevel int) int {
	if dwLevel < 4 {
		return -35
	}
	return -25
}

// dualWieldActiveWeaponCount returns how many weapons fire this round based on
// dual-wield skill level and whether both equipped weapons are claws.
func dualWieldActiveWeaponCount(dwLevel int, bothClaws bool) int {
	if bothClaws || dwLevel >= 3 {
		return 2
	}
	if dwLevel == 2 {
		if util.Rand(100) < 50 {
			return 2
		}
		return 1
	}
	return 1
}

// critDamageBonus returns the extra damage added to a hit that is a critical.
func critDamageBonus(dCount, dSides, dBonus int) int {
	return dCount*dSides + dBonus
}

// applyDefenseReduction applies a stochastic defense roll to incoming damage,
// returning the final damage and the amount reduced.
func applyDefenseReduction(damage, defenseRating int) (finalDamage, reduction int) {
	defenseAmt := util.Rand(defenseRating)
	if defenseAmt > 0 {
		reduction = int(math.Round((float64(defenseAmt) / 100) * float64(damage)))
		finalDamage = damage - reduction
		return finalDamage, reduction
	}
	return damage, 0
}

// damagePercentOfMax returns the dealt damage expressed as a percentage of the
// theoretical maximum damage for the given dice configuration.
func damagePercentOfMax(damage, dCount, dSides, dBonus int) int {
	maxDmg := dCount*dSides + dBonus
	if maxDmg < 1 {
		maxDmg = 1
	}
	return int(math.Ceil(float64(damage) / float64(maxDmg) * 100))
}

// tameSizeModifier returns the taming chance modifier based on the tamer's size.
func tameSizeModifier(size races.Size) int {
	switch size {
	case races.Large:
		return -25
	case races.Small:
		return 0
	default:
		return -10
	}
}

// tameHealthBonus returns the taming bonus granted when the tamer is injured.
// A tamer at full HP contributes 0; a tamer near death contributes up to 50.
func tameHealthBonus(currentHP, maxHP int) float64 {
	return 50 - math.Ceil(float64(currentHP)/float64(maxHP)*50)
}

// chanceToTame computes the raw taming probability from primitive parameters.
func chanceToTame(
	proficiency int,
	levelDiff int,
	currentHP int,
	maxHP int,
	tamerSize races.Size,
	targetIsAggro bool,
) int {
	const (
		modSkillMin       = 1
		modSkillMax       = 100
		modLevelDiffMin   = -25
		modLevelDiffMax   = 25
		factorIsAggro     = 0.50
	)

	if proficiency < modSkillMin {
		proficiency = modSkillMin
	} else if proficiency > modSkillMax {
		proficiency = modSkillMax
	}

	if levelDiff > modLevelDiffMax {
		levelDiff = modLevelDiffMax
	} else if levelDiff < modLevelDiffMin {
		levelDiff = modLevelDiffMin
	}

	sizeModifier := tameSizeModifier(tamerSize)
	healthModifier := tameHealthBonus(currentHP, maxHP)

	aggroModifier := 1.0
	if targetIsAggro {
		aggroModifier = factorIsAggro
	}

	return int(math.Ceil((float64(proficiency) + float64(levelDiff) + healthModifier + float64(sizeModifier)) * aggroModifier))
}

// ChanceToTame returns the probability that a user successfully tames a mob.
func ChanceToTame(s *users.UserRecord, t *mobs.Mob) int {
	raceInfo := races.GetRace(s.Character.GetRaceId())
	return chanceToTame(
		s.Character.MobMastery.GetTame(int(t.MobId)),
		s.Character.Level-t.Character.Level,
		s.Character.Health,
		s.Character.HealthMax.Value,
		raceInfo.Size,
		t.Character.IsAggro(s.UserId, 0),
	)
}

// AlignmentChange returns the alignment delta for a killer after slaying a target.
func AlignmentChange(killerAlignment int8, killedAlignment int8) int {

	isKillerGood := killerAlignment > characters.AlignmentNeutralHigh
	isKillerEvil := killerAlignment < characters.AlignmentNeutralLow
	isKillerNeutral := killerAlignment >= characters.AlignmentNeutralLow && killerAlignment <= characters.AlignmentNeutralHigh

	isKilledGood := killedAlignment > characters.AlignmentNeutralHigh
	isKilledEvil := killedAlignment < characters.AlignmentNeutralLow
	isKilledNeutral := killedAlignment >= characters.AlignmentNeutralLow && killedAlignment <= characters.AlignmentNeutralHigh

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
		if isKilledGood {
			factor = -2
			changeAmt = int(math.Max(float64(changeAmt), 1))
		} else if isKilledEvil {
			factor = 1
		} else if isKilledNeutral {
			factor = -1
		}
	} else if isKillerEvil {
		if isKilledGood {
			factor = -1
		} else if isKilledEvil {
			factor = 2
			changeAmt = int(math.Max(float64(changeAmt), 1))
		} else if isKilledNeutral {
			factor = -1
		}
	} else if isKillerNeutral {
		if isKilledGood {
			factor = -1
		} else if isKilledEvil {
			factor = 1
		} else if isKilledNeutral {
			factor = 0
		}
	}

	return factor * changeAmt
}

// expectedDPS estimates average damage per round without randomness.
func expectedDPS(atkChar characters.Character, defChar characters.Character) float64 {

	atkCount := combatAttackCount(atkChar, defChar)

	statDmgBonus := atkChar.StatMod(`damage`)

	attackWeapons := resolveAttackWeapons(atkChar)

	weaponWeight := make([]float64, len(attackWeapons))
	if len(attackWeapons) == 1 {
		weaponWeight[0] = 1.0
	} else {
		dwLevel := atkChar.GetSkillLevel(skills.DualWield)
		alwaysDual := atkChar.Equipment.Weapon.GetSpec().Subtype == items.Claws &&
			atkChar.Equipment.Offhand.GetSpec().Subtype == items.Claws
		switch {
		case alwaysDual || dwLevel >= 3:
			weaponWeight[0] = 1.0
			weaponWeight[1] = 1.0
		case dwLevel == 2:
			weaponWeight[0] = 1.0
			weaponWeight[1] = 0.5
		default:
			weaponWeight[0] = 1.0
			weaponWeight[1] = 0.0
		}
	}

	hitPct := float64(hitChance(atkChar.Stats.Speed.ValueAdj, defChar.Stats.Speed.ValueAdj)) / 100.0
	if hitPct < 0.05 {
		hitPct = 0.05
	}
	if hitPct > 0.95 {
		hitPct = 0.95
	}

	dwLevel := atkChar.GetSkillLevel(skills.DualWield)
	dwPenalty := 0.0
	if len(attackWeapons) > 1 {
		dwPenalty = float64(-dualWieldHitPenalty(dwLevel)) / 100.0
	}

	critPct := float64(critChance(
		atkChar.Stats.Strength.ValueAdj,
		atkChar.Stats.Speed.ValueAdj,
		atkChar.Level-defChar.Level,
		false,
		false,
	)) / 100.0

	defenseFraction := float64(defChar.GetDefense()) / 200.0
	if defenseFraction > 0.95 {
		defenseFraction = 0.95
	}

	totalDPS := 0.0

	for roundIdx := 0; roundIdx < atkCount; roundIdx++ {
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

			avgRoll := float64(dCount) * float64(dSides+1) / 2.0
			avgDmg := avgRoll + float64(dBonus)
			if avgDmg < 0 {
				avgDmg = 0
			}

			critBonus := float64(critDamageBonus(dCount, dSides, dBonus)) * critPct

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

// CombatOdds returns the ratio of rounds-for-attacker-to-kill-defender to
// rounds-for-defender-to-kill-attacker. Values above 1.0 favor the attacker.
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

	return defRoundsToKill / atkRoundsToKill
}
