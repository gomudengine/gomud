# GoMud Stats System Context

## Overview

The GoMud stats system provides a comprehensive character attribute framework with six core statistics that govern all character capabilities. It features level-based progression, racial bonuses, training point allocation, equipment modifications, and a sophisticated diminishing returns system for balanced character development.

## Architecture

The stats system is built around several key components:

### Core Components

**Six Primary Statistics:**
- **Strength**: Physical power affecting damage output and carrying capacity
- **Speed**: Agility and reflexes affecting combat speed and dodging
- **Smarts**: Intelligence and wisdom affecting magic power and problem-solving
- **Vitality**: Health and stamina affecting hit points and endurance
- **Mysticism**: Magical affinity affecting mana capacity and spell effectiveness
- **Perception**: Awareness affecting detection and observation abilities

**Multi-Layer Calculation System:**
- **Base Values**: Racial starting statistics
- **Racial Gains**: Level-based automatic progression
- **Training Points**: Player-allocated stat improvements
- **Equipment Modifiers**: Temporary bonuses from gear and effects
- **Diminishing Returns**: Balanced scaling for high-end statistics

## Key Features

### 1. **Balanced Progression System**
- **Level-Based Growth**: Automatic stat gains based on character level and racial base
- **Training Allocation**: Player choice in stat development through training points
- **Diminishing Returns**: Square root scaling prevents excessive stat stacking
- **Equipment Integration**: Temporary modifications from gear and effects

### 2. **Racial Differentiation**
- **Base Stat Variation**: Different races excel in different areas
- **Scaling Benefits**: Racial advantages compound with level progression
- **Natural Progression**: Automatic gains reflect racial strengths
- **Balanced Design**: No race is universally superior

### 3. **Flexible Modification System**
- **Equipment Bonuses**: Gear provides temporary stat improvements
- **Spell Effects**: Magic can enhance or reduce statistics
- **Buff Integration**: Status effects modify character capabilities
- **Dynamic Recalculation**: Stats update automatically when modifiers change

## Stat Structure

### Statistics Collection
```go
type Statistics struct {
    Strength   StatInfo // Physical power and damage
    Speed      StatInfo // Agility, reflexes, and combat speed
    Smarts     StatInfo // Intelligence, wisdom, and magical understanding
    Vitality   StatInfo // Health, stamina, and endurance
    Mysticism  StatInfo // Magical capacity and spell effectiveness
    Perception StatInfo // Awareness, detection, and observation
}
```

### Individual Stat Information
```go
type StatInfo struct {
    Training int // Player-allocated training points (persistent)
    Value    int // Final calculated value (runtime)
    ValueAdj int // Adjusted value with diminishing returns (runtime)
    Racial   int // Level-based racial bonus (runtime)
    Base     int // Racial base value (persistent)
    Mods     int // Equipment/effect modifiers (runtime)
}
```

## Stat Calculation System

### Core Calculation Constants
```go
const (
    BaseModFactor         = 0.3333333334 // Racial scaling per level (1/3)
    NaturalGainsModFactor = 0.5          // Free stat points per level (1/2)
)
```

### Level-Based Progression
```go
// Calculate automatic stat gains for a given level
func (si *StatInfo) GainsForLevel(level int) int
```

### Complete Stat Recalculation
```go
// Recalculate all stat components for current level
func (si *StatInfo) Recalculate(level int)
```

### Modifier Management
```go
// Set equipment and effect modifiers
func (si *StatInfo) SetMod(mod ...int)
```

## Progression Mathematics

### Racial Scaling Formula
```
Racial Gains = (Level - 1) × (1/3) × Racial_Base + Level × (1/2)

Components:
- (Level - 1) × (1/3) × Racial_Base: Racial advantage scaling
- Level × (1/2): Universal natural progression

Example for Level 10 character with Racial_Base = 15:
- Racial scaling: (10-1) × 0.333 × 15 = 45 points
- Natural gains: 10 × 0.5 = 5 points
- Total racial gains: 50 points
```

### Diminishing Returns System
```
For stats ≥ 105:
Adjusted_Value = 100 + sqrt(Value - 100) × 2

Examples:
- Value 105 → Adjusted 104 (√5 × 2 ≈ 4)
- Value 125 → Adjusted 110 (√25 × 2 = 10)
- Value 200 → Adjusted 120 (√100 × 2 = 20)

This prevents excessive stat stacking while maintaining progression.
```

## Stat Applications

### Combat Integration
```go
// Strength affects damage output
baseDamage := weaponDamage + (character.Stats.Strength.ValueAdj / 10)

// Speed affects hit chance and attack frequency
hitBonus := character.Stats.Speed.ValueAdj - target.Stats.Speed.ValueAdj
attacksPerRound := 1 + (character.Stats.Speed.ValueAdj / 50)

// Vitality affects health capacity
healthMax := baseHealth + (character.Stats.Vitality.ValueAdj * 2)
```

