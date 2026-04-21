# GoMud NPC Management System Context

## Overview

The GoMud mobs system provides comprehensive NPC (Non-Player Character) management with support for AI behaviors, scripting integration, conversation systems, pathfinding, shop management, and complex social dynamics. It features a dual-layer architecture with immutable mob specifications and mutable mob instances, supporting dynamic spawning, behavioral patterns, and sophisticated interaction systems.

## Architecture

The mobs system is built around several key components:

### Core Components

**Mob Specifications:**
- Immutable blueprint definitions for all NPC types
- YAML-based storage with automatic loading and validation
- Zone-based organization with hierarchical file structure
- Character integration for stats, equipment, and abilities

**Mob Instances:**
- Runtime instances with unique IDs and state management
- Dynamic spawning and despawning with memory management
- Behavioral state tracking and command scheduling
- Temporary data storage for scripting and AI systems

**AI Behavior System:**
- Activity-based idle command execution
- Combat command selection and execution
- Conversation system integration with multi-mob interactions
- Pathfinding and movement planning

**Social Dynamics:**
- Group-based allegiances and hostilities
- Race-based hatred and alliance systems
- Player relationship tracking and memory
- Alignment-based conflict resolution

## Key Features

### 1. **Dynamic Instance Management**
- Unique instance IDs for each spawned mob
- Automatic stat calculation and equipment validation
- Level scaling and stat point distribution
- Memory management with automatic cleanup

### 2. **Behavioral AI System**
- Activity level-based command frequency
- Idle, angry, and combat command sets
- Boredom tracking and player interaction memory
- Conversation participation with other NPCs

### 3. **Social and Combat Dynamics**
- Group-based allegiance system
- Race and alignment-based hostility
- Player attack tracking and memory
- Shop ownership and trading behavior

### 4. **Scripting Integration**
- JavaScript event handling for custom behaviors
- Script tag system for specialized mob variants
- Event-driven interaction with game systems
- Custom script path resolution

### 5. **Pathfinding and Movement**
- Pre-calculated path following system
- Waypoint-based navigation
- Wandering behavior with distance limits
- Room-based movement constraints

## Mob Structure

### Core Mob Properties
```go
type Mob struct {
    MobId           MobId                    // Unique mob type identifier
    Zone            string                   // Zone this mob belongs to
    InstanceId      int                      // Unique runtime instance ID
    HomeRoomId      int                      // Starting/home room
    Character       characters.Character     // Character stats and properties
    
    // Behavior Properties
    ActivityLevel   int                      // 1-100% activity frequency
    Hostile         bool                     // Attack players on sight
    MaxWander       int                      // Maximum rooms from home
    WanderCount     int                      // Current wander distance
    PreventIdle     bool                     // Disable idle behavior
    
    // AI Command Sets
    IdleCommands    []string                 // Commands executed when idle
    AngryCommands   []string                 // Commands when entering combat
    CombatCommands  []string                 // Commands during combat
    
    // Social Properties
    Groups          []string                 // Group allegiances
    Hates           []string                 // Groups/races this mob hates
    QuestFlags      []string                 // Quest flags for interactions
    
    // Economy
    ItemDropChance  int                      // Chance to drop items on death
    BuffIds         []int                    // Permanent buffs on spawn
    
    // Scripting
    ScriptTag       string                   // Custom script identifier
    
    // Runtime State
    LastIdleCommand uint8                    // Track last idle command used
    BoredomCounter  uint8                    // Rounds since seeing players
    tempDataStore   map[string]any           // Temporary data storage
    conversationId  int                      // Active conversation ID
    Path            PathQueue                // Movement pathfinding queue
    lastCommandTurn uint64                   // Command scheduling tracking
    playersAttacked map[int]struct{}         // Players this mob has attacked
}
```

### Mob Creation and Spawning
```go
// Create new mob instance from specification
func NewMobById(mobId MobId, homeRoomId int, forceLevel ...int) *Mob
```

## AI Behavior System

### Command Execution and Scheduling
```go
// Schedule commands with timing
func (m *Mob) Command(inputTxt string, waitSeconds ...float64)

// Sleep functionality
func (m *Mob) Sleep(seconds int)
```

### Behavioral Command Selection
```go
// Get idle command based on mob or race defaults
func (m *Mob) GetIdleCommand() string

// Get angry command when entering combat
func (m *Mob) GetAngryCommand() string
```

## Social Dynamics System

### Group Allegiances and Hostilities
```go
// Check if two mobs are allies
func (r *Mob) ConsidersAnAlly(m *Mob) bool

// Check race-based hatred
func (r *Mob) HatesRace(raceName string) bool

// Check alignment-based hostility
func (r *Mob) HatesAlignment(otherAlignment int8) bool
```

### Player Relationship Tracking
```go
// Track player attacks for memory system
func (m *Mob) PlayerAttacked(userId int)
func (m *Mob) HasAttackedPlayer(userId int) bool

// Global hostility tracking
func MakeHostile(groupName string, userId int, rounds int)
func IsHostile(groupName string, userId int) bool
```

## Conversation System Integration

### Multi-Mob Conversations
```go
// Check if mob is in conversation
func (m *Mob) InConversation() bool

// Set conversation participation
func (m *Mob) SetConversation(id int)

// Execute conversation actions
func (m *Mob) Converse()
```

