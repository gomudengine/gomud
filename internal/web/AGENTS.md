# GoMud Web Server and Admin Interface Context

## Overview

The GoMud web system provides a comprehensive HTTP/HTTPS server with both public web client functionality and a secure administrative interface. It supports WebSocket connections for real-time game clients, template-based HTML rendering, plugin integration, a versioned REST API for remote server management, and an in-process internal request dispatcher that allows core engine code to call API endpoints without network I/O or authentication overhead.

## Architecture

The web system is built around Go's standard `net/http` package with several key components:

### Core Components

**HTTP Server Management:**
- Dual HTTP/HTTPS server support with configurable ports
- TLS certificate validation and configuration
- Automatic HTTPS redirect capability
- WebSocket upgrade handling for real-time clients
- Graceful shutdown with timeout management
- Single `internalMux` (`*http.ServeMux`) shared by both live servers and the internal dispatcher
- Static asset serving for admin non-HTML files via `serveAdminStaticFile`

**Template System:**
- Go `text/template` based HTML rendering
- Automatic inclusion of `_*.html` template files
- Plugin template override capability
- Custom template functions for formatting and logic
- Dynamic navigation menu generation

**Authentication and Security:**
- HTTP Basic Authentication for admin areas
- Role-based access control (admin/user roles)
- Authentication caching (30-minute sessions)
- Game state mutex locking for concurrent access protection
- Directory traversal protection
- Internal requests bypass auth and locking via context flag

**Plugin Integration:**
- `WebPlugin` interface for module web extensions
- Dynamic navigation link management
- Custom request handling and template data injection

**Internal Request Dispatcher:**
- `InternalRequest` / `InternalRequestJSON` allow in-process callers to dispatch requests through the real handler pipeline
- No network I/O, no authentication required
- Auth and mud-lock wrappers detect internal requests and short-circuit automatically
- Handlers can inspect `IsInternalRequest(r)` to adjust behavior (e.g. skip audit logging)

**Test Mode:**
- `RunInTestMode` middleware wraps handlers so that requests carrying `X-Test-Mode: true` snapshot config overrides before the handler runs and restore them unconditionally afterwards
- Response carries `X-Test-Mode: true` header to confirm the mode was active
- `IsTestModeRequest(r)` lets handlers detect test-mode calls via context

## Go Source Files

| File | Purpose |
|---|---|
| `web.go` | Server startup, `internalMux`, nav types, `ModuleAdminRegistrar` impl, `buildAdminNav`, `GetAdminRegistrar`, `serveTemplate`, `serveAdminStaticFile`, `RunWithMUDLocked`, `Shutdown`, public route registration |
| `admin.go` | `adminIndex` handler - admin dashboard page |
| `admin_items.go` | `adminItems`, `adminItemsAPI`, `adminBuffs`, `adminBuffsAPI`, `adminQuests`, `adminQuestsAPI` handlers + `serveAdminTemplate` helper |
| `admin_stats.go` | `adminStatsAPI` handler — Stats API docs page |
| `admin_config_api.go` | `adminConfigAPI` handler - Config REST API docs page |
| `admin_routes.go` | `registerAdminRoutes(mux)` - registers all `/admin/` routes including static asset handler |
| `api.go` | `APIResponse[T]` generic envelope, `writeJSON`, `writeAPIError`, `RunInTestMode` middleware |
| `api_routes.go` | `registerAdminAPIRoutes(mux)` - registers all `/admin/api/` routes |
| `api_v1_config.go` | `apiV1GetConfig` and `apiV1PatchConfig` handlers |
| `api_v1_stats.go` | Memory stats endpoint (`/admin/api/v1/stats/memory`) |
| `api_v1_items.go` | Item CRUD + script endpoints (`/admin/api/v1/items/...`) |
| `api_v1_buffs.go` | Buff CRUD + script endpoints (`/admin/api/v1/buffs/...`) |
| `api_v1_quests.go` | Quest list / patch / delete endpoints (`/admin/api/v1/quests/...`) |
| `api_v1_users.go` | User search endpoint (`/admin/api/v1/users/...`) |
| `auth.go` | `doBasicAuth`, `handlerToHandlerFunc`, auth cache |
| `context.go` | `withInternalContext`, `IsInternalRequest`, `withTestModeContext`, `IsTestModeRequest` - request context flags |
| `internal.go` | `InternalRequest`, `InternalRequestJSON` - in-process API dispatcher |
| `stats.go` | `Stats`, `GetStats`, `UpdateStats` - server statistics with SSH/WebSocket/Telnet connection counts |
| `template_func.go` | `funcMap` - custom template functions |
| `web_test.go` | Unit tests for `buildHTTPSRedirectTarget` (IPv4, IPv6, port handling) |

