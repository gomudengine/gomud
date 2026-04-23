# Provisioning Guide

## Scope

- Use this file for Docker build files and helper scripts under `provisioning/`.
- These files define local container packaging and helper runtime images, not core game logic.

## Working Rules

- Keep the main server image aligned with `compose.yml` and `Makefile` entrypoints.
- Preserve the non-root runtime model unless the task explicitly requires a capability or ownership change.
- Keep the server image focused on the compiled binary plus bundled `_datafiles`; avoid pulling unrelated tooling into the runtime stage.
- Treat `provisioning/terminal/` as a helper client image for local development, not a production service.

## Verification

- For main image changes, prefer validating through `make run-docker` or the narrowest equivalent Docker build path.
- For terminal image changes, validate the affected Docker build or compose path.
- If you change shell helpers here, run `shellcheck` and `shfmt` if available.
