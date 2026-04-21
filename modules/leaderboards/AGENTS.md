# Leaderboards Module Context

## Overview

The `modules/leaderboards` module tracks and displays server-wide leaderboards for gold, experience, and kill counts. Rankings are recalculated periodically (every round) and cover all characters for all users, including offline players and alt characters. Leaderboard data is persisted across server restarts and is also exposed as a public web page.

## Key Components

### Module (`leaderboards.go`)

- Registered as plugin `leaderboards` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user command `leaderboard`.
- Registers `OnLoad`/`OnSave` callbacks for persistence via `plug.ReadIntoStruct` / `plug.WriteStruct`.
- Registers a `NewRound` event listener to trigger periodic recalculation.
- Exposes a web page at `/leaderboards` (`leaderboards.html`, public).

### Leaderboard Types

Three leaderboards are tracked (each independently toggleable via config):

| Leaderboard | Metric | Config Key |
|---|---|---|
| Gold | `character.Gold + character.Bank` | `GoldEnabled` |
| Experience | `character.Experience` | `ExperienceEnabled` |
| Kills | `character.KD.TotalKills` | `KillsEnabled` |

### Data Coverage

The `Update()` method scans:
- All currently active (online) users and their alt characters
- All offline users (via `users.SearchOfflineUsers`) and their alts

This ensures the leaderboard reflects the entire player base, not just online players.

### Configuration

Loaded from `Modules.leaderboards.*` in config:

| Key | Default | Description |
|---|---|---|
| `Size` | `10` | Number of entries per leaderboard |
| `GoldEnabled` | `true` | Enable gold leaderboard |
| `ExperienceEnabled` | `true` | Enable experience leaderboard |
| `KillsEnabled` | `true` | Enable kills leaderboard |

### Persistence

Leaderboard state is saved to the plugin data store under the key `latest-leaderboards` using YAML serialization. Loaded on server start, saved on each autosave cycle.

## File Structure

```
modules/leaderboards/
  leaderboards.go
  files/
    data-overlays/
      config.yaml
      keywords.yaml
    datafiles/html/public/
      leaderboards.html
    datafiles/templates/help/
      leaderboard.template
```

## Dependencies

- `internal/characters`: `LoadAlts` for alt character scanning
- `internal/events`: `NewRound` for periodic recalculation
- `internal/plugins`: Plugin, command, and web page registration
- `internal/skills`: Skill data for leaderboard entries
- `internal/templates`: Table rendering for in-game display
- `internal/users`: `GetAllActiveUsers`, `SearchOfflineUsers`
- `internal/util`: Number formatting
