# Config Package Guide

## Scope

- Use this file for configuration loading, validation, override persistence, reload behavior, and config-facing admin/API support in `internal/configs`.
- This package is easy to break with small path or merge-order mistakes. Keep changes precise.

## Working Rules

- Treat `_datafiles/config.yaml` as the bundled base config.
- In the current code, `CONFIG_PATH` selects the override file path. It does not switch the bundled base config file.
- The active override path is:
  - `CONFIG_PATH`, when set
  - otherwise `${DataFiles}/config-overrides.yaml`
- Preserve the existing `SetVal` and `ReloadConfig` contract:
  - runtime updates persist to the override file, not the bundled base config
  - reload starts from the bundled base config, then reapplies overrides, env assignments, and validation
- Be careful when changing `DataFiles`, override-path calculation, or reload order. Those changes can silently redirect where overrides are read or written.
- Respect `Server.Locked` and other validation-driven guardrails. Do not bypass them accidentally in new helper paths.
- Keep dot-path naming and type conversion behavior compatible with existing config callers unless the task explicitly changes that contract.

## Verification

- Run targeted tests in `internal/configs` for config-layering, reload, or persistence changes.
- Use broader repo checks such as `make validate` when a config change affects other packages.
- If you change override-path or reload semantics, verify both read and write behavior, not just one side.

## Documentation

- Keep only durable override and validation rules here.
- Put operator-facing config walkthroughs in README-style docs, not in this always-loaded file.
