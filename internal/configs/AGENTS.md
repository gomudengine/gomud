# GoMud Configuration Management System Context

## Overview

The GoMud configuration system provides comprehensive, type-safe configuration management with YAML-based storage, runtime overrides, environment variable support, and validation. It supports hierarchical configuration structures, dot-notation access, and secure handling of sensitive data through a sophisticated type system and validation framework.

## Architecture

The configuration system is built around a centralized `Config` struct with several key components:

### Core Components

**Configuration Structure:**
- Hierarchical configuration organized into logical subsections
- Type-safe configuration values using custom types (`ConfigString`, `ConfigInt`, `ConfigBool`, etc.)
- Automatic validation and default value enforcement
- Thread-safe access with read-write mutex protection

**Override System:**
- Runtime configuration overrides stored separately from base configuration
- Dot-notation path support for nested configuration access
- Persistent override storage in YAML format (path: `{DataFiles}/config-overrides.yaml` or `$CONFIG_PATH`)
- Automatic path correction and fuzzy matching for configuration keys

**Type System:**
- Custom configuration types with string conversion and validation
- Special `ConfigSecret` type that redacts sensitive values in output (`*** REDACTED ***`)
- Automatic type inference and conversion from string values via `StringToConfigValue`
- Support for complex types including slices and nested structures

**Validation Framework:**
- Per-subsection `Validate()` methods with range checking and defaults
- Locked configuration properties that cannot be changed at runtime
- Banned name patterns for user input validation
- Environment variable integration with automatic assignment

## Config Struct

```go
type Config struct {
    Server       Server
    Memory       Memory
    LootGoblin   LootGoblin
    Timing       Timing
    FilePaths    FilePaths
    GamePlay     GamePlay
    Integrations Integrations
    TextFormats  TextFormats
    Translation  Translation
    Network      Network
    Scripting    Scripting
    SpecialRooms SpecialRooms
    Validation   Validation
    Roles        Roles
    Modules      Modules
}
```

## Configuration Subsections

### Server Configuration
```go
type Server struct {
    MudName         ConfigString      // Name of the MUD
    CurrentVersion  ConfigString      // Current version string
    Seed            ConfigSecret      // Seed for content generation (redacted in output)
    MaxCPUCores     ConfigInt         // CPU cores for multi-core operations
    OnLoginCommands ConfigSliceString // Commands run on user login
    Motd            ConfigString      // Message of the day
    NextRoomId      ConfigInt         // Next room ID for room creation
    Locked          ConfigSliceString // Config properties locked from runtime changes
}
```

### Network Configuration
```go
type Network struct {
    MaxTelnetConnections ConfigInt         // Max telnet connections (default 50)
    TelnetPort           ConfigSliceString // One or more telnet ports
    LocalPort            ConfigInt         // Localhost-only admin port
    HttpPort             ConfigInt         // HTTP port (0 to disable)
    HttpsPort            ConfigInt         // HTTPS port (0 to disable)
    HttpsRedirect        ConfigBool        // Redirect HTTP to HTTPS
    SSHPort              ConfigInt         // SSH port (0 to disable)
    MaxSSHConnections    ConfigInt         // Max SSH connections (default 50)
    AfkSeconds           ConfigInt         // Seconds until AFK
    MaxIdleSeconds       ConfigInt         // Seconds before idle kick
    TimeoutMods          ConfigBool        // Whether to timeout admins/mods
    LinkDeadSeconds      ConfigInt         // Link-dead reconnect window
    LogoutRounds         ConfigInt         // Rounds of meditation required to log out
}
```

### GamePlay Configuration
```go
type GamePlay struct {
    AllowItemBuffRemoval     ConfigBool
    Death                    GameplayDeath
    Party                    GameplayParty
    LivesStart               ConfigInt    // Starting permadeath lives
    LivesMax                 ConfigInt    // Maximum permadeath lives
    LivesOnLevelUp           ConfigInt    // Lives gained on level up
    PricePerLife             ConfigInt    // Gold cost to buy a life
    ShopRestockRate          ConfigString // Default shop restock duration (e.g. "6 hours")
    ContainerSizeMax         ConfigInt    // Max items in a container
    ConsistentAttackMessages ConfigBool
    PVP                      ConfigString // "enabled", "disabled", "limited"
    PVPMinimumLevel          ConfigInt
    XPScale                  ConfigFloat  // XP difficulty multiplier (default 100)
    MobConverseChance        ConfigInt    // 0-100 chance of mob conversing when idle
}

type GameplayDeath struct {
    EquipmentDropChance ConfigFloat  // 0.0-1.0 chance to drop equipment on death
    AlwaysDropBackpack  ConfigBool
    XPPenalty           ConfigString // "none", "level", "10%", "25%", etc.
    ProtectionLevels    ConfigInt    // Levels protected from death penalties
    PermaDeath          ConfigBool
    CorpsesEnabled      ConfigBool
    CorpseDecayTime     ConfigString // Duration string (e.g. "1 hour")
}

type GameplayParty struct {
    MaxPlayerCount ConfigInt  // 0 = unlimited
    SameRoomOnly   ConfigBool
}
```

