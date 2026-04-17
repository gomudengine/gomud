# Repository Guidelines

## Goals

- Make the smallest correct change that satisfies the task.
- Keep PRs as small as practical to make review easier.
- Split Go code changes into a separate PR when they are not required for the task at hand.
- Preserve existing project conventions unless there is a clear reason to change them.
- Keep durable repo conventions here; put repeatable workflows in `.agents/`.

## Project Structure

- `main.go` and `world.go` are the main server entrypoints; most application code lives under `internal/`.
- Reusable game features live under `modules/`; bundled world data, HTML, localization, and sample scripts live under `_datafiles/`.
- Generated module wiring is handled by `cmd/generate/module-imports.go`.
- CI and contributor workflow docs live in `.github/`.
- `ai-context/*.md` is supplemental architecture/background context; treat code, config, and Makefile targets as the source of truth.

## Build, Test, and Run

Use the `Makefile` as the source of truth:

- `make help` lists supported developer targets.
- `make run` regenerates module imports and starts the server locally.
- `make build` validates and builds `./go-mud-server`.
- `make run-docker` starts the Docker Compose environment from `compose.yml`.
- `make validate` runs formatting checks and `go vet`.
- `make test` runs code generation, JavaScript lint, and `go test -race ./...`.
- `make ci-local` runs the local workflow/`act` validation path; use it when changing `.github/` files.

## Style

- Follow `.editorconfig`: 2-space indentation for Markdown, HTML, JavaScript, YAML, and CSS; tabs in `Makefile`.
- Keep Go code `gofmt`-clean and package names lowercase.
- Match existing naming patterns for Go identifiers, `*_test.go` tests, and `_datafiles/` names.

## Implementation Rules

- Keep secrets, tokens, and credentials out of code, logs, fixtures, and documentation.
- Prefer explicit error propagation or clear surfaced failures over silent fallbacks.
- When behavior changes, update or add relevant Go `testing` / `testify` tests.

## Verification

- Run the relevant existing checks for the files touched; prefer the smallest high-confidence set, then expand if needed.
- Use `make js-lint` for frontend JavaScript changes.
- Run `make test` before opening a PR.
- Do not claim success without verification.
- If checks cannot be run, state that clearly and explain why.

## Codex Skills

- Use `/review` before handoff when changes need bug/regression review or verification; follow `.agents/code_review.md`.
- Use `$gh-fix-ci` for failing GitHub Actions PR checks; inspect logs with `gh` (if available) and prefer repo-native fixes plus Makefile checks.
- Use `$security-review` only when explicitly requested for security review or threat modeling.

## Documentation

- Update docs when changes materially affect setup, usage, behavior, configuration, or developer workflows.
- Keep docs scoped and aligned with the actual Docker and Makefile workflow.
