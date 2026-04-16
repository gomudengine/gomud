package characters

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRoomBitset_SetAndHas(t *testing.T) {
	rb := make(RoomBitset)

	// Non-positive IDs are sentinel values and are always considered visited.
	assert.True(t, rb.Has(0))
	assert.True(t, rb.Has(-1))
	assert.False(t, rb.Has(63))
	assert.False(t, rb.Has(64))
	assert.False(t, rb.Has(512))

	rb.Set(0)  // sentinel — must be ignored (no-op)
	rb.Set(-1) // sentinel — must be ignored (no-op)
	rb.Set(63)
	rb.Set(64)
	rb.Set(512)

	assert.True(t, rb.Has(0))  // true — sentinel
	assert.True(t, rb.Has(-1)) // true — sentinel
	assert.True(t, rb.Has(63))
	assert.True(t, rb.Has(64))
	assert.True(t, rb.Has(512))

	// Neighbours should remain unset.
	assert.False(t, rb.Has(1))
	assert.False(t, rb.Has(62))
	assert.False(t, rb.Has(65))
	assert.False(t, rb.Has(511))
	assert.False(t, rb.Has(513))
}

func TestRoomBitset_Count(t *testing.T) {
	rb := make(RoomBitset)
	assert.Equal(t, 0, rb.Count())

	rb.Set(1)
	rb.Set(64)
	rb.Set(128)
	assert.Equal(t, 3, rb.Count())

	// Setting the same room twice should not change the count.
	rb.Set(1)
	assert.Equal(t, 3, rb.Count())
}

func TestRoomBitset_CountIn(t *testing.T) {
	rb := make(RoomBitset)
	rb.Set(1)
	rb.Set(2)
	rb.Set(300)

	valid := map[int]struct{}{
		1:   {},
		2:   {},
		3:   {},
		300: {},
	}

	assert.Equal(t, 3, rb.CountIn(valid))
}

func TestRoomBitset_IsComplete(t *testing.T) {
	rb := make(RoomBitset)

	zone := map[int]struct{}{
		10: {},
		11: {},
		12: {},
	}

	assert.False(t, rb.IsComplete(zone))

	rb.Set(10)
	rb.Set(11)
	assert.False(t, rb.IsComplete(zone))

	rb.Set(12)
	assert.True(t, rb.IsComplete(zone))
}

func TestRoomBitset_Prune(t *testing.T) {
	rb := make(RoomBitset)
	rb.Set(10)
	rb.Set(11)
	rb.Set(500) // will be pruned — not in valid set

	valid := map[int]struct{}{
		10: {},
		11: {},
	}

	rb.Prune(valid)

	assert.True(t, rb.Has(10))
	assert.True(t, rb.Has(11))
	assert.False(t, rb.Has(500))
	assert.Equal(t, 2, rb.Count())
}

func TestRoomBitset_Prune_EmptyBlockRemoved(t *testing.T) {
	rb := make(RoomBitset)
	rb.Set(640) // block 10
	rb.Set(641) // block 10

	// Valid set contains no rooms in block 10.
	valid := map[int]struct{}{
		1: {},
	}

	rb.Prune(valid)

	// Block 10 should be gone entirely.
	_, exists := rb[10]
	assert.False(t, exists)
}

func TestRoomBitset_YAMLRoundTrip(t *testing.T) {
	original := make(RoomBitset)
	original.Set(1)
	original.Set(63)
	original.Set(64)
	original.Set(300)
	original.Set(878)

	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	// Verify the output is human-readable hex, not a binary blob.
	assert.Contains(t, string(data), "0x")

	var restored RoomBitset
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.True(t, restored.Has(1))
	assert.True(t, restored.Has(63))
	assert.True(t, restored.Has(64))
	assert.True(t, restored.Has(300))
	assert.True(t, restored.Has(878))
	assert.False(t, restored.Has(2))
	assert.False(t, restored.Has(65))
	assert.Equal(t, original.Count(), restored.Count())
}

func TestRoomBitset_UnmarshalYAML_EmptyNode(t *testing.T) {
	var rb RoomBitset
	err := yaml.Unmarshal([]byte("null\n"), &rb)
	require.NoError(t, err)
	// A null node decodes to an empty map, which is valid and usable.
	assert.Equal(t, 0, rb.Count())
}