### File Paths Configuration
```go
type FilePaths struct {
    WebDomain        ConfigString // Web domain name
    WebCDNLocation   ConfigString // Optional CDN for static files
    DataFiles        ConfigString // Path to data files (default "_datafiles/world/default")
    PublicHtml       ConfigString // Public HTML directory
    AdminHtml        ConfigString // Admin HTML directory
    HttpsCertFile    ConfigString // TLS certificate path
    HttpsKeyFile     ConfigString // TLS private key path
    SSHHostKeyFile   ConfigString // SSH host private key (required for SSH)
    CarefulSaveFiles ConfigBool   // Write to .new file then rename
}
```

### Timing Configuration
```go
type Timing struct {
    TurnMs            ConfigInt // Milliseconds per turn (default 100, min 10)
    RoundSeconds      ConfigInt // Seconds per round (default 4)
    RoundsPerAutoSave ConfigInt // Rounds between auto-saves (default 900)
    RoundsPerDay      ConfigInt // Rounds per in-game day (default 20)
    NightHours        ConfigInt // Hours of night (0-24)
}

// Helper methods (calculated and cached on Validate):
func (e Timing) TurnsPerRound() int
func (e Timing) TurnsPerAutoSave() int
func (e Timing) TurnsPerSecond() int
func (e Timing) SecondsToRounds(seconds int) int
func (e Timing) SecondsToTurns(seconds int) int
func (e Timing) MinutesToRounds(minutes int) int
func (e Timing) MinutesToTurns(minutes int) int
func (e Timing) RoundsToSeconds(rounds int) int
```

### Validation Configuration
```go
type Validation struct {
    NameSizeMin      ConfigInt         // Min name length
    NameSizeMax      ConfigInt         // Max name length (max 80)
    PasswordSizeMin  ConfigInt         // Min password length
    PasswordSizeMax  ConfigInt         // Max password length
    NameRejectRegex  ConfigString      // Regex names must match
    NameRejectReason ConfigString      // Reason shown when name rejected
    EmailOnJoin      ConfigString      // "required", "optional", or "none"
    BannedNames      ConfigSliceString // Wildcard patterns for banned names
}
```

## Configuration Types

```go
type ConfigInt         int
type ConfigUInt64      uint64
type ConfigString      string
type ConfigSecret      string     // String() returns "*** REDACTED ***"
type ConfigFloat       float64
type ConfigBool        bool
type ConfigSliceString []string

type ConfigValue interface {
    String() string
    Set(string) error
}
```

`ConfigSecret` hides its value in all string representations. Use `configs.GetSecret(v ConfigSecret) string` to access the raw value.

## Configuration Access Patterns

### Reading Configuration
```go
config := configs.GetConfig()
serverConfig  := configs.GetServerConfig()
networkConfig := configs.GetNetworkConfig()
gameplayConfig := configs.GetGamePlayConfig()
timingConfig  := configs.GetTimingConfig()
filePathsConfig := configs.GetFilePathsConfig()
validationConfig := configs.GetValidationConfig()
rolesConfig   := configs.GetRolesConfig()
```

### Setting Configuration Values
```go
// Set by dot-path; validates, persists to override file, and reloads
err := configs.SetVal("Server.MudName", "New Name")
err := configs.SetVal("Network.HttpPort", "8080")
err := configs.SetVal("GamePlay.PVP", "enabled")
```

### Dot-Notation and Path Resolution
```go
// All config paths support dot notation
allConfig := config.AllConfigData()
// Returns map with keys like "Server.MudName", "Network.HttpPort", etc.

// Fuzzy path lookup (case-insensitive, partial match)
fullPath, typeName := configs.FindFullPath("mudname")
// Returns: "Server.MudName", "configs.ConfigString"
```

### Override System
```go
// Programmatically add overrides (e.g. from module config)
err := configs.AddOverlayOverrides(map[string]any{
    "Server.MudName": configs.ConfigString("Dev Server"),
})

// Get current overrides
overrides := configs.GetOverrides()

// Flatten/unflatten helpers
flat := configs.Flatten(nestedMap)
```

### Reload Configuration
```go
err := configs.ReloadConfig()
```

Reads `_datafiles/config.yaml`, applies overrides from `{DataFiles}/config-overrides.yaml` (or `$CONFIG_PATH`), applies environment variables, and validates.

## Validation System

Each subsection has a `Validate()` method enforcing defaults and ranges. The top-level `Config.Validate()` calls all subsection validators and caches computed values (e.g. `seedInt`, timing calculations).

### Banned Name Validation
```go
bannedPattern, isBanned := config.IsBannedName("testname")
```
Uses wildcard matching against `Validation.BannedNames`.

### Locked Configuration Properties
```go
// Properties listed in Server.Locked cannot be changed via SetVal at runtime
Server:
  Locked: ["Seed"]
```

## PVP Constants
```go
const (
    PVPEnabled  = "enabled"
    PVPDisabled = "disabled"
    PVPOff      = "off"      // normalized to "disabled" on Validate
    PVPLimited  = "limited"
)
```

## Performance Considerations

- Configuration is cached in memory with `sync.RWMutex` protection
- `validated` flag prevents redundant validation calls
- Key/type lookup tables (`keyLookups`, `typeLookups`) provide O(1) path resolution after initial build
- Timing helper values (`turnsPerRound`, etc.) are pre-calculated and cached on `Validate()`

## Dependencies

- `gopkg.in/yaml.v2` - YAML parsing and generation
- `internal/mudlog` - Logging and monitoring
- `internal/util` - File operations and utilities
- `sync` - Thread-safe access control
- `reflect` - Dynamic configuration introspection
- `os` - Environment variable access and file operations
