# Modules System Context

## Overview
The GoMud modules system provides a powerful plugin architecture that allows for extending the game engine with custom functionality without modifying the core codebase. Modules are self-contained packages that can add commands, handle events, provide web interfaces, and integrate deeply with all game systems through a comprehensive plugin API.

## Architecture Components

### Plugin Infrastructure (`internal/plugins`)

#### **Core Plugin System** (`plugins.go`)
- **Plugin struct**: Central plugin definition with callbacks, configuration, and file system integration
- **Plugin registry**: Global registry managing all loaded plugins with automatic discovery
- **Function export system**: Allows plugins to expose functions to other systems and JavaScript scripting
- **Embedded file system**: Each plugin can embed its own files (templates, web assets, data files)
- **Dependency management**: Plugin dependency resolution and version tracking

#### **Plugin Callbacks** (`plugincallbacks.go`)
- **Command registration**: Add new user commands and mob AI commands
- **Event handling**: IAC (telnet protocol) command processing via `SetIACHandler`
- **Text prefix handling**: WebSocket text prefix processing via `SetTextPrefixHandler` (used by GMCP for `!!GMCP(...)` messages)
- **Network callbacks**: Handle new connection events via `SetOnNetConnect`
- **Lifecycle hooks**: Load and save callbacks for plugin state management via `SetOnLoad`/`SetOnSave`
- **Script integration**: Expose plugin functions to JavaScript runtime via `AddScriptingFunction`
- **Function export**: Expose Go functions to other modules via `ExportFunction` / `GetExportedFunction`
- **Room tag registration**: Declare room tags the plugin recognises via `ReserveTags`; tags are listed by the `room tags` admin command

#### **Configuration System** (`pluginconfig.go`)
- **Plugin-specific config**: Each plugin gets its own configuration namespace
- **Dynamic configuration**: Runtime configuration changes through the main config system
- **Persistent settings**: Plugin configurations are saved with the main game configuration

#### **Web Integration** (`webconfig.go`)
- **Public web pages**: Plugins can add new pages to the public web interface via `plug.Web.WebPage`
- **Admin pages**: Plugins can add pages to the authenticated admin interface via `plug.Web.AdminPage`
- **Admin API endpoints**: Plugins can register REST API handlers under `/admin/api/v1/<slug>` via `plug.Web.AdminAPIEndpoint`
- **Navigation integration**: Add top-level or nested nav entries to the admin interface
- **Template system**: Use custom HTML templates with dynamic data supplied by a `func(*http.Request) map[string]any`
- **Asset serving**: Serve static files (CSS, JS, images) from plugin file systems
- **Acyclic wiring**: `main.go` calls `plugins.SetAdminRegistrar(web.GetAdminRegistrar())` before `plugins.Load()` to break the import cycle between `internal/web` and `internal/plugins`

#### **File System** (`pluginfiles.go`)
- **Embedded files**: Go embed.FS integration for packaging plugin assets
- **Virtual file system**: Plugins provide files through fs.ReadFileFS interface
- **File overlay system**: Plugin files can override core game files
- **Data file integration**: Plugin data files are merged with core data files

## Event System Integration

### **Event-Driven Architecture**
Modules extensively use the event system (`internal/events`) to integrate with game mechanics:

#### **Core Event Types**
- **NewRound**: Periodic processing (auctions, leaderboards, follow mechanics)
- **NewTurn**: Turn-based updates and maintenance
- **PlayerSpawn/PlayerDespawn**: Player login/logout handling
- **RoomChange**: Movement tracking and following behavior
- **PartyUpdated**: Party system integration
- **Communication**: Chat and communication system integration
- **CharacterVitalsChanged**: Character stat updates for GMCP
- **EquipmentChange**: Equipment updates for client sync
- **ItemOwnership**: Item tracking and quest integration

#### **Event Registration**
```go
events.RegisterListener(events.NewRound{}, module.handleNewRound)
events.RegisterListener(events.RoomChange{}, module.handleRoomChange)
```

#### **Event Priority System**
- **events.First**: High priority event handling
- **events.Last**: Final event processing
- **Default priority**: Standard event processing order

