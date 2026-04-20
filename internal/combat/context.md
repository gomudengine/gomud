# GoMud Combat System Context

## Overview

The GoMud combat system provides comprehensive turn-based combat mechanics with support for player vs player, player vs mob, and mob vs mob encounters. It features sophisticated damage calculations, dual wielding, critical hits, backstab mechanics, pet participation, alignment-based consequences, and detailed combat messaging with cross-room attack support.

## Architecture

The combat system is built around several key components:

### Core Components

**Combat Resolution Engine:**
- Turn-based combat with speed-based attack frequency
- Multi-attack system based on speed differentials
- Weapon-based damage calculations with racial bonuses
- Defense reduction and damage mitigation
- Critical hit system with buff effects

**Attack Result System:**
- Comprehensive result tracking for damage, hits, and effects
- Multi-target messaging system for attacker, defender, and rooms
- Support for cross-room combat with directional messaging
- Buff application tracking for combat effects

**Combat Calculations:**
- Hit chance calculations based on speed statistics
- Critical hit probability with level and stat modifiers
- Damage reduction through defense statistics
- Combat odds assessment via expected DPS modeling
- Alignment change calculations for PvP consequences

## Key Features

### 1. **Multi-Type Combat Support**
- Player vs Mob combat with damage tracking
- Player vs Player combat with alignment consequences
- Mob vs Player combat with AI integration
- Mob vs Mob combat with charm attribution

### 2. **Advanced Combat Mechanics**
- Speed-based multiple attacks per round
- Dual wielding with skill-based penalties
- Backstab mechanics with guaranteed critical hits
- Pet participation in combat (20% chance)
- Cross-room combat support with directional messaging

### 3. **Weapon and Equipment Integration**
- Weapon-specific damage dice and bonuses
- Racial weapon preferences and unarmed combat
- Equipment-based defense calculations
- Weapon subtype messaging and effects
- Stat modification integration

### 4. **Combat Messaging System**
- Dynamic message selection based on damage percentage
- Token-based message customization
- Separate messaging for same-room vs cross-room combat
- Critical hit and backstab message highlighting
- Damage reduction feedback

## Combat Structure

### Attack Result Data Structure
```go
type AttackResult struct {
    Hit                     bool     // Whether the attack connected
    Crit                    bool     // Whether it was a critical hit
    BuffSource              []int    // Buffs applied to attacker
    BuffTarget              []int    // Buffs applied to target
    DamageToTarget          int      // Total damage dealt to target
    DamageToTargetReduction int      // Damage blocked by target's defense
    DamageToSource          int      // Damage dealt to attacker (rare)
    DamageToSourceReduction int      // Damage blocked by attacker's defense
    MessagesToSource        []string // Messages sent to attacker
    MessagesToTarget        []string // Messages sent to target
    MessagesToSourceRoom    []string // Messages sent to attacker's room
    MessagesToTargetRoom    []string // Messages sent to target's room
}
```

### Combat Type Enumeration
```go
type SourceTarget string

const (
    User SourceTarget = "user"  // Player character
    Mob  SourceTarget = "mob"   // NPC character
)
```

## Combat Resolution System

### Player vs Mob Combat
```go
// Main combat function for player attacking mob
func AttackPlayerVsMob(user *users.UserRecord, mob *mobs.Mob) AttackResult
```

### Player vs Player Combat
```go
// PvP combat with alignment consequences
func AttackPlayerVsPlayer(userAtk *users.UserRecord, userDef *users.UserRecord) AttackResult
```

### Mob vs Player Combat
```go
// NPC attacking player with AI integration
func AttackMobVsPlayer(mob *mobs.Mob, user *users.UserRecord) AttackResult
```

### Mob vs Mob Combat
```go
// NPC vs NPC combat with charm attribution
func AttackMobVsMob(mobAtk *mobs.Mob, mobDef *mobs.Mob) AttackResult
```

## Combat Calculation Engine

### Main Combat Resolution
```go
func calculateCombat(sourceChar characters.Character, targetChar characters.Character,
                    sourceType SourceTarget, targetType SourceTarget) AttackResult
```

### Hit Calculation System
```go
// Calculate hit chance based on speed statistics
func hitChance(attackSpd, defendSpd int) int

// Determine if attack hits with modifiers
func Hits(attackSpd, defendSpd, hitModifier int) bool
```

### Critical Hit System
```go
// Calculate critical hit probability
func Crits(sourceChar characters.Character, targetChar characters.Character) bool
```

## Dual Wielding System

### Weapon Selection and Penalties
```go
// Determine weapons available for attack
func getAttackWeapons(sourceChar characters.Character) []items.Item

// Calculate dual wield penalty
func getDualWieldPenalty(weaponCount int, dualWieldLevel int) int
```

## Combat Messaging System

### Message Token System
```go
// Token replacement for dynamic combat messages
func buildTokenReplacements(sourceChar, targetChar characters.Character,
                          sourceType, targetType SourceTarget,
                          weaponName string, damage int) map[items.TokenName]string
```

### Cross-Room Combat Messaging
```go
// Handle messaging for cross-room combat
func handleCrossRoomMessages(sourceChar, targetChar characters.Character,
                           attackResult *AttackResult, msgs items.AttackMessages)

// Add exit/entrance direction tokens for cross-room combat
func addDirectionalTokens(sourceChar, targetChar characters.Character,
                         tokens map[items.TokenName]string)
```

