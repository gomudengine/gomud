# skill.search

## Skill Tag & Training

- Skill: `search` (`internal/skills/skills.go`)
- Trained at the Frostwarden Rangers (room 74), levels 1-4.

## Overview

`search` lets the player actively look for things that are not visible in the normal room description: secret exits, stashed items, and hidden players or mobs. Higher skill levels reveal more categories of hidden content.

## Preconditions

- Search skill level >= 1.
- Cooldown: **2 rounds** per use (keyed `search`).

## Success Odds

```
searchOddsIn100 = 10 + ceil(Perception(adj) / 2)
if skillLevel >= 4: searchOddsIn100 *= 2
```

A base 10% chance plus 1% for every 2 points of adjusted Perception, doubled at skill level 4. This roll is applied independently to each secret exit and each hidden entity.

| Level | Chance per roll (Perception ~5) | Chance per roll (Perception ~82) |
|---|---|---|
| 1-3 | ~13% | ~51% |
| 4 | ~26% | ~102% (effectively 100%) |

## Execution Flow

1. Skill level check — rejects if level 0.
2. Cooldown check.
3. Fires a `SkillUsed` event (`skill: search`).
4. Sends a "snooping around" message to the room.
5. Scans all room exits for secret ones and rolls individually for each.
6. If level > 2: scans stashed items, hidden players, and hidden mobs.
7. If level >= 3: prop/special-interest search (not yet implemented — empty block).

## What Is Revealed

### Level 1+ — Secret Exits

For each exit marked `Secret: true`, an independent roll is made. On success the player sees the exit name.

### Level 2+ — Stashed Items, Hidden Players, Hidden Mobs

**Stashed items**: All items in `room.Stash` are listed (no roll — stash is always revealed at level 2+). Invalid stash items are pruned.

**Hidden players**: For each other player in the room with the `Hidden` buff flag, an independent roll is made. On success they appear in a modified room "who" display tagged `(hiding)`.

**Hidden mobs**: Same roll logic as hidden players. On success they appear in a modified mob list tagged `(hiding)`.

Note: there is a bug in the hidden-mob loop — it calls `users.GetByUserId(mId)` instead of `mobs.GetInstance(mId)`, so hidden mobs are never actually revealed in practice.

### Level 3+ — Props / Things of Interest

Code block is present but empty (`// Find props`). No functionality is implemented yet.

## Notes

- The search action is always visible to the room regardless of the player's hidden status.
- Secret exits are only revealed to the searching player, not broadcast to the room.