## Pathfinding and Movement

### Path Queue System
```go
type PathQueue struct {
    roomQueue   []PathRoom
    currentRoom PathRoom
}

// Path management
func (p *PathQueue) SetPath(path []PathRoom)
func (p *PathQueue) Next() PathRoom

// Get remaining waypoints
func (p *PathQueue) Waypoints() []int
```

## Shop and Trading System

### NPC Merchant Behavior
```go
// Check if mob has shop
func (m *Mob) HasShop() bool

// Calculate sell price for items
func (m *Mob) GetSellPrice(item items.Item) int
```

## Scripting Integration

### Script System Support
```go
// Check for custom scripts
func (m *Mob) HasScript() bool

// Load mob script content
func (m *Mob) GetScript() string

// Generate script file path
func (m *Mob) GetScriptPath() string
```

### Temporary Data Storage
```go
// Runtime data storage for scripts and AI
func (m *Mob) SetTempData(key string, value any)
func (m *Mob) GetTempData(key string) any
```

## Special Mob Types and Behaviors

### Tameable Mobs
```go
// Check if mob can be tamed by players
func (m *Mob) IsTameable() bool
```

### Persistent vs Temporary Mobs
```go
// Check if mob should despawn when room unloads
func (m *Mob) Despawns() bool
```

## Memory and Performance Management

### Instance Tracking
```go
// Memory usage reporting
func GetMemoryUsage() map[string]util.MemoryResult

// Recent death tracking
func TrackRecentDeath(instanceId int)
func RecentlyDied(instanceId int) bool
```

### Hostility Management
```go
// Reduce hostility over time
func ReduceHostility()
```

## File Organization and Persistence

### Zone-Based File Structure
```go
// Automatic file organization by zone
func (m *Mob) Filepath() string
func (m *Mob) Filename() string

// Zone name sanitization
func ZoneNameSanitize(zone string) string
```

### Data Loading and Validation
```go
// Load all mob specifications from files
func LoadDataFiles()
```

## Integration Patterns

### Event System Integration
```go
// Buff application through events
func (m *Mob) AddBuff(buffId int, source string)
```

### Character System Integration
```go
// Mobs use the same character system as players
type Mob struct {
    Character characters.Character  // Full character integration
}

// Automatic stat training and equipment validation
// Level scaling and experience calculation
// Equipment bonuses and stat modifications
```

## Usage Examples

### Creating and Managing Mob Instances
```go
// Spawn mob in specific room
mob := mobs.NewMobById(mobs.MobId(123), roomId)
if mob != nil {
    // Mob spawned successfully
    room.AddMob(mob.InstanceId)
}

// Force specific level
highLevelMob := mobs.NewMobById(mobs.MobId(123), roomId, 25)

// Schedule mob commands
mob.Command("say Hello there!")
mob.Command("emote waves", 2.0) // Wait 2 seconds
mob.Command("look; smile", 1.0) // Multiple commands
```

### AI Behavior Implementation
```go
// Idle behavior processing
if mob.ActivityLevel > util.Rand(100) {
    idleCmd := mob.GetIdleCommand()
    if idleCmd != "" {
        mob.Command(idleCmd)
    }
}

// Combat initiation
if mob.Hostile && playerInRoom {
    angryCmd := mob.GetAngryCommand()
    if angryCmd != "" {
        mob.Command(angryCmd)
    }
    // Start combat...
}
```

### Social Dynamics
```go
// Check relationships before combat
func shouldAttack(attacker *Mob, target *Mob) bool
```

## Dependencies

- `internal/characters` - Character system integration for stats and equipment
- `internal/events` - Event system for command scheduling and buff application
- `internal/conversations` - Multi-mob conversation system
- `internal/items` - Item system for equipment and inventory management
- `internal/races` - Race system for default behaviors and restrictions
- `internal/buffs` - Status effect system for permanent and temporary effects
- `internal/configs` - Configuration management for file paths and timing
- `internal/util` - Utility functions for randomization, file operations, and validation
- `internal/fileloader` - YAML file loading and validation system

## Mob Creation and File Management

### New Mob Creation System
```go
// Create new mob with optional script template
func CreateNewMobFile(newMobInfo Mob, copyScript string) (MobId, error)

// Automatic ID assignment
func getNextMobId() MobId
```

### Script Templates
```go
// Available script templates for new mobs
var SampleScripts = map[string]string{
    "item and gold": "item-gold-quest.js",
}

const ScriptTemplateQuest = "item-gold-quest.js"

// Quest template automatically sets quest flags
if copyScript == ScriptTemplateQuest {
    newMobInfo.QuestFlags = []string{"1000000-start"}
}
```

### File System Integration
- **Automatic ID Assignment**: Sequential ID allocation to prevent conflicts
- **Template System**: Pre-built script templates for common mob behaviors
- **Careful Save Mode**: Optional backup creation during file operations
- **Directory Management**: Automatic creation of script directories
- **Cache Synchronization**: Immediate update of in-memory caches after creation

This comprehensive mob system provides sophisticated NPC management with AI behaviors, social dynamics, scripting integration, file management capabilities, and seamless integration with all other game systems.