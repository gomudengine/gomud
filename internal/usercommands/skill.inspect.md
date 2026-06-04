# skill.inspect

## Skill Tag & Training

- Skill: `inspect` (`internal/skills/skills.go`)
- Levels 1-4.

## Overview

`inspect` reveals progressively more detailed information about an item in the player's backpack. Each skill level unlocks an additional panel of information in the output.

## Preconditions

- Inspect skill level >= 1.
- A target item name argument is required.
- The item must be in the player's backpack (not worn).
- Cooldown: **3 rounds** per use (keyed `inspect`).

## Execution Flow

1. Skill level check — rejects if level 0.
2. Argument check.
3. Item lookup via `user.Character.FindInBackpack(rest)`.
4. Cooldown check.
5. Fires a `SkillUsed` event (`skill: inspect`).
6. Sends a brief action message to the player and room.
7. Calls `buildInspectPanel(skillLevel, &item, &iSpec)` and sends the result.

## Output Panels by Level

The output is built in `buildInspectPanel` in `descriptions.panels.go` and consists of up to four stacked panels.

### Panel 1 — Basic Info (always shown)

- Item name (uppercased)
- Description
- Type and subtype
- Gold value

### Panel 2 — Specific Stats (level >= 2)

- Damage dice and attack count (weapons only)
- One-handed or two-handed designation (weapons)
- Speed penalty if `WaitRounds > 0`
- Damage reduction percentage (armor)
- Uses remaining / max uses (consumables)
- Fragility / break chance on use

If level < 2, this panel shows "Unknown...".

### Panel 3 — Modifiers (level >= 3)

- Passive stat modifiers from `iSpec.StatMods`
- Worn buffs (`WornBuffIds`): buff name, stat modifiers, and flags
- On-use buffs (`BuffIds`): buff name, activation rate, stat modifiers, and flags

If level < 3, this panel shows "Unknown...".

### Panel 4 — Magical Effects (level >= 4)

- Curse status
- Elemental type
- Critical hit buffs (`Damage.CritBuffIds`): buff name and activation rate

If level < 4, this panel shows "Unknown...".

## Notes

- The item must be in the backpack, not currently equipped. A message hints at this if the item is not found.
