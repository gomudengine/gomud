# skill.protection.pray

## Skill Tag & Training

- Skill: `protection` (`internal/skills/skills.go`)
- `pray` is the level 4 ability of the protection skill tree.

## Overview

`pray` calls upon divine power to bestow one or more random beneficial buffs on the player or a target in the room. The number and quality of buffs scale with the caster's Mysticism stat.

## Preconditions

- Protection skill level >= 4.
- Cooldown: **5 real minutes** per use (keyed `protection`).

## Target Resolution

- No argument: targets the player themselves.
- With argument: resolves via `room.FindByName(rest)` — can target a player or mob.

## Buff Pool

The possible buffs are:

| Buff ID | Effect |
|---|---|
| 4 | (defined in buff data files) |
| 11 | (defined in buff data files) |
| 14 | (defined in buff data files) |
| 16 | (defined in buff data files) |
| 17 | (defined in buff data files) |
| 18 | (defined in buff data files) |

## Number of Buffs Applied

```
totalBuffCount = 1 + floor(Mysticism(adj) / 15) + rand(2)
totalBuffCount = min(totalBuffCount, len(possibleBuffIds))
```

At 0 Mysticism: 1-2 buffs. At 75 Mysticism: 6-7 buffs (capped at 6 — the size of the pool).

## Execution Flow

1. Skill level check — rejects if level < 4.
2. Cooldown check.
3. Target resolution.
4. Fires a `SkillUsed` event (`skill: protection`, `details: pray`).
5. Sends room flavor messages.
6. For each buff to apply: selects a random index from the remaining pool, queues a `Buff` event for the target, removes that buff ID from the pool (no duplicates).
7. A "glows for a moment" message is sent to the room for each buff applied.

## Notes

- Buffs are applied without duplicates — each buff ID can only be granted once per prayer.
- The buff pool is consumed in random order; with high Mysticism all 6 buffs can be granted in a single prayer.
- The cooldown key `protection` is shared with `aid`. Using either ability starts the shared timer.
