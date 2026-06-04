# skill.dualwield

## Skill Tag & Training

- Skill: `dualwield` (`internal/skills/skills.go`)
- Levels 1-4.

## Overview

`dualwield` is primarily a passive skill that modifies combat behavior when two weapons are equipped. The command itself only validates that the player knows the skill and then displays the help page.

## Skill Level Effects (passive, applied by the combat system)

| Level | Effect |
|---|---|
| 1 | Enables dual wielding weapons that would normally be incompatible. Attacks use a random weapon each round. |
| 2 | Occasionally attacks with both weapons in one round. |
| 3 | Always attacks with both weapons when dual wielding. |
| 4 | Dual wielding incurs fewer penalties. |

These effects are enforced in the combat resolution layer, not in this file.

## Command Behavior

Typing `dualwield` with the skill known simply delegates to `Help("dual-wield", ...)`, displaying the help entry for dual wielding. There is no active ability to trigger.

If the player does not know the skill (level 0), the command returns an error: "You haven't learned how to dual wield."

## Notes

- There is no cooldown, no stat roll, and no direct game-state change from invoking this command.
- The actual dual-wield attack logic is entirely within the combat system.
