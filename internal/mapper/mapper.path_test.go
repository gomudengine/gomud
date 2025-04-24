package mapper

import (
	"testing"
)

// buildTestMapper creates a simple 3-room chain: 1 -east-> 2 -south-> 3
func buildTestMapper() *mapper {
	m := NewMapper(1)
	m.crawledRooms = map[int]*mapNode{
		1: &mapNode{RoomId: 1, Pos: positionDelta{x: 0, y: 0, z: 0}, Exits: map[string]nodeExit{"east": {RoomId: 2}}},
		2: &mapNode{RoomId: 2, Pos: positionDelta{x: 1, y: 0, z: 0}, Exits: map[string]nodeExit{"south": {RoomId: 3}}},
		3: &mapNode{RoomId: 3, Pos: positionDelta{x: 1, y: 1, z: 0}, Exits: map[string]nodeExit{}},
	}
	return m
}

// buildCycleMapper creates a graph with a cycle: 1 <-> 2 -> 3
func buildCycleMapper() *mapper {
	m := NewMapper(1)
	m.crawledRooms = map[int]*mapNode{
		1: &mapNode{RoomId: 1, Pos: positionDelta{x: 0, y: 0, z: 0}, Exits: map[string]nodeExit{"to2": {RoomId: 2}}},
		2: &mapNode{RoomId: 2, Pos: positionDelta{x: 1, y: 0, z: 0}, Exits: map[string]nodeExit{"to1": {RoomId: 1}, "to3": {RoomId: 3}}},
		3: &mapNode{RoomId: 3, Pos: positionDelta{x: 2, y: 0, z: 0}, Exits: map[string]nodeExit{}},
	}
	return m
}

func Test_findPath_DestMatch(t *testing.T) {
	m := buildTestMapper()
	_, err := m.findPath(1, 1)
	if err != ErrPathDestMatch {
		t.Fatalf("expected ErrPathDestMatch, got %v", err)
	}
}

func Test_findPath_RoomNotFound(t *testing.T) {
	m := NewMapper(1)
	m.crawledRooms = map[int]*mapNode{
		1: &mapNode{RoomId: 1, Pos: positionDelta{}, Exits: map[string]nodeExit{}},
	}

	_, err := m.findPath(1, 2)
	if err != ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound for missing dest, got %v", err)
	}

	_, err = m.findPath(2, 1)
	if err != ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound for missing source, got %v", err)
	}
}

func Test_findPath_Simple(t *testing.T) {
	m := buildTestMapper()
	path, err := m.findPath(1, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(path) != 2 {
		t.Fatalf("expected path length 2, got %d", len(path))
	}

	if got := path[0]; got.exitName != "east" || got.roomId != 2 || got.waypoint {
		t.Errorf("step0 = %+v; want {ExitName:east RoomId:2 Waypoint:false}", got)
	}
	if got := path[1]; got.exitName != "south" || got.roomId != 3 || !got.waypoint {
		t.Errorf("step1 = %+v; want {ExitName:south RoomId:3 Waypoint:true}", got)
	}
}

func Test_findPath_Unreachable(t *testing.T) {
	m := NewMapper(1)
	m.crawledRooms = map[int]*mapNode{
		1: &mapNode{RoomId: 1, Pos: positionDelta{}, Exits: map[string]nodeExit{}},
		2: &mapNode{RoomId: 2, Pos: positionDelta{x: 1}, Exits: map[string]nodeExit{}},
	}

	_, err := m.findPath(1, 2)
	if err != ErrPathNotFound {
		t.Fatalf("expected ErrPathNotFound, got %v", err)
	}
}

func Test_findPath_Cycle(t *testing.T) {
	m := buildCycleMapper()
	path, err := m.findPath(1, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(path) != 2 {
		t.Fatalf("expected path length 2, got %d", len(path))
	}

	if got := path[0]; got.exitName != "to2" || got.roomId != 2 {
		t.Errorf("step0 = %+v; want {ExitName:to2 RoomId:2}", got)
	}
	if got := path[1]; got.exitName != "to3" || got.roomId != 3 {
		t.Errorf("step1 = %+v; want {ExitName:to3 RoomId:3}", got)
	}
}

func Test_GetPath_StartNotFound(t *testing.T) {
	_, err := GetPath(999999)
	if err != ErrPathNotFound {
		t.Fatalf("expected ErrPathNotFound for missing start, got %v", err)
	}
}
