# skill.brawling.throw

## Skill Tag & Training

- Skill: `brawling` (`internal/skills/skills.go`)
- Trained at the Soldiers Training Yard (room 829), levels 1-4.
- `throw` is the level 2 ability of the brawling skill tree.

## Overview

`throw` lets a player hurl an item from their backpack at a target in the same room, or lob it through an exit into an adjacent room. The item is removed from the player's inventory and placed on the floor of the destination.

## Preconditions

- Brawling skill level >= 2.
- Command requires at least two arguments: the item name and the destination.
- The item must be in the player's backpack.
- Cooldown: **4 rounds** per use (keyed `brawling:throw`).

## Argument Parsing

```
throw <item> <target|exit>
```

Arguments are split respecting quoted strings. The first token is the item name; the remainder is joined as the destination string.

## Execution Flow

1. Skill level check — returns `false` if level < 2.
2. Argument count check — returns `false` if fewer than two arguments.
3. Item lookup in backpack — returns early with a message if not found.
4. Cooldown check — rejects if too soon.
5. Fires a `SkillUsed` event (`skill: brawling`, `details: throw`).
6. Resolves the destination in priority order:
   - **Mob in the room** — removes item from inventory, fires `ItemOwnership` (lost), drops item on room floor.
   - **Player in the room** — same as mob, but checks PvP rules first.
   - **Named exit** — resolves via `room.FindExitByName`, then direction aliases. Checks if the exit is locked. Removes item, fires `ItemOwnership` (lost), places item in the destination room, notifies both rooms.
   - **Temporary exit** — same as named exit but using `room.ExitsTemp`. Requires at least 3 characters in the destination string for a close match.
7. If no destination matched, sends "You don't see a ... to throw it to."

## Grenade Note

The code checks `iSpec.Type == items.Grenade` on all throw paths and has a `// TODO: Grenade` comment. Grenade area-effect behavior is not yet implemented; grenades currently just land on the floor like any other item.

## Item Ownership Events

- A `ItemOwnership{Gained: false}` event is fired for the thrower on every successful throw path.
- No `ItemOwnership{Gained: true}` is fired for the recipient — the item lands on the floor, not in anyone's inventory.

## Notes

- Throwing at a mob or player does not initiate combat or deal damage; it is purely an item-movement mechanic.
- Hidden buff (`buffs.Hidden`) is cancelled when throwing through an exit (the player reveals themselves by the motion), but not when throwing at a target in the same room.
