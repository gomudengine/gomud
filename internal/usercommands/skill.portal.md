# skill.portal

## Skill Tag & Training

- Skill: `portal` (`internal/skills/skills.go`)
- Unlocked by touching the obelisk in room 871.
- Levels 1-4.

## Overview

`portal` is a teleportation skill. At low levels it instantly moves the player to a fixed destination. At higher levels the destination becomes configurable, and at level 4 a physical two-way portal can be opened for other players to use.

## Special Pre-check

If the player is in the death-recovery room, the command returns `false` (unhandled) immediately.

If `rest` is empty, the command first tries to treat "portal" as a movement direction via `Go("portal", ...)` — this allows players to walk through an existing portal exit by typing `portal` with no argument. If `Go` handles it, the skill command returns without doing anything else.

## Destination Resolution

The portal destination (`portalTargetRoomId`) is resolved in ascending level order, with each level overriding the previous:

| Level | Default destination |
|---|---|
| 1 | Start room (global `StartRoomIdAlias`) |
| 2 | Root room of the player's current zone |
| 3+ | Player's saved `portal` setting (if set), otherwise falls back to level 2 destination |

If at level 2 the zone root is the death-recovery room, it falls back to the start room.

## Portal Life Duration

```
portalLifeInSeconds = Mysticism(adj) * 10 + 60
portalLifeInMinutes = floor(portalLifeInSeconds / 60), minimum 1
```

At 0 Mysticism the portal lasts 1 minute. At 100 Mysticism it lasts ~17 minutes.

## Preconditions (all active uses)

- Portal skill level >= 1.
- Player must not be jailed (`TimerExpired("jail")` must be true).

## Instant Teleport (no argument)

- Player must not be in combat.
- Cooldown: **1 real minute** (keyed `portal`).
- Moves the player to `portalTargetRoomId` via `rooms.MoveToRoom`.
- Sends arrival/departure flavor messages to both rooms.

## Sub-commands (level 3+)

### `portal set`

Saves the current room ID to the player's `portal` setting. Future portal uses will teleport here by default.

### `portal unset` / `portal clear`

Clears the saved portal setting, reverting the destination to the level 2 zone-root default.

Both sub-commands fire a `SkillUsed` event and send flavor messages to the room.

## Physical Portal (level 4 only)

### `portal open`

Opens a two-way physical portal between the current room and the portal destination.

1. Cooldown: **1 real minute** (keyed `portal`).
2. If the player already has an open portal (`portal:open` setting is set), both ends of the old portal are closed and notified first.
3. Creates two `TemporaryRoomExit` entries — one in the current room pointing to the destination, one in the destination pointing back.
4. Portal title: `"glowing portal from <username>"` with rainbow color styling.
5. Portal expiry: `portalLifeInMinutes real minutes`.
6. Stores the two room IDs in the `portal:open` setting as `"roomId1:roomId2"` for later cleanup.
7. Fires a `SkillUsed` event (`details: open`).

The portal can be entered by any player using the exit name or by typing `portal` with no argument (caught by the pre-check at the top of the command).

## Notes

- The destination cannot be the same room the player is currently in.
- Only one physical portal can be open per player at a time; opening a new one closes the previous.
