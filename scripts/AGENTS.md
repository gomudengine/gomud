# Scripts Guide

## Scope

- Use this file for shell helpers and script-adjacent tests under `scripts/`.
- Keep shell changes compatible with both `bash` and `zsh`; prefer portable `sh` patterns unless the file already requires something else.

## Working Rules

- Preserve the script entrypoints used by the repo, especially `make https-setup`.
- For `https-setup.sh`, do not rewrite the bundled base config directly. The helper is meant to target config overrides and print or apply override-shaped updates.
- Keep operator-facing prompts terse and copy-paste friendly.
- When script behavior changes, update or add the nearby Go tests in `scripts/https_setup_test.go` where they cover the user-facing contract.

## Verification

- After editing shell scripts, run `shellcheck` and `shfmt` if available and fix relevant issues.
- Run the targeted script tests when behavior changes. For `https-setup`, prefer `go test ./scripts`.
- If a change is shell-only and not covered by Go tests, exercise the smallest practical command path and report what you verified.
