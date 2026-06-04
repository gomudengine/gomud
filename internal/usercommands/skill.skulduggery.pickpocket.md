# skill.skulduggery.pickpocket

## Skill Tag & Training

- Skill: `skulduggery` (`internal/skills/skills.go`)
- Trained at the Thieves Den (room 491), levels 1-4.
- `pickpocket` is the level 4 ability of the skulduggery skill tree.

## Overview

`pickpocket` attempts to steal gold and a random item directly from a target's inventory into the player's backpack. Unlike `bump`, the gold goes to the player rather than the floor. Failure against a mob causes it to attack; failure against a player reveals the attempt.

## Preconditions

- Skulduggery skill level >= 4.
- Player must not be in combat (`Aggro == nil`).
- No mobs may be attacking the player.
- A target name argument is required.
- Cooldown: **1 real minute** per use (keyed `skulduggery:pickpocket`), applied only if a valid target is found.

## Execution Flow

1. Skill level check — returns `false` if level < 4.
2. Combat and under-attack checks.
3. Argument parsing.
4. Target resolution via `room.FindByName`.
5. Cooldown check.
6. Fires a `SkillUsed` event (`skill: skulduggery`, `details: pickpocket`).
7. Rolls for success.
8. On success: steals gold and/or a random item.
9. On failure: reveals the player and triggers consequences.

## Success Roll

```
levelPenalty = max(0, target.Level - attacker.Level)
chanceIn100  = (Speed(adj) + Smarts(adj) + Perception(adj)) / 3 - target.Perception(adj) - levelPenalty
chanceIn100  = max(1, chanceIn100)
if isSneaking: chanceIn100 += 15
roll = rand(100)
success = roll < chanceIn100
```

The level penalty subtracts 1% per level the target is above the attacker, making higher-level targets harder to steal from. Attacking lower-level targets incurs no penalty. Being hidden adds a flat +15 to the chance.

## On Success

**Gold stolen:**
- Amount is a random value in the range `[target.Gold * 3/4, target.Gold]` (approximately the upper quarter of the target's gold).
- Deducted from target, added to the attacker.
- `EquipmentChange` events are fired for both parties (player targets only).

**Item stolen:**
- One random item from the target's inventory via `GetRandomItem()`.
- Removed from target, stored in attacker's backpack.
- `ItemOwnership` events are fired for both parties.

If the target has neither gold nor items, the player is informed they found nothing.

## On Failure

**Mob target:** The mob is notified ("catches you in the act"), the room is notified, the player's `Hidden` buff is cancelled, and the mob immediately attacks the player via `mob.Command("attack @userId")`.

**Player target:** The target is notified, the room is notified, and the player's `Hidden` buff is cancelled. No automatic combat is initiated against a player.

## Notes

- PvP rules are checked before attempting against a player target.
- The skill does not have a visible action message on success — only the player sees the result.
