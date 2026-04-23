# Storage Module Guide

## Scope

- Use this file for player storage persistence, storage-tag behavior, admin storage APIs, and storage-related autocomplete hooks in `modules/storage`.
- This module touches user persistence, items, room tags, and admin web behavior.

## Working Rules

- Preserve the `storage` room-tag contract and the module-owned storage command unless the task explicitly changes that ownership model.
- Be careful with legacy storage migration and plugin-data persistence. Data movement bugs here can silently drop player items.
- Keep room-tag behavior, command behavior, admin API behavior, and exported helper behavior aligned.
- When changing storage item access, inspect autocomplete and admin paths too.

## Verification

- Run targeted module tests for storage persistence or migration behavior.
- If the change affects admin storage pages or APIs, verify the exact route and deletion/listing behavior.
- Explicitly verify item movement semantics when changing add/remove/all flows.

## Documentation

- Keep this file about persistence and room-tag guardrails, not module history.
