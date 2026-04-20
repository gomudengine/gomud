# GoMud Network Layer - Comprehensive Analysis

## Overview

The GoMud network layer provides a sophisticated, multi-protocol networking system that supports traditional MUD clients via telnet, modern web-based clients via WebSocket, and terminal clients via SSH. The architecture is designed for high concurrency, robust protocol handling, and extensive customization through a plugin system.

## Network Architecture

### Core Components

**Main Entry Point (`main.go`)**
- Primary network initialization and server startup
- Manages multiple concurrent network listeners
- Handles graceful shutdown and connection cleanup
- Coordinates between telnet, WebSocket, and SSH protocols

**Connection Management (`internal/connections`)**
- Thread-safe connection tracking with unique IDs
- Triple protocol support: telnet (`net.Conn`), WebSocket (`*websocket.Conn`), SSH (`ssh.Channel`)
- Connection state management (Login, LoggedIn, LinkDead)
- Heartbeat monitoring for WebSocket connections
- Input processing pipeline with chainable handlers

**Web Server (`internal/web`)**
- HTTP/HTTPS server with WebSocket upgrade capability
- Administrative interface with authentication
- Template-based HTML rendering
- Plugin integration for custom web functionality

**Terminal Protocol (`internal/term`)**
- Comprehensive telnet protocol implementation
- ANSI escape sequence processing
- MUD Sound Protocol (MSP) support
- Cross-platform terminal compatibility

**Input Processing (`internal/inputhandlers`)**
- Multi-step authentication workflows
- System command processing
- Terminal protocol handling (IAC, ANSI)
- Input validation and sanitization

## Protocol Support

### Telnet Protocol
- **Default Ports**: 33333, 44444 (configurable)
- **Local Admin Port**: 9999 (localhost only, no connection limits)
- **Max Connections**: 100 (configurable)
- **Protocol Features**:
  - Full telnet option negotiation (WILL/WONT/DO/DONT)
  - Echo control for password input
  - Window size negotiation (NAWS)
  - Character encoding negotiation
  - Go-ahead suppression
  - Binary mode support

### WebSocket Protocol
- **Endpoint**: `/ws` on HTTP/HTTPS server
- **Upgrade**: Automatic HTTP to WebSocket upgrade
- **Heartbeat**: Ping/pong monitoring (60-second intervals)
- **Features**:
  - Real-time bidirectional communication
  - Cross-origin request support for development
  - Automatic connection health monitoring
  - Text masking for password input

### SSH Protocol
- **Port**: Configurable via `SSHPort` (0 to disable)
- **Max Connections**: 50 (configurable via `MaxSSHConnections`)
- **Authentication**: No client auth required (`NoClientAuth: true`)
- **Host Key**: RSA/EC private key file configured via `SSHHostKeyFile`; SSH is disabled if not set
- **Features**:
  - Full SSH handshake and channel negotiation via `golang.org/x/crypto/ssh`
  - Only `session` channel type accepted; others rejected
  - `pty-req` handling: parses terminal dimensions (cols/rows) from payload
  - `window-change` handling: live terminal resize updates to client settings
  - Out-of-band requests (keepalive, etc.) silently discarded
  - Same input handler chain and game integration as telnet
  - Connection type reported as `ssh` in `OnlineInfo` and Discord notifications
  - Tracked via `NewSSHConnectionDetails` with `ssh.Channel` and remote `net.Addr`

### HTTP/HTTPS Server
- **HTTP Port**: 80 (configurable, 0 to disable)
- **HTTPS Port**: 0 (disabled by default, requires certificates)
- **HTTPS Redirect**: Optional automatic HTTP to HTTPS redirection
- **Features**:
  - Static file serving
  - Template-based dynamic content
  - Administrative interface with authentication
  - Plugin web integration

## Connection Management

### Connection Lifecycle

**Connection Establishment**:
1. Accept incoming connection (telnet, WebSocket, or SSH)
2. Generate unique connection ID (atomic counter)
3. Initialize connection details structure
4. Set up input handler chain
5. Begin protocol negotiation (telnet) or heartbeat (WebSocket)

**Connection States**:
- **Login**: Initial state before authentication
- **LoggedIn**: Authenticated and active
- **LinkDead**: Disconnected but not yet cleaned up (configurable timeout)

**Connection Tracking**:
- Thread-safe connection registry with RWMutex
- Unique connection IDs for identification
- Connection statistics (total connects/disconnects)
- Active connection count monitoring

### Input Processing Pipeline

**Handler Chain Architecture**:
- Chainable input processors with configurable order
- Each handler can abort or continue processing
- Shared state map for handler communication
- Handler-specific error handling and recovery