## Routing Structure

All routes are registered on the package-level `internalMux`. Both live HTTP/HTTPS servers and `InternalRequest` use this same mux.

### Public Routes (registered inline in `Listen()`)
- `GET /favicon.ico` - favicon redirect
- `GET /` - public template server (`serveTemplate`)
- `GET /ws` - WebSocket upgrade endpoint

### Admin Routes (registered via `registerAdminRoutes`)
- `GET /admin/{file}` - static asset serving from admin HTML directory (auth required)
- `GET /admin/` - admin dashboard (auth required, mud-locked)
- `GET /admin/config` - live configuration editor (auth required, mud-locked)
- `GET /admin/config-api` - Config REST API docs page (auth required, mud-locked)
- `GET /admin/<slug>` - module-contributed admin pages, registered dynamically by `RegisterAdminPage`

### API Routes (registered via `registerAdminAPIRoutes`, called from `registerAdminRoutes`)
- `GET /admin/api/v1/config` - return all config as flat key/value map (auth required, mud-locked)
- `PATCH /admin/api/v1/config` - update one or more config values (auth required, mud-locked, test-mode aware)
- `GET /admin/api/v1/stats/memory` - return memory usage for all registered subsystems (auth required, mud-locked)
- `GET /admin/api/v1/items/types` - item types and subtypes
- `GET /admin/api/v1/items/attack-messages` - all weapon attack message groups
- `GET /admin/api/v1/items` - all item specs
- `POST /admin/api/v1/items` - create a new item spec
- `GET /admin/api/v1/items/{itemId}` - single item spec by id or name
- `PATCH /admin/api/v1/items/{itemId}` - update item spec properties
- `DELETE /admin/api/v1/items/{itemId}` - delete item spec and its script
- `GET /admin/api/v1/items/{itemId}/script` - item script contents
- `PUT /admin/api/v1/items/{itemId}/script` - replace (or delete) item script
- `GET /admin/api/v1/buffs` - all buff flags with descriptions
- `GET /admin/api/v1/buffs/{buffId}` - single buff spec
- `PATCH /admin/api/v1/buffs/{buffId}` - update buff spec properties
- `DELETE /admin/api/v1/buffs/{buffId}` - delete buff spec and its script
- `GET /admin/api/v1/buffs/{buffId}/script` - buff script contents
- `PUT /admin/api/v1/buffs/{buffId}/script` - replace (or delete) buff script
- `GET /admin/api/v1/quests` - all quests
- `PATCH /admin/api/v1/quests` - update a quest (questId in body)
- `DELETE /admin/api/v1/quests/{questId}` - delete a quest
- `GET /admin/api/v1/users/search?name={name}` - search users by username (exact or prefix match); returns `[]{ user_id, username, role, email }`
- `<METHOD> /admin/api/v1/<slug>` - module-contributed API endpoints, registered dynamically by `RegisterAdminAPIEndpoint`

All `/admin/` routes, including API routes, are wrapped with `RunWithMUDLocked` and `doBasicAuth`. Both wrappers short-circuit for internal requests.

## Internal Request Dispatcher

Core engine code can call any registered API endpoint in-process without credentials or network I/O.

### Functions

