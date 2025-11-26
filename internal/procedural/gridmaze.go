package procedural

import (
	"hash/fnv"
	"math/rand"
	"time"
)

// GridRoom implements the MazeRoom interface
type GridRoom struct {
	x, y, z           int
	connections       []*GridRoom
	step              int
	distanceFromStart int
	isStart           bool
	isEnd             bool
}

func (r *GridRoom) GetPosition() (int, int, int) {
	return r.x, r.y, r.z
}

func (r *GridRoom) GetConnections() []MazeRoom {
	connections := make([]MazeRoom, len(r.connections))
	for i, conn := range r.connections {
		connections[i] = conn
	}
	return connections
}

func (r *GridRoom) IsConnectedTo(room MazeRoom) bool {
	if room == nil {
		return false
	}
	rx, ry, rz := room.GetPosition()
	for _, conn := range r.connections {
		if conn.x == rx && conn.y == ry && conn.z == rz {
			return true
		}
	}
	return false
}

func (r *GridRoom) GetStep() int {
	return r.step
}

func (r *GridRoom) IsStart() bool {
	return r.isStart
}

func (r *GridRoom) IsEnd() bool {
	return r.isEnd
}

func (r *GridRoom) IsDeadEnd() bool {
	return len(r.connections) == 1
}

func (r *GridRoom) GetDistanceFromStart() int {
	return r.distanceFromStart
}

func (r *GridRoom) addConnection(other *GridRoom) {
	if r.IsConnectedTo(other) {
		return
	}
	r.connections = append(r.connections, other)
	other.connections = append(other.connections, r)
}

// GridMaze implements the Maze interface
type GridMaze struct {
	rooms          Maze2D
	startX, startY int
	endX, endY     int
	criticalPath   []*GridRoom
	rand           *rand.Rand
}

// NewGridMaze creates a new grid maze generator
func NewGridMaze() *GridMaze {
	return &GridMaze{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *GridMaze) Generate2D(xMax, yMax int, seed ...string) [][]*GridRoom {
	// Set up random generator with seed if provided
	if len(seed) > 0 && seed[0] != "" {
		h := fnv.New64a()
		h.Write([]byte(seed[0]))
		m.rand = rand.New(rand.NewSource(int64(h.Sum64())))
	}
	m.rooms = make(Maze2D, yMax)
	for y := range m.rooms {
		m.rooms[y] = make([]*GridRoom, xMax)
	}

	// Create all rooms initially
	m.createAllRooms(xMax, yMax)

	// Generate maze using recursive backtracking
	m.generateMazeRecursiveBacktracking(xMax, yMax)

	// Set start and end points
	m.setStartAndEnd(xMax, yMax)

	// Calculate distances from start
	m.calculateDistancesFromStart()

	// Find critical path and set steps
	m.findCriticalPath()

	// Add additional connections to create more branching and dead ends
	m.addAdditionalConnections(xMax, yMax)

	// Remove some rooms randomly, but protect critical path and preserve dead ends
	m.removeRoomsProtectingPath(xMax, yMax)

	// Remove any orphaned rooms (not reachable from start)
	m.removeOrphanedRooms(xMax, yMax)

	// Recalculate after room removal and connection addition
	m.calculateDistancesFromStart()
	m.findCriticalPath()

	// Convert to interface slice
	result := make([][]*GridRoom, yMax)
	for y := range result {
		result[y] = make([]*GridRoom, xMax)
		for x := range result[y] {
			if m.rooms[y][x] != nil {
				result[y][x] = m.rooms[y][x]
			}
		}
	}

	return result
}

func (m *GridMaze) createAllRooms(xMax, yMax int) {
	// Create all rooms initially - maze algorithm will determine connectivity
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			m.rooms[y][x] = &GridRoom{
				x: x, y: y,
				connections: make([]*GridRoom, 0),
			}
		}
	}
}

func (m *GridMaze) setStartAndEnd(xMax, yMax int) {
	// Collect all available rooms
	availableRooms := make([][2]int, 0)
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			if m.rooms[y][x] != nil {
				availableRooms = append(availableRooms, [2]int{x, y})
			}
		}
	}

	if len(availableRooms) == 0 {
		return
	}

	// Randomly select start position
	startIdx := m.rand.Intn(len(availableRooms))
	m.startX, m.startY = availableRooms[startIdx][0], availableRooms[startIdx][1]
	m.rooms[m.startY][m.startX].isStart = true

	if len(availableRooms) == 1 {
		return
	}

	// For end position, use a weighted random selection that favors distant rooms
	// but still allows for some variety
	weightedCandidates := make([]weightedCandidate, 0)
	for _, pos := range availableRooms {
		x, y := pos[0], pos[1]
		if x == m.startX && y == m.startY {
			continue // Skip start position
		}

		// Calculate Manhattan distance from start
		dist := abs(x-m.startX) + abs(y-m.startY)
		// Weight increases with distance, but add base weight for variety
		weight := dist + 1 // Base weight of 1 ensures all rooms have some chance
		weightedCandidates = append(weightedCandidates, weightedCandidate{x: x, y: y, weight: weight})
	}

	if len(weightedCandidates) > 0 {
		selected := m.selectWeightedRandom(weightedCandidates)
		m.endX, m.endY = selected.x, selected.y
		m.rooms[m.endY][m.endX].isEnd = true
	}
}

