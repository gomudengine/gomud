# Admin HTML Directory Context

## Overview

The `_datafiles/html/admin/` directory contains all front-end assets for the GoMud admin interface. The path is configured via `FilePaths.AdminHtml` in `config.yaml` and defaults to `_datafiles/html/admin`. Files are served by `internal/web`:

- HTML pages are rendered server-side via Go's `text/template` engine by the handlers in `internal/web/admin.go` and `internal/web/admin_config.go`.
- Non-HTML files (`.js`, images, etc.) are served as static assets via `serveAdminStaticFile` at `GET /admin/{file}`, authenticated but not mud-locked.

No external CDN dependencies. No Bootstrap, jQuery, or HTMX.

## Files

### `_header.html`

Defines the Go template named `header`. Included by every admin page via `{{template "header" .}}`.

Contents:
- Minimal HTML5 shell (`<!DOCTYPE html>`, `<head>`, `<body>`)
- Inline CSS reset and base styles (system font stack, nav bar, `<main>` container)
- Top navigation bar with links to **Dashboard** (`/admin/`) and **Config** (`/admin/config`)
- `<script src="/admin/api.js"></script>` — loads the `AdminAPI` client library before page scripts run

Template data used: none (the header is layout-only).

### `_footer.html`

Defines the Go template named `footer`. Included by every admin page via `{{template "footer" .}}`.

Contents:
- Closing `</main>`, `</body>`, `</html>` tags

### `index.html`

Admin dashboard page. Rendered by `adminIndex` at `GET /admin/`.

Template data:
- `.CONFIG` — `configs.Config` struct
- `.STATS` — `web.Stats` struct

Sections:
- **Header**: MUD name (`CONFIG.Server.MudName`) and version (`CONFIG.Server.CurrentVersion`)
- **Stats grid** (3 cards): Telnet ports with connection count, WebSocket port with connection count, SSH port with connection count — all sourced from `STATS`
- **REST API reference**: Expandable `<details>` entries for each API endpoint, showing method badge, path, description, curl examples, and collapsible response examples. The PATCH entry includes a **Test Mode** checkbox that toggles the `X-Test-Mode: true` curl flag via inline JavaScript (`toggleTestMode`)

Inline JavaScript: `toggleTestMode(blockId, enabled)` — shows or hides `.test-mode-flag` spans within a named curl block.

### `config.html`

Live configuration editor page. Rendered by `adminConfig` at `GET /admin/config`.

Template data:
- `.CONFIG` — `configs.Config` struct (used only for the page subtitle; the table is populated via API)
- `.STATS` — `web.Stats` struct

Behaviour:
1. On page load, calls `AdminAPI.get('/admin/api/v1/config')` and builds a sortable, section-grouped table from the returned flat key/value map.
2. **Inline editing**: clicking a non-redacted value hides the display span and shows an `<input>` with Save/Cancel buttons. Pressing Enter stages the edit; Escape cancels.
3. **Staging**: edits are accumulated in a `pending` map and highlighted in the table row (yellow background). A **Pending Changes** panel appears listing all staged changes with individual remove buttons.
4. **Apply**: `applyChanges()` calls `AdminAPI.patch('/admin/api/v1/config', pending)`. Applied keys are committed to the local `configData` snapshot; rejected (locked) keys remain staged and are reported in a status banner.
5. **Discard**: `discardAll()` reverts all staged changes and restores display values from the local snapshot.
6. **Filtering**: a search input filters rows by key or value substring; a section dropdown filters by config section prefix (e.g. `Server`, `Network`). Section header rows are hidden when all their children are filtered out.
7. **Locked keys**: displayed with a `locked` badge and are not clickable (no inline edit).
8. **Redacted secrets**: displayed in italic grey with no click handler.

JavaScript functions (all `window`-scoped for template event-handler compatibility):
- `filterTable()` — applies search and section filters
- `startEdit(key)` / `cancelEdit(key)` / `editKeydown(event, key)` / `stageEdit(key)` — inline edit lifecycle
- `removePending(key)` / `discardAll()` — pending change management
- `applyChanges()` — submits staged changes via `AdminAPI`

Utility helpers (module-private):
- `escHtml(s)` / `escAttr(s)` — HTML and attribute escaping
- `cssId(key)` — converts a dot-path config key to a safe CSS id fragment (e.g. `Server.MudName` → `Server_MudName`)

### `api.js`

Served as a static asset at `/admin/api.js` (loaded by `_header.html`). Provides the `AdminAPI` global object — a thin, promise-based HTTP client for admin pages.

#### `AdminAPI` API

```js
// Single requests — return Promise<APIResult>
AdminAPI.get(path)
AdminAPI.post(path, body)
AdminAPI.put(path, body)
AdminAPI.patch(path, body)
AdminAPI.delete(path, body?)

// Parallel dispatch — resolves when all settle
AdminAPI.all([promise, ...])

// Fluent queue builder — dispatches all in parallel on .run()
AdminAPI.queue()
  .get(path)
  .patch(path, body)
  .run()  // returns Promise<APIResult[]>
```

#### `APIResult` shape

```js
{
  ok: boolean,     // true when HTTP status is 2xx
  status: number,  // HTTP status code
  data: any,       // parsed JSON body (null on parse failure)
  error: string,   // error message, or empty string on success
}
```

Implementation details:
- Uses `fetch` with `credentials: 'same-origin'` and `Content-Type: application/json`.
- Never rejects — network errors are caught and returned as `APIResult` with `ok: false`.
- Response body is parsed as JSON; falls back to raw text on parse failure.
- `AdminAPI.all` delegates to `Promise.all` (already non-rejecting because each request handles its own errors).
- `AdminAPI.queue().run()` splices the internal pending array so the same queue instance can be reused.

## Template Rendering Flow

Both admin pages follow the same pattern:

```
handler parses:
  _header.html  (defines "header")
  <page>.html   (defines page body, calls {{template "header" .}} and {{template "footer" .}})
  _footer.html  (defines "footer")

template.New("<page>.html").Funcs(funcMap).ParseFiles(...).Execute(w, templateData)
```

`Cache-Control: no-store` is set on every response to prevent stale admin pages.

## Relationship to Go Source

| HTML file | Go handler | Route |
|---|---|---|
| `index.html` | `internal/web/admin.go` `adminIndex` | `GET /admin/` |
| `config.html` | `internal/web/admin_config.go` `adminConfig` | `GET /admin/config` |
| `api.js` | `internal/web/web.go` `serveAdminStaticFile` | `GET /admin/api.js` |
| `_header.html`, `_footer.html` | included by both page handlers | — |
