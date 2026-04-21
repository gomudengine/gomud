# GoMud Hooks System Context

## Overview

The GoMud hooks system provides comprehensive event-driven game logic through a collection of specialized event listeners that handle everything from combat rounds to quest progression. It serves as the primary integration layer between the event system and game mechanics, implementing core gameplay features like combat resolution, mob AI, player lifecycle management, and system maintenance tasks.

> **Note:** This package (`internal/hooks`) contains *event listeners* - asynchronous, fire-and-forget handlers wired to the event queue. It is distinct from the synchronous `util.Hook[T]` callback chain described below, which is used when a caller needs to receive a modified return value.

## Two Hook Concepts

### 1. Event Listeners (`internal/hooks` package)
Asynchronous, queue-based. Used when you want to *notify* that something happened. Handlers do not return values to the original caller. Registered via `events.RegisterListener`.

### 2. Synchronous Data Hooks (`util.Hook[T]`)
Synchronous, return-value-carrying callback chains. Used when a function needs to allow external code to *modify data* before it is returned. Defined as package-level variables on the packages that own the data. See `internal/util/hook.go`.

Currently defined synchronous hooks:
- **`rooms.OnGetDetails`** (`util.Hook[RoomTemplateDetails]`) - fired at the end of `rooms.GetDetails`, giving modules the opportunity to modify the fully-built room details (e.g. add room alerts) before they reach the caller.

## Architecture

The hooks system is built around several key categories:

### Core Components

**Event Registration System:**
- Centralized listener registration in `RegisterListeners()`
- Type-safe event handling with proper casting
- Ordered execution with priority support (`events.Last`)
- Comprehensive coverage of all game events

**Game Loop Hooks:**
- **NewRound Events**: Combat, healing, mob AI, player ticks
- **NewTurn Events**: Autosave, cleanup, buff management
- **Player Lifecycle**: Spawn, despawn, character changes
- **System Maintenance**: VM pruning, link-dead cleanup, respawns

**Gameplay Integration:**
- **Combat System**: Full combat round processing with multi-target support
- **Quest System**: Progress tracking and reward distribution
- **Buff System**: Application, expiration, and effect processing
- **Audio System**: MSP sound effects and location-based music

## Complete Listener Registration

All listeners are registered in `RegisterListeners()` in `hooks.go`:

### Buff Handlers
```go
events.RegisterListener(events.Buff{}, ApplyBuffs)
```

### RoomChange Handlers
```go
events.RegisterListener(events.RoomChange{}, LocationMusicChange)
events.RegisterListener(events.RoomChange{}, CleanupEphemeralRooms)
events.RegisterListener(events.RoomChange{}, SpawnGuide)
```

### NewRound Handlers (11 handlers)
```go
events.RegisterListener(events.NewRound{}, PruneVMs)
events.RegisterListener(events.NewRound{}, InactivePlayers)
events.RegisterListener(events.NewRound{}, UpdateZoneMutators)
events.RegisterListener(events.NewRound{}, CheckNewDay)
events.RegisterListener(events.NewRound{}, SpawnLootGoblin)
events.RegisterListener(events.NewRound{}, UserRoundTick)
events.RegisterListener(events.NewRound{}, MobRoundTick)
events.RegisterListener(events.NewRound{}, HandleRespawns)
events.RegisterListener(events.NewRound{}, DoCombat)   // Combat goes here
events.RegisterListener(events.NewRound{}, AutoHeal)
events.RegisterListener(events.NewRound{}, IdleMobs)
events.RegisterListener(events.MobIdle{}, HandleIdleMobs)
```

### NewTurn Handlers (4 handlers)
```go
events.RegisterListener(events.NewTurn{}, CleanupLinkDead)
events.RegisterListener(events.NewTurn{}, AutoSave)
events.RegisterListener(events.NewTurn{}, PruneBuffs)
events.RegisterListener(events.NewTurn{}, ActionPoints)
```

### Player Lifecycle Handlers
```go
events.RegisterListener(events.PlayerSpawn{}, HandleJoin)
events.RegisterListener(events.PlayerDespawn{}, HandleLeave, events.Last) // final listener
events.RegisterListener(events.PlayerDrop{}, HandlePlayerDrop)
events.RegisterListener(events.CharacterCreated{}, BroadcastNewChar)
events.RegisterListener(events.CharacterChanged{}, BroadcastNewChar)
```

### Game Mechanics Handlers
```go
events.RegisterListener(events.ItemOwnership{}, CheckItemQuests)
events.RegisterListener(events.MSP{}, PlaySound)
events.RegisterListener(events.Quest{}, HandleQuestUpdate)
events.RegisterListener(events.LevelUp{}, SendLevelNotifications)
events.RegisterListener(events.LevelUp{}, CheckGuide)
events.RegisterListener(events.DayNightCycle{}, NotifySunriseSunset)
events.RegisterListener(events.Looking{}, HandleLookHints)
```

### Messaging and UI Handlers
```go
events.RegisterListener(events.Message{}, Message_SendMessage)
events.RegisterListener(events.RedrawPrompt{}, RedrawPrompt_SendRedraw)
events.RegisterListener(events.UserSettingChanged{}, ClearSettingCaches)
events.RegisterListener(events.WebClientCommand{}, WebClientCommand_SendWebClientCommand)
events.RegisterListener(events.Broadcast{}, Broadcast_SendToAll)
```

