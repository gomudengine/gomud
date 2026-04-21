# Zombie Mode Module Context

## Overview

The `modules/zombiemode` module adds an AFK automation ("zombie mode") system to GoMud. When active, it takes over a player's character and autonomously performs configurable actions: attacking mobs, looting items, roaming within a radius, and resting when HP is low. Any real player input immediately deactivates zombie mode. The module also handles the existing disconnection-triggered zombie adjective (idle flavor messages).

## Key Components

### Module (`zombiemode.go`)

- Registered as plugin `zombiemode` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user commands `zombie` and `zombieact`.
- Registers `OnSave` callback for config persistence.

### State

```go
type ZombieConfig struct {
    CombatTargets []string                // Mob name substrings to attack; "*" = all
    RoamRadius    int                     // Max rooms from home room to roam; 0 = disabled
    RestThreshold int                     // HP% below which to rest/flee; 0 = disabled
    LootTargets   []string                // Item name substrings to pick up; "*" = all
    Profiles      map[string]ZombieConfig // Named saved profiles (max 5)
}

type zombieRuntime struct {
    HomeRoom int         // Room where zombie mode was started
    Stats    zombieStats // Session kill/loot/XP counters
}
```

- `configs`: per-userId `ZombieConfig` — loaded on `PlayerSpawn`, saved on `PlayerDespawn` and autosave.
- `active`: per-userId `zombieRuntime` — present only while voluntary zombie mode is running.

### AI Behavior (`zombieact.go`)

The `zombieact` command (called each round by the game loop for players with the `zombie` adjective) implements a priority-ordered decision tree:

1. **Rest check**: If HP% < `RestThreshold`, flee if in combat, otherwise idle.
2. **Continue combat**: If already in combat, keep attacking current target.
3. **Initiate combat**: Scan room for mobs matching `CombatTargets`, attack first match.
4. **Loot**: Pick up gold or items matching `LootTargets` from the floor.
5. **Roam**: Move to a random exit within `RoamRadius` rooms of home room (uses mapper for pathfinding; falls back to heading toward home if out of radius).
6. **Idle**: 20% chance to emit a flavor message.

AI-issued commands are tagged with `cmdZombieAI` (`EventFlag = 0b00100000`) so the input interceptor can distinguish them from real player input.

### Command: `zombie` (`command.zombie.go`)

Subcommands:
- `zombie start` — activates voluntary zombie mode; sets `zombie` adjective and records home room.
- `zombie set <combat|roam|rest|loot> <value>` — configures AI behavior.
- `zombie unset <combat|roam|rest|loot> [name]` — removes a setting.
- `zombie save <name>` — saves current config as a named profile (max 5 profiles).
- `zombie load <name>` — loads a saved profile.
- `zombie list` — lists all saved profiles with their settings.
- `zombie delete <name>` — deletes a saved profile.
- `zombie` (no args) — displays current config and active status.

### Event Listeners

| Event | Behavior |
|---|---|
| `events.PlayerSpawn` | Load user config; clear any stale active entry |
| `events.PlayerDespawn` | Persist config; exit zombie mode |
| `events.PlayerDrop` | Exit zombie mode on death; enqueue session summary |
| `events.AggroChanged` | When a mob targets a zombie player, auto-attack back |
| `events.Input` (First) | Any non-AI input wakes the player and exits zombie mode |
| `events.MobDeath` | Record mob kills in session stats |
| `events.EquipmentChange` | Track gold gained in session stats |
| `events.ItemOwnership` | Track items looted in session stats |
| `events.GainExperience` | Track XP gained in session stats |
| `events.LevelUp` | Track levels gained in session stats |
| `zombieSummary` (local event) | Display session stats to the player after mode ends |

### Configuration

From `Modules.zombiemode.*` in config:

| Key | Default | Description |
|---|---|---|
| `Enabled` | `true` | Server-side toggle; disables voluntary zombie mode if false |

## File Structure

```
modules/zombiemode/
  zombiemode.go       # Module setup, state, event listeners
  zombieact.go        # AI decision tree (zombieact command)
  command.zombie.go   # zombie command (config/profile management)
  files/
    data-overlays/
      config.yaml
    datafiles/templates/help/
      zombie.template
```

## Dependencies

- `internal/events`: All event subscriptions and AI command injection
- `internal/mapper`: Pathfinding for roam radius enforcement
- `internal/mobs`: Mob instance lookup for combat targeting
- `internal/plugins`: Plugin and command registration
- `internal/rooms`: Room loading, exit enumeration, floor item scanning
- `internal/users`: User record access
- `internal/util`: Random number generation, argument parsing
