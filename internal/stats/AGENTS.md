# Stats Package Guide

## Scope

- Use this file for base statistics, derived stat helpers, and stat-structure behavior in `internal/stats`.
- This package underpins many other systems, so compatibility matters more than cleanup.

## Working Rules

- Preserve field meaning and derived-stat semantics unless the task explicitly changes gameplay rules.
- Be careful with default values, derived calculations, and serialization-visible changes.
- If a change affects combat, vitals, leveling, or statmods, inspect the consumer path rather than assuming the package boundary is sufficient.
- Avoid adding unrelated game policy here when it belongs in a higher-level system.

## Verification

- Run targeted package tests for stat calculation or representation changes.
- Use a higher-level caller check if the change affects gameplay-visible outcomes.
- Call out any broad downstream packages not retested.

## Documentation

- Keep this file focused on stat semantics and blast-radius warnings.
