# Suggestions Guide

## Scope

- `internal/suggestions` covers two things only: cycling UI suggestions and autocomplete result generation for player input.
- Modules extend autocomplete through `OnAutoComplete`; keep this package as the shared coordination point rather than a grab bag of module logic.

## Files

- `suggestions.go`: simple stateful suggestion cycling used by the input UI.
- `autocomplete.go`: built-in completion logic plus the `OnAutoComplete` hook contract.

## Working Rules

- `GetAutoComplete` returns suffixes to append to the current input, not full replacement strings. Preserve that contract when adding cases.
- The hook is additive. Module handlers should append to `Results` and return the request unchanged when they do not own the command.
- Built-in completion order matters because results are sorted and used interactively. Avoid expensive scans or duplicate-heavy output on the hot path.
- Alias handling happens before most command-specific logic. If you add command branches, use the resolved command name, not only the raw first token.

## Verification

- Run `go test ./internal/suggestions`.
- If you change autocomplete behavior, sanity-check one built-in command and one module-provided completion path.