## Combat Calculations and Utilities

### Combat Odds System

`CombatOdds(atkChar, defChar)` returns the ratio of rounds-for-defender-to-kill-attacker divided by rounds-for-attacker-to-kill-defender. A value above `1.0` means the attacker wins faster; below `1.0` means the defender wins faster; `1.0` is an even fight.

It is backed by `expectedDPS(attacker, defender)`, which computes deterministic expected damage per round by mirroring the live combat logic:
- Attack count from speed differential (`ceil((atkSpd - defSpd) / 25)`, min 1) plus `StatMod("attacks")`
- Weapon selection and dual-wield probability weights (0%, 50%, or 100% offhand contribution based on skill level; always 100% for dual claws)
- Hit chance from `hitChance()`, clamped 5-95%, with dual-wield penalty on the offhand slot
- Crit chance: `5 + round((Strength + Speed) / levelDiff)`, clamped to min 5%
- Average dice damage: `N*(S+1)/2 + bonus`, with crit bonus added as `(dCount*dSides+dBonus) * critPct`
- Defense reduction as expected value: `GetDefense() / 200` fraction of damage absorbed

```go
// Returns ratio of (rounds defender takes to kill attacker) / (rounds attacker takes to kill defender).
// > 1.0 favors the attacker; < 1.0 favors the defender; 1.0 is even.
func CombatOdds(atkChar characters.Character, defChar characters.Character) float64
```

### Taming Mechanics
```go
// Calculate chance to tame a mob
func ChanceToTame(s *users.UserRecord, t *mobs.Mob) int
```

### Alignment Change System
```go
// Calculate alignment change from PvP combat
func AlignmentChange(killerAlignment int8, killedAlignment int8) int
```

## Pet Combat System

### Pet Participation
```go
// Handle pet joining combat (20% chance per round)
func processPetAttack(sourceChar characters.Character, targetChar characters.Character,
                     attackResult *AttackResult, sourceType, targetType SourceTarget)
```

## Integration Patterns

### Character System Integration
```go
// Combat integrates deeply with character stats and equipment
- sourceChar.Stats.Speed.ValueAdj  // Speed for hit chance and attack frequency
- sourceChar.Stats.Strength.ValueAdj  // Strength for critical hit chance
- sourceChar.Equipment.Weapon.GetDiceRoll()  // Weapon damage calculations
- sourceChar.GetDefense()  // Defense for damage reduction
- sourceChar.StatMod("damage")  // Stat modifications for bonus damage
- sourceChar.HasBuffFlag(buffs.Accuracy)  // Buff effects on combat
```

### Event System Integration
```go
// Combat results trigger various game events
- user.WimpyCheck()  // Automatic flee on low health
- mob.Character.TrackPlayerDamage()  // Damage tracking for loot distribution
- user.PlaySound()  // Audio feedback for combat actions
- sourceChar.SetAggro()  // Aggression state management
```

### Item System Integration
```go
// Weapons provide combat capabilities and messaging
- weapon.GetDiceRoll()  // Damage calculation from weapon stats
- weapon.StatMod()  // Racial bonuses and special modifiers
- items.GetAttackMessage()  // Dynamic combat messaging
- weapon.DisplayName()  // Weapon identification in messages
```

## Usage Examples

### Basic Combat Initiation
```go
// Player attacks mob
user := users.GetByUserId(userId)
mob := mobs.GetInstance(mobInstanceId)

if user != nil && mob != nil {
    result := combat.AttackPlayerVsMob(user, mob)
    
    // Send messages to all participants
    for _, msg := range result.MessagesToSource {
        user.SendText(msg)
    }
    
    // Check if mob died
    if mob.Character.Health <= 0 {
        handleMobDeath(mob, user)
    }
}
```

### PvP Combat with Consequences
```go
// Player vs player combat
attacker := users.GetByUserId(attackerId)
defender := users.GetByUserId(defenderId)

result := combat.AttackPlayerVsPlayer(attacker, defender)

// Calculate alignment change
alignmentChange := combat.AlignmentChange(attacker.Character.Alignment, defender.Character.Alignment)
attacker.Character.Alignment += int8(alignmentChange)

// Handle death consequences
if defender.Character.Health <= 0 {
    handlePlayerDeath(defender, attacker)
}
```

### Combat Assessment
```go
// Assess relative combat odds
player := users.GetByUserId(userId)
mob := mobs.GetInstance(mobInstanceId)

odds := combat.CombatOdds(*player.Character, mob.Character)

if odds >= 1.75 {
    player.SendText("This should be an easy fight.")
} else if odds < 0.9 {
    player.SendText("This looks very dangerous!")
} else {
    player.SendText("This should be a fair fight.")
}
```

## Dependencies

- `internal/characters` - Character stats, equipment, and abilities
- `internal/items` - Weapon specifications and combat messaging
- `internal/users` - Player character management and state
- `internal/mobs` - NPC character management and AI integration
- `internal/buffs` - Status effects that modify combat
- `internal/skills` - Skill system for dual wielding and combat abilities
- `internal/races` - Racial bonuses and unarmed combat specifications
- `internal/rooms` - Room management for cross-room combat
- `internal/util` - Dice rolling and random number generation
- `internal/configs` - Configuration for combat behavior and messaging

This comprehensive combat system provides sophisticated turn-based combat mechanics with support for all character types, advanced weapon systems, detailed messaging, and seamless integration with all other game systems.