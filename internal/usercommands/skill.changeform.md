# skill.changeform

## Skill Tag & Training

- Skill: `changeform` (`internal/skills/skills.go`)
- Levels 1-4.

## Overview

`changeform` temporarily transforms the player into another race, applying that race's stats, abilities, and physical traits for a limited number of rounds. The transformation is reverted manually or when the buff expires.

## Preconditions

- ChangeForm skill level >= 1.
- Player must not already be transformed (`IsFormChanged()` must be false), unless using `changeform revert`.
- Cooldown: **20 rounds** per use (keyed `changeform`). Not applied when reverting.

## Revert

```
changeform revert
```

If the player is currently transformed, this reverts them to their original race via `user.Character.RevertFormChange()` and removes buffs 41 and 42. No cooldown is consumed.

## Transformation

### Target Race

The argument is matched against available races via `races.FindRace(rest)`.

- The target race must not be the player's current race.
- At levels 1-3, only selectable races (`raceInfo.Selectable == true`) are allowed.
- At level 4, all races (including non-selectable/monster races) are available.

### Duration by Level

| Level | Duration (rounds) |
|---|---|
| 1 | 10 |
| 2 | 20 |
| 3 | 40 |
| 4 | 80 |

### Execution

1. Cooldown check.
2. Race lookup and validation.
3. Calls `user.Character.ApplyFormChange(raceInfo.RaceId)` — applies the new race's stats and traits.
4. Applies **buff 41** ("Form Change") with the level-appropriate duration via `user.Character.AddBuff(41, false, duration)`.
5. Fires a `SkillUsed` event (`skill: changeform`).
6. Sends transformation messages to the player and room.

## Buff 41 — Form Change

| Field | Value |
|---|---|
| Name | Form Change |
| Trigger rate | 1 round |
| Trigger count | 20 (base; overridden by the duration argument) |

The buff tracks the transformed state. When it expires, the form-change is expected to be reverted by the buff expiry handler.

## Buff 42 — Polymorphed

Buff 42 ("Polymorphed") is referenced in the revert path (`user.Character.RemoveBuff(42)`) but is not applied by this command. It is likely applied by external polymorph effects (e.g., mob spells) and is cleaned up together with buff 41 on revert.

## Notes

- The transformation changes the player's race-derived stats, equipment slot availability, and racial buffs for the duration.
- There is no stat roll; transformation always succeeds if preconditions are met.
