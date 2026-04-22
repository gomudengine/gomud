
# Storage Module Context

## Overview

The `modules/storage` module provides the player item storage system. Players can deposit and retrieve items at rooms tagged with `storage`. Storage data is persisted in the plugin data store (not in `UserRecord`), and legacy data from `UserRecord.ItemStorage` is automatically migrated on first login after upgrading.

## Room Tag

Rooms opt into storage functionality by adding the `storage` tag:

```yaml
tags:
  - storage
```

The module registers an `OnRoomLook` hook that injects the storage alert into the room description when the tag is present. The core `roomdetails.go` no longer has a hardcoded `IsStorage` alert.

## Key Components

### Module (`storage.go`)

- Registered as plugin `storage` version `1.0`.
- Embeds data files from `files/` using `//go:embed files/*`.
- Registers user command `storage`.
- Registers `OnSave` callback for persistence.
- Exports functions `GetStorageItems`, `AddStorageItem`, `RemoveStorageItem` for cross-module use (e.g. autocomplete).
- Registers admin page at `/admin/storage` ("View / Edit" sub-item under "Storage" nav entry).
- Registers admin API docs page at `/admin/storage-api` ("API Docs" sub-item under "Storage" nav entry).
- Registers admin API endpoints: `GET` and `DELETE` `/admin/api/v1/storage`.
- Listens to `PlayerSpawn` to load (and migrate legacy) storage data.
- Listens to `PlayerDespawn` to save and unload storage data.
- Registers `rooms.OnRoomLook` hook to inject the storage room alert.
- Registers `suggestions.OnAutoComplete` hook to provide tab-completion for `storage add` and `storage remove`.

### Data Structures

```go
type StorageData struct {
    Items []items.Item `yaml:"items,omitempty"`
}

type StorageModule struct {
    plug    *plugins.Plugin
    storage map[int]StorageData // keyed by userId; loaded on PlayerSpawn
}
```

### Command: `storage`

- `storage` — lists items in storage.
- `storage add <item>` — moves item from backpack to storage (max 20 items).
- `storage add all` — moves all backpack items to storage.
- `storage remove <item>` — retrieves item from storage to backpack.
- `storage remove <number>` — retrieves item by position number.
- `storage remove all` — retrieves all items from storage.

### Admin API (`admin.go`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/admin/api/v1/storage?user_id=<id>` | List all stored items for a user. |
| `DELETE` | `/admin/api/v1/storage?user_id=<id>&short_id=<sid>` | Remove a specific item by short ID. |

### Legacy Migration

On `PlayerSpawn`, if the plugin data store is empty for the user but `UserRecord.ItemStorage` contains items, those items are migrated to the plugin store and `UserRecord.ItemStorage` is cleared.

### Exported Functions

- `GetStorageItems(userId int) []items.Item` — returns a copy of the user's stored items. Used by `internal/suggestions/autocomplete.go` for `storage remove` tab completion.
- `AddStorageItem(userId int, itm items.Item) bool` — adds an item to the user's storage.
- `RemoveStorageItem(userId int, itm items.Item) bool` — removes an item from the user's storage.

## File Structure

```
modules/storage/
  storage.go
  admin.go
  files/
    data-overlays/
      keywords.yaml
    datafiles/
      html/admin/
        storage.html
        storage-api.html
      templates/help/
        storage.template
```

## Changes to Core Engine

- `internal/usercommands/storage.go` — deleted; command is now owned by the module.
- `internal/usercommands/usercommands.go` — `storage` entry removed from the core command map.
- `internal/usercommands/default.go` — uses `room.Tags` loop instead of `room.IsStorage`; dispatches via `user.Command("storage")` to avoid init cycles.
- `internal/rooms/roomdetails.go` — `IsStorage` alert block removed; the module's `OnRoomLook` hook handles it.
- `modules/gmcp/gmcp.Room.go` — `IsStorage` detail block removed; the existing `room.Tags` loop already adds `storage` to GMCP details.
- `modules/gmcp/gmcp.World.go` — `IsStorage` detail block replaced with a `room.Tags` loop.
- `internal/suggestions/autocomplete.go` — `storage remove` completion now calls the exported `GetStorageItems` function instead of accessing `user.ItemStorage` directly.

## Dependencies

- `internal/events`: `PlayerSpawn`, `PlayerDespawn`, `ItemOwnership`
- `internal/items`: Item type for storage data
- `internal/plugins`: Plugin registration, data store
- `internal/rooms`: `OnRoomLook` hook, `Room` type
- `internal/users`: User record access, `SaveUser`, `NewUserIndex`
- `internal/util`: `SplitButRespectQuotes`
