# GoMud Users System Context

## Overview

The GoMud users system provides comprehensive user account management with support for authentication, character association, connection tracking, item storage, messaging, and configuration management. It features a sophisticated indexing system for fast user lookups, link-dead connection handling, role-based permissions, and persistent user data storage with YAML serialization.

## Architecture

The users system is built around several key components:

### Core Components

**User Management:**
- Active user tracking with connection mapping
- Role-based permission system (guest, user, admin)
- Link-dead connection handling for graceful disconnections
- Thread-safe user operations with proper cleanup

**User Index System:**
- High-performance binary index for user lookups
- Fixed-width record format for fast seeking
- Username-to-UserID mapping with collision handling
- Automatic index rebuilding and maintenance

**User Storage:**
- YAML-based persistent user data storage
- Item storage system for player belongings
- Configuration options and customization settings

**Connection Integration:**
- Connection ID to User ID mapping
- Real-time connection state tracking
- Input handling and prompt system integration
- Client settings and display preferences

## Key Features

### 1. **Comprehensive User Management**
- **Authentication**: bcrypt password hashing with migration from legacy SHA256 hashes
- **Role System**: Guest, user, and admin roles with permissions
- **Connection Tracking**: Real-time user connection mapping
- **Link-Dead Handling**: Graceful disconnection and cleanup

### 2. **High-Performance Index System**
- **Binary Index**: Fast username lookups with O(log n) performance
- **Fixed Records**: 89-byte fixed-width records for efficient seeking
- **Automatic Maintenance**: Index rebuilding and corruption recovery
- **Version Management**: Versioned index format with migration support

### 3. **Rich User Data Model**
- **Character Integration**: Full character system association
- **Item Storage**: Personal item storage separate from inventory
- **Customization**: Macros, aliases, and configuration options

### 4. **Advanced Features**
- **Screen Reader Support**: Accessibility features for visually impaired users
- **Audio Integration**: Music and sound effect tracking via `PlayMusic`/`PlaySound`
- **Tip System**: Tutorial completion tracking
- **Temporary Data**: Session-based data storage for scripting
- **Alt Characters**: Support for multiple characters per account via `SwapToAlt`
- **Wimpy System**: Auto-flee at configurable health percentage via `WimpyCheck`
- **Connection Type Tracking**: Telnet, WebSocket, and SSH connection type reported in `OnlineInfo`

## User Structure

### User Record Structure
```go
type UserRecord struct {
    UserId         int                   // Unique user identifier
    Role           string                // Permission role (guest/user/admin)
    Username       string                // Login username
    Password       string                // bcrypt-hashed password
    Joined         time.Time             // Account creation date
    Macros         map[string]string     // User-defined command macros
    Aliases        map[string]string     // Command aliases and shortcuts
    Character      *characters.Character // Associated character data
    ItemStorage    Storage               // Personal item storage
    ConfigOptions  map[string]any        // User configuration preferences
    Inbox          Inbox                 // Message inbox with attachments
    Muted          bool                  // Communication restrictions
    Deafened       bool                  // Communication filtering
    ScreenReader   bool                  // Accessibility mode
    EmailAddress   string                // Contact email (optional)
    TipsComplete   map[string]bool       // Tutorial completion tracking

    // Runtime fields (not persisted)
    EventLog       UserLog               // Session event logging
    LastMusic      string                // Audio state tracking
    connectionId   uint64                // Current connection ID
    unsentText     string                // Buffered output
    suggestText    string                // Input suggestions
    connectionTime time.Time             // Connection timestamp
    lastInputRound uint64                // Last input round number
    tempDataStore  map[string]any        // Temporary session data
    activePrompt   *prompt.Prompt        // Current prompt state
    isLinkDead     bool                  // Link-dead connection flag
    inputBlocked   bool                  // Input processing control
}
```

