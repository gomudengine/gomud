# skill.enchant

## Skill Tag & Training

- Skill: `enchant` (`internal/skills/skills.go`)
- Levels 1-4.

## Overview

`enchant` applies magical enhancements to weapons and armor in the player's backpack. Each application risks destroying the item, and that risk increases with the number of existing enchantments. Level 4 also unlocks the ability to remove enchantments and curses.

Two convenience aliases exist in the same file:
- `uncurse <item>` → delegates to `enchant uncurse <item>`
- `unenchant <item>` → delegates to `enchant remove <item>`

## Preconditions

- Enchant skill level >= 1.
- Target item must be in the player's backpack (not worn).
- Target must be a weapon (`items.Weapon`) or wearable armor (`items.Wearable` subtype).
- Cooldown: **15 real minutes** per enchantment attempt (keyed `enchant`). Does not apply to `chance`, `remove`, or `uncurse` sub-commands.

## Sub-commands

### `enchant chance <item>`

Shows the current destruction probability without attempting the enchantment. No cooldown consumed.

```
chanceToDestroy = 50
chanceToDestroy -= skillLevel * 10          (10–40% reduction)
chanceToDestroy += item.Enchantments * 20   (20% per existing enchantment)
chanceToDestroy -= Mysticism(adj) / 4       (1% per 4 Mysticism points)
```

### `enchant remove <item>` (level 4 only)

Removes all enchantments from the item. No destruction risk. Fires `SkillUsed` with `details: remove`.

### `enchant uncurse <item>` (level 4 only)

Removes the curse from a cursed item. The item must be cursed. No destruction risk. Fires `SkillUsed` with `details: uncurse`.

## Enchantment Attempt

### Destruction Roll

```
roll = rand(100)
if roll < chanceToDestroy: item is destroyed
```

The item is removed from inventory. If destroyed, an `ItemOwnership{Gained: false}` event is fired and the item is gone permanently.

### Curse Roll (independent of destruction)

```
roll = rand(100)
if roll < 25: enchantment is cursed
```

A 25% chance the enchantment is cursed regardless of skill level or stats.

### Bonuses Applied

| Level | Bonus |
|---|---|
| 1+ | Weapon damage bonus: `ceil(sqrt(Mysticism(adj)))` |
| 2+ | Armor defense bonus: `ceil(sqrt(Mysticism(adj)))` (wearable only) |
| 3+ | Random stat bonuses: `skillLevel - 1` stats chosen randomly, each `ceil(sqrt(Mysticism(adj)))` |

At level 3 the player gets 2 random stat bonuses; at level 4 they get 3. Stats are chosen from: strength, speed, smarts, vitality, mysticism, perception.

The enchantment is applied via `item.Enchant(damageBonus, defenseBonus, statBonus, cursed)`.

## Notes

- Enchanting a weapon at level 2+ also applies the defense bonus to armor, but the damage bonus only applies if the item is a weapon.
- The stat bonus selection is random — the same stat can be selected twice (the map will just overwrite with the same value).
- Enchantments stack: each call to `enchant` increments `item.Enchantments`, increasing future destruction risk.
