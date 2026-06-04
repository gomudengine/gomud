# skill.skulduggery.bump

## Skill Tag & Training

- Skill: `skulduggery` (`internal/skills/skills.go`)
- Trained at the Thieves Den (room 491), levels 1-4.
- `bump` is the level 2 ability of the skulduggery skill tree (despite the file comment saying level 3).

## Overview

`bump` is a non-combat gold-extraction move. The player "accidentally" bumps into a target, with a chance of causing some of their gold to spill onto the floor. Unlike `pickpocket`, the gold goes to the room floor rather than to the player's inventory, and the action is always visible to the room.

## Preconditions

- Skulduggery skill level >= 2.
- Player must not be in combat (`Aggro == nil`).
- No mobs may be attacking the player.
- A target name argument is required.
- Cooldown: **1 real minute** per use (keyed `skulduggery:bump`), applied only if a valid target is found.
- The `Hidden` buff is **always cancelled** on a successful bump attempt (the action is inherently visible).

## Execution Flow

1. Skill level check — returns `false` if level < 2.
2. Combat and under-attack checks.
3. Argument check.
4. Target resolution via `room.FindByName`.
5. Cooldown check and hidden-buff cancellation.
6. Fires a `SkillUsed` event (`skill: skulduggery`, `details: bump`).
7. Rolls for gold drop.
8. Sends bump messages to the player, target (PvP), and room.
9. If gold was dropped, adds it to `room.Gold` and notifies the room.

## Success Roll

```
levelPenalty = max(0, target.Level - attacker.Level)
chanceIn100  = attacker.Strength(adj) / 2 - levelPenalty
chanceIn100  = max(1, chanceIn100)
roll = rand(100)
success = roll < chanceIn100
```

The level penalty subtracts 1% per level the target is above the attacker, making higher-level targets harder to bump. Attacking lower-level targets incurs no penalty.

## Gold Amount

On success, the gold dropped is a random value between 0 and `target.Gold / 4`. The gold is deducted from the target and added to the room floor — it is not given directly to the player.

For player targets, an `EquipmentChange` event is fired to reflect the gold loss.

## Notes

- The bump message is always sent to the room regardless of success or failure — the action itself is not hidden.
- `bump` does not steal items, only gold.
- The player must pick up the dropped gold separately.
