# Users Package Guide

## Scope

- Use this file for account records, indexing, storage, search, authentication helpers, copyover support, and user persistence in `internal/users`.
- This package carries user data and login-critical behavior. Favor compatibility and safety over cleanup.

## Working Rules

- Preserve on-disk compatibility for user records and user index behavior unless the task explicitly includes a migration.
- Be careful with password and login changes. Existing authentication, legacy-password migration, and reconnect/copyover flows are easy to break indirectly.
- Keep user search and indexing changes aligned with the existing storage/index flow rather than introducing a second lookup path.
- If a change affects item storage, online info, or account migration behavior, review the corresponding persistence code together.
- Avoid mixing unrelated gameplay refactors into user-data changes.

## Verification

- Run targeted `internal/users` tests for indexing, password, migration, or persistence changes.
- If the change affects login or reconnect behavior, verify the exact code path changed and note any live flows not exercised.
- Use broader repo checks when the user-package change crosses into web auth, copyover, or item storage.

## Documentation

- Keep this file focused on persistence, auth, and compatibility guardrails.
- Put deeper subsystem walkthroughs elsewhere if they are still needed.