```go
// InternalRequest dispatches method+path through internalMux, bypassing auth
// and the mud lock. body may be nil. Returns raw response bytes.
func InternalRequest(method, path string, body io.Reader) (statusCode int, responseBody []byte, err error)

// InternalRequestJSON marshals reqBody as JSON, calls InternalRequest, and
// unmarshals the response into dst. Pass nil for either to skip that step.
func InternalRequestJSON(method, path string, reqBody any, dst any) (int, error)
```

### Usage example

```go
var result web.APIResponse[web.patchConfigResult]
status, err := web.InternalRequestJSON(
    http.MethodPatch,
    "/admin/api/v1/config",
    map[string]string{"GamePlay.PVP": "enabled"},
    &result,
)
```

### Locking contract

`RunWithMUDLocked` skips `util.LockMud()` for internal requests. The caller is responsible for holding the lock when calling from outside the game loop. Callers already inside the game loop (e.g. event handlers, hooks) hold the lock implicitly and need not acquire it again.

### Handler behavior for internal requests

Handlers can detect internal calls via `IsInternalRequest(r)` and adjust behavior accordingly (skip audit logging, rate limiting, etc.):

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    if !web.IsInternalRequest(r) {
        // log external call
    }
    // ... shared logic
}
```

## REST API

### Response Envelope

Every API response uses the same JSON structure:

```go
type APIResponse[T any] struct {
    Success  bool   `json:"success"`
    Data     T      `json:"data,omitempty"`
    Error    string `json:"error,omitempty"`
    TestMode bool   `json:"test_mode,omitempty"`
}
```

### `GET /admin/api/v1/config`

Returns the full current configuration as a flat dot-path key/value map. Secrets are automatically redacted by `ConfigSecret.String()`.

**Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "Server.MudName": "GoMud",
    "Network.HttpPort": "80"
  }
}
```

### `PATCH /admin/api/v1/config`

Updates one or more configuration values. Request body is a flat `map[string]string`.

- Locked keys are silently skipped and returned in `rejected`.
- Unknown keys return `400 Bad Request`.
- Malformed body returns `400 Bad Request`.
- Unexpected errors return `500 Internal Server Error`.
- Supports `X-Test-Mode: true` header: changes are applied then rolled back; response includes `"test_mode": true`.

**Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "applied": ["Server.MudName"],
    "rejected": ["Server.Seed"]
  }
}
```

## Stats Structure

```go
type Stats struct {
    OnlineUsers          []users.OnlineInfo
    TelnetPorts          []int
    WebSocketPort        int
    SSHPort              int
    TelnetConnections    int
    WebSocketConnections int
    SSHConnections       int
}
```

Thread-safe via `sync.RWMutex`. Updated by the main game loop via `UpdateStats`. Read by template handlers via `GetStats`.

## Admin HTML Templates

Located in `_datafiles/html/admin/` (path configured via `FilePaths.AdminHtml`). See `_datafiles/html/admin/AGENTS.md` for full details.

| File | Purpose |
|---|---|
| `_header.html` | Defines `{{define "header"}}` - HTML5 shell, inline CSS, dropdown nav bar driven by `.NAV`, loads `api.js` |
| `_footer.html` | Defines `{{define "footer"}}` - closing tags |
| `index.html` | Dashboard: server name, version, port stats, link card to API docs |
| `config.html` | Live config editor: inline editing, pending-changes panel, section filter, search |
| `config-api.html` | Config REST API reference (GET and PATCH `/admin/api/v1/config` with curl examples) |
| `items.html` | Items editor: searchable sidebar list, full item form with stat mods, buff chips, script editor |
| `items-api.html` | Items REST API reference |
| `buffs.html` | Buffs editor: searchable sidebar list, flag checkboxes, stat mods, script editor |
| `buffs-api.html` | Buffs REST API reference |
| `quests.html` | Quests editor: searchable sidebar list, step cards, reward fields |
| `quests-api.html` | Quests REST API reference |
| `users.html` | Users search page: username search with results table |
| `users-api.html` | Users REST API reference |
| `stats-api.html` | Stats API reference (`GET /admin/api/v1/stats/memory`) |
| `api.js` | `AdminAPI` JS client library served as a static asset at `/admin/api.js` |

Template data passed to all admin page handlers:
- `CONFIG` - `configs.Config` struct
- `STATS` - `web.Stats` struct
- `NAV` - `[]web.WebNavItem` from `buildAdminNav()`

## Admin Navigation

The admin nav is driven by `buildAdminNav()` which returns `[]WebNavItem`. Each handler passes `NAV: buildAdminNav()` in its template data. The nav supports two-level dropdowns (core items) and three-level flyout groups (module items).

```go
type WebNavItem struct {
    Name     string
    Target   string       // primary href; empty if dropdown-only
    SubItems []WebNavSub  // two-level dropdown items
    SubMenus []WebNavItem // three-level: group contains sub-menus, each with SubItems
}

