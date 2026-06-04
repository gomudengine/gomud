# skill.tame

## Skill Tag & Training

- Skill: `tame` (`internal/skills/skills.go`)
- Trained by giving a mushroom to the fairie in room 558, then training in room 830.
- Levels 1-4.

## Overview

`tame` allows a player to attempt to charm a mob, turning it into a temporary companion. Higher skill levels allow more simultaneously charmed creatures. The actual taming resolution is handled by the `tameskill` spell script, not within this command file.

## Charmed Creature Limits by Level

| Level | Max charmed creatures |
|---|---|
| 1 | 2 |
| 2 | 3 |
| 3 | 4 |
| 4 | 5 |

These limits are enforced by the spell/charm system, not directly in this command.

## Sub-command: `tame list` (or no argument)

Displays a table of the player's taming proficiency for each mob species they have previously tamed. Data comes from `user.Character.MobMastery.GetAllTame()`, which maps mob IDs to proficiency percentages. The table is rendered via the generic table template.

## Active Tame Attempt

### Preconditions

- Tame skill level >= 1.
- A mob name argument is required.
- The target mob must be in the current room.
- The target mob must not already be charmed by the player.

### Execution Flow

1. Skill level check — rejects if level 0.
2. If argument is `list` or empty, shows proficiency table and returns.
3. Target resolution via `room.FindByName` (mob only; player targets are ignored).
4. Fires a `SkillUsed` event (`skill: tame`).
5. Checks if the mob is already charmed by the player — rejects if so.
6. Builds a `SpellAggroInfo` targeting the mob with spell ID `tameskill`.
7. Calls `scripting.TrySpellScriptEvent("onCast", ...)` — allows scripts to intercept.
8. If not intercepted, cancels the `Hidden` buff and calls `user.Character.SetCast(waitRounds, spellAggro)` to queue the tame cast.

## Spell Resolution

The tame is resolved as a spell cast (`tameskill`). The wait rounds, success probability, and charm application are all handled by the `tameskill` spell script and the spell resolution system — not in this command file.

## Notes

- Only mob targets are valid; the command ignores player IDs returned by `FindByName`.
- Taming proficiency (`MobMastery`) tracks how experienced the player is with each mob type, likely influencing success rates in the spell script.