// weightedCandidate represents a position with an associated weight for selection
type weightedCandidate struct {
	x, y   int
	weight int
}

// selectWeightedRandom selects a candidate based on weighted probability
func (m *GridMaze) selectWeightedRandom(candidates []weightedCandidate) weightedCandidate {
	// Calculate total weight
	totalWeight := 0
	for _, candidate := range candidates {
		totalWeight += candidate.weight
	}

	// Select random value within total weight
	randomValue := m.rand.Intn(totalWeight)

	// Find the candidate that corresponds to this random value
	currentWeight := 0
	for _, candidate := range candidates {
		currentWeight += candidate.weight
		if randomValue < currentWeight {
			return candidate
		}
	}

	// Fallback (should never reach here)
	return candidates[len(candidates)-1]
}

func (m *GridMaze) generateMazeRecursiveBacktracking(xMax, yMax int) {
	// Recursive backtracking algorithm for maze generation
	visited := make([][]bool, yMax)
	for y := range visited {
		visited[y] = make([]bool, xMax)
	}

	// Start from a random position
	startX := m.rand.Intn(xMax)
	startY := m.rand.Intn(yMax)

	m.backtrack(startX, startY, visited, xMax, yMax)
}

func (m *GridMaze) backtrack(x, y int, visited [][]bool, xMax, yMax int) {
	visited[y][x] = true

	// Get all valid neighbors in random order
	directions := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	m.shuffleDirections(directions)

	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]

		// Check if neighbor is valid and unvisited
		if nx >= 0 && nx < xMax && ny >= 0 && ny < yMax && !visited[ny][nx] {
			// Connect current room to neighbor
			m.rooms[y][x].addConnection(m.rooms[ny][nx])

			// Recursively visit neighbor
			m.backtrack(nx, ny, visited, xMax, yMax)
		}
	}
}

func (m *GridMaze) shuffleDirections(directions [][2]int) {
	for i := len(directions) - 1; i > 0; i-- {
		j := m.rand.Intn(i + 1)
		directions[i], directions[j] = directions[j], directions[i]
	}
}

func (m *GridMaze) removeRoom(x, y int) {
	if m.rooms[y][x] == nil {
		return
	}

	// Remove all connections to this room
	for _, conn := range m.rooms[y][x].connections {
		for i, backConn := range conn.connections {
			if backConn == m.rooms[y][x] {
				conn.connections = append(conn.connections[:i], conn.connections[i+1:]...)
				break
			}
		}
	}

	// Remove the room
	m.rooms[y][x] = nil
}

func (m *GridMaze) removeRoomsProtectingPath(xMax, yMax int) {
	// Create a set of critical path rooms to protect
	criticalPathSet := make(map[*GridRoom]bool)
	for _, room := range m.criticalPath {
		criticalPathSet[room] = true
	}

	// Remove some rooms
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			room := m.rooms[y][x]
			if room == nil {
				continue
			}

			if criticalPathSet[room] {
				continue
			}

			if m.rand.Float32() < 0.5 {
				// Only remove rooms that are not adjacent to critical path
				if m.canRemoveWithoutDisconnectingPath(x, y) {
					m.removeRoom(x, y)
				}
			}

		}
	}
}

func (m *GridMaze) canRemoveWithoutDisconnectingPath(x, y int) bool {
	room := m.rooms[y][x]
	if room == nil {
		return false
	}

	// Don't remove rooms that are directly on the critical path
	if room.GetStep() > 0 {
		return false
	}

	// Don't remove rooms that would isolate other rooms
	// Check if removing this room would create orphaned areas
	if len(room.connections) > 1 {
		// Temporarily remove connections and check connectivity
		originalConnections := make([]*GridRoom, len(room.connections))
		copy(originalConnections, room.connections)

		// If this room bridges different areas, keep it
		for i, conn1 := range room.connections {
			for j, conn2 := range room.connections {
				if i != j && !m.areConnectedWithoutRoom(conn1, conn2, room) {
					return false
				}
			}
		}
	}

	return true
}