type WebNavSub struct {
    Label  string
    Target string
}
```

Core nav entries (Dashboard, Config, Items, Buffs, Quests, Users, Colors, Races, Keywords) are hardcoded in `buildAdminNav()`. Module entries are appended from `defaultRegistrar.navItems`.

## Module Admin Page and API Registration

Modules register admin pages and API endpoints via `plugins.WebConfig` without importing `internal/web`:

```go
// In a module init():
// navGroup places the entry inside a top-level group dropdown (e.g. "Modules").
// navParent nests it as a sub-item under that parent within the group.
plugin.Web.AdminPage("View / Edit", "storage", "html/admin/storage.html", true, "Modules", "Storage",
    func(r *http.Request) map[string]any {
        return map[string]any{"ITEM_COUNT": getCount()}
    },
)

// Second sub-item under the same group+parent
plugin.Web.AdminPage("API Docs", "storage-api", "html/admin/storage-api.html", true, "Modules", "Storage",
    nil,
)

plugin.Web.AdminAPIEndpoint("GET", "storage", func(r *http.Request) (int, bool, any) {
    return http.StatusOK, true, getStats()
})
```

This produces a nav entry like:
```
Modules
  +-> Storage
        +-> View / Edit
        +-> API Docs
```

To place a page at the top level (no group), pass empty strings for both `navGroup` and `navParent`.

`main.go` wires the registrar before `plugins.Load()`:

```go
plugins.SetAdminRegistrar(web.GetAdminRegistrar())
plugins.Load(dataFilesPath)
```

This keeps the import graph acyclic: `web` does not import `plugins`; `plugins` does not import `web`; `main` wires them.

## Security

- All `/admin/` paths require HTTP Basic Authentication via `doBasicAuth`.
- Users must have a role other than `user` (i.e., admin or higher).
- Successful auth results are cached for 30 minutes.
- All admin handlers are wrapped with `RunWithMUDLocked` to serialize access to shared game state.
- Internal requests (via `InternalRequest`) bypass both auth and locking; they are identified by a context value set in `withInternalContext` and checked by `IsInternalRequest`.
- Static admin assets (`/admin/{file}`) are auth-gated but not mud-locked.

## Configuration

```yaml
network:
  http_port: 80
  https_port: 443
  https_redirect: true

file_paths:
  public_html: "_datafiles/html/public"
  admin_html: "_datafiles/html/admin"
  https_cert_file: "cert.pem"
  https_key_file: "key.pem"
```

## Dependencies

- `net/http` - HTTP server and routing
- `net/http/httptest` - In-process request dispatching for `InternalRequest`
- `encoding/json` - API response serialization
- `github.com/gorilla/websocket` - WebSocket upgrade and handling
- `text/template` - HTML template processing
- `crypto/tls` - HTTPS certificate management
- `internal/configs` - Configuration management and `SetVal`/`AllConfigData`/`GetOverrides`/`RestoreOverrides`
- `internal/users` - Authentication and user management
- `internal/mudlog` - Logging and monitoring
- `internal/util` - Game state mutex protection

