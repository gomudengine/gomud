# Zombiemode Module Guide

## Scope

- Use this file for AFK automation behavior, zombie-mode config/profile handling, and zombie event interception in `modules/zombiemode`.
- This module combines per-user persistence with live event-driven automation.

## Working Rules

- Preserve the distinction between saved zombie configuration and live active runtime state.
- Be careful with the wake-up path: real player input must still disable automation reliably.
- If a change affects the AI decision order, treat it as player-visible behavior rather than an internal refactor.
- Keep session-stat tracking and config persistence aligned with player spawn/despawn behavior.
- When roam or combat targeting logic changes, inspect mapper, mobs, or rooms assumptions too.

## Verification

- Run targeted module tests for config, activation, or automation behavior changes.
- Verify the exact automation path changed: wake-up on input, target selection, loot, roam, rest, or profile handling.
- Call out any live round-loop or event-interception behavior not exercised directly.

## Documentation

- Keep this file about automation and persistence guardrails, not a full AI spec.
