# GoMud Connections System Context

## Overview

The GoMud connections system provides comprehensive network connection management with support for telnet, WebSocket, and SSH connections. It features connection lifecycle management, input handling with history, heartbeat monitoring for WebSocket connections, client settings management, and thread-safe connection operations with graceful shutdown support.

## Architecture

The connections system is built around several key components:

### Core Components

**Connection Management:**
- Thread-safe connection tracking with unique IDs
- Triple protocol support: telnet (`net.Conn`), WebSocket (`*websocket.Conn`), and SSH (`ssh.Channel`)
- Connection state management (Login, LoggedIn, LinkDead)
- Automatic connection cleanup and resource management

**Input Processing:**
- Chainable input handler system with state management
- Command history with navigation (up/down arrows)
- Special key handling (Enter, Backspace, Tab)
- Input buffering and clipboard support

**Client Settings:**
- Screen dimension tracking and defaults
- MSP (MUD Sound Protocol) support detection
- Telnet protocol option management
- Display preference configuration

**Heartbeat System:**
- WebSocket connection monitoring with ping/pong
- Configurable timeout and interval settings
- Automatic connection cleanup on timeout
- Thread-safe heartbeat management

## Key Features

### 1. **Triple Protocol Support**
- **Telnet Connections**: Traditional MUD protocol with full terminal control
- **WebSocket Connections**: Modern web-based connections with heartbeat monitoring
- **SSH Connections**: Full SSH channel support via `golang.org/x/crypto/ssh`
- **Unix Socket Support**: Local connections for development and administration
- **Protocol Detection**: `IsWebSocket()` and `IsSSH()` methods on `ConnectionDetails`

### 2. **Advanced Input Management**
- **Handler Chaining**: Multiple input processors with configurable order
- **Command History**: 10-command history with navigation
- **Special Key Support**: Enter, Backspace, Tab handling
- **Input Buffering**: Real-time input processing with buffer management

### 3. **Connection State Management**
- **Login State**: Initial connection before authentication
- **LoggedIn State**: Authenticated and active connections
- **LinkDead State**: Disconnected but not yet cleaned up
- **Thread-Safe Operations**: All connection operations are mutex-protected

### 4. **Robust Heartbeat System**
- **WebSocket Monitoring**: Automatic ping/pong for connection health
- **Configurable Timeouts**: Customizable ping intervals and pong wait times
- **Graceful Degradation**: Automatic cleanup on connection failure
- **Resource Management**: Proper goroutine cleanup on disconnection

## Connection Structure

### ConnectionDetails Structure
```go
type ConnectionDetails struct {
    connectionId      ConnectionId    // Unique connection identifier
    state             ConnectState    // Current connection state
    lastInputTime     time.Time       // Last input received timestamp
    conn              net.Conn        // Raw network connection (telnet/unix)
    wsConn            *websocket.Conn // WebSocket connection (if applicable)
    wsLock            sync.Mutex      // WebSocket write synchronization
    sshChannel        ssh.Channel     // SSH channel (if applicable)
    sshRemoteAddr     net.Addr        // SSH remote address
    handlerMutex      sync.Mutex      // Input handler synchronization
    inputHandlerNames []string        // Handler names for management
    inputHandlers     []InputHandler  // Handler function chain
    inputDisabled     bool            // Input processing toggle
    clientSettings    ClientSettings  // Client configuration
    heartbeat         *heartbeatManager // WebSocket heartbeat manager
}
```

### Construction
```go
// Create a telnet or WebSocket connection (heartbeat auto-started for WebSocket)
func NewConnectionDetails(connId ConnectionId, c net.Conn, wsC *websocket.Conn, config *HeartbeatConfig) *ConnectionDetails

// Create an SSH connection
func NewSSHConnectionDetails(connId ConnectionId, ch ssh.Channel, remoteAddr net.Addr) *ConnectionDetails
```

### Client Input Structure
```go
type ClientInput struct {
    ConnectionId  ConnectionId // Connection identifier
    DataIn        []byte       // Raw input data
    Buffer        []byte       // Current input buffer
    Clipboard     []byte       // Clipboard content for paste operations
    LastSubmitted []byte       // Previous command for reference
    EnterPressed  bool         // Enter key detection
    BSPressed     bool         // Backspace key detection
    TabPressed    bool         // Tab key detection
    History       InputHistory // Command history management
}
```

### Client Settings Structure
```go
type ClientSettings struct {
    Display           DisplaySettings // Screen dimensions and display options
    MSPEnabled        bool            // MUD Sound Protocol support
    SendTelnetGoAhead bool            // Telnet IAC GA after prompts
}

type DisplaySettings struct {
    ScreenWidth  uint32 // Terminal width (default: 80)
    ScreenHeight uint32 // Terminal height (default: 40)
}
```

## Connection Management

### Adding Connections
```go
// Add a telnet or WebSocket connection
func Add(conn net.Conn, wsConn *websocket.Conn) *ConnectionDetails

// Add an SSH connection
func AddSSH(ch ssh.Channel, remoteAddr net.Addr) *ConnectionDetails
```

### Connection Lifecycle
```go
// Remove connection and cleanup resources
func Remove(id ConnectionId) error

// Forcibly disconnect with reason (does not delete from map)
func Kick(id ConnectionId, reason string) error

// Cleanup all connections
func Cleanup()
```

### Connection Discovery
```go
// Get connection by ID
func Get(id ConnectionId) *ConnectionDetails

// Check if connection is WebSocket
func IsWebsocket(id ConnectionId) bool

// Get all active connection IDs
func GetAllConnectionIds() []ConnectionId

// Get active connection count
func ActiveConnectionCount() int

// Get connection statistics
func Stats() (connections uint64, disconnections uint64)
```

