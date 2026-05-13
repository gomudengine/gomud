package automission

import (
	"embed"
	"fmt"
	"slices"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	//go:embed files/*
	files embed.FS
)

// One tag per mission type, plus a convenience check helper.
const (
	tagKillMob  = "missionboard kill_mob"
	tagFindItem = "missionboard find_item"
	tagExplore  = "missionboard explore"
	tagEscort   = "missionboard escort"

	acceptedMissionsKey = "automission_accepted"
)

// allBoardTags is the full set of tags this module reserves.
var allBoardTags = []string{tagKillMob, tagFindItem, tagExplore, tagEscort}

// isBoardRoom returns true if the room has at least one mission board tag.
func isBoardRoom(room *rooms.Room) bool {
	for _, tag := range allBoardTags {
		if room.HasTag(tag) {
			return true
		}
	}
	return false
}

// acceptedEntry records a mission ID together with the board room it came from,
// so turn-in can enforce that the player returns to the correct board.
type acceptedEntry struct {
	MissionId   int
	BoardRoomId int
}

func init() {
	m := &AutoMissionModule{
		plug:   plugins.New("automission", "1.0"),
		boards: make(map[int][]*Mission),
	}

	if err := m.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	m.plug.ReserveTags(allBoardTags...)

	m.plug.Web.AdminPage("Config", "automission-config",
		"html/admin/automission-config.html", true, "Modules", "Auto Mission", nil)
	m.plug.Web.AdminPage("About", "automission-about",
		"html/admin/automission-about.html", true, "Modules", "Auto Mission", nil)

	m.plug.Web.AdminAPIEndpoint("GET", "automission-config", m.apiGetConfig)
	m.plug.Web.AdminAPIEndpoint("PATCH", "automission-config", m.apiPatchConfig)

	m.plug.AddUserCommand("mission", m.missionCommand, true, false)

	rooms.OnRoomLook.Register(m.onRoomLook)

	events.RegisterListener(events.NewRound{}, m.onNewRound)
	events.RegisterListener(events.MobDeath{}, m.onMobDeath)
	events.RegisterListener(events.ItemOwnership{}, m.onItemOwnership)
	events.RegisterListener(events.RoomChange{}, m.onRoomChange)
	events.RegisterListener(events.PlayerDespawn{}, m.onPlayerDespawn)
}

// AutoMissionModule holds all runtime state for the mission board system.
type AutoMissionModule struct {
	plug             *plugins.Plugin
	boards           map[int][]*Mission // boardRoomId -> missions for that board
	lastRestockRound uint64
	missionIdCounter int
}

// userRecord is a convenience alias used throughout the module.
type userRecord = users.UserRecord

// allMissions returns a flat iterator over every mission across all boards.
func (m *AutoMissionModule) allMissions() []*Mission {
	var all []*Mission
	for _, ms := range m.boards {
		all = append(all, ms...)
	}
	return all
}

// onRoomLook injects a mission board alert when the room has any board tag.
func (m *AutoMissionModule) onRoomLook(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	for _, tag := range allBoardTags {
		for _, t := range d.Tags {
			if strings.EqualFold(t, tag) {
				d.RoomAlerts = append(d.RoomAlerts,
					`<ansi fg="yellow-bold">There's a mission board here!</ansi> <ansi fg="command">mission list</ansi> lists missions.`,
				)
				return d
			}
		}
	}
	return d
}

// ensureBoardMissions generates missions for a board room if none exist yet.
// Called lazily when a player interacts with a board that was missed during restock.
func (m *AutoMissionModule) ensureBoardMissions(room *rooms.Room) {
	if _, ok := m.boards[room.RoomId]; ok {
		return
	}
	maxMissions := 6
	if v, ok := m.plug.Config.Get("MaxMissions").(int); ok && v > 0 {
		maxMissions = v
	}
	missions := m.generateMissionsForBoard(room, maxMissions)
	if len(missions) > 0 {
		m.boards[room.RoomId] = missions
	}
}

// missionsForBoard returns the mission slice for a given board room,
// generating them on demand if the board has not been stocked yet.
func (m *AutoMissionModule) missionsForBoard(roomId int) []*Mission {
	if ms, ok := m.boards[roomId]; ok {
		return ms
	}
	room := rooms.LoadRoom(roomId)
	if room != nil && isBoardRoom(room) {
		m.ensureBoardMissions(room)
	}
	return m.boards[roomId]
}

