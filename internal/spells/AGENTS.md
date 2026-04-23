# Spells Package Guide

## Scope

- Use this file for spell metadata, spell lookup, and spell-side persistence/helpers in `internal/spells`.
- Spell behavior usually crosses into scripting, combat, buffs, and commands.

## Working Rules

- Preserve spell identifiers and lookup behavior unless the task explicitly changes spell contracts.
- Be careful with changes that affect castability, targeting, cooldown metadata, or script linkage.
- If spell behavior is script-backed, verify the scripting-side expectation before changing how spell data is loaded or exposed.
- Keep spell metadata changes separate from unrelated combat or command refactors when possible.

## Verification

- Run targeted package tests for spell lookup or persistence changes.
- Use a higher-level validation path if the change affects casting flow, combat, or script-backed spells.
- Call out compatibility risk when changing how existing spells are loaded or resolved.

## Documentation

- Keep this file about spell contract and integration guardrails.