### **Custom Event Types**
Modules can define and emit custom events:
```go
type GMCPOut struct {
    ConnectionId uint64
    Data         []byte
}
```

## Plugin Capabilities

### Room Tag Registration
```go
// Declare the room tags this plugin recognises. Listed by "room tags" admin command.
plugin.ReserveTags("storage")
```

### Command System Integration
```go
// Add user commands
plugin.AddUserCommand("auction", auctionCommand, allowWhenDowned, adminOnly)

// Add mob AI commands  
plugin.AddMobCommand("customai", aiCommand, allowWhenDowned)
```

### **Scripting Integration**
```go
// Expose functions to JavaScript
plugin.AddScriptingFunction("GetFollowers", getFollowersFunc)

// Available in scripts as:
// modules.follow.GetFollowers()
```

### **Function Export (cross-module)**
```go
// Export a Go function for other modules to call by name
plugin.ExportFunction("SendInboxMessage", func(userId int, from, msg string, gold int, itm *items.Item) {
    // ...
})

// Consuming module retrieves and calls it
if fn, ok := usercommands.GetExportedFunction("SendInboxMessage"); ok {
    if f, ok := fn.(func(int, string, string, int, *items.Item)); ok {
        f(userId, "System", "Your message", 0, nil)
    }
}
```

### **Public Web Page Integration**
```go
// Add a public web page (no auth required)
plugin.Web.WebPage("Leaderboards", "/leaderboards", "leaderboards.html", true,
    func(r *http.Request) map[string]any {
        return map[string]any{"DATA": getLeaderboardData()}
    },
)

// Add navigation links to the public interface
plugin.Web.NavLink("Leaderboards", "/leaderboards")
```

### **Admin Page Integration**
```go
// Add an authenticated admin page at /admin/<slug>
// navGroup: if non-empty, places the entry inside a top-level group dropdown
// navParent: if non-empty, nests this page as a sub-item under that parent within the group
plugin.Web.AdminPage("Mudmail", "mudmail", "html/admin/mudmail.html", true, "Modules", "Mudmail",
    nil,
)

// Sub-item within the same group and parent
plugin.Web.AdminPage("API Docs", "mudmail-api", "html/admin/mudmail-api.html", true, "Modules", "Mudmail",
    nil,
)
```

### **Admin API Endpoint Registration**
```go
// Register a REST handler at /admin/api/v1/<slug>
// All admin API routes are automatically auth-gated and mud-locked
plugin.Web.AdminAPIEndpoint("GET", "mudmail", func(r *http.Request) (int, bool, any) {
    return http.StatusOK, true, getStats()
})

plugin.Web.AdminAPIEndpoint("POST", "mudmail", func(r *http.Request) (int, bool, any) {
    // parse r.Body, send mail ...
    return http.StatusOK, true, map[string]any{"sent": true}
})

plugin.Web.AdminAPIEndpoint("DELETE", "mudmail", func(r *http.Request) (int, bool, any) {
    return http.StatusOK, true, nil
})
```

The handler signature is `func(r *http.Request) (status int, success bool, data any)`. The framework wraps the return values in the standard `APIResponse[T]` JSON envelope.

### **Configuration Management**
```go
// Plugin-specific configuration (read from Modules.<pluginname>.* in config.yaml)
value := plugin.Config.Get("maxAuctions")
```

### **Plugin Data Persistence**
```go
// Save/load arbitrary bytes
plugin.WriteBytes("mydata", rawBytes)
rawBytes, err := plugin.ReadBytes("mydata")

// Save/load a struct (YAML serialization)
plugin.WriteStruct("auctionhistory", &auctionData)
plugin.ReadIntoStruct("auctionhistory", &auctionData)
```

Data is stored under `<datafiles>/plugin-data/<name>-v<version>/` as `<identifier>.plugin.dat` files.

### **File System Integration**
```go
//go:embed files/*
var files embed.FS

// Attach to plugin — walks the embed.FS and maps paths for later lookup
plugin.AttachFileSystem(files)

// Files available as overlays to core system
```