### Active Users Management
```go
type ActiveUsers struct {
    Users               map[int]*UserRecord                 // userId -> UserRecord
    Usernames           map[string]int                      // username -> userId
    Connections         map[connections.ConnectionId]int    // connectionId -> userId
    UserConnections     map[int]connections.ConnectionId    // userId -> connectionId
    LinkDeadConnections map[connections.ConnectionId]uint64 // connectionId -> turn they became link-dead
}
```

## User Index System

### Index Structure
```go
type IndexMetaData struct {
    MetaDataSize uint64 // Header size in bytes (100)
    IndexVersion uint64 // Index format version (1)
    RecordCount  uint64 // Number of user records
    RecordSize   uint64 // Fixed record size (89 bytes)
}

type IndexUserRecord struct {
    UserID   int64     // 8 bytes - User identifier
    Username [80]byte  // 80 bytes - Fixed-width username
                       // 1 byte - Line terminator
}
```

## User Management Operations

### User Creation and Authentication
```go
// Create new user record
func NewUserRecord(userId int, connectionId uint64) *UserRecord

// Password validation: tries bcrypt first, then migrates legacy SHA256 hashes to bcrypt
func (u *UserRecord) PasswordMatches(input string) bool

// Set password (bcrypt-hashes the provided value)
func (u *UserRecord) SetPassword(pw string) error
```

### Connection Management
```go
// Get connection ID for user
func GetConnectionId(userId int) connections.ConnectionId

// Get multiple connection IDs
func GetConnectionIds(userIds []int) []connections.ConnectionId

// Get all active (non-link-dead) users
func GetAllActiveUsers() []*UserRecord

// Get all online user IDs (including link-dead)
func GetOnlineUserIds() []int
```

### Link-Dead Connection Handling
```go
// Mark user as link-dead
func SetLinkDeadUser(userId int)

// Clear link-dead status for a user
func RemoveLinkDeadUser(userId int)

// Remove a link-dead connection entry directly
func RemoveLinkDeadConnection(connectionId connections.ConnectionId)

// Check if connection is link-dead
func IsLinkDeadConnection(connectionId connections.ConnectionId) bool

// Get expired link-dead users for cleanup
func GetExpiredLinkDeadUsers(expirationTurn uint64) []int
```

### Login / Logout
```go
// Log in a user (handles link-dead reconnect)
func LoginUser(user *UserRecord, connectionId connections.ConnectionId) (*UserRecord, string, error)

// Log in a user reconnecting after a copyover
func CopyoverReconnectUser(user *UserRecord, connectionId connections.ConnectionId) (*UserRecord, string, error)

// Log out a user by connection ID (saves data and cleans up maps)
func LogOutUserByConnectionId(connectionId connections.ConnectionId) error

// Create a brand-new user account
func CreateUser(u *UserRecord) error
```

## Storage Systems

### Item Storage
```go
type Storage struct {
    Items []items.Item // Personal item storage
}

func (s *Storage) FindItem(itemName string) (items.Item, bool)
func (s *Storage) AddItem(i items.Item) bool
func (s *Storage) RemoveItem(i items.Item) bool
```

### Migration Helper

`MigrateInbox(userId int) []LegacyMessage` reads any `inbox:` data from a user's YAML file written by a previous server version, before the mudmail module took ownership of inbox data. Called once per user on their first `PlayerSpawn` after upgrading. `LegacyMessage` mirrors the old message struct for unmarshalling only.

### User Search (`search.go`)

```go
type UserSearchResult struct {
    UserId   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    Email    string `json:"email"`
}

// SearchUsers returns an exact match (sole result) or all prefix matches for
// searchName. The search is case-insensitive and reads the binary user index
// directly to avoid loading full user records for every candidate. Role and
// email are populated from the in-memory user map when the user is online, or
// from a skip-validation disk load otherwise.
func SearchUsers(searchName string) []UserSearchResult
```

## User Data Management

### Temporary Data Storage
```go
func (u *UserRecord) SetTempData(key string, value any)
func (u *UserRecord) GetTempData(key string) any
```

### Configuration Options
```go
func (u *UserRecord) SetConfigOption(key string, value any)
func (u *UserRecord) GetConfigOption(key string) any

// Client display settings (screen width/height, MSP, etc.)
func (u *UserRecord) ClientSettings() connections.ClientSettings
```

