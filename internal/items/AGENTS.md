# GoMud Game Items System Context

## Overview

The GoMud items system provides a comprehensive item management framework with support for equipment, consumables, weapons, and special objects. It features a dual-layer architecture with immutable item specifications and mutable item instances, supporting enchantments, durability, scripting integration, and complex item behaviors through type-based categorization and attribute systems.

## Architecture

The items system is built around two main components:

### Core Components

**Item Specifications (`ItemSpec`):**
- Immutable blueprint definitions for all item types
- YAML-based storage with automatic loading and validation
- Hierarchical organization by item type and subtype
- Automatic value calculation based on item properties

**Item Instances (`Item`):**
- Mutable runtime instances based on specifications
- UUID-based unique identification for each instance
- Support for enchantments, durability, and temporary modifications
- Blob storage for custom data and scripting integration

**Type System:**
- Primary types (weapon, armor, consumables, etc.) with ID ranges
- Subtypes for specialized behaviors (wearable, usable, throwable, etc.)
- Element types for magical damage and effects
- Weapon classification for combat message selection

**Attack Message System:**
- Dynamic combat message generation based on weapon subtypes
- Intensity-based message selection (miss, weak, normal, heavy, critical)
- Token replacement system for personalized combat text
- Separate messages for attacker, defender, and room observers

## Key Features

### 1. **Hierarchical Item Classification**
- Type-based organization with reserved ID ranges for different categories
- Subtype system for specialized behaviors and interactions
- Element system for magical properties and damage types
- Automatic categorization and validation

### 2. **Instance Management**
- UUID-based unique identification for every item instance
- Temporary data storage for runtime modifications
- Enchantment system with stat bonuses and curse mechanics
- Durability and usage tracking with break chance mechanics

### 3. **Dynamic Item Modification**
- Runtime enchantment system with stat modifications
- Temporary adjective system for visual effects
- Blob storage for custom content and scripting data
- Override specifications for personalized item properties

### 4. **Combat Integration**
- Weapon damage calculation with dice roll systems
- Attack message generation based on weapon type and damage intensity
- Critical hit mechanics with buff application
- Backstab compatibility based on weapon subtype

### 5. **Scripting Support**
- JavaScript integration for custom item behaviors
- Event-driven item interactions (onFound, onLost, onUse, onPurchase)
- Template data storage for dynamic content
- Script path resolution and loading

## Item Types and Categories

### Equipment Types (ID Ranges)
```go
// Weapons: 10000-19999
Weapon ItemType = "weapon"

// Armor: 20000-29999
Head    ItemType = "head"
Neck    ItemType = "neck"
Body    ItemType = "body"
Belt    ItemType = "belt"
Gloves  ItemType = "gloves"
Ring    ItemType = "ring"
Legs    ItemType = "legs"
Feet    ItemType = "feet"
Offhand ItemType = "offhand"

// Consumables: 30000-39999
Potion     ItemType = "potion"
Food       ItemType = "food"
Drink      ItemType = "drink"
Botanical  ItemType = "botanical"

// Other: 0-9999
Scroll     ItemType = "scroll"
Readable   ItemType = "readable"
Key        ItemType = "key"
Object     ItemType = "object"
Gemstone   ItemType = "gemstone"
Lockpicks  ItemType = "lockpicks"
Grenade    ItemType = "grenade"
Junk       ItemType = "junk"
```

### Item Subtypes
```go
// Behavior Subtypes
Wearable  ItemSubType = "wearable"
Drinkable ItemSubType = "drinkable"
Edible    ItemSubType = "edible"
Usable    ItemSubType = "usable"
Throwable ItemSubType = "throwable"
Mundane   ItemSubType = "mundane"

// Weapon Subtypes (for combat messages)
Generic     ItemSubType = "generic"
Bludgeoning ItemSubType = "bludgeoning"
Cleaving    ItemSubType = "cleaving"
Stabbing    ItemSubType = "stabbing"
Slashing    ItemSubType = "slashing"
Shooting    ItemSubType = "shooting"
Claws       ItemSubType = "claws"
Whipping    ItemSubType = "whipping"
```

