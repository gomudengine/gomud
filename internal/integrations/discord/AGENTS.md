# Discord Integration Context

## Overview

The `internal/integrations/discord` package provides outbound Discord webhook integration for GoMud. It listens for game events and forwards notifications to a configured Discord channel via webhook. It does **not** receive messages from Discord.

## Key Components

### Client (`client.go`)

- **`Init(webhookUrl string)`**: Initializes the client, stores the webhook URL, and registers all event listeners. Safe to call multiple times (no-ops after first call).
- **`SendMessage(message string)`**: Sends a plain-text webhook message.
- **`SendRichMessage(message string, color Color)`**: Sends a Discord embed with a colored sidebar.
- **`SendPayload(payload webHookPayload)`**: Sends a fully constructed payload (used for structured embeds with fields).
- **Backoff mechanism**: On HTTP failure or non-204 response, requests are suppressed for `RequestFailureBackoffSeconds` (30s) to avoid hammering Discord during outages.
- All HTTP sends are non-blocking (executed in a goroutine). Timeouts are 3 seconds for dial, TLS, response headers, and 1 second for continue.

### Event Listeners (`listeners.go`)

Registered automatically by `Init`:

| Game Event | Discord Message |
|---|---|
| `events.PlayerSpawn` | Player connected (with connection type: telnet/websocket/ssh) |
| `events.PlayerDespawn` | Player disconnected (with time online) |
| `events.Log` (ERROR level only) | Error notification; script timeout errors are suppressed |
| `events.LevelUp` | Level-up announcement |
| `events.PlayerDeath` | Death announcement |
| `events.Broadcast` (IsCommunication only) | In-game broadcast channel message |
| `AuctionUpdate` (GenericEvent) | Auction start, bid, and end events |

### Types (`types.go`)

- **`Color int`**: Named Discord embed color constants (e.g., `Green`, `Red`, `Gold`, `Purple`, `Grey`, `DarkOrange`)
- **`webHookPayload`**: Top-level Discord webhook JSON structure (`username`, `avatar_url`, `content`, `embeds`)
- **`embed`**: Discord embed object with `title`, `description`, `color`, `fields`, `image`, `thumbnail`, `footer`
- **`embedField`**: Field within an embed (`name`, `value`, `inline`)

## Configuration

Discord integration is enabled by calling `discord.Init(webhookUrl)` from the application startup code. The webhook URL is sourced from `config.yaml` (`Integrations.Discord.WebhookUrl`). If the URL is empty, `Init` is not called and no events are forwarded.

## Usage Patterns

```go
// Initialize once at startup
discord.Init(cfg.Integrations.Discord.WebhookUrl)

// The package then handles all events automatically via registered listeners
// Direct calls are also possible:
discord.SendRichMessage("Server restarting", discord.Red)
```

## Dependencies

- `internal/events`: Event system for all game event subscriptions
- `internal/users`: Resolving user data for player events
- `internal/connections`: Determining connection type (telnet/websocket/ssh)
- `internal/util`: ANSI stripping for clean Discord output
- `github.com/GoMudEngine/ansitags`: Stripping ANSI tags from broadcast messages
- Standard library: `net/http`, `encoding/json`, `sync`, `time`

## Special Considerations

- **No inbound messages**: This package only sends to Discord; it cannot receive messages.
- **Thread safety**: The backoff timer uses `sync.RWMutex` for concurrent access safety.
- **ANSI stripping**: All game text passed to Discord has ANSI/color tags stripped before sending.
- **Auction integration**: The `AuctionUpdate` listener uses `events.GenericEvent` with typed data fields rather than a concrete event struct, since the auction module is a plugin.
