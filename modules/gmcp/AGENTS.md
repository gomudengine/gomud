# GMCP Module Context

## Overview

The `modules/gmcp` module implements the Generic MUD Communication Protocol (GMCP) for GoMud. GMCP is a Telnet sub-negotiation protocol (IAC SB 201 ... IAC SE) that allows structured JSON data to be exchanged between the server and MUD clients. The module also supports a web-socket variant using `!!GMCP(...)` text prefixes for the built-in web client.

## Key Components

### Core Module (`gmcp.go`)

- Registered as plugin `gmcp` version `1.1`.
- Uses an LRU cache (128 entries) keyed by `connectionId` to store per-connection `GMCPSettings`.
- WebSocket connections are automatically treated as GMCP-enabled with a synthetic `WebClient` identity.
- Telnet connections receive a GMCP enable request (`IAC WILL 201`) on connect.
- Exports scripting functions:
  - `SendGMCPEvent(userId, module, payload)` — send arbitrary GMCP data from scripts
  - `IsMudlet(connectionId)` — returns true if the client identified itself as Mudlet

### GMCP Namespaces

| File | Namespace | Description |
|---|---|---|
| `gmcp.Char.go` | `Char.*` | Character stats, vitals, equipment, skills, quests, cooldowns |
| `gmcp.Comm.go` | `Comm.*` | Channel communication events |
| `gmcp.Game.go` | `Game.*` | Server info, online player list |
| `gmcp.Gametime.go` | `Gametime.*` | In-game time and calendar |
| `gmcp.Help.go` | `Help.*` | Help topic data for client-side help browsers |
| `gmcp.Mudlet.go` | `MudletMap.*` | Mudlet-specific map data (room info, exits, area names) |
| `gmcp.Party.go` | `Party.*` | Party membership and member stats |
| `gmcp.Room.go` | `Room.*` | Room info, exits, contents, map data |
| `gmcp.Suggestion.go` | `Suggestion.*` | Command auto-complete suggestions |
| `gmcp.World.go` | `World.*` | World-level events (broadcasts, deaths, level-ups) |

### Protocol Details

- **Telnet GMCP**: IAC SB 201 `<module> <json>` IAC SE
- **WebSocket GMCP**: Text prefix `!!GMCP(<module> <json>)`
- **Inbound negotiation**: Handles `Core.Hello`, `Core.Supports.Set`, `Core.Supports.Remove`, `Core.Login`, `Core.Request`
- **`GMCPOut` event**: Other packages fire `GMCPOut{UserId, Module, Payload}` events to send GMCP data; the module's `dispatchGMCP` listener serializes and delivers them.

### Settings

```go
type GMCPSettings struct {
    Client struct {
        Name     string
        Version  string
        IsMudlet bool
    }
    GMCPAccepted   bool
    EnabledModules map[string]int
}
```

## File Structure

```
modules/gmcp/
  gmcp.go
  gmcp.Char.go
  gmcp.Comm.go
  gmcp.Game.go
  gmcp.Gametime.go
  gmcp.Help.go
  gmcp.Mudlet.go
  gmcp.Party.go
  gmcp.Room.go
  gmcp.Suggestion.go
  gmcp.World.go
  files/
    data-overlays/
      config.yaml
      keywords.yaml
    datafiles/templates/help/
      gmcp.template, gmcp-char.template, gmcp-room.template, ...
```

## Dependencies

- `internal/connections`: Connection type detection and sending raw bytes
- `internal/events`: `GMCPOut` event dispatch and `PlayerSpawn` for initial data push
- `internal/plugins`: Plugin registration, IAC handler, text prefix handler, net-connect callback
- `internal/term`: Telnet IAC constants and `TerminalCommand` wrappers
- `internal/users`: User data for character GMCP payloads
- `github.com/hashicorp/golang-lru/v2`: LRU cache for per-connection settings

## Special Considerations

- **Mudlet compatibility**: When `IsMudlet` is true, certain ANSI/escape sequences must be suppressed. The `IsMudlet` scripting export allows other systems to check this.
- **Web client**: WebSocket connections bypass Telnet negotiation and are always GMCP-enabled.
- **Module versioning**: `EnabledModules` tracks client-declared GMCP module versions, though not all clients populate this correctly.
