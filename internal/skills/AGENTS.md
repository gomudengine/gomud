# Skills Package Guide

## Scope

- Use this file for skill definitions, progression behavior, and skill-side helpers in `internal/skills`.
- Skill changes are player-facing and often couple to commands, combat, and spells.

## Working Rules

- Preserve compatibility of existing skill identifiers and progression semantics unless the task explicitly changes gameplay balance.
- Be careful with skill-level math and thresholds. Small numeric changes can have broad balance impact.
- If a change affects a skill used by commands or combat, review the caller path as well rather than assuming the package boundary is enough.
- Avoid mixing unrelated gameplay refactors into a skill-table or progression fix.

## Verification

- Run targeted package tests for skill lookup or progression changes.
- Use at least one caller-level check if the change affects combat, usercommands, or spell gating.
- Note balance-sensitive changes explicitly if broad gameplay validation was not run.

## Documentation

- Keep only durable skill-system rules here.