**Standard Handler Chain (Telnet and SSH)**:
1. **TelnetIACHandler**: Telnet protocol command processing
2. **AnsiHandler**: ANSI escape sequence processing
3. **CleanserInputHandler**: Input sanitization
4. **LoginPromptHandler**: Multi-step authentication (initial)
5. **EchoInputHandler**: Terminal echo management (post-login)
6. **HistoryInputHandler**: Command history tracking
7. **SystemCommandInputHandler**: System commands (admin only)
8. **SignalHandler**: Terminal signal processing

**WebSocket Processing**:
- Simplified handler chain (no telnet/ANSI processing)
- Direct message processing
- Text masking for password fields
- Real-time input handling

**SSH Processing**:
- Same full handler chain as telnet
- Terminal dimensions initialized from `pty-req` payload
- Live resize via `window-change` channel requests
- Reads/writes directly on the SSH channel

## Authentication and Security

### Multi-Step Authentication
- **Username Validation**: Existence checking and format validation
- **Password Authentication**: Secure bcrypt password verification
- **Account Creation**: New user registration workflow
- **Duplicate Login Handling**: Detection and management of concurrent sessions

### Security Features
- **Input Sanitization**: Protection against injection attacks
- **Rate Limiting**: Protection against input flooding
- **Authentication Caching**: 30-minute session caching for admin interface
- **Role-Based Access**: Admin/user role verification
- **Connection Limits**: Configurable maximum connections per protocol

### System Commands
- **Administrative Commands**: `/quit`, `/reload`, `/shutdown`
- **Permission Checking**: Admin role verification required
- **Graceful Operations**: Countdown timers for shutdown operations
- **Audit Logging**: Comprehensive logging of administrative actions

## Network Configuration

### Port Configuration
```yaml
Network:
  MaxTelnetConnections: 100
  TelnetPort: [33333, 44444]    # Multiple ports supported
  LocalPort: 9999               # Localhost admin access
  HttpPort: 80                  # Web server port
  HttpsPort: 0                  # HTTPS port (0 = disabled)
  HttpsRedirect: false          # Auto-redirect HTTP to HTTPS
  SSHPort: 0                    # SSH port (0 = disabled)
  MaxSSHConnections: 50         # Max SSH connections
  LinkDeadSeconds: 60           # Link-dead connection timeout
  AfkSeconds: 1800              # AFK timeout
  TimeoutMods: false            # Timeout moderators/admins
```

### File Paths
```yaml
FilePaths:
  WebDomain: "localhost"
  WebCDNLocation: ""            # Optional CDN for static files
  PublicHtml: "_datafiles/html/public"
  AdminHtml: "_datafiles/html/admin"
  HttpsCertFile: ""             # TLS certificate
  HttpsKeyFile: ""              # TLS private key
  SSHHostKeyFile: ""            # SSH host private key (required for SSH)
```

## Advanced Features

### Heartbeat System (WebSocket)
- **Ping Interval**: 54 seconds (90% of pong timeout)
- **Pong Timeout**: 60 seconds
- **Write Timeout**: 10 seconds for control messages
- **Automatic Cleanup**: Connection removal on heartbeat failure
- **Thread Safety**: Goroutine-safe ping/pong handling

### Client Settings Management
- **Screen Dimensions**: Width/height tracking (default 80x40)
- **Protocol Capabilities**: MSP support detection
- **Display Preferences**: Color and formatting support
- **Terminal Type**: Client terminal identification
- **Connection Type**: Tracked as `telnet`, `websocket`, or `ssh` in `OnlineInfo`

### Command History
- **History Size**: 10 commands maximum
- **Navigation**: Up/down arrow key support
- **Position Tracking**: Current history position management
- **Session Persistence**: History maintained per connection

### Input Buffer Management
- **Buffer Size**: 1024 bytes read buffer
- **Real-time Processing**: Character-by-character input handling
- **Special Keys**: Enter, Backspace, Tab detection
- **Clipboard Support**: Paste operation handling

## Plugin Integration

### Network Plugin Hooks
- **Connection Events**: `OnNetConnect` for new connections
- **IAC Command Processing**: Custom telnet command handling
- **Web Interface Extensions**: Custom admin pages and navigation
- **Command Registration**: Add custom user and system commands

### Plugin Capabilities
- **Web Pages**: Custom HTML pages with template processing
- **Navigation Links**: Add menu items to web interface
- **Static Assets**: Serve CSS, JS, images from plugin filesystem
- **Template Data**: Inject custom data into web templates

## Performance Characteristics

