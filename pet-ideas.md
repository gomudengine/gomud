# Pet Ideas - One Per Profession

Reference: existing pets (cat, dog, mule, owl) each have a single ability at level 1. These proposals use the full level 1-10 range with multiple ability tiers that replace each other as the pet levels up.

Each pet is designed to complement and reinforce the playstyle of a specific profession. There are 10 professions: Warrior, Paladin, Ranger, Assassin, Monster Hunter, Sorcerer, Arcane Scholar, Explorer, Treasure Hunter, and Merchant.

---

## 1. War Hound (Warrior)

**Profession skills:** Brawling, Dual Wield

Warriors live and die by melee output. The war hound is a pure damage multiplier -- at high levels it boosts the *owner's* attacks and flat damage in addition to dealing its own.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | strength +3 | 10% chance, 1d4 | |
| 4 | strength +5, speed +3 | 15% chance, 1d6 | |
| 7 | strength +7, speed +5, damage +1 | 20% chance, 2d4 | |
| 10 | strength +10, speed +7, damage +2, attacks +1 | 25% chance, 2d6 | |

**Why it fits:** Warriors stack strength and dual-wield for raw DPS. The hound amplifies that by adding strength (more melee damage), speed (more hits land), and eventually `damage +2` and `attacks +1` which directly buff the warrior's own combat output. No utility, no magic -- just relentless aggression. The highest combat-chance and damage of any pet.

---

## 2. Tortoise (Paladin)

**Profession skills:** Protection, Brawling

Paladins protect and endure. The tortoise is a slow, ancient guardian that makes its owner extraordinarily hard to kill through massive health regeneration and vitality.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | vitality +4, healthrecovery +1 | | |
| 5 | vitality +7, healthrecovery +2, healthmax +10 | | |
| 10 | vitality +10, healthrecovery +3, healthmax +25 | | capacity 3 |

**Why it fits:** Paladins use the Protection skill to absorb hits. The tortoise's vitality and healthmax stack makes the paladin nearly unkillable, and healthrecovery +3 at max level means constant regeneration between fights. Only 3 ability tiers (big jumps at 1, 5, 10) which thematically mirrors the tortoise's slow but inevitable growth. The small capacity 3 at max level lets the paladin carry a few extra healing items. Zero combat from the pet itself -- the paladin handles that.

---

## 3. Hawk (Ranger)

**Profession skills:** Map, Search, Track

Rangers scout and track. The hawk is their eyes in the sky -- boosting perception and speed, striking from above, and eventually revealing hidden enemies.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | perception +4 | 5% chance, 1d3 | |
| 3 | perception +6, speed +2 | 10% chance, 1d4 | |
| 6 | perception +8, speed +4 | 15% chance, 1d6 | |
| 9 | perception +10, speed +5 | 20% chance, 2d4 | buff: see-hidden |

**Why it fits:** Rangers use Search to find things and Track to follow targets. The hawk's scaling perception directly supports this playstyle, and the see-hidden buff at level 9 is a game-changer for spotting stealthed assassins or hidden passages. The combat is moderate (hawk swoops and rakes) but the real value is as a scout. Perception +10 at max level is the highest perception bonus of any pet.

---

## 4. Viper (Assassin)

**Profession skills:** Skulduggery, Dual Wield, Track

Assassins strike from shadows and apply debilitating effects. The viper complements this with speed, poison-on-crit, and late-game perception for tracking targets.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | speed +2 | 8% chance, 1d3 | |
| 3 | speed +4 | 12% chance, 1d4 (crit: poison) | |
| 6 | speed +6, perception +2 | 15% chance, 1d6 (crit: poison) | |
| 9 | speed +8, perception +3, damage +1 | 20% chance, 2d4 (crit: poison) | |

**Why it fits:** Assassins rely on Skulduggery for stealth and Dual Wield for burst damage. The viper's speed bonus helps the assassin land hits and dodge, while the poison-on-crit adds a DoT that punishes enemies after the assassin's opening strike. Less raw damage than the war hound, but the poison stacks with the assassin's own attacks for sustained pressure. The late perception bonus helps with Track.

