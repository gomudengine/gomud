# skill.protection.rank

## Skill Tag & Training

- Skill: `protection` (`internal/skills/skills.go`)
- `rank` is available at protection skill level >= 1.

## Overview

`rank` sets the player's tactical position within their party. Position affects how likely the player is to be targeted by enemies in combat.

## Preconditions

- Protection skill level >= 1.
- Player must be in a party (`parties.Get(user.UserId)` must return a non-nil party).

## Positions

| Argument | Position | Targeting chance |
|---|---|---|
| `front` | Front | 2 (high — likely to be targeted) |
| `back` | Back | 0 (protected — cannot be directly targeted) |
| anything else | Middle | 1 (moderate) |

Targeting chance values are returned by `party.ChanceToBeTargetted(userId)` and used by the combat system.

## Execution Flow

1. Skill level check — rejects if level < 1.
2. Party check — rejects if not in a party.
3. Sets the rank on the party via `party.SetRank(user.UserId, position)`.
4. Fires a `SkillUsed` event (`skill: protection`, `details: rank`).
5. Sends confirmation to the player and room.

## Notes

- There is no cooldown on `rank`.
- The rank persists for the duration of the party session.
- Rank only has effect when the player is in a party; solo players are always targetable.