### Unsent Text (prompt redraw support)SendMudMail
```go
func (u *UserRecord) SetUnsentText(t string, suggest string)
func (u *UserRecord) GetUnsentText() (unsent string, suggestion string)
```

## Online User Information

### OnlineInfo Structure
```go
type OnlineInfo struct {
    Username      string // Login username
    CharacterName string // Character display name
    Level         int    // Character level
    Alignment     string // Character alignment
    Profession    string // Character profession
    OnlineTime    int64  // Seconds online
    OnlineTimeStr string // Formatted time string (e.g. "2h30m")
    IsAFK         bool   // Away from keyboard status
    Role          string // User role (guest/user/admin)
    ConnType      string // Connection type: "telnet", "websocket", or "ssh"
}
```

`ConnType` is determined at runtime from the underlying `ConnectionDetails` via `cd.IsWebSocket()` / `cd.IsSSH()`.

## Integration Patterns

### Character System Integration
```go
- user.Character                                          // Full character data
- user.Character.Name                                    // Character name
- user.Character.Level                                   // Character progression
- user.Character.RoomId                                  // Current location
- user.Character.HasVisitedRoom(roomId, zone)            // Permanent room visit check
- user.Character.MarkVisitedRoom(roomId, zone)           // Record a room visit
- user.Character.ZoneVisitProgress(zone, validRoomIds)   // Visited/total count for a zone
```

### Connection System Integration
```go
- user.connectionId                // Current connection
- connections.GetClientSettings()  // Display preferences
- connections.SendTo()             // Send messages to user
- connections.Get(id).IsSSH()      // Detect SSH connection type
```

### Prompt System Integration
```go
- user.activePrompt               // Current prompt state
- user.inputBlocked               // Input processing control
- user.StartPrompt(cmd, rest)     // Start or reuse a prompt
- user.GetPrompt()                // Retrieve current prompt
- user.ClearPrompt()              // Clear active prompt
```

### Event System Integration
```go
- events.AddToQueue()             // Queue user actions
- user.EventLog                   // Track user events
- user.lastInputRound             // Input timing
- user.Command(inputTxt)          // Queue a command for the user
- user.CommandFlagged(inputTxt, flags) // Queue a command with event flags
- user.SendText(txt)              // Queue a Message event to the user
- user.AddBuff(buffId, source)    // Queue a Buff event for the user
- user.GrantXP(amt, source)       // Grant XP and fire LevelUp events if needed
```

## Usage Examples

### User Authentication
```go
user, err := users.LoadUser(username)
if err != nil {
    return errors.New("user not found")
}
if !user.PasswordMatches(password) {
    return errors.New("invalid password")
}
// User authenticated successfully
```

### Zombie Connection Cleanup
```go
currentTurn := util.GetTurnCount()
expirationTurn := currentTurn - linkDeadTurns

expiredUsers := users.GetExpiredLinkDeadUsers(expirationTurn)
for _, userId := range expiredUsers {
    user := users.GetByUserId(userId)
    if user != nil {
        user.Save()
        users.LogOutUserByConnectionId(user.ConnectionId())
    }
}
```

### Alt Character Swap
```go
if user.SwapToAlt("AltName") {
    user.SendText("Swapped to alt character.")
} else {
    user.SendText("Alt character not found.")
}
```

## Dependencies

- `internal/characters` - Character system integration for user avatars
- `internal/connections` - Network connection management and client settings
- `internal/items` - Item system for storage and inventory management
- `internal/configs` - Configuration management for user settings
- `internal/prompt` - Interactive prompt system for user input
- `internal/util` - Utility functions for hashing, file operations, and validation
- `internal/mudlog` - Logging system for user events and debugging
- `internal/audio` - Audio file lookup for MSP sound/music events
- `internal/skills` - Profession lookup for online info display
- `golang.org/x/crypto/bcrypt` - Secure password hashing
- `gopkg.in/yaml.v2` - YAML serialization for user data persistence