*Note: crit-poison requires a poison buff ID. The dice roll format supports this via `2d4#13` where 13 is the poison buff ID.*

---

## 5. Lynx (Monster Hunter)

**Profession skills:** Tame, Track, ChangeForm

Monster Hunters tame wild creatures and hunt beasts. The lynx is a fierce, semi-wild predator that directly boosts the owner's tame chance and provides solid combat with tracking instincts.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | tame +1, perception +2 | 5% chance, 1d3 | |
| 3 | tame +2, perception +3, speed +2 | 10% chance, 1d4 | |
| 6 | tame +3, perception +5, speed +3 | 15% chance, 1d6 | |
| 8 | tame +4, perception +6, speed +4, strength +3 | 20% chance, 1d8 | |
| 10 | tame +5, perception +7, speed +5, strength +5 | 25% chance, 2d6 | |

**Why it fits:** The lynx is the only pet that boosts the `tame` stat mod, making it essential for Monster Hunters who want to tame more difficult creatures. The perception and speed help with Track. Five ability tiers give it the most frequent upgrades, reflecting the bond between a Monster Hunter and their companion deepening over time. At max level, tame +5 is a significant boost, and the lynx becomes a dangerous combatant in its own right.

---

## 6. Toad (Sorcerer)

**Profession skills:** Cast, Enchant

Sorcerers channel raw magical power. The toad is a classic arcane familiar -- pulsing with energy, boosting casting success, mana pool, and mana regeneration.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | mysticism +3, manamax +5 | | |
| 3 | mysticism +5, manamax +10, casting +3 | | |
| 6 | mysticism +7, manamax +15, casting +5, manarecovery +1 | | |
| 10 | mysticism +10, manamax +25, casting +8, manarecovery +2 | | |

**Why it fits:** Sorcerers live and die by their spell casting. The toad directly boosts casting success (+8 at max level), expands the mana pool (+25 manamax), and reduces downtime with mana recovery. The mysticism stat further amplifies mana-related calculations. No combat, no utility -- pure magical amplification. The toad is to sorcerers what the war hound is to warriors: the obvious, powerful choice for the profession's core mechanic.

---

## 7. Raven (Arcane Scholar)

**Profession skills:** Enchant, Scribe, Inspect

Arcane Scholars are intellectual mages focused on knowledge, enchantment, and analysis. The raven sharpens the mind with smarts and perception while carrying scrolls and tomes.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | smarts +3, perception +2 | | |
| 3 | smarts +5, perception +3, mysticism +2 | | capacity 1 |
| 6 | smarts +7, perception +5, mysticism +3, casting +2 | | capacity 2 |
| 9 | smarts +10, perception +6, mysticism +4, casting +4 | | capacity 3 |

**Why it fits:** Arcane Scholars use Inspect to analyze and Scribe to create. The raven's smarts bonus is the highest of any pet (+10 at max), directly supporting skill learning and spell power. The mysticism and casting bonuses distinguish it from the toad: less raw mana power, but more intellectual breadth. The small capacity lets the scholar's raven carry a few extra items (scrolls, reagents). Perception helps with Inspect. No combat -- scholars fight with their mind.

---

## 8. Fox (Explorer)

**Profession skills:** Map, Portal, Scribe

Explorers chart unknown territory and tear open portals between locations. The fox is a quick, cunning pathfinder that sharpens the explorer's instincts and keeps them alive in unfamiliar terrain.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | speed +3, perception +2 | | |
| 3 | speed +4, perception +3, smarts +2 | | capacity 1 |
| 5 | speed +6, perception +4, smarts +3, healthrecovery +1 | | capacity 2 |
| 8 | speed +8, perception +6, smarts +4, healthrecovery +1 | | capacity 3 |
| 10 | speed +10, perception +7, smarts +5, healthrecovery +2 | | capacity 4 |

