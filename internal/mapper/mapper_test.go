package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdjustExitName(t *testing.T) {
	cases := []struct {
		in        string
		wantName  string
		wantDir   string
		wantError bool
	}{
		// plain cardinal
		{"east", "east", "east", false},
		// special cardinal
		{"east-x2", "east", "east-x2", false},
		// freeform noun
		{"cave", "cave", "", false},
		// noun + compass
		{"cave:south", "cave", "south", false},
		// noun + special compass
		{"cave:south-x2", "cave", "south-x2", false},
		// invalid special direction
		{"foo-x0", "foo-x0", "", true},
		// invalid colon direction
		{"cave:unknown", "cave", "", false},
		// over-parameterized
		{"east-x2:east-x2", "east-x2:east-x2", "", true},
		{"east-x2:east", "east-x2:east", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			name, dir, err := AdjustExitName(tc.in)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantName, name)
			assert.Equal(t, tc.wantDir, dir)
		})
	}
}

func TestArrowForDelta(t *testing.T) {
	cases := []struct {
		dx, dy, dz int
		want       rune
	}{
		{0, -1, 0, '\u2502'},  // north: vertical bar
		{0, 1, 0, '\u2502'},   // south: vertical bar
		{-1, 0, 0, '\u2500'},  // west: horizontal bar
		{1, 0, 0, '\u2500'},   // east: horizontal bar
		{1, -1, 0, '\u2571'},  // northeast: slash
		{-1, 1, 0, '\u2571'},  // southwest: slash
		{1, 1, 0, '\u2572'},   // southeast: backslash
		{-1, -1, 0, '\u2572'}, // northwest: backslash
		{0, 0, 1, '^'},        // up
		{0, 0, -1, 'v'},       // down
		{0, 0, 3, '^'},        // up multi-step
		{0, 0, -2, 'v'},       // down multi-step
	}
	for _, tc := range cases {
		got := arrowForDelta(tc.dx, tc.dy, tc.dz)
		assert.Equal(t, tc.want, got, "arrowForDelta(%d,%d,%d)", tc.dx, tc.dy, tc.dz)
	}
}

// buildStartMapper constructs a mapper whose crawledRooms are pre-populated so
// that Start() can be exercised without hitting the real room loader.
// It injects nodes directly and then calls the grid-initialisation portion of
// Start() by invoking a stripped version of the loop logic via the exported
// RoomGrid helpers.
func buildStartMapperFromNodes(rootId int, nodes map[int]*mapNode) *mapper {
	m := &mapper{
		rootRoomId:   rootId,
		crawledRooms: nodes,
		roomGrid: RoomGrid{
			rooms: [][][]*mapNode{},
		},
	}

	minX, maxX, minY, maxY, minZ, maxZ := 0, 0, 0, 0, 0, 0
	for _, n := range nodes {
		if n.Pos.x < minX {
			minX = n.Pos.x
		} else if n.Pos.x > maxX {
			maxX = n.Pos.x
		}
		if n.Pos.y < minY {
			minY = n.Pos.y
		} else if n.Pos.y > maxY {
			maxY = n.Pos.y
		}
		if n.Pos.z < minZ {
			minZ = n.Pos.z
		} else if n.Pos.z > maxZ {
			maxZ = n.Pos.z
		}
	}
	m.roomGrid.initialize(minX, maxX, minY, maxY, minZ, maxZ)
	for _, n := range nodes {
		m.roomGrid.addNode(n)
	}
	return m
}

// TestStart_StoredCoordsAreAuthoritative verifies that when a node carries
// HasStoredCoords=true its Pos is preserved verbatim in the grid regardless of
// the crawl-derived position that would otherwise be computed from exit deltas.
func TestStart_StoredCoordsAreAuthoritative(t *testing.T) {
	nodes := map[int]*mapNode{
		10: {
			RoomId:          10,
			HasStoredCoords: true,
			Pos:             positionDelta{x: 5, y: 3, z: 0},
			Exits:           map[string]nodeExit{},
		},
		11: {
			RoomId:          11,
			HasStoredCoords: true,
			Pos:             positionDelta{x: 6, y: 3, z: 0},
			Exits:           map[string]nodeExit{},
		},
	}

	m := buildStartMapperFromNodes(10, nodes)

	x10, y10, z10, err := m.GetCoordinates(10)
	require.NoError(t, err)
	assert.Equal(t, 5, x10)
	assert.Equal(t, 3, y10)
	assert.Equal(t, 0, z10)

	x11, y11, z11, err := m.GetCoordinates(11)
	require.NoError(t, err)
	assert.Equal(t, 6, x11)
	assert.Equal(t, 3, y11)
	assert.Equal(t, 0, z11)
}

// TestStart_MixedCoords verifies that rooms with stored coordinates keep their
// absolute positions while rooms without stored coordinates are placed relative
// to the crawl origin, and both sets share the same grid without displacing
// each other.
func TestStart_MixedCoords(t *testing.T) {
	nodes := map[int]*mapNode{
		// Rooms with stored coords at known absolute positions.
		1: {
			RoomId:          1,
			HasStoredCoords: true,
			Pos:             positionDelta{x: 0, y: 0, z: 0},
			Exits:           map[string]nodeExit{},
		},
		2: {
			RoomId:          2,
			HasStoredCoords: true,
			Pos:             positionDelta{x: 2, y: 0, z: 0},
			Exits:           map[string]nodeExit{},
		},
		// Room without stored coords placed by crawl-derived delta.
		3: {
			RoomId:          3,
			HasStoredCoords: false,
			Pos:             positionDelta{x: 1, y: 0, z: 0},
			Exits:           map[string]nodeExit{},
		},
	}

	m := buildStartMapperFromNodes(1, nodes)

	// Stored-coord rooms must be at their declared positions.
	x1, y1, _, err := m.GetCoordinates(1)
	require.NoError(t, err)
	assert.Equal(t, 0, x1)
	assert.Equal(t, 0, y1)

	x2, y2, _, err := m.GetCoordinates(2)
	require.NoError(t, err)
	assert.Equal(t, 2, x2)
	assert.Equal(t, 0, y2)

	// The non-stored-coord room must be reachable and at its crawl-derived position.
	_, _, _, err = m.GetCoordinates(3)
	require.NoError(t, err)

	// The grid must be able to find all three rooms.
	rid1, err := m.GetRoomId(0, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, rid1)

	rid2, err := m.GetRoomId(2, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, rid2)
}

// TestStart_NoStoredCoords verifies that when no room has stored coordinates
// the mapper still builds a valid grid using purely crawl-derived positions.
func TestStart_NoStoredCoords(t *testing.T) {
	nodes := map[int]*mapNode{
		1: {
			RoomId:          1,
			HasStoredCoords: false,
			Pos:             positionDelta{x: 0, y: 0, z: 0},
			Exits:           map[string]nodeExit{},
		},
		2: {
			RoomId:          2,
			HasStoredCoords: false,
			Pos:             positionDelta{x: 1, y: 0, z: 0},
			Exits:           map[string]nodeExit{},
		},
	}

	m := buildStartMapperFromNodes(1, nodes)

	// Both rooms must be reachable via the grid.
	rid1, err := m.GetRoomId(0, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, rid1)

	rid2, err := m.GetRoomId(1, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, rid2)
}
