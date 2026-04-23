# Keywords Package Guide

## Scope

- Use this file for aliases, help-topic keyword mapping, direction aliases, and map-legend overrides in `internal/keywords`.
- Changes here are user-facing and can affect command parsing and help lookup broadly.

## Working Rules

- Preserve compatibility of existing aliases unless the task explicitly changes player-facing terminology.
- Be careful with normalization and lookup behavior. Small changes can quietly break help, directions, or command shortcuts.
- Keep keyword data and runtime lookup behavior aligned; do not patch one side without checking the other.
- Prefer additive changes over silent replacement when expanding alias coverage.

## Verification

- Run targeted tests for alias-resolution or lookup changes.
- If the change affects help topics or direction parsing, verify the exact user-facing resolution path involved.
- Use a higher-level check if the change affects command lookup outside this package.

## Documentation

- Keep only durable lookup and compatibility rules here.