### Connection Properties
```go
func (cd *ConnectionDetails) IsWebSocket() bool
func (cd *ConnectionDetails) IsSSH() bool
func (cd *ConnectionDetails) IsLocal() bool      // loopback or unix socket
func (cd *ConnectionDetails) RemoteAddr() net.Addr
func (cd *ConnectionDetails) ConnectionId() ConnectionId
func (cd *ConnectionDetails) State() ConnectState
func (cd *ConnectionDetails) SetState(state ConnectState)
func (cd *ConnectionDetails) InputDisabled(setTo ...bool) bool
```

## Input Handling System

### Input Handler Chain
```go
type InputHandler func(ci *ClientInput, handlerState map[string]any) (doNextHandler bool)

// Add input handler (optionally after a named handler)
func (cd *ConnectionDetails) AddInputHandler(name string, newInputHandler InputHandler, after ...string)

// Remove input handler by name
func (cd *ConnectionDetails) RemoveInputHandler(name string)

// Process input through handler chain
func (cd *ConnectionDetails) HandleInput(ci *ClientInput, handlerState map[string]any) (doNextHandler bool, lastHandler string, err error)
```

### Input Processing
```go
// Reset the client input to "no current input"
func (ci *ClientInput) Reset()
```

## Command History System

### History Management
```go
type InputHistory struct {
    inhistory bool     // Currently navigating history
    position  int      // Current position in history
    history   [][]byte // Command history buffer (max 10)
}

func (ih *InputHistory) Add(input []byte)
func (ih *InputHistory) Previous()
func (ih *InputHistory) Next()
func (ih *InputHistory) Get() []byte
func (ih *InputHistory) Position() int
func (ih *InputHistory) ResetPosition()
func (ih *InputHistory) InHistory() bool
```

## Communication System

### Broadcasting and Messaging
```go
// Broadcast message to all LoggedIn connections (with optional exclusions)
func Broadcast(colorizedText []byte, skipConnectionIds ...ConnectionId) []ConnectionId

// Send message to specific connections
func SendTo(b []byte, ids ...ConnectionId)
```

### Protocol-Specific I/O
```go
// Write data (handles SSH, WebSocket, and telnet)
// - SSH: writes to ssh.Channel
// - WebSocket: uses wsLock, rejects telnet IAC bytes
// - Telnet/Unix: writes to net.Conn
// Line endings \n -> \r\n are applied for all protocols
func (cd *ConnectionDetails) Write(p []byte) (n int, err error)

// Read data (handles SSH, WebSocket, and telnet)
func (cd *ConnectionDetails) Read(p []byte) (n int, err error)

// Close the connection (stops heartbeat, closes underlying transport)
func (cd *ConnectionDetails) Close()
```

## Heartbeat System

### WebSocket Heartbeat Management
```go
type HeartbeatConfig struct {
    PongWait   time.Duration // Maximum time to wait for pong
    PingPeriod time.Duration // Ping interval (90% of PongWait)
    WriteWait  time.Duration // Write timeout for control messages
}

var DefaultHeartbeatConfig = HeartbeatConfig{
    PongWait:   60 * time.Second,
    PingPeriod: 54 * time.Second, // (60s * 9) / 10
    WriteWait:  10 * time.Second,
}

// Start heartbeat monitoring for a WebSocket connection
func (cd *ConnectionDetails) StartHeartbeat(config HeartbeatConfig) error
```

## Client Settings Management

```go
// Get client settings for connection
func GetClientSettings(id ConnectionId) ClientSettings

// Update client settings
func OverwriteClientSettings(id ConnectionId, cs ClientSettings)

// Display settings with defaults
func (c DisplaySettings) GetScreenWidth() int   // default 80
func (c DisplaySettings) GetScreenHeight() int  // default 40
```

## Connection State Management

```go
type ConnectState uint32

const (
    Login    ConnectState = iota // Initial connection state
    LoggedIn                     // Authenticated and active
    LinkDead                     // Disconnected but not cleaned up
)
```

## Shutdown and Cleanup

```go
// Set shutdown signal channel (call once at startup)
func SetShutdownChan(osSignalChan chan os.Signal)

// Signal shutdown to all systems
func SignalShutdown(s os.Signal)
```

## Usage Examples

### Telnet Connection Setup
```go
conn, err := listener.Accept()
connDetails := connections.Add(conn, nil) // nil = telnet
connDetails.AddInputHandler("auth", authHandler)
connDetails.SetState(connections.LoggedIn)
```

### WebSocket Connection Setup
```go
wsConn, err := upgrader.Upgrade(w, r, nil)
connDetails := connections.Add(nil, wsConn) // heartbeat auto-started
```

### SSH Connection Setup
```go
// After SSH handshake and channel acceptance:
connDetails := connections.AddSSH(sshChannel, remoteAddr)
// Terminal dimensions set separately via OverwriteClientSettings
```

### Input Handler Implementation
```go
func gameInputHandler(ci *connections.ClientInput, handlerState map[string]any) bool
connDetails.AddInputHandler("game", gameInputHandler)
```

## Dependencies

- `net` - Network connection handling and address management
- `sync` - Thread synchronization and atomic operations
- `github.com/gorilla/websocket` - WebSocket protocol implementation
- `golang.org/x/crypto/ssh` - SSH protocol implementation
- `internal/mudlog` - Logging system for connection events and debugging
- `internal/term` - Terminal control codes and telnet protocol handling
