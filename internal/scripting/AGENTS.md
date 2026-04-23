# Scripting Package Guide

## Scope

- Use this file for the Goja-backed scripting runtime in `internal/scripting`.
- This package powers room, mob, item, spell, and buff scripts and exposes engine APIs to JavaScript.

## Working Rules

- Preserve the existing separation between script categories and their wrappers. Do not collapse room, mob, item, spell, and buff behavior into a single ad hoc path.
- Keep script-exposed APIs stable unless the task explicitly changes the scripting contract.
- Be careful with timeout, VM reuse, and wrapper behavior. Small runtime changes here can affect all scripted content.
- Prefer extending existing script helpers and wrapper methods rather than adding one-off special cases in individual call paths.
- When changing what scripts can do, consider both engine safety and content compatibility with existing world scripts.

## Verification

- Run targeted `internal/scripting` tests for runtime, wrapper, or helper changes.
- If the change affects a specific script surface such as room or item scripts, verify the nearest package tests and any directly affected integration path.
- Call out any compatibility risk for existing scripts if behavior changed but broad world-content testing was not run.

## Documentation

- Keep this file focused on runtime contracts and safety rules.
- Put long API catalogs in narrower docs if they are still needed.
