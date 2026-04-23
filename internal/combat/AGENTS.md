# Combat Package Guide

## Scope

- Use this file for combat resolution, attack calculations, combat messaging, and combat-side integration work in `internal/combat`.
- Changes here affect gameplay balance and can ripple into buffs, items, stats, mobs, pets, and PvP behavior.

## Working Rules

- Keep combat behavior changes narrowly scoped. Avoid mixing mechanical changes with unrelated cleanup.
- Preserve the distinction between hit chance, critical logic, damage calculation, mitigation, and message rendering unless the task explicitly restructures that boundary.
- Be careful with changes that alter combat odds or pacing. Small math changes can become player-visible balance shifts.
- If a change affects combat messaging, verify both the mechanic and the text path together.
- When combat behavior depends on stats, buffs, weapons, pets, or alignment, inspect the connected package rather than patching assumptions locally.

## Verification

- Run targeted `internal/combat` tests when behavior changes here.
- If the change alters balance-sensitive behavior, add or update focused tests around the exact mechanic changed.
- Use at least one higher-level validation path when the change crosses into mobs, usercommands, or items.

## Documentation

- Keep this file about balance and integration guardrails, not a combat reference manual.
