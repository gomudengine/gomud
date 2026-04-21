# GoMud Buffs System Context

## Overview

The GoMud buffs system provides comprehensive temporary status effects for characters with support for stat modifications, behavioral flags, round-based triggers, duration management, and scripting integration. It features a dual-layer architecture with immutable buff specifications and mutable buff instances, supporting complex timing mechanics, permanent buffs, and sophisticated flag-based behavior modification.

## Architecture

The buffs system is built around several key components:

### Core Components

**Buff Specifications (BuffSpec):**
- Immutable blueprint definitions for all buff types
- YAML-based storage with automatic loading and validation
- Time-based trigger rate calculations with game time integration
- Stat modification definitions and behavioral flags
- JavaScript scripting support for custom buff behaviors

**Buff Instances (Buff):**
- Runtime instances with unique state tracking
- Round-based trigger counters and expiration management
- Source tracking for buff origin identification
- Permanent buff support for equipment and racial effects
- Start event queuing for delayed activation

**Buffs Collection (Buffs):**
- Efficient collection management with flag indexing
- Fast lookup maps for buff IDs and flags
- Automatic validation and rebuilding of internal indexes
- Batch operations for triggering and pruning

## Key Features

### 1. **Comprehensive Flag System**
- Behavioral modification flags (combat, movement, fleeing restrictions)
- Death prevention and revival mechanics
- Equipment interaction flags (permanent gear, curse removal)
- Status effect flags (poison, accuracy, stealth, vision enhancement)
- Environmental interaction flags (water cancellation, light emission)

### 2. **Advanced Timing System**
- Round-based trigger intervals with game time integration
- Flexible trigger rates using time string parsing
- Unlimited duration support for permanent effects
- Precise expiration tracking and automatic cleanup
- Trigger counting with configurable limits

### 3. **Stat Modification Integration**
- Dynamic stat bonuses and penalties
- Cumulative effects from multiple buffs
- Integration with character stat system
- Racial and equipment stat modifications
- Combat effectiveness modifiers

### 4. **Scripting and Customization**
- JavaScript event handling for complex buff behaviors
- Custom script path resolution and loading
- Event-driven interaction with game systems
- Flexible buff value calculations for balance

## Buff Structure

### Buff Specification Structure
```go
type BuffSpec struct {
    BuffId        int               // Unique identifier
    Name          string            // Display name
    Description   string            // Description text
    Secret        bool              // Hidden from player view
    TriggerNow    bool              // Immediate trigger on application
    TriggerRate   string            // Time-based trigger frequency
    RoundInterval int               // Calculated round interval
    TriggerCount  int               // Total trigger limit
    StatMods      statmods.StatMods // Stat modifications
    Flags         []Flag            // Behavioral flags
}
```

### Buff Instance Structure
```go
type Buff struct {
    BuffId         int    // Reference to BuffSpec
    Source         string // Origin identifier (spell, item, area)
    OnStartWaiting bool   // Pending start event
    PermaBuff      bool   // Permanent buff flag
    RoundCounter   int    // Elapsed rounds
    TriggersLeft   int    // Remaining triggers
}
```

### Buffs Collection Structure
```go
type Buffs struct {
    List      []*Buff           // Active buff instances
    buffFlags map[Flag][]int    // Flag to buff index mapping
    buffIds   map[int]int       // BuffId to index mapping
}
```

## Flag System

### Behavioral Flags
```go
const (
    // Combat and Movement Restrictions
    NoCombat       Flag = "no-combat"        // Prevents combat initiation
    NoMovement     Flag = "no-go"            // Prevents movement
    NoFlee         Flag = "no-flee"          // Prevents fleeing combat
    
    // Cancellation Conditions
    CancelIfCombat Flag = "cancel-on-combat" // Removes buff when combat starts
    CancelOnAction Flag = "cancel-on-action" // Removes buff on any action
    CancelOnWater  Flag = "cancel-on-water"  // Removes buff in water
    
    // Death and Revival
    ReviveOnDeath  Flag = "revive-on-death"  // Prevents death once
    
    // Equipment Interaction
    PermaGear      Flag = "perma-gear"       // Equipment cannot be removed
    RemoveCurse    Flag = "remove-curse"     // Allows cursed item removal
    
    // Status Effects
    Poison         Flag = "poison"           // Harmful poison effect
    Drunk          Flag = "drunk"            // Intoxication effects
    Hidden         Flag = "hidden"           // Stealth/invisibility
    Accuracy       Flag = "accuracy"         // Enhanced hit chance
    Blink          Flag = "blink"            // Dodge enhancement
    
    // Sensory Enhancement
    EmitsLight     Flag = "lightsource"      // Provides illumination
    SuperHearing   Flag = "superhearing"     // Enhanced hearing
    NightVision    Flag = "nightvision"      // See in darkness
    SeeHidden      Flag = "see-hidden"       // Detect hidden entities
    SeeNouns       Flag = "see-nouns"        // Enhanced object identification
    
    // Environmental Status
    Warmed         Flag = "warmed"           // Temperature regulation
    Hydrated       Flag = "hydrated"         // Hydration status
    Thirsty        Flag = "thirsty"          // Dehydration status
)
```

### Flag Usage Patterns
```go
// Check for specific behavioral flags
func (bs *Buffs) HasFlag(action Flag, expire bool) bool

// Get all buff IDs with specific flag
func (bs *Buffs) GetBuffIdsWithFlag(action Flag) []int
```

## Timing and Trigger System

### Round-Based Triggers
```go
// Trigger buffs based on round intervals
func (bs *Buffs) Trigger(buffId ...int) (triggeredBuffs []*Buff)
```

