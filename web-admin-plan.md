# Web Admin Refactor Plan

## Overview

This plan covers three distinct changes to `internal/web` and `_datafiles/html/admin/`:

1. Strip the existing admin UI down to a single index page (header/footer paradigm preserved).
2. Reorganize the Go source so routing lives in dedicated files.
3. Add a versioned REST API at `/admin/api/v1/` with a `config` resource as the first endpoint.

---

## Part 1 - Remove Existing Admin Pages

### HTML/CSS/JS to delete

All files under `_datafiles/html/admin/` except the three that will be kept or replaced:

| Path | Action |
|---|---|
| `_datafiles/html/admin/items/` (entire directory) | Delete |
| `_datafiles/html/admin/mobs/` (entire directory) | Delete |
| `_datafiles/html/admin/mutators/` (entire directory) | Delete |
| `_datafiles/html/admin/races/` (entire directory) | Delete |
| `_datafiles/html/admin/rooms/` (entire directory) | Delete |
| `_datafiles/html/admin/static/js/htmx.2.0.3.js` | Delete |
| `_datafiles/html/admin/static/js/scripts.js` | Delete |
| `_datafiles/html/admin/static/css/styles.css` | Delete |
| `_datafiles/html/admin/static/images/gopher-dance.gif` | Delete |
| `_datafiles/html/admin/_header.html` | Replace (simplified, no Bootstrap CDN, no sidebar) |
| `_datafiles/html/admin/_footer.html` | Replace (simplified) |
| `_datafiles/html/admin/index.html` | Replace (new minimal dashboard) |

The `_datafiles/html/admin/static/` directory itself will be empty after the deletions above and can be removed as well. The `/admin/static/` HTTP route will be removed from Go routing.

### New admin HTML

Three files remain, all using the `{{define "header"}}` / `{{define "footer"}}` template pattern already established:

**`_header.html`** - Minimal semantic HTML5 shell. No external CDN dependencies. Inline a small amount of CSS sufficient for a readable dashboard (a top nav bar with the server name and a "Dashboard" link, plus basic typography). No jQuery, no Bootstrap, no HTMX.

**`_footer.html`** - Closing tags, optional small script block.

**`index.html`** - Calls `{{template "header" .}}`, renders a simple dashboard showing:
- Server name and version (from `{{.CONFIG}}`)
- Online player count (from `{{.STATS}}`)
- Telnet ports and websocket port
- A section listing available API endpoints (static informational text)
- Calls `{{template "footer" .}}`

The template data passed to the index handler will use the same `map[string]any` shape as today (`CONFIG`, `STATS`), so the existing `adminIndex` handler logic is largely preserved - it just needs to point at the new template files.

---

## Part 2 - Go Source Reorganization

### Files to delete

| File | Reason |
|---|---|
| `internal/web/admin.items.go` | Handlers removed |
| `internal/web/admin.mobs.go` | Handlers removed |
| `internal/web/admin.mutators.go` | Handlers removed |
| `internal/web/admin.races.go` | Handlers removed |
| `internal/web/admin.rooms.go` | Handlers removed |

### Files to keep (unchanged)

- `internal/web/auth.go` - `doBasicAuth`, `handlerToHandlerFunc`, auth cache. No changes needed.
- `internal/web/stats.go` - `Stats`, `GetStats`, `UpdateStats`. No changes needed.
- `internal/web/template_func.go` - `funcMap`. No changes needed.
- `internal/web/web.go` - Server startup, `serveTemplate`, `RunWithMUDLocked`, `Shutdown`. The routing block inside `Listen()` will be reduced (all the deleted page routes are removed) and two new routing calls are added (see below). Everything else stays.

### Files to create

**`internal/web/admin.go`** (replaces the current minimal file)

Owns the single admin page handler:

```go
func adminIndex(w http.ResponseWriter, r *http.Request)
```

Loads `_header.html`, `index.html`, `_footer.html` from `AdminHtml` path, passes `CONFIG` and `STATS` template data. Identical in shape to the current `adminIndex` but pointing at the new simplified templates.

**`internal/web/admin_routes.go`** (new)

Contains a single exported function called from `Listen()`:

```go
func registerAdminRoutes(mux *http.ServeMux)
```

This function registers all `/admin/` routes in one place:

```
GET /admin/              -> RunWithMUDLocked(doBasicAuth(adminIndex))
GET /admin/static/       -> RunWithMUDLocked(doBasicAuth(...FileServer...))  [only if static dir exists]
```

The API routes are also registered here by delegating to `registerAPIRoutes` (see Part 3).

**`internal/web/api_routes.go`** (new)

Contains:

```go
func registerAPIRoutes(mux *http.ServeMux)
```

Registers all `/admin/api/` routes. Initially:

```
GET   /admin/api/v1/config  -> RunWithMUDLocked(doBasicAuth(apiV1GetConfig))
PATCH /admin/api/v1/config  -> RunWithMUDLocked(doBasicAuth(apiV1PatchConfig))
```

**`internal/web/api.go`** (new)

Owns shared API infrastructure:

- `APIResponse[T]` - the generic envelope used by every API endpoint.
- `writeJSON` - helper that sets `Content-Type: application/json` and encodes a response.
- `writeAPIError` - helper that writes an `APIResponse` with an error payload and the given HTTP status code.

**`internal/web/api_v1_config.go`** (new)

Owns the config resource handlers:

- `apiV1GetConfig(w, r)` - GET handler.
- `apiV1PatchConfig(w, r)` - PATCH handler.

---

## Part 3 - REST API Design

### Routing integration

`Listen()` in `web.go` currently registers routes directly against the default `http.ServeMux`. The call site will change to:

