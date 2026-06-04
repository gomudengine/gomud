# skill.brawling.recover

## Skill Tag & Training

- Skill: `brawling` (`internal/skills/skills.go`)
- Trained at the Soldiers Training Yard (room 829), levels 1-4.
- `recover` is the level 1 ability of the brawling skill tree.

## Overview

`recover` puts the player into a short regeneration state outside of combat. It applies buff 23 ("Warriors Respite"), which accelerates health recovery for up to 20 rounds but cancels immediately on combat entry or any player action.

## Preconditions

- Brawling skill level >= 1.
- Player must not be in combat (`Aggro == nil`).
- Cooldown: **2 real minutes** per use (keyed `brawling:recover`).

## Execution Flow

1. Skill level check — returns `false` (unhandled) if level < 1, making the command invisible.
2. Combat check — rejects with a message if the player is currently in combat.
3. Cooldown check — rejects if the cooldown has not expired.
4. Fires a `SkillUsed` event (`skill: brawling`, `details: recover`).
5. Applies **buff 23** (`Warriors Respite`) via `user.AddBuff(23, "skill")`.

## Buff 23 — Warriors Respite

| Field | Value |
|---|---|
| Name | Warriors Respite |
| Trigger rate | 1 round |
| Trigger count | 20 (up to 20 rounds duration) |
| Flags | `cancel-on-combat`, `cancel-on-action` |

The buff fires every round for up to 20 rounds, providing accelerated health regeneration. It is cancelled immediately if the player enters combat or takes any action.

## Notes

- No stat roll is involved; the buff is always applied on a successful use.
- The `cancel-on-action` flag means even moving or typing a command will end the recovery state early.