func TestMarkVisitedRoom_ZoneCompletion(t *testing.T) {
	c := New()

	zone := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}

	// Non-completing visits return false.
	assert.False(t, c.MarkVisitedRoom(1, "testzone", zone))
	assert.False(t, c.MarkVisitedRoom(2, "testzone", zone))

	// Visiting an already-visited room is never the completing visit.
	assert.False(t, c.MarkVisitedRoom(1, "testzone", zone))

	// The last unvisited room completes the zone.
	assert.True(t, c.MarkVisitedRoom(3, "testzone", zone))

	// Revisiting after completion still returns false.
	assert.False(t, c.MarkVisitedRoom(3, "testzone", zone))

	// Nil validRoomIds never signals completion.
	c2 := New()
	assert.False(t, c2.MarkVisitedRoom(1, "testzone", nil))
}

func TestRoomBitset_CharacterIntegration(t *testing.T) {
	c := New()

	assert.False(t, c.HasVisitedRoom(1, "frostfang"))

	c.MarkVisitedRoom(1, "frostfang", nil)
	c.MarkVisitedRoom(64, "frostfang", nil)
	c.MarkVisitedRoom(300, "dark_forest", nil)

	assert.True(t, c.HasVisitedRoom(1, "frostfang"))
	assert.True(t, c.HasVisitedRoom(64, "frostfang"))
	assert.False(t, c.HasVisitedRoom(2, "frostfang"))
	assert.True(t, c.HasVisitedRoom(300, "dark_forest"))
	assert.False(t, c.HasVisitedRoom(1, "dark_forest"))
}

func TestRoomBitset_ZoneVisitProgress(t *testing.T) {
	c := New()

	zone := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
	}

	visited, total := c.ZoneVisitProgress("frostfang", zone)
	assert.Equal(t, 0, visited)
	assert.Equal(t, 4, total)

	c.MarkVisitedRoom(1, "frostfang", nil)
	c.MarkVisitedRoom(3, "frostfang", nil)

	visited, total = c.ZoneVisitProgress("frostfang", zone)
	assert.Equal(t, 2, visited)
	assert.Equal(t, 4, total)
}

func TestCharacter_ZoneVisitPercent(t *testing.T) {
	tests := []struct {
		name         string
		visitedRooms []int
		zone         map[int]struct{}
		want         int
	}{
		{
			name:         "none visited",
			visitedRooms: []int{},
			zone:         map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}},
			want:         0,
		},
		{
			name:         "half visited",
			visitedRooms: []int{1, 2},
			zone:         map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}},
			want:         50,
		},
		{
			name:         "all visited",
			visitedRooms: []int{1, 2, 3, 4},
			zone:         map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}},
			want:         100,
		},
		{
			name:         "empty zone",
			visitedRooms: []int{},
			zone:         map[int]struct{}{},
			want:         0,
		},
		{
			name:         "one of three visited",
			visitedRooms: []int{1},
			zone:         map[int]struct{}{1: {}, 2: {}, 3: {}},
			want:         33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			for _, roomId := range tt.visitedRooms {
				c.MarkVisitedRoom(roomId, "testzone", nil)
			}
			got := c.ZoneVisitPercent("testzone", tt.zone)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoomBitset_CharacterYAMLRoundTrip(t *testing.T) {
	c := New()
	c.Name = "Tester"
	c.MarkVisitedRoom(1, "frostfang", nil)
	c.MarkVisitedRoom(300, "dark_forest", nil)

	data, err := yaml.Marshal(c)
	require.NoError(t, err)

	// The zonesvisited key must appear in the output.
	assert.Contains(t, string(data), "zonesvisited")

	var c2 Character
	err = yaml.Unmarshal(data, &c2)
	require.NoError(t, err)

	assert.True(t, c2.HasVisitedRoom(1, "frostfang"))
	assert.True(t, c2.HasVisitedRoom(300, "dark_forest"))
	assert.False(t, c2.HasVisitedRoom(2, "frostfang"))

	// Verify no extra bytes were introduced (idempotent marshal).
	data2, err := yaml.Marshal(&c2)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(data, data2))
}
