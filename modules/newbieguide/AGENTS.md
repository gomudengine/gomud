# Newbie Guide Module Context

## Overview

The `modules/newbieguide` module automatically spawns a guide mob to accompany new players (level 1–5) as they explore the world. The guide provides contextual help and offers a portal back to Town Square. It despawns itself when the player reaches level 5.

## Key Components

### Module (`newbieguide.go`)

- Registered as plugin `newbieguide` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- No user commands registered.
- Registers two event listeners.

### Behavior

**On `events.RoomChange`** (`spawnGuide`):
- Only fires for users (not mobs).
- Skips players above level 5.
- Skips movement into/out of rooms 900–999 (tutorial/special rooms).
- Rate-limited: will not spawn more frequently than once every 300 seconds (converted to rounds) per player, tracked via player temp data (`lastGuideRound`).
- Checks if the player already has a guide mob charmed; skips if so.
- Spawns a new instance of the guide mob (default mob ID 38, configurable via `GuideMobId` plugin config key).
- Names the mob `"<PlayerName>'s Guide"`.
- Permanently charms the mob to the player (`CharmPermanent`, `CharmExpiredDespawn`).
- The guide greets the player and offers to portal them back to Town Square.

**On `events.LevelUp`** (`checkGuide`):
- When a player reaches level 5 or higher, finds any charmed guide mob and commands it to say farewell, emote a disappearance, and then `suicide vanish`.

### Configuration

From `Modules.newbieguide.*` in config:

| Key | Default | Description |
|---|---|---|
| `GuideMobId` | `38` | Mob ID to use as the guide |

## File Structure

```
modules/newbieguide/
  newbieguide.go
  files/
    data-overlays/
      config.yaml
```

## Dependencies

- `internal/characters`: `CharmPermanent`, `CharmExpiredDespawn` constants
- `internal/configs`: `GetTimingConfig().SecondsToRounds` for rate limiting
- `internal/events`: `RoomChange` and `LevelUp` event subscriptions
- `internal/mobs`: `NewMobById`, `GetInstance` for guide mob management
- `internal/plugins`: Plugin registration
- `internal/rooms`: `LoadRoom`, `GetOriginalRoom` for room validation
- `internal/users`: User record access
- `internal/util`: `GetRoundCount` for rate limiting