**Why it fits:** Explorers use Map to chart rooms, Portal to jump across the world, and Scribe to record what they find. The fox's speed bonus is the highest of any pet (+10), letting the explorer move and dodge faster. Perception helps reveal hidden exits and details in new areas. Smarts supports Scribe and general skill learning. The healthrecovery keeps the explorer topped up between encounters in remote zones where there's no healer. Five ability tiers mirror the fox's nature -- always adapting, always one step ahead. Small but growing capacity lets the explorer stash maps and supplies.

---

## 9. Ferret (Treasure Hunter)

**Profession skills:** Map, Search, Peep, Inspect, Trading

Treasure Hunters find hidden things and get past locked doors. The ferret is the ultimate infiltration companion -- wriggling into locks, carrying loot, and sharpening the owner's senses.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | speed +2, picklock +1 | | |
| 3 | speed +3, picklock +2, perception +2 | | capacity 2 |
| 6 | speed +5, picklock +3, perception +4 | | capacity 4 |
| 9 | speed +7, picklock +4, perception +6 | | capacity 7 |

**Why it fits:** Treasure Hunters use Search to find hidden caches and need to bypass locks to reach them. The ferret provides the best picklock bonus in the game (+4) and the most carrying capacity (7 items) for hauling loot. Perception helps with Search and Inspect. Speed aids exploration. Zero combat -- Treasure Hunters avoid fights when they can. At max level, this is the definitive "open anything, carry everything" pet.

---

## 10. Magpie (Merchant)

**Profession skills:** Peep, Trading

Merchants appraise, trade, and profit. The magpie is a shrewd collector -- boosting experience gain, carrying trade goods, and sharpening the owner's eye for value.

| Level | Stats | Combat | Other |
|-------|-------|--------|-------|
| 1 | smarts +2, xpscale +2 | | capacity 2 |
| 3 | smarts +3, xpscale +4, perception +2 | | capacity 3 |
| 6 | smarts +5, xpscale +6, perception +3 | | capacity 5 |
| 10 | smarts +7, xpscale +10, perception +5 | | capacity 7 |

**Why it fits:** Merchants use Peep to appraise and Trading to profit. The magpie boosts smarts (better learning), perception (sharper appraisals), and provides the best xpscale of any pet (+10 at max). The high capacity (7 items) lets the merchant carry more trade goods. No combat -- merchants hire muscle for that. The xpscale +10 makes this the best progression pet in the game, rewarding the merchant playstyle of grinding smart rather than fighting hard.

---

## Comparison Matrix

| Pet | Profession | Primary Role | Combat | Carry | Unique Angle |
|-----|-----------|-------------|--------|-------|--------------|
| War Hound | Warrior | Offense | Best | No | +damage, +attacks for owner |
| Tortoise | Paladin | Survival | None | Small | Best healthrecovery (+3) |
| Hawk | Ranger | Scout/Combat | Medium | No | See-hidden buff at lv9 |
| Viper | Assassin | Debuff/Combat | Medium | No | Poison on crit |
| Lynx | Monster Hunter | Tame/Combat | Strong | No | Only pet with +tame |
| Toad | Sorcerer | Mage Support | None | No | Best casting (+8) and mana |
| Raven | Arcane Scholar | Intellect | None | Small | Best smarts (+10) |
| Fox | Explorer | Pathfinder | None | Medium | Best speed (+10) |
| Ferret | Treasure Hunter | Thief/Utility | None | Best | Best picklock (+4) |
| Magpie | Merchant | Commerce/XP | None | High | Best xpscale (+10) |

### Design Notes

- Every pet has a clear niche tied to its profession's core mechanic.
- Combat pets (War Hound, Lynx, Hawk, Viper) never overlap fully: hound has raw damage, lynx has tame utility, hawk has scouting, viper has poison.
- Non-combat pets (Tortoise, Toad, Raven, Fox, Ferret, Magpie) each excel at something different: survival, mana, intellect, speed/exploration, locks, or XP.
- Ability tier counts vary from 3 (Tortoise -- slow growth, big jumps) to 5 (Lynx, Fox -- constant bonding progression).
- The two highest-capacity pets (Ferret and Magpie, both 7 at max) serve different professions and stat profiles.
- The Fox is the only pet that combines best-in-class speed with healthrecovery, making it uniquely suited for long solo exploration runs.