### System Handlers
```go
events.RegisterListener(events.RebuildMap{}, HandleMapRebuild)
events.RegisterListener(events.Log{}, FollowLogs)
```

## Handler Files

Each listener is implemented in its own file, named `EventType_HandlerName.go`:

| File | Event | Handler |
|------|-------|---------|
| `Buff_ApplyBuffs.go` | `Buff` | `ApplyBuffs` |
| `RoomChange_LocationMusicChange.go` | `RoomChange` | `LocationMusicChange` |
| `RoomChange_CleanupEphemeralRooms.go` | `RoomChange` | `CleanupEphemeralRooms` |
| `RoomChange_SpawnGuide.go` | `RoomChange` | `SpawnGuide` |
| `NewRound_PruneVMs.go` | `NewRound` | `PruneVMs` |
| `NewRound_InactivePlayers.go` | `NewRound` | `InactivePlayers` |
| `NewRound_UpdateZoneMutators.go` | `NewRound` | `UpdateZoneMutators` |
| `NewRound_CheckNewDay.go` | `NewRound` | `CheckNewDay` |
| `NewRound_SpawnLootGoblin.go` | `NewRound` | `SpawnLootGoblin` |
| `NewRound_UserRoundTick.go` | `NewRound` | `UserRoundTick` |
| `NewRound_MobRoundTick.go` | `NewRound` | `MobRoundTick` |
| `NewRound_HandleRespawns.go` | `NewRound` | `HandleRespawns` |
| `NewRound_DoCombat.go` | `NewRound` | `DoCombat` |
| `NewRound_AutoHeal.go` | `NewRound` | `AutoHeal` |
| `NewRound_IdleMobs.go` | `NewRound` | `IdleMobs` |
| `MobIdle_HandleIdleMobs.go` | `MobIdle` | `HandleIdleMobs` |
| `NewTurn_CleanupLinkDead.go` | `NewTurn` | `CleanupLinkDead` |
| `NewTurn_AutoSave.go` | `NewTurn` | `AutoSave` |
| `NewTurn_PruneBuffs.go` | `NewTurn` | `PruneBuffs` |
| `NewTurn_ActionPoints.go` | `NewTurn` | `ActionPoints` |
| `ItemOwnership_CheckItemQuests.go` | `ItemOwnership` | `CheckItemQuests` |
| `MSP_PlaySound.go` | `MSP` | `PlaySound` |
| `Quest_HandleQuestUpdate.go` | `Quest` | `HandleQuestUpdate` |
| `PlayerSpawn_HandleJoin.go` | `PlayerSpawn` | `HandleJoin` |
| `PlayerDespawn_HandleLeave.go` | `PlayerDespawn` | `HandleLeave` |
| `PlayerDrop_HandlePlayerDrop.go` | `PlayerDrop` | `HandlePlayerDrop` |
| `LevelUp_SendLevelNotifications.go` | `LevelUp` | `SendLevelNotifications` |
| `LevelUp_CheckGuide.go` | `LevelUp` | `CheckGuide` |
| `DayNightCycle_NotifySunriseSunset.go` | `DayNightCycle` | `NotifySunriseSunset` |
| `Looking_HandleLookHints.go` | `Looking` | `HandleLookHints` |
| `Message_SendMessages.go` | `Message` | `Message_SendMessage` |
| `RedrawPrompt_SendRedraw.go` | `RedrawPrompt` | `RedrawPrompt_SendRedraw` |
| `UserSettingChanged_ClearSettingCaches.go` | `UserSettingChanged` | `ClearSettingCaches` |
| `WebClientCommand_SendWebClientCommand.go` | `WebClientCommand` | `WebClientCommand_SendWebClientCommand` |
| `CharacterCreated_BroadcastNewChar.go` | `CharacterCreated` | `BroadcastNewChar` |
| `Broadcast_SendToAll.go` | `Broadcast` | `Broadcast_SendToAll` |
| `RebuildMap_HandleMapRebuild.go` | `RebuildMap` | `HandleMapRebuild` |
| `Log_FollowLogs.go` | `Log` | `FollowLogs` |

## Integration Patterns

### Event System Integration
```go
// All hooks integrate with the event system
- events.RegisterListener()        // Register event handlers
- events.AddToQueue()             // Queue new events from handlers
- events.Continue / events.Cancel // Control event processing flow
- events.Last                     // Priority flag: run this listener last
```

### Cross-System Communication
```go
// Hooks coordinate between systems
- users.GetByUserId()             // User management integration
- rooms.LoadRoom()                // Room system integration
- mobs.GetInstance()              // Mob system integration
- combat.AttackPlayerVsMob()      // Combat system integration
- scripting.TryRoomScriptEvent()  // Scripting system integration
```

## Dependencies

- `internal/events` - Event system for listener registration and event processing
- `internal/users` - User management for player-related hooks
- `internal/mobs` - NPC management for mob-related hooks
- `internal/combat` - Combat system for battle resolution
- `internal/quests` - Quest system for progression tracking
- `internal/rooms` - Room management for location-based events
- `internal/scripting` - JavaScript runtime for script execution
- `internal/buffs` - Status effects for buff management
- `internal/configs` - Configuration management for system settings
- `internal/mudlog` - Logging system for debugging and monitoring
