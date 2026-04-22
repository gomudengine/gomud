# Admin HTML Directory Context

## Overview

The `_datafiles/html/admin/` directory contains all front-end assets for the GoMud admin interface. The path is configured via `FilePaths.AdminHtml` in `config.yaml` and defaults to `_datafiles/html/admin`. Files are served by `internal/web`:

- HTML pages are rendered server-side via Go's `text/template` engine by the handlers in `internal/web/admin.go` and `internal/web/admin_config.go`.
- Non-HTML files (`.js`, images, etc.) are served as static assets via `serveAdminStaticFile` at `GET /admin/{file}` (single-segment) and `GET /admin/static/{path...}` (multi-segment), authenticated but not mud-locked. JavaScript files live in `_datafiles/html/admin/static/js/`.

No external CDN dependencies. No Bootstrap, jQuery, or HTMX.

## Files

### `_header.html`

Defines the Go template named `header`. Included by every admin page via `{{template "header" .}}`.

Contents:
- Minimal HTML5 shell (`<!DOCTYPE html>`, `<head>`, `<body>`)
- Inline CSS reset and base styles (system font stack, nav bar, `<main>` container)
- Top navigation bar with links to **Dashboard** (`/admin/`) and **Config** (`/admin/config`)
- `<script>` tags loading shared JS libraries from `/admin/static/js/` — `api.js`, `ansi-colors.js`, `select-dialog.js`, `script-editor.js`, and `highlight.js`

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

### `static/js/api.js`

Served as a static asset at `/admin/static/js/api.js` (loaded by `_header.html`). Provides the `AdminAPI` global object — a thin, promise-based HTTP client for admin pages.

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

### `static/js/ansi-colors.js`

Served as a static asset at `/admin/static/js/ansi-colors.js` (loaded by `_header.html` after `api.js`). Provides the `AnsiColors` global object for converting ANSI 256-color codes to CSS hex values and rendering color previews.

#### `AnsiColors` API

```js
AnsiColors.toHex(n)                    // ANSI code (0–255) → CSS hex string e.g. '#ff0000'
AnsiColors.swatchHtml(colors, max?)    // array of ANSI codes → HTML string of colored <span>s
AnsiColors.previewTextHtml(text, colors)  // colorize a text string using bounce pattern → HTML
```

- `toHex(n)` handles all 256 codes: 0–15 (standard), 16–231 (6×6×6 cube), 232–255 (grayscale).
- `swatchHtml(colors, max?)` returns `<span style="background:#hex"></span>` for each code, optionally limited to `max` entries.
- `previewTextHtml(text, colors)` applies the default bounce pattern (per-character, reversing at ends) and returns HTML with inline `color:` styles.

### `static/js/select-dialog.js`

Served as a static asset at `/admin/static/js/select-dialog.js` (loaded by `_header.html`). Provides the `SelectDialog` global object — a reusable search-and-select modal dialog for admin pages.

#### `SelectDialog` API

```js
SelectDialog.open({
  title: 'Dialog title',          // string shown in modal header
  apiPath: '/admin/api/v1/...',   // GET endpoint to fetch data from
  transform: (data) => [...],     // maps API .data into [{label, value}, ...]
  onSelect: (value) => { ... },   // called with selected value (string)
                                  // or array of values if multi:true
  multi: false,                   // optional, enables multi-select with checkboxes
});

SelectDialog.close();             // close programmatically
```

`transform(data)` receives the parsed `APIResult.data` field and must return an array of `{label: string, value: string}` objects.

In **single-select** mode (default), clicking an item calls `onSelect(value)` and closes the dialog immediately.

In **multi-select** mode, items show checkboxes and a "Select" button appears in the footer. Clicking "Select" calls `onSelect([...values])`.

A search/filter input is always shown. The dialog closes on Cancel, the × button, or clicking the overlay backdrop.

#### Common usage: inject a color pattern name into a text input

```js
SelectDialog.open({
  title: 'Pick a Color Pattern',
  apiPath: '/admin/api/v1/colorpatterns',
  transform: data => Object.keys(data).sort().map(k => ({ label: k, value: k })),
  onSelect: value => {
    myInput.value += ':' + value;  // inject as :patternname
  },
});
```

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
| `static/js/*.js` | `internal/web/web.go` `serveAdminStaticFile` | `GET /admin/static/{path...}` |
| `_header.html`, `_footer.html` | included by both page handlers | — |

## API Documentation Page Specification

Every `-api.html` page documents the REST endpoints for its resource. New API doc pages must follow this structure so the admin interface stays consistent.

### Page structure

1. **Title**: `<h1>` with the resource name + "API Reference" (e.g. "Items API Reference").
2. **Subtitle**: `<p class="subtitle">` — one sentence describing what the endpoints manage.
3. **Endpoint list**: a `<div class="api-list">` containing one `<details class="api-entry">` per endpoint.

### Per-endpoint entry

Each `<details class="api-entry">` contains:

1. **Summary line** (always visible):
   - **Method badge**: `<span class="method method-{method}">` — one of `method-get`, `method-post`, `method-patch`, `method-put`, `method-delete`.
   - **Path**: `<span class="api-path">` — the full route (e.g. `/admin/api/v1/items/{itemId}`).
   - **Description**: `<span class="api-desc">` — short phrase describing the action.

2. **Body** (inside `<div class="api-body">`):
   - **Description paragraph**: `<p>` explaining what the endpoint does, any special parameters, and behavior notes.
   - **curl example**: a `<div class="curl-block">` showing a complete curl command. Use `{{.CONFIG.FilePaths.WebDomain}}` for the host. Color spans: `.kw` for the command name, `.flag` for flags, `.str` for string values.
   - **Response examples** (required): a `<div class="response-examples">` containing one or more `<details class="resp-entry">` blocks:
     - Each has a **summary** with a status badge (`status-ok` or `status-err`) and a label (e.g. "Success", "Not Found").
     - Each has a **body** with a `<div class="curl-block">` showing an example JSON response.

### Required response examples

Every endpoint must include at minimum:

| Method | Required responses |
|---|---|
| GET (collection) | 200 Success |
| GET (single) | 200 Success, 404 Not Found |
| POST | 200/201 Success, 400 Bad Request |
| PATCH | 200 Success, 400 Bad Request, 404 Not Found |
| PUT | 200 Success, 400 Bad Request, 404 Not Found |
| DELETE | 200 Success, 404 Not Found |

Additional error codes (409 Conflict, 500 Internal Server Error) should be documented when the endpoint can produce them.

### CSS classes

All `-api.html` pages share the same `<style>` block. The required classes are:

| Class | Purpose |
|---|---|
| `.api-list` | Flex column container for all entries |
| `.api-entry` | Collapsible `<details>` wrapper |
| `.method` | Base style for HTTP method badges |
| `.method-get`, `.method-post`, `.method-patch`, `.method-put`, `.method-delete` | Color variants |
| `.api-path` | Monospace route path |
| `.api-desc` | Right-aligned description text |
| `.api-body` | Expanded content area |
| `.curl-block` | Dark code block for curl commands and JSON |
| `.response-examples` | Container for response entries |
| `.resp-entry` | Collapsible response example |
| `.status-ok`, `.status-err` | Green/red status badges |
