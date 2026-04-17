# GoMud Code Review

Use this checklist for `/review` and pre-PR local review.

## Review Focus

- Prioritize bugs, behavior regressions, missing tests, unsafe defaults, and user-visible breakage.
- Lead with findings ordered by severity; include file and line references.
- If there are no findings, say that clearly and list any checks not run.
- Prefer concrete fixes over broad style advice.
- When JavaScript files change, check whether the current lint rules and lint scope still match the intended review contract.

## Verification Map

- Go runtime or config changes: run `make validate`; add targeted `go test ./internal/<pkg>` or `make test` for broad/shared changes.
- Frontend JavaScript changes: run `make js-lint`; use browser or Playwright checks for user-visible UI behavior.
- `.github/` workflow changes: run `make ci-local` when practical.
- Docs-only changes: skip full tests unless behavior or commands changed; use `make help` when command docs changed.
- Shell script changes: run `shellcheck` and `shfmt` if available, plus the relevant Makefile target.

## Repo Rules

- Prefer Makefile targets over ad hoc commands.
- Treat `make js-lint` as the JavaScript lint source of truth.
- If JavaScript files changed, update the lint rules or lint scope when the code change makes the current policy incomplete or misleading.
- Remember GitHub required check names are `lint` and `test`.
- Update docs only when setup, usage, behavior, configuration, or developer workflow changed.
