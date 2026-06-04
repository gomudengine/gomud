# skill.track

## Skill Tag & Training

- Skill: `track` (`internal/skills/skills.go`)
- Trained at the Frostwarden Rangers (room 74), levels 1-4.

## The Visitor System (underlying data)

Every time a player or mob enters or leaves a room, `room.MarkVisited(id, VisitorType)` is called. This stamps a turn-based expiry timestamp on the room:

```
expires = currentTurn + (180 seconds * turnsPerSecond)
```

When leaving, `MarkVisited(..., 1)` is called with a `subtrackTurns` argument that reduces the expiry by 1 turn (makes the trail slightly staler immediately on departure).

`room.Visitors(vType)` converts raw turn counts back into a `float64` trail strength in the range `[0.0, 1.0]`:

```
strength = (expires - currentTurn) / (180 * turnsPerSecond)
```

Trails expire after **180 seconds (3 real minutes)**. Entries are pruned lazily; currently-present entities always get their timestamp refreshed in `PruneVisitors`.

## Trail Strength Labels

| Numeric strength | Label |
|---|---|
| < 0.15 | Dead |
| < 0.50 | Weak |
| < 0.70 | Good |
| < 0.90 | Warm |
| >= 0.90 | Hot |

## Skill Levels

All levels share a **1-round cooldown** enforced via `TryCooldown`. A `SkillUsed` event is fired on each successful use.

### Level 1 — Single strongest trail

`track` with no argument. Scans all mob and user visitor records in the current room. Skips entities currently present. Keeps only the **single entry with the highest trail strength**. Outputs a "Recent Visitors" panel with that one entry, no exit direction.

### Level 2 — All recent trails

Same as level 1 but shows **all** past visitors, not just the strongest one. Still no exit direction information.

### Level 3 — Trails with exit directions

No-argument usage: same as level 2 but each entry also shows **which exit the target likely used** via `findExited`.

`findExited` loads each adjacent non-secret room and checks whether the target's visitor record is strongest there, returning the exit name with the best trail.

With an argument (`track <name>`): first checks if the target is in the current room. If not, searches current-room visitor records for names matching the prefix, then checks adjacent rooms for the best exit direction. Results are de-duplicated per name, keeping the strongest trail.

### Level 4 — Active / persistent tracking

`track <name>` with a target found in the current room's visitor records.

- Stores the target name in character misc data as `tracking-mob` or `tracking-user`.
- Applies **buff 26** ("Actively Tracking") to the player.
- On every subsequent room look (`GetDetails`), the `TrackingString` field is populated:
  - Target is **here**: "Tracking \<name\>... They are here!" — buff removed (tracking complete).
  - Target has a trail in an adjacent room: "Tracking \<name\>... They went \<exit\>".
  - No trail found: "You lost the trail of \<name\>" — buff removed.
- The tracking string is rendered below the room description in ANSI color 182.

## Buff 26 — Actively Tracking

| Field | Value |
|---|---|
| Name | Actively Tracking |
| Trigger rate | 1 real minute |
| Trigger count | 3 (up to ~3 minutes duration) |
| Flags | `cancel-on-combat` |

## Secret Exit Handling

Both `findExited` and the level 3 named-target search skip exits marked `Secret: true`. Tracking cannot reveal hidden passages.
