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

## Go Source Files

| File | Purpose |
|---|---|
| `web.go` | Server startup, `internalMux`, `serveTemplate`, `RunWithMUDLocked`, `Shutdown`, public route registration |
| `admin.go` | `adminIndex` handler - single admin dashboard page |
| `admin_routes.go` | `registerAdminRoutes(mux)` - registers all `/admin/` routes in one place |
| `api.go` | `APIResponse[T]` generic envelope, `writeJSON`, `writeAPIError` helpers |
| `api_routes.go` | `registerAdminAPIRoutes(mux)` - registers all `/admin/api/` routes |
| `api_v1_config.go` | `apiV1GetConfig` and `apiV1PatchConfig` handlers |
| `auth.go` | `doBasicAuth`, `handlerToHandlerFunc`, auth cache |
| `context.go` | `withInternalContext`, `IsInternalRequest` - internal request context flag |
| `internal.go` | `InternalRequest`, `InternalRequestJSON` - in-process API dispatcher |
| `stats.go` | `Stats`, `GetStats`, `UpdateStats` |
| `template_func.go` | `funcMap` - custom template functions |

## Routing Structure

All routes are registered on the package-level `internalMux`. Both live HTTP/HTTPS servers and `InternalRequest` use this same mux.

### Public Routes (registered inline in `Listen()`)
- `GET /favicon.ico` - favicon redirect
- `GET /` - public template server (`serveTemplate`)
- `GET /ws` - WebSocket upgrade endpoint

### Admin Routes (registered via `registerAdminRoutes`)
- `GET /admin/` - admin dashboard (auth required)

### API Routes (registered via `registerAdminAPIRoutes`, called from `registerAdminRoutes`)
- `GET /admin/api/v1/config` - return all config as flat key/value map (auth required)
- `PATCH /admin/api/v1/config` - update one or more config values (auth required)

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
    Success bool   `json:"success"`
    Data    T      `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
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

## Admin HTML Templates

Located in `_datafiles/html/admin/` (path configured via `FilePaths.AdminHtml`):

| File | Purpose |
|---|---|
| `_header.html` | Defines `{{define "header"}}` - minimal HTML5 shell, inline CSS, top nav bar |
| `_footer.html` | Defines `{{define "footer"}}` - closing tags |
| `index.html` | Dashboard: server name, version, online count, ports, API endpoint listing |

No external CDN dependencies. No Bootstrap, jQuery, or HTMX.

Template data passed to `adminIndex`:
- `CONFIG` - `configs.Config` struct
- `STATS` - `web.Stats` struct (online users, telnet ports, websocket port)

## Plugin Integration

### WebPlugin Interface
```go
type WebPlugin interface {
    NavLinks() map[string]string
    WebRequest(r *http.Request) (html string, templateData map[string]any, ok bool)
}
```

Plugins can add navigation links and handle custom public web requests via `SetWebPlugin`. The admin interface does not use plugin templates.

## Security

- All `/admin/` paths require HTTP Basic Authentication via `doBasicAuth`.
- Users must have a role other than `user` (i.e., admin or higher).
- Successful auth results are cached for 30 minutes.
- All admin handlers are wrapped with `RunWithMUDLocked` to serialize access to shared game state.
- Internal requests (via `InternalRequest`) bypass both auth and locking; they are identified by a context value set in `withInternalContext` and checked by `IsInternalRequest`.

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
- `internal/configs` - Configuration management and `SetVal`/`AllConfigData`
- `internal/users` - Authentication and user management
- `internal/mudlog` - Logging and monitoring
- `internal/util` - Game state mutex protection

