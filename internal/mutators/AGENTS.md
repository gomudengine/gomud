# Mutators Guide

## Scope

- `internal/mutators` manages room/world mutator specs plus runtime lifecycle state for spawned, despawned, decaying, and respawning mutators.
- This package is tightly coupled to game time, exits, buffs, and room presentation. Small field changes can have broad gameplay effects.

## Files

- `mutators.go`: spec structs, runtime list operations, update/decay logic, YAML loading, and text/exits helpers.

## Working Rules

- Keep spec YAML tags, load logic, and runtime assumptions in sync. This package is data-driven and has very little defensive layering.
- `Mutator.Update` is the critical path. Changes to decay or respawn rules must preserve initialization, special time-of-day respawns, and `DecayIntoId` transitions.
- `MutatorList.Update` removes entries only after per-mutator updates. Be careful not to introduce slice mutation bugs when changing lifecycle handling.
- Exit and buff fields are consumed outside this package. Search for `PlayerBuffIds`, `MobBuffIds`, `NativeBuffIds`, `LightMod`, and `Exits` before changing semantics.

## Verification

- There are no package-local tests here today, so use focused repo validation for any behavior change.
- At minimum, run the relevant Go test target that exercises rooms or gameplay paths affected by the mutator change.
