package automission

import (
	"errors"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/telemetry"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// MissionType identifies the kind of objective.
type MissionType string

const (
	MissionTypeKillMob  MissionType = "kill_mob"
	MissionTypeFindItem MissionType = "find_item"
	MissionTypeExplore  MissionType = "explore"
	MissionTypeEscort   MissionType = "escort"
)

// boardTagForType maps each mission type to the room tag that enables it.
var boardTagForType = map[MissionType]string{
	MissionTypeKillMob:  tagKillMob,
	MissionTypeFindItem: tagFindItem,
	MissionTypeExplore:  tagExplore,
	MissionTypeEscort:   tagEscort,
}

// MissionDifficulty is either easy or hard.
type MissionDifficulty string

const (
	DifficultyEasy MissionDifficulty = "easy"
	DifficultyHard MissionDifficulty = "hard"
)

// Mission holds all state for one active mission board entry.
type Mission struct {
	Id          int
	BoardRoomId int // room the mission was generated for; turn-in must happen here
	Type        MissionType
	Difficulty  MissionDifficulty
	Title       string
	Description string

	TargetMobId     int
	TargetItemId    int
	TargetRoomId    int
	TargetZone      string
	EscortMobSpecId int

	EscortMobs     map[int]int    // userId -> mobInstanceId
	EscortExpiries map[int]uint64 // userId -> round when escort expires

	Reward RewardConfig

	AcceptedBy    []int
	CompletedBy   []int
	ReadyToTurnIn []int
}

// generateMissionsForBoard creates up to count missions for a single board room,
// restricting generation to the mission types whose tags are present on that room.
// Slots are distributed evenly across available types first, then remaining
// slots are filled randomly. A generator is retired once it returns nil.
func (m *AutoMissionModule) generateMissionsForBoard(boardRoom *rooms.Room, count int) []*Mission {
	type generatorFunc func(MissionDifficulty, *usedTargets) *Mission

	type namedGen struct {
		gen  generatorFunc
		fail int // consecutive failures
	}

	var gens []namedGen

	if boardRoom.HasTag(tagKillMob) {
		gens = append(gens, namedGen{gen: m.generateKillMob})
	}
	if boardRoom.HasTag(tagFindItem) {
		gens = append(gens, namedGen{gen: m.generateFindItem})
	}
	if boardRoom.HasTag(tagExplore) {
		gens = append(gens, namedGen{gen: m.generateExploreRoom(boardRoom.RoomId)})
	}
	if boardRoom.HasTag(tagEscort) && len(m.escortMobIds()) > 0 {
		gens = append(gens, namedGen{gen: m.generateEscort(boardRoom.RoomId)})
	}

	if len(gens) == 0 {
		return nil
	}

	difficulties := []MissionDifficulty{DifficultyEasy, DifficultyHard}
	used := &usedTargets{
		mobIds:  make(map[int]bool),
		itemIds: make(map[int]bool),
		roomIds: make(map[int]bool),
	}

	var result []*Mission

	// Phase 1: round-robin across all available generators to ensure even
	// distribution. Each generator gets one turn per round until count is met
	// or all generators are exhausted.
	for len(result) < count && len(gens) > 0 {
		// Shuffle each round so the order varies.
		for i := len(gens) - 1; i > 0; i-- {
			j := util.Rand(i + 1)
			gens[i], gens[j] = gens[j], gens[i]
		}

		progress := false
		i := 0
		for i < len(gens) && len(result) < count {
			diff := difficulties[util.Rand(len(difficulties))]
			mission := gens[i].gen(diff, used)
			if mission == nil {
				gens[i].fail++
				if gens[i].fail >= 3 {
					// Generator exhausted — retire it.
					gens = append(gens[:i], gens[i+1:]...)
					continue
				}
			} else {
				gens[i].fail = 0
				m.missionIdCounter++
				mission.Id = m.missionIdCounter
				mission.BoardRoomId = boardRoom.RoomId
				mission.EscortMobs = make(map[int]int)
				mission.EscortExpiries = make(map[int]uint64)
				mission.Reward = m.selectReward(mission.Type, diff)
				result = append(result, mission)
				progress = true
			}
			i++
		}
		// If no generator produced anything this round, stop.
		if !progress {
			break
		}
	}

	return result
}

// usedTargets tracks which mob IDs, item IDs, and room IDs have already been
// assigned to missions on the current board to prevent repeats.
type usedTargets struct {
	mobIds  map[int]bool
	itemIds map[int]bool
	roomIds map[int]bool
}

// generateKillMob creates a kill_mob mission using telemetry.
// Falls back to all known mob specs when telemetry has no kill data.
func (m *AutoMissionModule) generateKillMob(diff MissionDifficulty, used *usedTargets) *Mission {
	results := telemetry.Query().
		Category(telemetry.CatMobKill).
		GroupBy(telemetry.GroupByMobId).
		SortDesc().
		Results()

	var pool []int

	if len(results) == 0 {
		// No telemetry yet — use all known mob specs as candidates.
		for _, mob := range mobs.GetAllMobInfo() {
			id := int(mob.MobId)
			if id > 0 && !used.mobIds[id] {
				pool = append(pool, id)
			}
		}
	} else {
		bucket := len(results) / 10
		if bucket < 1 {
			bucket = 1
		}
		if diff == DifficultyEasy {
			for i := 0; i < bucket; i++ {
				id := results[i].MobId
				if !used.mobIds[id] {
					pool = append(pool, id)
				}
			}
		} else {
			start := len(results) - bucket
			for i := start; i < len(results); i++ {
				id := results[i].MobId
				if !used.mobIds[id] {
					pool = append(pool, id)
				}
			}
		}
	}

	if len(pool) == 0 {
		return nil
	}

	mobId := pool[util.Rand(len(pool))]
	spec := mobs.GetMobSpec(mobs.MobId(mobId))
	if spec == nil {
		return nil
	}

	used.mobIds[mobId] = true
	name := spec.Character.Name

	var title string
	if diff == DifficultyHard {
		title = `Defeat a <ansi fg="mobname">` + name + `</ansi> (Hard)`
	} else {
		title = `Defeat a <ansi fg="mobname">` + name + `</ansi>`
	}

	return &Mission{
		Type:        MissionTypeKillMob,
		Difficulty:  diff,
		Title:       title,
		Description: `Hunt down and kill a <ansi fg="mobname">` + name + `</ansi>.`,
		TargetMobId: mobId,
	}
}

// generateFindItem creates a find_item mission using telemetry.
// Skips quest items and already-used item IDs.
func (m *AutoMissionModule) generateFindItem(diff MissionDifficulty, used *usedTargets) *Mission {
	results := telemetry.Query().
		Category(telemetry.CatItemDrop).
		GroupBy(telemetry.GroupByItemId).
		SortDesc().
		Results()

	if len(results) == 0 {
		return nil
	}

	bucket := len(results) / 10
	if bucket < 1 {
		bucket = 1
	}

	// Build a candidate list from the appropriate difficulty bucket,
	// excluding quest items and already-used item IDs.
	var pool []int
	if diff == DifficultyEasy {
		for i := 0; i < bucket; i++ {
			id := results[i].ItemId
			if used.itemIds[id] {
				continue
			}
			spec := items.GetItemSpec(id)
			if spec == nil || spec.QuestToken != "" {
				continue
			}
			pool = append(pool, id)
		}
	} else {
		start := len(results) - bucket
		for i := start; i < len(results); i++ {
			id := results[i].ItemId
			if used.itemIds[id] {
				continue
			}
			spec := items.GetItemSpec(id)
			if spec == nil || spec.QuestToken != "" {
				continue
			}
			pool = append(pool, id)
		}
	}
	if len(pool) == 0 {
		return nil
	}

	itemId := pool[util.Rand(len(pool))]
	spec := items.GetItemSpec(itemId)
	if spec == nil {
		return nil
	}

	used.itemIds[itemId] = true
	name := spec.Name

	var title string
	if diff == DifficultyHard {
		title = `Acquire a <ansi fg="itemname">` + name + `</ansi> (Hard)`
	} else {
		title = `Acquire a <ansi fg="itemname">` + name + `</ansi>`
	}

	return &Mission{
		Type:         MissionTypeFindItem,
		Difficulty:   diff,
		Title:        title,
		Description:  `Find and bring back a <ansi fg="itemname">` + name + `</ansi>.`,
		TargetItemId: itemId,
	}
}

// generateExploreRoom returns a generator that picks a reachable, unused room.
func (m *AutoMissionModule) generateExploreRoom(boardRoomId int) func(MissionDifficulty, *usedTargets) *Mission {
	return func(diff MissionDifficulty, used *usedTargets) *Mission {
		zoneNames := rooms.GetAllZoneNames()
		if len(zoneNames) == 0 {
			return nil
		}

		shuffled := make([]string, len(zoneNames))
		copy(shuffled, zoneNames)
		for i := len(shuffled) - 1; i > 0; i-- {
			j := util.Rand(i + 1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}

		var candidates []roomCandidate
		zonesChecked := 0
		for _, zone := range shuffled {
			if zonesChecked >= 5 {
				break
			}
			zonesChecked++

			roomIds := rooms.GetAllZoneRoomsIds(zone)
			if len(roomIds) == 0 {
				continue
			}

			sample := roomIds
			if len(sample) > 20 {
				perm := make([]int, len(roomIds))
				copy(perm, roomIds)
				for i := len(perm) - 1; i > 0; i-- {
					j := util.Rand(i + 1)
					perm[i], perm[j] = perm[j], perm[i]
				}
				sample = perm[:20]
			}

			for _, rid := range sample {
				if used.roomIds[rid] {
					continue
				}
				steps, err := mapper.GetPath(boardRoomId, rid)
				if err != nil {
					if errors.Is(err, mapper.ErrPathNotFound) || errors.Is(err, mapper.ErrPathDestMatch) {
						continue
					}
					continue
				}
				candidates = append(candidates, roomCandidate{roomId: rid, dist: len(steps)})
			}
		}

		if len(candidates) == 0 {
			return nil
		}

		median := medianDist(candidates)

		var easy, hard []roomCandidate
		for _, c := range candidates {
			if c.dist < median {
				easy = append(easy, c)
			} else {
				hard = append(hard, c)
			}
		}

		pool := easy
		if diff == DifficultyHard {
			pool = hard
		}
		if len(pool) == 0 {
			pool = candidates
		}

		chosen := pool[util.Rand(len(pool))]
		room := rooms.LoadRoom(chosen.roomId)
		if room == nil {
			return nil
		}

		used.roomIds[chosen.roomId] = true

		var title string
		if diff == DifficultyHard {
			title = `Explore <ansi fg="cyan">` + room.Title + `</ansi> (Hard)`
		} else {
			title = `Explore <ansi fg="cyan">` + room.Title + `</ansi>`
		}

		return &Mission{
			Type:         MissionTypeExplore,
			Difficulty:   diff,
			Title:        title,
			Description:  `Travel to <ansi fg="cyan">` + room.Title + `</ansi> in the ` + room.Zone + ` region.`,
			TargetRoomId: chosen.roomId,
		}
	}
}

// generateEscort returns a generator that creates escort missions whose
// destination zone is reachable from boardRoomId and is not the same zone
// as the board room itself.
func (m *AutoMissionModule) generateEscort(boardRoomId int) func(MissionDifficulty, *usedTargets) *Mission {
	boardRoom := rooms.LoadRoom(boardRoomId)
	boardZone := ""
	if boardRoom != nil {
		boardZone = boardRoom.Zone
	}

	return func(diff MissionDifficulty, used *usedTargets) *Mission {
		ids := m.escortMobIds()
		if len(ids) == 0 {
			return nil
		}

		// Shuffle so we don't always try the same mob first.
		shuffled := make([]int, len(ids))
		copy(shuffled, ids)
		for i := len(shuffled) - 1; i > 0; i-- {
			j := util.Rand(i + 1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}

		for _, specId := range shuffled {
			if used.mobIds[specId] {
				continue
			}

			spec := mobs.GetMobSpec(mobs.MobId(specId))
			if spec == nil {
				continue
			}

			zone := spec.Zone

			if zone == "" {
				continue
			}

			// Skip mobs whose destination is the same zone as the board room.
			if boardZone != "" && zone == boardZone {
				continue
			}

			if !zoneIsReachable(boardRoomId, zone) {
				continue
			}

			name := spec.Character.Name
			used.mobIds[specId] = true
			var title string
			if diff == DifficultyHard {
				title = `Escort <ansi fg="mobname">` + name + `</ansi> to ` + zone + ` (Hard)`
			} else {
				title = `Escort <ansi fg="mobname">` + name + `</ansi> to ` + zone
			}

			return &Mission{
				Type:            MissionTypeEscort,
				Difficulty:      diff,
				Title:           title,
				Description:     `<ansi fg="mobname">` + name + `</ansi> needs to reach ` + zone + `. Keep them safe and get them there in time.`,
				TargetZone:      zone,
				EscortMobSpecId: specId,
			}
		}

		return nil
	}
}

// zoneIsReachable returns true if any room in zone has a path from boardRoomId.
// Shuffles the candidate rooms so the same rooms aren't always checked first.
func zoneIsReachable(boardRoomId int, zone string) bool {
	roomIds := rooms.GetAllZoneRoomsIds(zone)
	if len(roomIds) == 0 {
		return false
	}
	// Shuffle and sample up to 10 rooms.
	perm := make([]int, len(roomIds))
	copy(perm, roomIds)
	for i := len(perm) - 1; i > 0; i-- {
		j := util.Rand(i + 1)
		perm[i], perm[j] = perm[j], perm[i]
	}
	sample := perm
	if len(sample) > 10 {
		sample = perm[:10]
	}
	for _, rid := range sample {
		_, err := mapper.GetPath(boardRoomId, rid)
		if err == nil {
			return true
		}
	}
	return false
}

// escortMobIds reads the EscortMobIds config key and returns []int.
func (m *AutoMissionModule) escortMobIds() []int {
	raw := m.configSlice("EscortMobIds")
	if len(raw) == 0 {
		return nil
	}
	ids := make([]int, 0, len(raw))
	for _, v := range raw {
		if id := toInt(v); id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

// medianDist returns the median distance from a roomCandidate slice.
func medianDist(cs []roomCandidate) int {
	if len(cs) == 0 {
		return 0
	}
	sorted := make([]int, len(cs))
	for i, c := range cs {
		sorted[i] = c.dist
	}
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	return sorted[len(sorted)/2]
}

type roomCandidate struct {
	roomId int
	dist   int
}
