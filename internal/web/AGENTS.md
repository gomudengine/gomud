# Web Package Guide

## Scope

- Use this file for `internal/web`, including public routes, admin pages, admin APIs, auth, internal requests, and HTTPS/admin helpers.
- This package sits on the boundary between HTTP behavior, config mutation, and module-contributed web surfaces.

## Permission System

The admin area uses a two-role model:

- `admin` — always has full access to everything. No permission checks apply.
- `mod` — can access the admin panel and in-game admin commands only for the permissions stored in `UserRecord.Permissions`.

### How permissions are enforced

**Web pages** — all admin HTML pages are accessible to any authenticated mod/admin. No per-page read permission is required.

**API GET endpoints** — all GET (read) API endpoints are accessible to any authenticated mod/admin. No per-endpoint read permission is required.

**API mutating endpoints** — POST, PATCH, PUT, and DELETE methods require a `*.write` permission checked via `RequirePermission` in `api_routes.go`. A 403 returns JSON `{"success":false,"error":"forbidden: requires permission <key>"}` with the specific key named.

**In-game commands** — `TryCommand` in `internal/usercommands/usercommands.go` checks `user.HasRolePermission(cmd, true)` before executing any `AdminOnly` command. A mod without permission receives `You don't have permission to use <command>.`

### Permission key conventions

- Web write keys use `<resource>.write` dot notation: `mobs.write`, `rooms.write`.
- In-game command keys match the command name: `teleport`, `mob.create`, `room.edit.exits`.
- Prefix-match is supported: granting `room` implicitly grants `room.edit`, `room.edit.exits`, etc.
- There are no `*.read` permission keys. Read access is unconditional for any authenticated mod/admin.

### Adding a new admin route

1. Register GET page and GET API routes in `admin_routes.go` / `api_routes.go` with just `doBasicAuth(RunWithMUDLocked(handler))` — no permission wrapper needed for reads.
2. Register mutating API methods with `doBasicAuth(RequirePermission("resource.write", RunWithMUDLocked(handler)))`.
3. Add the `resource.write` key to `allPermissions` in `api_v1_permissions.go` with the correct `Category`.
4. The key is automatically added to the validation set by the `init()` function.

### Permission catalog

The canonical catalog lives in `internal/web/api_v1_permissions.go` as `allPermissions []PermissionDef`. Each entry has:
- `Key` — the permission string checked at runtime.
- `Description` — shown in the admin UI picker.
- `Category` — groups entries in the picker (e.g. `Mobs`, `Rooms`, `Server`).

Only write and command permissions are in the catalog. There are no read permission keys.

The API endpoint `GET /admin/api/v1/permissions` returns the full catalog sorted by category then key.

## Working Rules

- Preserve the distinction between public routes, admin HTML routes, admin API routes, and static admin asset routes.
- Be careful with auth and mud-lock behavior:
  - admin HTML and admin API routes are generally auth-gated and mud-locked
  - admin static assets are auth-gated but not mud-locked
  - internal requests intentionally bypass auth and mud-lock wrappers
- Keep internal-request behavior explicit. If a handler must behave differently for in-process callers, use the existing internal-request helpers rather than inventing a parallel path.
- Preserve test-mode behavior for config-affecting API work. The existing test-mode middleware snapshots and restores overrides for dry-run style requests.
- When adding module web surfaces, follow the existing registrar pattern for admin pages and admin API endpoints instead of hard-coding module routes here.
- Keep route additions and auth-wrapper changes narrow. Small mistakes here can expose admin behavior or break module pages.

## Verification

- Run targeted package tests for `internal/web` when behavior changes here.
- If the change affects admin APIs or config patch behavior, verify the exact route and method involved.
- If the change affects redirects, HTTPS mode, or host/port handling, exercise the specific redirect or HTTPS path you changed.
- If package tests are not enough, note the manual route checks you performed.