func (m *GridMaze) removeOrphanedRooms(xMax, yMax int) {
	if m.rooms[m.startY][m.startX] == nil {
		return
	}

	// Use BFS to find all rooms reachable from the start room
	reachable := make(map[*GridRoom]bool)
	queue := []*GridRoom{m.rooms[m.startY][m.startX]}
	reachable[m.rooms[m.startY][m.startX]] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, conn := range current.connections {
			if !reachable[conn] {
				reachable[conn] = true
				queue = append(queue, conn)
			}
		}
	}

	// Remove any rooms that are not reachable from the start
	for x := 0; x < xMax; x++ {
		for y := 0; y < yMax; y++ {
			room := m.rooms[y][x]
			if room != nil && !reachable[room] {
				m.removeRoom(x, y)
			}
		}
	}
}

func (m *GridMaze) calculateDistancesFromStart() {
	if m.rooms[m.startY][m.startX] == nil {
		return
	}

	// Use BFS to calculate distances from start
	queue := []*GridRoom{m.rooms[m.startY][m.startX]}
	visited := make(map[*GridRoom]bool)

	m.rooms[m.startY][m.startX].distanceFromStart = 0
	visited[m.rooms[m.startY][m.startX]] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, conn := range current.connections {
			if !visited[conn] {
				visited[conn] = true
				conn.distanceFromStart = current.distanceFromStart + 1
				queue = append(queue, conn)
			}
		}
	}
}

func (m *GridMaze) findCriticalPath() {
	if m.rooms[m.startY][m.startX] == nil || m.rooms[m.endY][m.endX] == nil {
		return
	}

	// Use BFS to find shortest path
	queue := []*GridRoom{m.rooms[m.startY][m.startX]}
	visited := make(map[*GridRoom]bool)
	parent := make(map[*GridRoom]*GridRoom)

	visited[m.rooms[m.startY][m.startX]] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == m.rooms[m.endY][m.endX] {
			// Reconstruct path
			path := make([]*GridRoom, 0)
			for room := current; room != nil; room = parent[room] {
				path = append([]*GridRoom{room}, path...)
			}

			// Set steps
			for i, room := range path {
				room.step = i + 1
			}

			m.criticalPath = path
			return
		}

		for _, conn := range current.connections {
			if !visited[conn] {
				visited[conn] = true
				parent[conn] = current
				queue = append(queue, conn)
			}
		}
	}
}

func (m *GridMaze) GetStart() (int, int) {
	return m.startX, m.startY
}

func (m *GridMaze) GetEnd() (int, int) {
	return m.endX, m.endY
}

func (m *GridMaze) GetCriticalPath() []MazeRoom {
	path := make([]MazeRoom, len(m.criticalPath))
	for i, room := range m.criticalPath {
		path[i] = room
	}
	return path
}

// areConnectedWithoutRoom checks if two rooms are connected without going through a specific room
func (m *GridMaze) areConnectedWithoutRoom(room1, room2, excludeRoom *GridRoom) bool {
	if room1 == room2 {
		return true
	}

	visited := make(map[*GridRoom]bool)
	queue := []*GridRoom{room1}
	visited[room1] = true
	visited[excludeRoom] = true // Exclude the room we're testing

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == room2 {
			return true
		}

		for _, conn := range current.connections {
			if !visited[conn] {
				visited[conn] = true
				queue = append(queue, conn)
			}
		}
	}

	return false
}

// addAdditionalConnections creates more branching paths to increase dead ends
func (m *GridMaze) addAdditionalConnections(xMax, yMax int) {
	// Add some additional connections to create more complex paths
	// This increases the number of potential dead ends
	connectionAttempts := (xMax * yMax) / 10 // Attempt connections for 10% of grid size

	for i := 0; i < connectionAttempts; i++ {
		x := m.rand.Intn(xMax)
		y := m.rand.Intn(yMax)

		if m.rooms[y][x] == nil {
			continue
		}

		// Try to connect to a nearby room that's not already connected
		directions := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
		m.shuffleDirections(directions)

		for _, dir := range directions {
			nx, ny := x+dir[0], y+dir[1]

			if nx >= 0 && nx < xMax && ny >= 0 && ny < yMax &&
				m.rooms[ny][nx] != nil && !m.rooms[y][x].IsConnectedTo(m.rooms[ny][nx]) {

				// Only add connection if it doesn't create too dense connectivity
				if len(m.rooms[y][x].connections) < 3 && len(m.rooms[ny][nx].connections) < 3 {
					m.rooms[y][x].addConnection(m.rooms[ny][nx])
					break
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