## Item Specification Structure

### Basic Item Properties
```go
type ItemSpec struct {
    ItemId          int
    Name            string
    DisplayName     string        // Formatted display name with colors
    NameSimple      string        // Simple name for matching
    Description     string
    Value           int           // Gold value (auto-calculated if 0)
    Type            ItemType
    Subtype         ItemSubType
    
    // Usage Properties
    Uses            int           // Number of uses before consumption
    BuffIds         []int         // Buffs applied when used
    WornBuffIds     []int         // Buffs applied while worn
    QuestToken      string        // Quest progress granted when obtained
    
    // Combat Properties
    Damage          Damage        // Weapon damage specification
    DamageReduction int           // Armor damage reduction percentage
    WaitRounds      int           // Extra combat rounds required
    Hands           WeaponHands   // 1 or 2 handed weapon
    Element         Element       // Magical element type
    
    // Enhancement Properties
    StatMods        statmods.StatMods  // Stat modifications when worn
    BreakChance     uint8              // Chance to break on use (0-100)
    Cursed          bool               // Cannot be removed when equipped
    KeyLockId       string             // Lock ID this key opens
}
```

### Damage System
```go
type Damage struct {
    Attacks     int      // Number of attacks per round
    DiceCount   int      // Number of dice to roll
    SideCount   int      // Sides per die
    BonusDamage int      // Flat damage bonus
    DiceRoll    string   // Formatted dice roll (e.g., "2d6+3")
    CritBuffIds []int    // Buffs applied on critical hits
}
```

## Item Instance Management

### Item Creation and Validation
```go
// Create new item instance
func New(itemId int) Item

// Item validation ensures consistency
func (i *Item) Validate()
```

### Item Identification and Matching
```go
// Multiple identification methods
func (i *Item) ShorthandId() string

// Name matching with partial and full match support
func (i *Item) NameMatch(input string, allowContains bool) (partialMatch bool, fullMatch bool)
```

## Enchantment and Modification System

### Dynamic Item Enhancement
```go
// Enchant item with bonuses
func (i *Item) Enchant(damageBonus int, defenseBonus int, statBonus map[string]int, cursed bool)

// Curse management
func (i *Item) IsCursed() bool
func (i *Item) Uncurse()
```

### Adjective System
```go
// Visual effects through adjectives
func (i *Item) SetAdjective(adj string, addToList bool)

// Display name with adjectives
func (i *Item) DisplayName() string
```

## Combat Message System

### Attack Message Structure
```go
type WeaponAttackMessageGroup struct {
    OptionId ItemSubType
    Options  AttackTypes
}

type AttackTypes map[Intensity]AttackOptions

type AttackOptions struct {
    Together TogetherMessages  // Same room messages
    Separate SeparateMessages  // Different room messages
}

type TogetherMessages struct {
    ToAttacker MessageOptions  // Messages to attacker
    ToDefender MessageOptions  // Messages to defender
    ToRoom     MessageOptions  // Messages to room observers
}
```

### Message Selection and Token Replacement
```go
// Get attack message based on damage percentage
func GetAttackMessage(subType ItemSubType, pctDamage int) AttackOptions

// Token replacement in messages
func (am ItemMessage) SetTokenValue(tokenName TokenName, tokenValue string) ItemMessage
```

## Durability and Usage System

### Break Mechanics
```go
// Break chance testing
func (i *Item) BreakTest(increaseChance ...int) bool

// Usage tracking
func (i *Item) UseItem() bool
```

## Data Storage and Persistence

### Blob Content System
```go
// Store custom data in items
func (i *Item) SetBlob(blob string)
func (i *Item) GetBlob() string

// Temporary data storage
func (i *Item) SetTempData(key string, value any)
```

### File Organization
```go
// Automatic file organization by item ID ranges
func (i *ItemSpec) ItemFolder(baseonly ...bool) string
```