// onNewRound handles restock timing and escort expiry each game round.
func (m *AutoMissionModule) onNewRound(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.NewRound)
	if !ok {
		return events.Continue
	}
	currentRound := evt.RoundNumber

	if m.lastRestockRound == 0 {
		m.restock(currentRound)
		return events.Continue
	}

	nextRestockAt := gametime.GetDate(m.lastRestockRound).AddPeriod(m.restockPeriod())
	if currentRound >= nextRestockAt {
		m.restock(currentRound)
		return events.Continue
	}

	m.checkEscortExpiry(currentRound)
	return events.Continue
}

// restock regenerates missions for every board room in the world.
func (m *AutoMissionModule) restock(currentRound uint64) {
	// Fail all active escort missions.
	for _, mission := range m.allMissions() {
		if mission.Type != MissionTypeEscort {
			continue
		}
		for userId, instanceId := range mission.EscortMobs {
			if mob := mobs.GetInstance(instanceId); mob != nil {
				mob.Command("emote collapses and perishes.;suicide vanish")
				if user := users.GetByUserId(userId); user != nil {
					user.Character.TrackCharmed(instanceId, false)
				}
			}
			if user := users.GetByUserId(userId); user != nil {
				user.SendText(buildBanner(mission.Title, "", "red"))
			}
		}
	}

	// Clear accepted state from all online users.
	for _, uid := range users.GetOnlineUserIds() {
		if user := users.GetByUserId(uid); user != nil {
			user.SetTempData(acceptedMissionsKey, nil)
		}
	}

	m.boards = make(map[int][]*Mission)

	maxMissions := 6
	if v, ok := m.plug.Config.Get("MaxMissions").(int); ok && v > 0 {
		maxMissions = v
	}

	// Discover all board rooms and generate missions for each one.
	boardRoomIds := m.findAllBoardRoomIds()
	for _, roomId := range boardRoomIds {
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}
		missions := m.generateMissionsForBoard(room, maxMissions)
		if len(missions) > 0 {
			m.boards[roomId] = missions
		}
	}

	m.lastRestockRound = currentRound

	// Notify players standing at any board room.
	for _, uid := range users.GetOnlineUserIds() {
		user := users.GetByUserId(uid)
		if user == nil {
			continue
		}
		room := rooms.LoadRoom(user.Character.RoomId)
		if room != nil && isBoardRoom(room) {
			user.SendText(`<ansi fg="yellow">The mission board has been refreshed with new missions!</ansi>`)
		}
	}
}

// checkEscortExpiry fails expired escort missions each round.
func (m *AutoMissionModule) checkEscortExpiry(currentRound uint64) {
	for _, mission := range m.allMissions() {
		if mission.Type != MissionTypeEscort {
			continue
		}
		for userId, instanceId := range mission.EscortMobs {
			expiry, hasExpiry := mission.EscortExpiries[userId]
			if !hasExpiry || currentRound < expiry {
				continue
			}
			if slices.Contains(mission.CompletedBy, userId) {
				continue
			}
			if mob := mobs.GetInstance(instanceId); mob != nil {
				mob.Command("emote collapses and perishes.;suicide vanish")
			}
			if user := users.GetByUserId(userId); user != nil {
				user.Character.TrackCharmed(instanceId, false)
				user.SendText(buildBanner(mission.Title, "", "red"))
				m.removeAcceptedEntry(user, mission.Id)
			}
			delete(mission.EscortMobs, userId)
			delete(mission.EscortExpiries, userId)
			mission.AcceptedBy = removeInt(mission.AcceptedBy, userId)
		}
	}
}

// onMobDeath handles kill_mob completion and escort mob death.
func (m *AutoMissionModule) onMobDeath(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.MobDeath)
	if !ok {
		return events.Continue
	}

	for _, mission := range m.allMissions() {
		if mission.Type == MissionTypeKillMob && mission.TargetMobId == evt.MobId {
			for _, userId := range evt.KilledByUsers {
				if slices.Contains(mission.AcceptedBy, userId) && !slices.Contains(mission.CompletedBy, userId) {
					m.markReadyToTurnIn(mission, userId)
				}
			}
		}

		if mission.Type == MissionTypeEscort {
			for userId, instanceId := range mission.EscortMobs {
				if instanceId == evt.InstanceId {
					m.failMission(mission, userId)
					delete(mission.EscortMobs, userId)
					delete(mission.EscortExpiries, userId)
				}
			}
		}
	}

	return events.Continue
}

// onItemOwnership handles find_item completion.
func (m *AutoMissionModule) onItemOwnership(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.ItemOwnership)
	if !ok || !evt.Gained || evt.UserId == 0 {
		return events.Continue
	}

	for _, mission := range m.allMissions() {
		if mission.Type != MissionTypeFindItem {
			continue
		}
		if evt.Item.ItemId != mission.TargetItemId {
			continue
		}
		if slices.Contains(mission.AcceptedBy, evt.UserId) && !slices.Contains(mission.CompletedBy, evt.UserId) {
			m.markReadyToTurnIn(mission, evt.UserId)
		}
	}

	return events.Continue
}

