# Modernize GitHub Actions and Release Builds

## Summary

- Implement the work across four focused PRs to reduce release risk.
- Split the current draft branch into separate runtime and release-automation PRs because the XP clamp and GitHub Actions changes have different reviewers, risks, and rollback paths.
- Support official release binaries for `linux/amd64`, `linux/arm64`, `linux/arm/v7`, `windows/amd64`, `windows/arm64`, `darwin/amd64`, and `darwin/arm64`.
- Keep XP values as the current `int` domain and clamp XP formulas with `math.MaxInt`.
- Split mutable prereleases from protected stable releases, then add targeted CI speedups.

## PR 1: Clamp XP Progression Values

- Replace `Character.XPTL`'s `math.MaxInt64` clamp with `math.MaxInt`.
- Add the same `math.MaxInt` clamp to `internal/web/api_v1_progression.go`'s `xpTLWithCfg`.
- Add high-XP progression tests proving engine and admin preview clamp consistently.

## PR 2: Update Current Release Asset Matrix

- Replace `linux/arm GOARM=5` with `linux/arm GOARM=7`.
- Add `linux/arm64`.
- Preserve existing `x64` asset names and add ARM assets:
  - `gomud-linux_x64`
  - `gomud-linux_arm64`
  - `gomud-linux_armv7`
  - `gomud-windows_x64.exe`
  - `gomud-windows_arm64.exe`
  - `gomud-darwin_x64`
  - `gomud-darwin_arm64`
- Update current release workflows' build steps, checksum generation, and upload lists for the seven binaries.
- Add CI cross-compile checks for all seven release targets

## PR 3: Redesign Release Publishing

- If possible: reviewers should not be notified for draft PRs, only when the PR is set to ready for review.
- Split the current mixed release flow into:
  - `Prerelease`: push to `master`, mutable rolling `prerelease`, no protected environment.
  - `Stable Release`: manual `workflow_dispatch` with required semver `release_tag`, protected GitHub Environment, no tag moves or asset clobbering.
- Keep repo-wide immutable releases disabled while rolling `prerelease` stays mutable.
- Enforce stable-release immutability by workflow policy: fail if the semver tag or release already exists.
- Publish stable releases draft-first: create draft, upload assets, attach notes, then publish.
- Define release targets once in `.github/scripts/release-assets.sh`.
- Build release binaries from that target table and upload one
  `release-binaries` artifact.
- Assemble all binaries, `gomud-ALL-datafiles.zip`, and `SHA256SUMS.txt` in the publish job.
- Add artifact attestations for each binary and `SHA256SUMS.txt` with a single
  multi-subject attestation step.
- Use job-level permissions:
  - build/test jobs: `contents: read`;
  - publish jobs: `contents: write`, `id-token: write`, `attestations: write`.
- Centralize release note generation in `scripts/release-notes.sh` or a composite action.
- Release notes include `Overview`, `Downloads`, `Install From Source`, `Manual Binary Install`, `Verify Provenance`, and `Changes`.
- Insert GitHub auto-generated notes from `releases/generate-notes`.
- Avoid shell-injection risk by writing generated notes to files and not evaluating generated release text as shell input.
- Drop `go build -v` from release builds unless debugging.
- Avoid duplicate full test runs on `master` by making prerelease publishing depend on the normal test workflow success or by combining test/build/release with a single test job.

## PR 4: Installers and CI Optimization

- Keep current installers as source-build installers.
- Add release ref support:
  - `scripts/install.sh`: support `GOMUD_VER=v0.9.9` or `GOMUD_VER=prerelease` before build.
  - `scripts/install.ps1`: support `GOMUD_VER` environment variable and `-version` parameter.
- Update `scripts/install.sh` architecture detection to recognize `armv7l` as ARMv7.
- Do not claim install scripts download release binaries.
- Release notes state official Windows binary assets are `windows/amd64` and `windows/arm64`.
- Add a docs-only change detector so Go tests, release cross-compiles, and Docker builds can be skipped when changes are limited to docs or non-runtime metadata.
- Keep checks enabled for runtime/build inputs: `*.go`, `go.mod`, `go.sum`, `Makefile`, `_datafiles/**`, `modules/**`, `cmd/**`, `internal/**`, `scripts/**`, `provisioning/**`, `compose.yml`, and relevant workflow files.
- Make Docker PR builds conditional on Docker/runtime/build input changes.
- Keep `go test -race ./...` on PRs and `master` initially; revisit PR race tests only if CI time remains a problem.
- Keep `cancel-in-progress: true` for PR test/lint runs.
- Consider `cancel-in-progress: true` for Docker `master` image builds so only the newest image publishes.
- Keep stable release publishing `cancel-in-progress: false`.
- Use separate Docker cache scopes for PR and master if cache contention appears.
- Fix `fmtcheck` so it checks formatting without first mutating files.
- Fix `make js-lint` so missing Docker does not get masked as success; prefer
  local `npx`, fall back to Docker, and fail if neither is available.
- Correct Dependabot Docker paths to `/provisioning` and `/provisioning/terminal`.
- Keep `go generate` behavior explicit: run once in CI and verify generated output is clean before tests/builds depend on it.

## Test Plan

- PR 1:
  - `go test ./internal/characters ./internal/web`
  - `make validate`
- PR 2:
  - `make validate`
  - `go test -race ./...`
  - all seven supported cross-compiles
  - verify current release workflow dry-run/checksums reference all seven binaries
- PR 3:
  - workflow lint / `make ci-local` if Docker is available
  - dry-run release note generation
  - prerelease run on `master`
  - stable release dispatch against a test semver tag, including existing tag/release failure behavior
  - verify binaries, datafiles archive, checksums, and attestations
- PR 4:
  - installer source-build paths with and without `GOMUD_VER` / `-Ref`
  - Linux `armv7l` detection logic
  - docs-only and Docker-conditional skip behavior with representative changed-file sets
  - `make ci-local` if Docker is available

## Assumptions

- XP remains an `int` domain; no `int64` persistence/API migration is in scope.
- 32-bit official binary support means Linux ARMv7 only, not ARMv5 or `linux/386`.
- Existing `x64` artifact names are preserved for compatibility.
- Release asset renames apply to new releases; stale asset cleanup is not required for this branch's release path.
- Repo-wide immutable releases remain disabled while rolling `prerelease` stays mutable.