```go
registerAdminRoutes(http.DefaultServeMux)
```

This keeps `web.go` clean and makes future route additions a matter of editing only the relevant `*_routes.go` file.

### Authentication

All `/admin/` paths, including `/admin/api/`, continue to use the existing `doBasicAuth` middleware unchanged. No new auth mechanism is introduced.

### Response envelope

Every API response, success or error, uses the same JSON structure:

```go
type APIResponse[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

- On success: `success: true`, `data` populated, `error` omitted.
- On error: `success: false`, `data` omitted, `error` is a human-readable message.

This is a generic type so handlers can be strongly typed without casting at the call site.

### `GET /admin/api/v1/config`

Returns the full current configuration as a flat key/value map.

**Request:** No body. No query parameters.

**Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "Server.MudName": "GoMud",
    "Network.HttpPort": "80",
    ...
  }
}
```

Implementation calls `configs.GetConfig().AllConfigData()` which already returns a `map[string]any` of all config values keyed by dot-path. Secrets are automatically redacted by `ConfigSecret.String()`.

### `PATCH /admin/api/v1/config`

Updates one or more configuration values.

**Request body (JSON):**
```json
{
  "Server.MudName": "My MUD",
  "Network.HttpPort": "8080"
}
```

The body is a flat `map[string]string` - keys are dot-path config names, values are always strings (matching what `configs.SetVal` expects).

**Processing:**
- Decode the JSON body. Return `400 Bad Request` with an error message if the body is malformed.
- For each key/value pair, call `configs.SetVal(key, value)`.
- If `SetVal` returns `ErrLockedConfig`, skip that key and record it in a `rejected` list returned to the caller.
- If `SetVal` returns `ErrInvalidConfigName`, treat as a `400 Bad Request` for that field.
- Any other error from `SetVal` is treated as a `500 Internal Server Error`.
- If all updates succeed (or are locked-skipped), return `200 OK`.

**Response `200 OK` (all applied):**
```json
{
  "success": true,
  "data": {
    "applied": ["Server.MudName"],
    "rejected": []
  }
}
```

**Response `200 OK` (some rejected due to lock):**
```json
{
  "success": true,
  "data": {
    "applied": ["Server.MudName"],
    "rejected": ["Server.Seed"]
  }
}
```

**Response `400 Bad Request` (invalid key or malformed body):**
```json
{
  "success": false,
  "error": "unknown config key: Server.Bogus"
}
```

**Response `500 Internal Server Error`:**
```json
{
  "success": false,
  "error": "internal error applying config"
}
```

Note: locked keys are not an error condition from the caller's perspective - the server simply cannot change them at runtime. Returning them in `rejected` gives the caller visibility without treating it as a failure.

---

## Part 4 - `web.go` Routing Changes

The `Listen()` function currently has an inline block of ~20 lines of `http.HandleFunc` calls for admin routes. After this change, that block is replaced with:

```go
registerAdminRoutes(http.DefaultServeMux)
```

The public routes (`/favicon.ico`, `/`, `/ws`) remain inline in `Listen()` as they are today - they are not admin routes and do not need to move.

The `/admin/static/` file server route is removed entirely (no static assets remain in the admin directory).

---

## File Change Summary

| File | Action |
|---|---|
| `internal/web/admin.go` | Rewrite - single `adminIndex` handler |
| `internal/web/admin.items.go` | Delete |
| `internal/web/admin.mobs.go` | Delete |
| `internal/web/admin.mutators.go` | Delete |
| `internal/web/admin.races.go` | Delete |
| `internal/web/admin.rooms.go` | Delete |
| `internal/web/admin_routes.go` | Create - `registerAdminRoutes` |
| `internal/web/api.go` | Create - `APIResponse`, `writeJSON`, `writeAPIError` |
| `internal/web/api_routes.go` | Create - `registerAPIRoutes` |
| `internal/web/api_v1_config.go` | Create - GET/PATCH config handlers |
| `internal/web/web.go` | Edit - replace inline admin route block with `registerAdminRoutes(...)` |
| `internal/web/auth.go` | No change |
| `internal/web/stats.go` | No change |
| `internal/web/template_func.go` | No change |
| `_datafiles/html/admin/_header.html` | Replace - simplified, no CDN deps |
| `_datafiles/html/admin/_footer.html` | Replace - simplified |
| `_datafiles/html/admin/index.html` | Replace - minimal dashboard |
| `_datafiles/html/admin/items/` | Delete (directory) |
| `_datafiles/html/admin/mobs/` | Delete (directory) |
| `_datafiles/html/admin/mutators/` | Delete (directory) |
| `_datafiles/html/admin/races/` | Delete (directory) |
| `_datafiles/html/admin/rooms/` | Delete (directory) |
| `_datafiles/html/admin/static/` | Delete (directory) |
| `internal/web/AGENTS.md` | Update to reflect new structure |

---

## Notes and Constraints

- The `AdminHtml` config path (`FilePaths.AdminHtml`) is unchanged. The new templates live in the same location, just with fewer files.
- The `RunWithMUDLocked` wrapper is kept on all `/admin/` routes including API routes. Config reads and writes touch shared state and must be mutex-protected.
- `configs.SetVal` already handles persistence (writing to the override file) and validation internally. The API handler does not need to call any additional save function.
- The `APIResponse` type uses Go generics (`[T any]`), requiring Go 1.18+. The project already requires Go 1.24, so this is fine.
- No new external dependencies are introduced. The API uses only `encoding/json` from the standard library.
- The AGENTS.md for `internal/web` will need to be updated to document the new file layout and API structure.
