# Procedural Package Context

## Overview

The `internal/procedural` package provides procedural content generation for GoMud, currently focused on maze generation. It generates 2D grid mazes using recursive backtracking and can convert them into live ephemeral rooms in the game world.

## Key Components

### Interfaces (`procedural.go`)

- **`MazeRoom`**: Interface for a single cell in a generated maze.
  - `GetPosition() (x, y, z int)`
  - `GetConnections() []MazeRoom`
  - `IsConnectedTo(MazeRoom) bool`
  - `GetStep() int` — step number on the critical path (0 if not on critical path)
  - `IsStart() bool`, `IsEnd() bool`, `IsDeadEnd() bool`
  - `GetDistanceFromStart() int`
- **`Maze`**: Interface for a maze generator.
  - `Generate2D(xMax, yMax int, seed ...string) Maze2D`
  - `GetStart() (int, int)`, `GetEnd() (int, int)`
  - `GetCriticalPath() []MazeRoom`
- **`Maze2D`**: `[][]*GridRoom` — a 2D slice indexed `[y][x]`; `nil` entries indicate removed/absent cells.
- **`Maze3D`**: `[][][]*GridRoom` — reserved for future 3D maze support.

### Grid Maze Implementation (`gridmaze.go`)

- **`GridRoom`**: Concrete `MazeRoom` implementation. Stores position, connections (bidirectional), step, distance from start, and start/end flags.
- **`GridMaze`**: Concrete `Maze` implementation.
  - **`NewGridMaze() *GridMaze`**: Creates a new generator with a time-seeded RNG.
  - **`Generate2D(xMax, yMax int, seed ...string) [][]*GridRoom`**: Full maze generation pipeline:
    1. Creates all rooms in the grid.
    2. Generates connectivity via recursive backtracking (`backtrack`).
    3. Randomly selects start position; selects end position using weighted random selection favoring distance from start.
    4. Calculates BFS distances from start.
    5. Finds the critical path (BFS shortest path from start to end) and sets `step` values.
    6. Adds extra connections (`addAdditionalConnections`) to increase branching (up to 3 connections per room, ~10% of grid size attempts).
    7. Removes rooms randomly while protecting the critical path (`removeRoomsProtectingPath`, ~50% removal chance for non-critical rooms).
    8. Removes orphaned rooms unreachable from start.
    9. Recalculates distances and critical path after modifications.
  - Deterministic generation when a non-empty seed string is provided (FNV-64a hash of seed string).

### Room Instantiation (`rooms.go`)

- **`CreateEphemeralMaze2D(mazeRooms [][]*GridRoom) (allRoomIds []int, startRoomId int, endRoomId int)`**: Converts a generated maze into live ephemeral rooms.
  - Allocates the required number of ephemeral room IDs via `rooms.CreateEmptyEphemeralRooms`.
  - Iterates the maze grid and wires up north/south/east/west exits between connected rooms.
  - Returns all room IDs, the start room ID, and the end room ID.

## Usage Pattern

```go
// Generate a 10x10 maze with a deterministic seed
maze := procedural.NewGridMaze()
grid := maze.Generate2D(10, 10, "my-seed-string")

// Instantiate as ephemeral rooms in the game world
allIds, startId, endId := procedural.CreateEphemeralMaze2D(grid)

// Teleport a player to the start room
// ... rooms.LoadRoom(startId), user.Character.RoomId = startId, etc.
```

## Dependencies

- `internal/rooms`: `CreateEmptyEphemeralRooms`, `LoadRoom` for ephemeral room creation
- `internal/exit`: `RoomExit` struct for wiring room exits
- Standard library: `hash/fnv`, `math/rand`, `time`

## Testing

Test coverage in `*_test.go` files covers:
- Maze generation dimensions and connectivity
- Critical path existence and correctness
- Start/end room assignment
- Deterministic generation with seeds
- Dead-end detection
- Orphan removal

## Special Considerations

- **Nil cells**: The `Maze2D` grid may contain `nil` entries after room removal. All consumers must nil-check before accessing a cell.
- **Bidirectional connections**: `addConnection` on `GridRoom` is always bidirectional; the grid-to-room wiring in `rooms.go` handles each pair once by only connecting in the south and east directions during iteration.
- **Critical path protection**: Rooms on the critical path (step > 0) are never removed, guaranteeing the maze is always solvable.
- **Ephemeral room limits**: The number of rooms generated is bounded by the ephemeral room system limits (100 chunks × 250 rooms = 25,000 max ephemeral rooms).