## Data File Integration

### **File Overlay System**
Modules can provide files that override or extend core game data:

#### **Data Overlays** (`files/data-overlays/`)
- **config.yaml**: Module-specific configuration additions
- **keywords.yaml**: Help system keyword additions
- **ansi-aliases.yaml**: Color scheme additions

#### **Data Files** (`files/datafiles/`)
- **templates/**: Custom message templates
- **html/**: Web interface files
- **help/**: Help system documentation

### **Template System Integration**
Modules can provide custom templates for:
- **Auction notifications**: Bid updates, auction start/end messages
- **Help documentation**: Command help and feature documentation  
- **Web interfaces**: Custom HTML pages with dynamic data

## Module Development Patterns

### **Basic Module Structure**
```go
package mymodule

import (
    "embed"
    "github.com/GoMudEngine/GoMud/internal/plugins"
    "github.com/GoMudEngine/GoMud/internal/events"
)

//go:embed files/*
var files embed.FS

func init()
```

### **State Management**
```go
type ModuleData struct {
    SomeState map[string]interface{} `json:"somestate"`
}

func (m *MyModule) save()
func (m *MyModule) load()
```

### **Event Handling**
```go
func (m *MyModule) handleNewRound(e events.Event) events.ListenerReturn
```

## Integration Points

### **Core System Hooks**
The engine provides numerous hooks for module integration:

#### **Auto-Save Integration**
- Modules are automatically saved during system save operations
- `plugins.Save()` called from `internal/hooks/NewTurn_AutoSave.go`

#### **Command System**
- User commands integrated into `internal/usercommands` system
- Mob commands integrated into `internal/mobcommands` system
- Full access to command infrastructure and permissions

#### **Event System**
- Full access to all game events
- Ability to register listeners with priority control
- Custom event types supported

#### **Web System**
- Integration with public and admin web interfaces
- Admin pages require `plugins.SetAdminRegistrar(web.GetAdminRegistrar())` called in `main.go` before `plugins.Load()`
- Custom pages, navigation, and REST API endpoints
- Template system access with per-request data functions
- All admin routes are automatically auth-gated and mud-locked by the framework

### **JavaScript Runtime Integration**
- Module functions exposed to scripting system
- Available in scripts as `modules.modulename.functionname()`
- Full integration with game scripting capabilities

## Performance Considerations

### **Event Processing**
- Event listeners are called synchronously
- Heavy processing should be deferred or optimized
- Event priority system allows control over execution order

### **File System**
- Embedded files are loaded at compile time
- File overlay system provides efficient file serving
- Plugin files cached in memory for performance

### **State Management**
- Plugin state saved/loaded with main game state
- JSON serialization for complex data structures
- Efficient binary serialization available

## Security and Isolation

### **Sandboxed Execution**
- Plugins run in the same process but with controlled access
- No direct file system access outside of embedded files
- Configuration changes go through main config system

### **API Boundaries**
- Well-defined interfaces for all plugin interactions
- Type-safe event system
- Controlled access to core game systems

## Module Ecosystem

### **Official Modules**
- **GMCP**: Essential for modern MUD clients
- **Auctions**: Player economy features
- **Cleanup**: World maintenance (trash and bury commands)
- **Follow**: Social and AI mechanics
- **Gambling**: Room-based slot machines and claw machines
- **Leaderboards**: Player engagement and competition
- **Mudmail**: Player inbox and admin mass-mail system
- **Newbie Guide**: Automatic guide companion for new players
- **Storage**: Player item storage system with room-tag-based activation, legacy migration, and admin web interface
- **Time**: Basic utility functionality
- **Web Help**: Web-based help browser
- **Zombie Mode**: AFK automation system

### **Module Discovery**
- Automatic module loading via `modules/all-modules.go`
- Code generation for module imports
- Compile-time module inclusion

This comprehensive plugin architecture allows GoMud to be extended with sophisticated functionality while maintaining clean separation between core engine and optional features. The event-driven design ensures modules can integrate deeply with game mechanics while remaining modular and maintainable.