### Magic System Integration
```go
// Smarts affects spell success and power
spellBonus := character.Stats.Smarts.ValueAdj / 5
spellSuccessChance += spellBonus

// Mysticism affects mana capacity and regeneration
manaMax := baseMana + (character.Stats.Mysticism.ValueAdj * 3)
manaRegen := 1 + (character.Stats.Mysticism.ValueAdj / 25)
```

### Skill System Integration
```go
// Perception affects detection and awareness
detectionRange := baseRange + (character.Stats.Perception.ValueAdj / 20)
hiddenItemChance := baseChance + (character.Stats.Perception.ValueAdj / 10)

// Stats can modify skill effectiveness
skillBonus := (relevantStat.ValueAdj - 50) / 10 // Bonus/penalty from stat
```

## Character Development Patterns

### Racial Specialization Examples
```go
// Example racial base values
Human := Statistics{
    Strength:   StatInfo{Base: 10}, // Balanced
    Speed:      StatInfo{Base: 10}, // Balanced
    Smarts:     StatInfo{Base: 10}, // Balanced
    Vitality:   StatInfo{Base: 10}, // Balanced
    Mysticism:  StatInfo{Base: 10}, // Balanced
    Perception: StatInfo{Base: 10}, // Balanced
}

Elf := Statistics{
    Strength:   StatInfo{Base: 8},  // Lower physical strength
    Speed:      StatInfo{Base: 12}, // Higher agility
    Smarts:     StatInfo{Base: 12}, // Higher intelligence
    Vitality:   StatInfo{Base: 8},  // Lower health
    Mysticism:  StatInfo{Base: 15}, // Much higher magic affinity
    Perception: StatInfo{Base: 11}, // Better awareness
}

Dwarf := Statistics{
    Strength:   StatInfo{Base: 15}, // Much higher strength
    Speed:      StatInfo{Base: 8},  // Lower speed
    Smarts:     StatInfo{Base: 9},  // Slightly lower intelligence
    Vitality:   StatInfo{Base: 15}, // Much higher health
    Mysticism:  StatInfo{Base: 6},  // Lower magic affinity
    Perception: StatInfo{Base: 10}, // Average perception
}
```

### Training Point Allocation Strategies
```go
// Warrior build (focusing on combat effectiveness)
warrior.Stats.Strength.Training = 50   // Maximum damage output
warrior.Stats.Vitality.Training = 40   // Survivability
warrior.Stats.Speed.Training = 30       // Combat speed
// Total: 120 training points

// Mage build (focusing on magical power)
mage.Stats.Smarts.Training = 50        // Spell effectiveness
mage.Stats.Mysticism.Training = 50     // Mana capacity
mage.Stats.Perception.Training = 20    // Awareness
// Total: 120 training points

// Balanced build (versatile character)
balanced.Stats.Strength.Training = 20
balanced.Stats.Speed.Training = 20
balanced.Stats.Smarts.Training = 20
balanced.Stats.Vitality.Training = 20
balanced.Stats.Mysticism.Training = 20
balanced.Stats.Perception.Training = 20
// Total: 120 training points
```

## Equipment and Modifier Integration

### Dynamic Stat Modification
```go
// Apply equipment bonuses
func ApplyEquipmentBonuses(character *Character)
```

### Temporary Stat Changes
```go
// Spell effect example: Bull's Strength (+10 Strength for 30 rounds)
func CastBullsStrength(caster *Character, target *Character)
```

## Integration Patterns

### Character System Integration
```go
// Stats are core to character functionality
- character.Stats.Strength.ValueAdj    // Combat damage calculations
- character.Stats.Speed.ValueAdj       // Combat speed and dodging
- character.Stats.Vitality.ValueAdj    // Health capacity calculation
- character.Stats.Mysticism.ValueAdj   // Mana capacity calculation
```

### Equipment System Integration
```go
// Equipment modifies stats temporarily
- item.StatMod("strength")             // Equipment stat bonuses
- character.Stats.Strength.SetMod()    // Apply equipment modifiers
- character.Stats.Strength.Recalculate() // Update final values
```

### Buff System Integration
```go
// Buffs can modify stats
- buff.StatMod("strength")             // Buff stat modifications
- character.Buffs.StatMod("strength")  // Total buff modifications
- character.Stats.Strength.Recalculate() // Update with buff effects
```

## Usage Examples

### Character Creation
```go
// Create new character with racial stats
func CreateCharacter(raceId int) *Character
```

### Training Point Allocation
```go
// Spend training points on a stat
func TrainStat(character *Character, statName string, points int) error
```

### Level Up Processing
```go
// Handle character level increase
func LevelUp(character *Character)
```

### Equipment Change Handling
```go
// Update stats when equipment changes
func EquipItem(character *Character, item *Item)

func ApplyAllModifiers(character *Character)
```

## Dependencies

- `math` - Mathematical functions for diminishing returns calculations

This comprehensive stats system provides balanced character development with meaningful choices, racial differentiation, and flexible modification support while maintaining game balance through sophisticated progression mathematics and diminishing returns.