// onRoomChange handles explore completion and escort zone arrival.
func (m *AutoMissionModule) onRoomChange(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.RoomChange)
	if !ok || evt.UserId == 0 {
		return events.Continue
	}

	toRoom := rooms.LoadRoom(evt.ToRoomId)
	if toRoom == nil {
		return events.Continue
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	for _, mission := range m.allMissions() {
		if !slices.Contains(mission.AcceptedBy, evt.UserId) {
			continue
		}
		if slices.Contains(mission.CompletedBy, evt.UserId) {
			continue
		}

		switch mission.Type {
		case MissionTypeExplore:
			if evt.ToRoomId == mission.TargetRoomId {
				m.markReadyToTurnIn(mission, evt.UserId)
			}

		case MissionTypeEscort:
			instanceId, hasEscort := mission.EscortMobs[evt.UserId]
			if !hasEscort {
				continue
			}
			mob := mobs.GetInstance(instanceId)
			if mob == nil {
				continue
			}
			if mob.Character.IsCharmed(evt.UserId) && toRoom.Zone == mission.TargetZone {
				mob.Command(fmt.Sprintf(`say Thank you for escorting me! I am finally home!`))
				mob.Command(`emote bows graciously and disappears.;suicide vanish`, 1)
				user.Character.TrackCharmed(instanceId, false)
				delete(mission.EscortMobs, evt.UserId)
				delete(mission.EscortExpiries, evt.UserId)
				m.markReadyToTurnIn(mission, evt.UserId)
			}
		}
	}

	// After processing all mission completions, check whether the player
	// has just arrived at a board room with missions ready to turn in.
	if isBoardRoom(toRoom) {
		var readyTitles []string
		for _, mission := range m.missionsForBoard(toRoom.RoomId) {
			if slices.Contains(mission.ReadyToTurnIn, evt.UserId) {
				readyTitles = append(readyTitles, mission.Title)
			}
		}
		if len(readyTitles) > 0 {
			user.SendText(
				`<ansi fg="yellow-bold">You have completed missions ready to turn in here!</ansi>` +
					` Type <ansi fg="command">mission turn in</ansi> to claim your reward.`,
			)
		}
	}

	return events.Continue
}
func (m *AutoMissionModule) onPlayerDespawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerDespawn)
	if !ok {
		return events.Continue
	}

	for _, mission := range m.allMissions() {
		if mission.Type == MissionTypeEscort {
			if instanceId, has := mission.EscortMobs[evt.UserId]; has {
				if mob := mobs.GetInstance(instanceId); mob != nil {
					mob.Command("emote collapses and perishes.;suicide vanish")
				}
				delete(mission.EscortMobs, evt.UserId)
				delete(mission.EscortExpiries, evt.UserId)
			}
		}
		mission.AcceptedBy = removeInt(mission.AcceptedBy, evt.UserId)
		mission.ReadyToTurnIn = removeInt(mission.ReadyToTurnIn, evt.UserId)
	}

	return events.Continue
}

// markReadyToTurnIn adds the user to ReadyToTurnIn and notifies them.
func (m *AutoMissionModule) markReadyToTurnIn(mission *Mission, userId int) {
	if slices.Contains(mission.ReadyToTurnIn, userId) {
		return
	}
	mission.ReadyToTurnIn = append(mission.ReadyToTurnIn, userId)

	user := users.GetByUserId(userId)
	if user == nil {
		return
	}
	boardRoom := rooms.LoadRoom(mission.BoardRoomId)
	boardDesc := "a mission board"
	if boardRoom != nil {
		boardDesc = boardRoom.Title
	}
	banner := buildBanner("MISSION COMPLETE!", mission.Title, "yellow") +
		fmt.Sprintf("\n"+`<ansi fg="yellow">Return to <ansi fg="cyan">%s</ansi> to claim your reward.</ansi>`, boardDesc)
	user.SendText(banner)
}

// failMission notifies the user of failure and removes their accepted state.
func (m *AutoMissionModule) failMission(mission *Mission, userId int) {
	if user := users.GetByUserId(userId); user != nil {
		user.SendText(buildBanner("MISSION FAILED", mission.Title, "red"))
		m.removeAcceptedEntry(user, mission.Id)
	}
	mission.AcceptedBy = removeInt(mission.AcceptedBy, userId)
	mission.ReadyToTurnIn = removeInt(mission.ReadyToTurnIn, userId)
}

