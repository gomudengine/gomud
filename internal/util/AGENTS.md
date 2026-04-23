# Util Package Guide

## Scope

- Use this file for shared helpers in `internal/util`, including matching helpers, hashing helpers, turn/round utilities, copyover helpers, and generic hooks.
- This package is widely imported. Small behavior changes can have large blast radius.

## Working Rules

- Prefer changing the smallest helper that solves the task. Do not refactor unrelated utilities while touching a widely shared function.
- Preserve user-facing semantics for generic helpers such as matching, parsing, and formatting unless the task explicitly changes them.
- Be especially careful with helpers used by persistence, auth, command parsing, or timing-sensitive code.
- Keep generic hook behavior simple and predictable. If a change affects hook ordering or value propagation, treat it as a compatibility-sensitive change.
- Avoid adding game-specific policy to `internal/util` when it belongs in a narrower package.

## Verification

- Run targeted `internal/util` tests for any changed helper.
- If the helper is used broadly and the change is behaviorally significant, run at least one higher-level check that exercises a real caller.
- Call out callers you did not retest when a utility change has broad reach.

## Documentation

- Keep this file about blast-radius and compatibility rules.
- Do not turn it back into a catalog of unrelated helper functions.
