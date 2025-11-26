package procedural

// [y][x]
type Maze2D [][]*GridRoom

// [z][y][x]
type Maze3D [][][]*GridRoom

// MazeRoom represents a single room/space in the maze
type MazeRoom interface {
	// GetPosition returns the x, y, z coordinates of this room
	GetPosition() (int, int, int)

	// GetConnections returns all rooms this room connects to
	GetConnections() []MazeRoom

	// IsConnectedTo checks if this room is connected to another room
	IsConnectedTo(room MazeRoom) bool

	// GetStep returns the step number on the critical path (0 if not on critical path)
	GetStep() int

	// IsStart returns true if this is the start room
	IsStart() bool

	// IsEnd returns true if this is the end room
	IsEnd() bool

	// IsDeadEnd returns true if this room is a dead end (only one connection)
	IsDeadEnd() bool

	// GetDistanceFromStart returns the number of steps from the start room
	GetDistanceFromStart() int
}

// Maze represents a 2D maze generator
type Maze interface {
	// Generate creates a new maze with the specified dimensions
	// Optional seed string can be provided for deterministic generation
	// Returns a 2D slice where [x][y] contains a MazeRoom or nil
	Generate2D(xMax, yMax int, seed ...string) Maze2D

	// GetStart returns the starting room position
	GetStart() (int, int)

	// GetEnd returns the ending room position
	GetEnd() (int, int)

	// GetCriticalPath returns the rooms on the critical path from start to end
	GetCriticalPath() []MazeRoom
}
