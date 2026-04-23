# Usercommands Package Guide

## Scope

- Use this file for player command registration and command-side behavior in `internal/usercommands`.
- This package is heavily user-facing and tightly coupled to many gameplay systems.

## Working Rules

- Keep command changes narrowly scoped. Avoid bundling broad gameplay refactors into a command fix.
- Preserve command names, argument expectations, and admin/user availability unless the task explicitly changes the player-facing contract.
- If a command interacts with another subsystem, review the subsystem rule as well instead of open-coding assumptions in the command.
- Prefer existing prompt, parsing, and helper patterns over inventing a command-local flow.
- Be careful with command registration or removal; that can affect help, aliases, autocomplete, and admin tooling.

## Verification

- Run targeted tests for the changed command path where available.
- Use at least one higher-level validation path that exercises the actual command behavior.
- If a command change is mostly prompt-driven or integration-heavy, note what was not exercised live.

## Documentation

- Keep this file about command-contract and integration rules, not a catalog of all commands.
