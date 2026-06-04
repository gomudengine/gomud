# skill.peep

## Skill Tag & Training

- Skill: `peep` (`internal/skills/skills.go`)
- Levels 1-4.
- Level 1 is a passive ability; levels 2-4 are active commands.

## Overview

`peep` is a reconnaissance skill. At level 1 it passively shows health percentages on NPCs in room descriptions. At higher levels it becomes an active command that reveals detailed stats, inventory, and drop chances of a target.

## Level 1 — Passive Health Display

Level 1 requires no command invocation. The peep skill level is checked elsewhere in the room description rendering pipeline to determine whether to show health percentages next to mob names. Calling `peep` at level 1 returns an error telling the player it is passive.

## Active Command (Levels 2-4)

### Preconditions

- Skulduggery skill level >= 2 to use the active command.
- A target name argument is required.
- Cooldown: **1 round** per use (keyed `peep`).

### Execution Flow

1. Skill level check — rejects with an error if level 0.
2. Argument check.
3. Level check for active use — rejects at level 1 with a passive-skill message.
4. Cooldown check.
5. Target resolution via `room.FindByName`.
6. Fires a `SkillUsed` event (`skill: peep`).
7. Builds and sends output panels based on skill level.

### Output by Level

#### Level 2 — Detailed Stats

Calls `buildPeepPanel(character)` which renders a panel showing the target's core stats, health, mana, and other character details.

#### Level 3 — Stats + Inventory

Adds `buildPeepInventoryPanel(character, itemNames)` showing all items in the target's backpack. Consumable items with limited uses show their remaining use count.

#### Level 4 — Stats + Inventory + Drop Chance

- For **players**: always reports 100% drop chance for equipment on death.
- For **mobs**: reports the mob's actual `ItemDropChance` percentage.

### Visibility

- Peeping at a player sends them a notification: "\<user\> is peeping at you."
- Peeping at a mob sends a room-wide message (excluding the peeper).
- Peeping at a player also sends a room-wide message (excluding both parties).