### Concurrency Model
- **Goroutine per Connection**: Each connection handled in separate goroutine
- **Thread-Safe Operations**: All connection operations use mutex protection
- **Non-Blocking I/O**: Asynchronous network operations
- **Resource Pooling**: Efficient buffer and resource management

### Scalability Features
- **Connection Limits**: Configurable per-protocol connection limits
- **Resource Management**: Automatic cleanup of failed connections
- **Memory Efficiency**: Minimal per-connection memory overhead
- **CPU Utilization**: Configurable CPU core usage

### Monitoring and Statistics
- **Connection Tracking**: Total connections and disconnections
- **Active Connections**: Real-time active connection count
- **Error Logging**: Comprehensive error tracking and reporting

## Error Handling and Recovery

### Connection Error Management
- **Graceful Degradation**: Automatic handling of connection failures
- **LinkDead State**: Temporary preservation of disconnected users
- **Automatic Cleanup**: Failed connection removal and resource cleanup
- **Error Logging**: Detailed error reporting with context

### Protocol Error Handling
- **Malformed Input**: Safe handling of invalid protocol data
- **Timeout Management**: Connection and operation timeouts
- **Buffer Overflow Protection**: Safe buffer handling
- **Recovery Mechanisms**: Automatic recovery from protocol errors

## Integration with Game Systems

### User Management Integration
- **Authentication**: Seamless integration with user database
- **Session Management**: Connection association with user accounts
- **Character Loading**: Automatic character data loading on login
- **Duplicate Detection**: Prevention of multiple concurrent logins
- **Connection Type**: SSH/WebSocket/telnet type exposed via `UserRecord.GetOnlineInfo()`

### Event System Integration
- **Connection Events**: Login/logout event generation
- **Input Events**: User command event processing
- **Broadcast Events**: Server-wide message distribution
- **Custom Events**: Plugin-generated network events

### World Manager Integration
- **Input Processing**: User command forwarding to game engine
- **Output Handling**: Game output formatting and transmission
- **State Synchronization**: Connection state with game state
- **LinkDead Management**: Temporary character preservation

## User Input Processing Flow

### Input Flow Stages

#### Stage 1: Network Input Reception
Raw bytes received from the connection and passed through the input handler chain (see Handler Chain above).

#### Stage 2: World Input Channel
```go
type WorldInput struct {
    FromId    int     // User ID
    InputText string  // Raw command text
    ReadyTurn uint64  // Turn number when ready to process
}
```
- **Channel**: `worldInput chan WorldInput`
- Synchronous communication between network and game layers

#### Stage 3: Input Worker Processing
- Receives `WorldInput` from network layer
- Converts to `events.Input` event
- Adds to event queue with current turn number

#### Stage 4: Event System Processing
```go
type Input struct {
    UserId        int
    MobInstanceId int
    InputText     string
    ReadyTurn     uint64
    WaitTurns     int
    Flags         EventFlag
}
```
- Turn-based queuing: commands wait until their `ReadyTurn`
- One command per user per turn
- Commands not ready are requeued

#### Stage 5: Command Execution (`processInput()` in `world.go`)
1. User validation
2. Prompt handling
3. Macro expansion
4. Command parsing
5. Execution via `usercommands.TryCommand()`

### Channel Architecture

| Channel | Type | Purpose |
|---------|------|---------|
| World Input | `chan WorldInput` | User commands from network layer |
| Enter World | `chan [2]int` | User login completion (userId, roomId) |
| Leave World | `chan int` | User logout/disconnect (userId) |
| Logout Connection | `chan connections.ConnectionId` | Connection-specific logout |
| LinkDead Flag | `chan [2]int` | LinkDead state management |

### Worker Goroutines

**InputWorker**: Converts network input to game events. Single goroutine, no shared state.

**MainWorker**: Game state management and event processing. Handles event loop, turn/round timing, user lifecycle, room maintenance.

### Input Processing Features

- **Turn-Based Processing**: Global turn counter; one command per user per turn
- **Input Blocking**: Commands can block/unblock further user input via `CmdBlockInput` flag
- **Macro System**: Two-character macros expand to semicolon-delimited command sequences
- **Alias Resolution**: Commands resolved through keyword alias system

## Deployment and Operations

### Server Startup Process
1. Configuration loading and network settings validation
2. Port binding: Telnet, SSH, and HTTP/HTTPS server startup
3. Worker initialization: InputWorker and MainWorker goroutines
4. Plugin loading and network plugin initialization
5. Signal handling for graceful shutdown

### Graceful Shutdown
1. Broadcast shutdown message to all connections
2. Block new connections
3. Close all active connections
4. Release network resources
5. Wait for worker completion
