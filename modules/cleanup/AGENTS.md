# Cleanup Module Context

## Overview

The `modules/cleanup` module adds `trash` and `bury` commands to GoMud, allowing players and mobs to permanently destroy items from their inventory or remove corpses from rooms. It optionally awards experience points for trashing items based on their gold value.

## Key Components

### Module (`cleanup.go`)

- Registered as plugin `cleanup` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user commands `trash` and `bury`, and mob commands `trash` and `bury`.

### Commands

#### `trash <item>`

- Removes a matching item from the player's backpack permanently.
- Fires `events.ItemOwnership` with `Gained: false` to trigger quest/ownership hooks.
- If `TrashExperienceEnabled` is true, grants XP equal to `max(ExperienceValue, itemValue/10)`.
- Sneaking players suppress the room-visible destruction message.
- Mob version silently removes the item from mob inventory.

#### `bury <corpse>`

- Removes a matching corpse from the current room.
- Works for both player and mob corpses.
- Mob version silently removes the corpse.

### Configuration

Loaded from `Modules.cleanup.*` in config:

| Key | Default | Description |
|---|---|---|
| `TrashExperienceEnabled` | `false` | Whether trashing items grants XP |
| `ExperienceValue` | `1` | Minimum XP granted per trash (actual is `max(this, itemValue/10)`) |

## File Structure

```
modules/cleanup/
  cleanup.go
  files/
    data-overlays/
      config.yaml
      keywords.yaml
    datafiles/templates/help/
      bury.template
      trash.template
```

## Dependencies

- `internal/buffs`: Checking `buffs.Hidden` flag for sneak suppression
- `internal/events`: `ItemOwnership` event for quest/hook integration
- `internal/mobs`, `internal/rooms`, `internal/users`: Command execution context
- `internal/plugins`: Plugin and command registration
- `internal/util`: Argument parsing