## Integration Patterns

### Scripting Integration
```go
// JavaScript event integration
func (i *Item) GetScript() string
func (i *ItemSpec) GetScriptPath() string

// Script events: onFound, onLost, onUse, onPurchase
// Called from various game systems when items are manipulated
```

### Character Equipment Integration
```go
// Stat modification when equipped
func (i *Item) StatMod(statName ...string) int

// Equipment comparison
func (i *Item) IsBetterThan(otherItm Item) bool
```

### Quest System Integration
```go
// Automatic quest progress when item obtained
type ItemSpec struct {
    QuestToken string  // Quest progress granted when obtained
}

// Quest integration happens through event system
// when ItemOwnership events are fired
```

## Admin / Persistence Helpers (`admin.go`)

```go
// GetAllAttackMessages returns a copy of the loaded attack message groups keyed
// by weapon subtype.
func GetAllAttackMessages() map[ItemSubType]*WeaponAttackMessageGroup

// SaveItemSpec validates, saves to disk, and updates the in-memory cache.
// ItemId must already be set.
func SaveItemSpec(spec *ItemSpec) error

// DeleteItemSpec removes the YAML file and any associated JS script from disk
// and purges the entry from the in-memory cache.
func DeleteItemSpec(itemId int) error

// SaveItemScript writes content to the item's script path.  An empty content
// string deletes the script file.
func SaveItemScript(itemId int, content string) error
```

## Search and Discovery

### Item Finding Functions
```go
// Multiple search methods
func FindItem(nameOrId string) int
func FindItemByName(name string) int
```

### Advanced Item Matching
```go
// Find items in collections with numbering support
func FindMatchIn(itemName string, items ...Item) (pMatch Item, fMatch Item)
```

## Performance Considerations

### Memory Management
- Item specifications loaded once at startup and cached
- Item instances use minimal memory with spec references
- Temporary data cleared automatically on item destruction
- UUID generation optimized for performance

### File Loading Optimization
```go
// Batch loading of all item specifications
func LoadDataFiles()
```

## Dependencies

- `internal/buffs` - Status effect integration for item usage and worn effects
- `internal/configs` - Configuration management for file paths and settings
- `internal/statmods` - Stat modification system for equipment bonuses
- `internal/uuid` - Unique identification system for item instances
- `internal/util` - Utility functions for dice rolls, compression, and string processing
- `internal/colorpatterns` - Color pattern application for display names
- `internal/fileloader` - YAML file loading and validation system

## Usage Examples

### Creating and Modifying Items
```go
// Create new item instance
sword := items.New(12345)

// Enchant the sword
sword.Enchant(5, 0, map[string]int{"strength": 2}, false)

// Add visual effect
sword.SetAdjective("glowing", true)

// Check if cursed
if sword.IsCursed() {
    sword.Uncurse()
}

// Test for breakage
if sword.BreakTest(10) {
    // Item broke with 10% increased chance
}
```

### Item Searching and Matching
```go
// Find item by name or ID
itemId := items.FindItem("steel sword")

// Find in player inventory
inventory := []items.Item{sword, shield, potion}
partial, exact := items.FindMatchIn("sw", inventory...)

if exact.ItemId > 0 {
    // Found exact match
} else if partial.ItemId > 0 {
    // Found partial match
}
```

### Combat Integration
```go
// Get weapon damage
attacks, dCount, dSides, bonus, critBuffs := weapon.GetDiceRoll()

// Get attack messages
messages := items.GetAttackMessage(items.Slashing, 85) // 85% damage = Heavy

// Apply token replacements
message := messages.Together.ToAttacker.Get()
message = message.SetTokenValue(items.TokenDamage, "15")
message = message.SetTokenValue(items.TokenTarget, "orc")
```

This comprehensive item system provides the foundation for all equipment, consumables, and objects in GoMud, supporting complex interactions, modifications, and integration with all other game systems.