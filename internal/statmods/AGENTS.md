# Statmods Package Guide

## Scope

- Use this file for stat modification composition and application rules in `internal/statmods`.
- This package influences combat, equipment, buffs, races, and derived character state.

## Working Rules

- Preserve aggregation semantics unless the task explicitly changes how stat modifiers stack.
- Be careful with ordering and source attribution when combining modifiers from equipment, buffs, racial traits, or skills.
- Prefer changing the narrow statmod rule that is wrong instead of refactoring all modifier plumbing at once.
- If a change affects derived combat or vitals behavior, inspect the consuming package too.

## Verification

- Run targeted tests for modifier merge or calculation behavior.
- Use a higher-level check when the change affects player-visible stats or combat outcomes.
- Note any untested cross-source stacking cases if behavior changed.

## Documentation

- Keep only stacking and compatibility rules here.
