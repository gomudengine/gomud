# skill.skulduggery.sneak

## Skill Tag & Training

- Skill: `skulduggery` (`internal/skills/skills.go`)
- Trained at the Thieves Den (room 491), levels 1-4.
- `sneak` is the level 1 ability of the skulduggery skill tree.

## Overview

`sneak` applies buff 9 ("Hidden") to the player, concealing them from normal room observation. Being hidden is a prerequisite for `backstab` and provides a bonus to `pickpocket` success.

## Preconditions

- Skulduggery skill level >= 1.
- Player must not already have the `Hidden` buff flag.
- Player must not be in combat (`Aggro == nil`).
- The current room must be calm (`room.IsCalm()`).

## Execution Flow

1. Skill level check — returns `false` if level < 1.
2. Already-hidden check — rejects if the player already has the `Hidden` buff flag.
3. Combat check — rejects if in combat.
4. Room calm check — rejects if the room is not calm.
5. Applies **buff 9** via `user.AddBuff(9, "skill")`.
6. Fires a `SkillUsed` event (`skill: skulduggery`, `details: sneak`).

Note: the `SkillUsed` event is fired after the buff is applied, which is the reverse of most other skills.

## Buff 9 — Hidden

| Field | Value |
|---|---|
| Name | Hidden |
| Trigger rate | 1 round |
| Trigger count | 15 (up to 15 rounds duration) |
| Flags | `hidden`, `cancel-on-combat` |

The `hidden` flag causes the player to be omitted from normal room descriptions. The buff cancels on combat entry. It also cancels explicitly when actions like `bump` or `pickpocket` fail.

## Notes

- There is no cooldown on `sneak`.
- The skill has no stat roll; it always succeeds if preconditions are met.
- Other skulduggery abilities (`backstab`, `pickpocket`) check for the `Hidden` buff flag directly — they do not require `sneak` to have been the source.
