## Release Process

[Semantic Versioning 2.0.0 reference](https://github.com/semver/semver/blob/master/semver.md)

This downstream fork is the proving ground for release automation changes before they
are proposed upstream.

### 1. New Feature or Breaking‑Change Release (Minor/Major)

1. **Merge & Verify**
- Merge all feature or breaking‑change PRs into `master`.
- Ensure CI (tests, linter, codegen) all pass on `master`.

2. **Determine Version Bump**
- **Major** (`X.0.0`) when you make incompatible changes
- **Minor** (`0.Y.0`) when you add functionality in a backward compatible manner
- **Patch** (`0.0.Z`) when you make backward compatible bug fixes

3. **Create Git Tag**
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
   This triggers the `Release` workflow.

4. **Monitor Release**
   - GitHub Actions will:
     - Run `go generate ./...`
     - Build artifacts with `main.version=vX.Y.Z`
     - Zip as `go-mud-release-vX.Y.Z.zip`
     - Publish a GitHub prerelease for `vX.Y.Z`
     - Leave the release unmarked as `Latest`

5. **Announce**
   - After review, the upstream owner can edit the release in GitHub and promote it
     to `Latest`.
   - Share the release link with the team or via configured notifications.

---

### 2. Basic Patch Release (x.y.Z)

1. **Merge Bug‑Fix PR**
   - Once the fix is in `master` and CI is green.

2. **Determine Patch Bump**
   ```bash
   # if current version is vX.Y.Z:
   git tag vX.Y.(Z+1)
   git push origin vX.Y.(Z+1)
   ```

3. **Tag & Push**
   - Pushing the tag triggers the same release workflow.

4. **Publish**
   - The workflow publishes the release automatically as a prerelease after the build
     completes.

---

### 3. Downstream First-Test Flow

1. **Validate in this fork first**
   - Use this downstream repo to verify any release automation change before opening
     an upstream PR.

2. **Push a disposable test tag**
   ```bash
   git tag vX.Y.Z-test.1
   git push origin vX.Y.Z-test.1
   ```

3. **Verify the GitHub release**
   - Confirm the workflow succeeds.
   - Confirm the zip asset is attached.
   - Confirm GitHub marks the release as a prerelease.
   - Confirm GitHub does not mark it as `Latest`.

4. **Clean up if needed**
   - Delete the test tag and release after validation if you do not want to keep them
     in downstream history.

---

### FAQ / Guidelines

- **Does every merge to `master` trigger a release?**
  No - only pushing a Git tag matching `v*.*.*` triggers a release.

- **Is auto-tagging enabled?**
  No - releases are manual. Create and push the version tag yourself when you want to publish.

- **Are workflow-created releases stable releases?**
  No - the workflow creates prereleases. A repo owner must manually promote a release
  to `Latest` in GitHub when it is approved.

- **When should I bump minor vs. patch?**
  - **Minor** for new, backward‑compatible features.
  - **Patch** for bug fixes or documentation tweaks.

- **What about `go generate` directives?**
  The workflow runs `go generate ./...` automatically before each build.