### Duration Calculations
```go
// Calculate remaining and total duration
func GetDurations(buff *Buff, spec *BuffSpec) (roundsLeft int, totalRounds int)

// Check if buff has expired
func (b *Buff) Expired() bool
```

### Time String Processing
```go
// Validate and convert time strings to round intervals
func (b *BuffSpec) Validate() error
```

## Stat Modification System

### Individual Buff Stat Modifications
```go
// Get stat modification from single buff
func (b *Buff) StatMod(statName string) int
```

### Cumulative Stat Modifications
```go
// Calculate total stat modification from all active buffs
func (bs *Buffs) StatMod(statName string) int
```

### Buff Value Calculation
```go
// Calculate relative power/value of a buff for balance
func (b *BuffSpec) GetValue() int
```

## Buff Management Operations

### Adding Buffs
```go
// Add new buff or refresh existing buff
func (bs *Buffs) AddBuff(buffId int, isPermanent bool) bool
```

### Removing Buffs
```go
// Remove specific buff by ID
func (bs *Buffs) RemoveBuff(buffId int) bool

// Mark buff as started (no longer waiting for start event)
func (bs *Buffs) Started(buffId int)
```

### Pruning Expired Buffs
```go
// Remove all expired buffs and rebuild indexes
func (bs *Buffs) Prune() (prunedBuffs []*Buff)
```

## Collection Validation and Indexing

### Index Management
```go
// Validate and rebuild internal indexes
func (bs *Buffs) Validate(forceRebuild ...bool)
```

### Query Operations
```go
// Check if specific buff exists
func (bs *Buffs) HasBuff(buffId int) bool

// Get remaining triggers for buff
func (bs *Buffs) TriggersLeft(buffId int) int

// Get all active buffs (optionally filtered by ID)
func (bs *Buffs) GetBuffs(buffId ...int) []*Buff
```

## Scripting Integration

### Script System Support
```go
// Get buff script content
func (b *BuffSpec) GetScript() string

// Generate script file path
func (b *BuffSpec) GetScriptPath() string
```

### Display and Visibility
```go
// Get visible name and description (handles secret buffs)
func (b *BuffSpec) VisibleNameDesc() (name, description string)

// Get buff display name
func (bs *Buff) Name() string
```

## Data Management and Search

### Buff Discovery
```go
// Search buffs by name or description
func SearchBuffs(searchTerm string) []int

// Get all available buff IDs
func GetAllBuffIds() []int
```

### File Management
```go
// Generate filename for buff specification
func (b *BuffSpec) Filename() string

// Load all buff specifications from files
func LoadDataFiles()
```

## Integration Patterns

### Character System Integration
```go
// Buffs integrate with character stats and behavior
- character.Buffs.StatMod("strength")     // Stat modifications
- character.Buffs.HasFlag(buffs.NoCombat) // Behavioral restrictions
- character.Buffs.Trigger()               // Round-based processing
- character.Buffs.Prune()                 // Cleanup expired buffs
```

### Combat System Integration
```go
// Combat checks buff flags for behavior modification
if sourceChar.HasBuffFlag(buffs.Accuracy) {
    critChance *= 2 // Double crit chance
}

if targetChar.HasBuffFlag(buffs.Blink) {
    critChance /= 2 // Half crit chance against blink
}

if !sourceChar.HasBuffFlag(buffs.Hidden) {
    // Send visible combat messages
}
```

### Event System Integration
```go
// Buffs trigger events for start, effect, and end
events.AddToQueue(events.Buff{
    MobInstanceId: mobInstanceId,
    BuffId:        buffId,
    Source:        source,
})
```

## Usage Examples

### Basic Buff Management
```go
// Create new buff collection
buffs := buffs.New()

// Add temporary buff
buffs.AddBuff(poisonBuffId, false)

// Add permanent buff (from equipment)
buffs.AddBuff(strengthBuffId, true)

// Check for specific behavior
if buffs.HasFlag(buffs.NoCombat, false) {
    user.SendText("You cannot engage in combat right now.")
    return
}

// Process round-based triggers
triggeredBuffs := buffs.Trigger()
for _, buff := range triggeredBuffs {
    // Handle buff effects
    processBuff(buff)
}

// Clean up expired buffs
prunedBuffs := buffs.Prune()
for _, buff := range prunedBuffs {
    // Send buff expiration messages
    notifyBuffExpired(buff)
}
```

### Stat Modification Usage
```go
// Calculate total stat bonuses from all buffs
strengthBonus := character.Buffs.StatMod("strength")
speedBonus := character.Buffs.StatMod("speed")
healthBonus := character.Buffs.StatMod("health")

// Apply to character stats
character.Stats.Strength.ValueAdj += strengthBonus
character.Stats.Speed.ValueAdj += speedBonus
character.HealthMax.Value += healthBonus
```

### Flag-Based Behavior Control
```go
// Check movement restrictions
if character.Buffs.HasFlag(buffs.NoMovement, false) {
    user.SendText("You are unable to move.")
    return
}

// Check combat restrictions with expiration
if character.Buffs.HasFlag(buffs.CancelOnAction, true) {
    user.SendText("Your concentration is broken!")
    // Buff automatically expired by HasFlag call
}

// Environmental interactions
if character.Buffs.HasFlag(buffs.EmitsLight, false) {
    room.LightLevel += 1 // Provide illumination
}
```

## Dependencies

- `internal/statmods` - Stat modification system integration
- `internal/configs` - Configuration management for file paths and timing
- `internal/gametime` - Game time system for trigger rate calculations
- `internal/fileloader` - YAML file loading and validation system
- `internal/util` - Utility functions for file operations and validation
- `internal/mudlog` - Logging system for debugging and monitoring

This comprehensive buffs system provides sophisticated temporary status effects with precise timing control, behavioral modification, stat integration, and seamless integration with all other game systems.