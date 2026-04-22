
# Mudmail Module Context

## Overview

The `modules/mudmail` module provides the player inbox system and the admin mass-mail command. Players use `inbox` to read messages sent to them (including attached gold and items). Admins use `mudmail` to compose and send a message to every player account, both online and offline.

## Key Components

### Module (`mudmail.go`)

- Registered as plugin `mudmail` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user command `inbox` (available when downed, non-admin).
- Registers user command `mudmail` (available when downed, admin-only).
- Exports function `SendMudMail` via `plug.ExportFunction` for use by other modules.
- Registers admin page at `/admin/mudmail` (Send Mudmail UI + message list).
- Registers admin API docs page at `/admin/mudmail-api` (nested under Mudmail nav).
- Registers admin API endpoints: `GET`, `POST`, and `DELETE` `/admin/api/v1/mudmail`.

### Admin API (`admin.go`)

Four handler methods on `MudmailModule`:

| Method | Path | Description |
|---|---|---|
| `GET` | `/admin/api/v1/mudmail?user_id=<id>` | List summary entries (no body) for one user's inbox. `user_id` is required. |
| `GET` | `/admin/api/v1/mudmail-body/{user_id}/{timestamp}` | Fetch full message data (including body) for a single message identified by user ID and microsecond timestamp. |
| `POST` | `/admin/api/v1/mudmail` | Send a message to one user (`user_id`) or broadcast to everyone (`user_id` omitted/0). |
| `DELETE` | `/admin/api/v1/mudmail` | Delete one message by `?user_id=<id>&date_sent_us=<us>`. |

All endpoints require admin authentication and the mud lock (applied automatically by the framework).

### Inbox Management

All inbox data and business logic lives in this module. `Message` and `Inbox` types are defined here. Per-user inboxes are stored as plugin data files (`plugin-data/mudmail-v1-0/inbox-user-<id>.plugin.dat`) and are not part of `UserRecord`.

- In-memory store: `map[int]Inbox` keyed by userId, loaded on `PlayerSpawn`, flushed on `PlayerDespawn` and `onSave`.
- Offline delivery: reads the plugin data file directly, appends, and writes back without requiring the user to be online.
- Migration: on `PlayerSpawn`, calls `users.MigrateInbox(userId)` to import any messages stored in the legacy `inbox:` field of old user YAML files. After import the user record is re-saved so the field is absent going forward.

### Exported Function: `SendInboxMessage`
SendMudMail
Signature: `func(userId int, fromName string, message string, gold int, itm *items.Item)`

Other modules that need to deliver a message to a player's inbox must use this exported function rather than accessing inbox data directly:

```go
if sendInbox, ok := usercommands.GetExportedFunction(`SendInboxMessage`); ok {
    if fn, ok := sendInbox.(func(int, string, string, SendMudMail)); ok {
        fn(userId, `System`, `Your message here`, 0, nil)
    }
}
```

If the mudmail module is not loaded, `GetExportedFunction` returns `false` and the call is silently skipped.

## File Structure

```
modules/mudmail/
  mudmail.go
  files/
    data-overlays/
      keywords.yaml
    datafiles/templates/
      admincommands/help/
        command.mudmail.template
      help/
        inbox.md
      mail/
        message.template
```

## Dependencies

- `internal/configs`: `TextFormats.Time` for message date formatting
- `internal/events`: `PlayerSpawn`, `PlayerDespawn`, `EquipmentChange` events
- `internal/items`: `Item` type for message attachments
- `internal/language`: Inbox localisation strings
- `internal/plugins`: Plugin registration, data store (`WriteStruct`/`ReadIntoStruct`)
- `internal/rooms`, `internal/users`: Command execution context
- `internal/templates`: `mail/message` template rendering
- `internal/term`: CRLF constant for mudmail prompt

### Commands

#### `inbox`

Displays unread messages in the player's inbox. On first read, any attached gold is deposited to the player's bank and any attached item is added to inventory.

Subcommands:
- `inbox` - show all unread messages
- `inbox old` - show already-read messages
- `inbox clear` - delete all messages
- `inbox check` - print unread/read counts without displaying messages

#### `mudmail`

Admin-only interactive prompt to compose and send a mass mail to all player accounts.

Prompts: from name, message body, optional gold attachment, optional item attachment, confirmation.

## File Structure

```
modules/mudmail/
  mudmail.go
  files/
    data-overlays/
      keywords.yaml
    datafiles/templates/
      admincommands/help/
        command.mudmail.template
      help/
        inbox.md
      mail/
        message.template
```

## Dependencies

- `internal/events`: `EquipmentChange` for gold deposited from mail
- `internal/language`: Inbox localisation strings
- `internal/plugins`: Plugin and command registration
- `internal/rooms`, `internal/users`: Command execution context
- `internal/templates`: `mail/message` template rendering
- `internal/term`: CRLF constant for mudmail prompt
