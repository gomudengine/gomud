## Release Process

[Semantic Versioning 2.0.0 reference](https://github.com/semver/semver/blob/master/semver.md)

Recommended versioning for the next release line:

- Start the next stable release at `v0.10.0`.
- Keep the binary's embedded version in source as a normal semver such as `0.10.0`.
- Let merge-driven prereleases use generated tags like `pre-YYYYMMDDHHMMSS-<sha7>`.
- Reserve manual semver tags like `v0.10.x` for a future stable-release process if needed.

### 1. New Feature or Breaking‑Change Release (Minor/Major)

1. **Merge & Verify**

- Merge all feature or breaking‑change PRs into `master`.
- Ensure CI (tests, linter, codegen) all pass on `master`.

1. **Determine Version Bump**

- **Major** (`X.0.0`) when you make incompatible changes
- **Minor** (`0.Y.0`) when you add functionality in a backward compatible manner
- **Patch** (`0.0.Z`) when you make backward compatible bug fixes

1. **Merge to `master`**
   - Merging to `master` triggers the `Release` workflow automatically.

   Or, for a manual test without merging:
   - Run the `Release` workflow with `workflow_dispatch`.
   - Optionally provide a short `release_label` such as `test`.

2. **Monitor Release**
   - GitHub Actions will:
     - Run `go generate ./...`
     - Build per-platform binaries with `main.version` set from `main.go`
     - Create a generated prerelease tag like `pre-YYYYMMDDHHMMSS-<sha7>`
     - Archive `_datafiles` as `gomud-ALL-datafiles-pre-YYYYMMDDHHMMSS-<sha7>.zip`
     - Generate `gomud-pre-YYYYMMDDHHMMSS-<sha7>-SHA256SUMS.txt`
     - Publish a GitHub prerelease for that generated tag
     - Leave the release unmarked as `Latest`

3. **Announce**
   - After review, a repo owner can edit the release in GitHub and promote it to
     `Latest`.
   - Share the release link with the team or via configured notifications.

---

### 2. Merge-Driven Prerelease Policy

1. **Pull requests do not publish release binaries**
   - PRs should run normal CI only.

2. **Merges to `master` do publish release binaries**
   - A push to `master` runs the `Release` workflow and publishes a prerelease.

3. **Manual test runs can also publish prereleases**
   - `workflow_dispatch` can be used to create a test prerelease without merging.
   - An optional `release_label` is appended to the generated prerelease tag.

4. **Generated release naming**
   - The release tag is generated automatically from UTC time plus the merge commit SHA.
   - Example: `pre-20260417021530-1a2b3c4`
   - With a manual label: `pre-20260417021530-1a2b3c4-test`

---

### 3. Manual Test Release Flow

1. **Run the release workflow manually**
   - Use `workflow_dispatch` when you want a test release
     without merging to `master`.
   - Optionally set `release_label=test` or similar to make the generated tag clearer.

2. **Verify the GitHub release**
   - Confirm the workflow succeeds.
   - Confirm the per-platform binaries are attached.
   - Confirm the `_datafiles` zip asset is attached.
   - Confirm the checksum manifest asset is attached.
   - Confirm GitHub marks the release as a prerelease.
   - Confirm GitHub does not mark it as `Latest`.

3. **Clean up if needed**
   - Delete the test tag and release after validation if you do not want to keep them
     in repository history.

---

### FAQ / Guidelines

- **Does every merge to `master` trigger a release?**
  Yes - every push to `master` runs the release workflow and publishes a prerelease.

- **Is auto-tagging enabled?**
  Stable semver tags are not generated automatically. The merge-driven workflow creates
  its own prerelease tag from UTC time plus commit SHA.

- **Can I create a test release without merging to `master`?**
  Yes - run the `Release` workflow manually with `workflow_dispatch`. That keeps PR
  submissions clean while still allowing on-demand test releases.

- **Are workflow-created releases stable releases?**
  No - the workflow creates prereleases. A repo owner must manually promote a release
  to `Latest` in GitHub when it is approved.

- **What assets should a release include?**
  Each release should include separate per-platform binaries, a `_datafiles` zip,
  and a checksum manifest so testers can download only what they need and still
  verify the assets.

- **What tag format should we use going forward?**
  Keep the source/binary version on the `0.10.x` line. Let the workflow generate
  prerelease tags like `pre-YYYYMMDDHHMMSS-<sha7>` on merges to `master`.

- **When should I bump minor vs. patch?**
  - **Minor** for new, backward‑compatible features.
  - **Patch** for bug fixes or documentation tweaks.

- **What about `go generate` directives?**
  The workflow runs `go generate ./...` automatically before each build.
