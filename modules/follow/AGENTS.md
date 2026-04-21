# Follow Module Context

## Overview

The `modules/follow` module adds a follow system to GoMud, allowing players and mobs to follow other players or mobs between rooms. When a followed entity moves, all followers are automatically moved to the same destination. Follow relationships can be time-limited (auto-expiring after a set number of rounds).

## Key Components

### Module (`follow.go`)

- Registered as plugin `follow` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user command `follow` and mob command `follow`.
- Exports scripting function `GetFollowers` (callable from JS as `module.follow.GetFollowers(actor)`).

### Internal State

```go
type followId struct {
    userId        int
    mobInstanceId int
}

type FollowModule struct {
    plug         *plugins.Plugin
    followed     map[followId][]followId  // target -> list of followers
    followers    map[followId]followId    // follower -> target
    followLimits map[followId]uint64      // follower -> expiry round number
}
```

State is in-memory only; follow relationships do not persist across server restarts.

### Event Listeners

| Event | Behavior |
|---|---|
| `events.RoomChange` | When the followed entity moves, move all followers to the same room |
| `events.PlayerDespawn` | Stop all follows involving the disconnecting player |
| `events.MobDeath` | Stop all follows involving the dead mob |
| `events.PlayerDeath` | Stop all follows involving the dead player |
| `events.MobIdle` (First priority) | Handle mob following behavior during idle ticks |
| `events.PartyUpdated` | Stop all follows for users whose party membership changed |
| `events.NewRound` | Expire time-limited follows that have passed their cutoff round |

### Scripting API

`module.follow.GetFollowers(actor)` — returns a `[]*ScriptActor` list of all entities currently following the given actor. Usable from mob/room scripts.

## Commands

- **`follow <target>`** — player begins following the named player or mob.
- **`follow`** (no args) — player stops following their current target.
- Mob command `follow` — allows mob AI scripts to issue follow commands.

## File Structure

```
modules/follow/
  follow.go
  files/
    data-overlays/
      config.yaml
      keywords.yaml
    datafiles/templates/help/
      follow.template
```

## Dependencies

- `internal/events`: Event subscriptions for room changes, deaths, party updates
- `internal/parties`: Party change detection
- `internal/plugins`: Plugin and command registration
- `internal/rooms`: Room loading for movement
- `internal/users`, `internal/mobs`: Resolving follow targets
- `internal/scripting`: Exported scripting function
- `internal/gametime`: Round number tracking for expiry
