# Time Module Context

## Overview

The `modules/time` module adds a `time` command to GoMud that displays the current in-game date and time to the player. It is a minimal single-command module with no persistent state.

## Key Components

### Module (`time.go`)

- Registered as plugin `time` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user command `time` via `plug.AddUserCommand`.

### Command: `time`

Displays the current in-game time, including:
- Formatted time string (e.g., "dawn", "mid-morning", etc.)
- Day/night indicator with colored output
- Day number within the year
- Year number
- Month name
- Zodiac sign for the year

Accepts an optional argument for testing: `time <period>` — looks up the game date for the specified period string via `gametime.GetLastPeriod`.

## File Structure

```
modules/time/
  time.go
  files/
    data-overlays/
      keywords.yaml
    datafiles/templates/help/
      time.template
```

## Dependencies

- `internal/events`: `EventFlag` for command signature
- `internal/gametime`: `GetDate`, `GetLastPeriod`, `MonthName`, `GetZodiac`
- `internal/plugins`: Plugin and command registration
- `internal/rooms`, `internal/users`: Command signature requirements