// spawnEscortMob spawns the escort mob and charms it to the user.
func (m *AutoMissionModule) spawnEscortMob(mission *Mission, user *userRecord, room *rooms.Room) {
	mob := mobs.NewMobById(mobs.MobId(mission.EscortMobSpecId), room.RoomId)
	if mob == nil {
		return
	}
	room.AddMob(mob.InstanceId)
	mob.Character.Charm(user.UserId, characters.CharmPermanent, "emote collapses and perishes.;suicide vanish")
	user.Character.TrackCharmed(mob.InstanceId, true)

	mission.EscortMobs[user.UserId] = mob.InstanceId

	gd := gametime.GetDate(util.GetRoundCount())
	mission.EscortExpiries[user.UserId] = gd.AddPeriod(m.escortTimeLimit())

	mob.Command(fmt.Sprintf(`say Thank you! Please escort me to %s!`, mission.TargetZone))
}

// restockPeriod returns the configured restock period string.
func (m *AutoMissionModule) restockPeriod() string {
	if v, ok := m.plug.Config.Get("RestockPeriod").(string); ok && v != "" {
		return v
	}
	return "1 real day"
}

// escortTimeLimit returns the configured escort time limit string.
func (m *AutoMissionModule) escortTimeLimit() string {
	if v, ok := m.plug.Config.Get("EscortTimeLimit").(string); ok && v != "" {
		return v
	}
	return "1 hour"
}

// findAllBoardRoomIds returns room IDs of every loaded room that has at least
// one mission board tag. Uses online players to seed room loading; falls back
// to zone root rooms so boards are discovered even with no players online.
func (m *AutoMissionModule) findAllBoardRoomIds() []int {
	seen := make(map[int]bool)
	var result []int

	check := func(roomId int) {
		if seen[roomId] {
			return
		}
		seen[roomId] = true
		room := rooms.LoadRoom(roomId)
		if room != nil && isBoardRoom(room) {
			result = append(result, roomId)
		}
	}

	for _, uid := range users.GetOnlineUserIds() {
		if user := users.GetByUserId(uid); user != nil {
			check(user.Character.RoomId)
		}
	}
	for _, zone := range rooms.GetAllZoneNames() {
		if rootId, err := rooms.GetZoneRoot(zone); err == nil {
			check(rootId)
		}
	}
	return result
}

// -------------------------------------------------------------------------
// Per-user accepted entry helpers
// -------------------------------------------------------------------------

func getAcceptedEntries(user *userRecord) []acceptedEntry {
	v := user.GetTempData(acceptedMissionsKey)
	if v == nil {
		return nil
	}
	if entries, ok := v.([]acceptedEntry); ok {
		return entries
	}
	return nil
}

func addAcceptedEntry(user *userRecord, missionId, boardRoomId int) {
	entries := getAcceptedEntries(user)
	entries = append(entries, acceptedEntry{MissionId: missionId, BoardRoomId: boardRoomId})
	user.SetTempData(acceptedMissionsKey, entries)
}

func (m *AutoMissionModule) removeAcceptedEntry(user *userRecord, missionId int) {
	entries := getAcceptedEntries(user)
	for i, e := range entries {
		if e.MissionId == missionId {
			entries = append(entries[:i], entries[i+1:]...)
			break
		}
	}
	user.SetTempData(acceptedMissionsKey, entries)
}

// acceptedMissionIds returns just the IDs from the user's accepted entries.
func acceptedMissionIds(user *userRecord) []int {
	entries := getAcceptedEntries(user)
	ids := make([]int, len(entries))
	for i, e := range entries {
		ids[i] = e.MissionId
	}
	return ids
}

// boardRoomForAcceptedMission returns the board room ID the user accepted a
// given mission from, or 0 if not found.
func boardRoomForAcceptedMission(user *userRecord, missionId int) int {
	for _, e := range getAcceptedEntries(user) {
		if e.MissionId == missionId {
			return e.BoardRoomId
		}
	}
	return 0
}

// -------------------------------------------------------------------------
// Misc helpers
// -------------------------------------------------------------------------

// missionTypeTag returns a short display label for a mission type tag.
func missionTypeLabel(tag string) string {
	switch tag {
	case tagKillMob:
		return "Kill"
	case tagFindItem:
		return "Find"
	case tagExplore:
		return "Explore"
	case tagEscort:
		return "Escort"
	}
	return tag
}

// boardTypeDescription returns a human-readable list of mission types a room offers.
func boardTypeDescription(room *rooms.Room) string {
	var types []string
	for _, tag := range allBoardTags {
		if room.HasTag(tag) {
			types = append(types, missionTypeLabel(tag))
		}
	}
	if len(types) == 0 {
		return "missions"
	}
	return strings.Join(types, "/") + " missions"
}

// removeInt removes the first occurrence of v from s.
func removeInt(s []int, v int) []int {
	for i, x := range s {
		if x == v {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
