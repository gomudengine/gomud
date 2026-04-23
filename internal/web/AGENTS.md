# Web Package Guide

## Scope

- Use this file for `internal/web`, including public routes, admin pages, admin APIs, auth, internal requests, and HTTPS/admin helpers.
- This package sits on the boundary between HTTP behavior, config mutation, and module-contributed web surfaces.

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

## Documentation

- Keep this file focused on handler and routing guardrails, not a full route catalog.
- Put deeper protocol or UI details in narrower docs when they are useful.
