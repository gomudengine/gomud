# skill.protection.aid

## Skill Tag & Training

- Skill: `protection` (`internal/skills/skills.go`)
- `aid` is the level 1 ability of the protection skill tree. Level 3 removes the calm-room restriction.

## Overview

`aid` revives a downed (0 HP) player in the same room by initiating a cast of the `aidskill` spell. The actual revival effect is handled by the spell script.

## Preconditions

- Protection skill level >= 1.
- At levels 1-2: the room must be calm (`room.IsCalm()`). At level 3+, aid can be used in any room including during combat.
- The player themselves cannot be the target.
- The target must be downed (`FindDowned` flag used in target search).
- The caster must not be in combat (`Aggro == nil`).

## Target Resolution

`room.FindByName(rest, rooms.FindDowned)` — only finds players with 0 or fewer HP. If the found player ID is the caster's own ID, it is rejected.

## Execution Flow

1. Skill level check — rejects if level 0.
2. Calm-room check (levels 1-2 only).
3. Target resolution.
4. Fires a `SkillUsed` event (`skill: protection`, `details: aid`).
5. Validates the target player exists and has 0 HP.
6. Combat check on the caster — rejects if in combat.
7. Builds a `SpellAggroInfo` targeting the downed player with spell ID `aidskill`.
8. Calls `scripting.TrySpellScriptEvent("onCast", ...)` — scripts can intercept.
9. If not intercepted: cancels the `Hidden` buff and calls `user.Character.SetCast(waitRounds, spellAggro)` to queue the cast.

## Spell Resolution

The revival is resolved as a spell cast (`aidskill`). The wait rounds and revival effect are defined in the spell data and script — not in this command file.

## Notes

- The cooldown key `protection` is shared with `pray`. Using either ability starts the shared timer.
- There is no cooldown enforced in this command file; the shared cooldown (if any) would be managed by `pray`.
- The `aidskill` spell's `WaitRounds` determines how long the cast takes before the target is revived.
