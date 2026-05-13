package automission

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// missionCommand is the top-level handler for the "mission" user command.
func (m *AutoMissionModule) missionCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	args := strings.Fields(rest)

	if len(args) == 0 {
		return m.cmdList("", user, room)
	}

	sub := strings.ToLower(args[0])
	subRest := strings.TrimSpace(strings.TrimPrefix(rest, args[0]))

	switch sub {
	case "list":
		return m.cmdList(subRest, user, room)
	case "accept":
		return m.cmdAccept(subRest, user, room)
	case "turn", "turnin":
		if sub == "turn" && len(args) >= 2 && strings.ToLower(args[1]) == "in" {
			return m.cmdTurnIn(user, room)
		}
		if sub == "turnin" {
			return m.cmdTurnIn(user, room)
		}
		return m.cmdList("", user, room)
	default:
		return m.cmdAccept(rest, user, room)
	}
}

// cmdList shows missions.
// At a board room: shows that board's missions.
// Elsewhere: shows only the user's accepted missions across all boards.
func (m *AutoMissionModule) cmdList(_ string, user *users.UserRecord, room *rooms.Room) (bool, error) {
	atBoard := isBoardRoom(room)

	var displayMissions []*Mission
	var tableTitle string

	if atBoard {
		displayMissions = m.missionsForBoard(room.RoomId)
		tableTitle = room.Title + " — " + boardTypeDescription(room)
		if len(displayMissions) == 0 {
			user.SendText(`There are no missions available here right now. Check back later!`)
			return true, nil
		}
	} else {
		acceptedIds := acceptedMissionIds(user)
		for _, mission := range m.allMissions() {
			if slices.Contains(acceptedIds, mission.Id) {
				displayMissions = append(displayMissions, mission)
			}
		}
		tableTitle = "Your Active Missions"
		if len(displayMissions) == 0 {
			user.SendText(`You have no active missions. Visit a mission board to browse available missions.`)
			return true, nil
		}
	}

	headers := []string{"#", "Title", "Difficulty", "Reward", "Status"}
	formatting := []string{
		`<ansi fg="cyan">%s</ansi>`,
		`<ansi fg="yellow">%s</ansi>`,
		`<ansi fg="white">%s</ansi>`,
		`<ansi fg="gold">%s</ansi>`,
		`<ansi fg="green">%s</ansi>`,
	}

	acceptedIds := acceptedMissionIds(user)
	rows := [][]string{}
	for i, mission := range displayMissions {
		status := ""
		if slices.Contains(mission.ReadyToTurnIn, user.UserId) {
			status = "[READY]"
		} else if slices.Contains(acceptedIds, mission.Id) {
			status = "[ACCEPTED]"
		} else if slices.Contains(mission.CompletedBy, user.UserId) {
			status = "[COMPLETED]"
		}
		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			mission.Title,
			string(mission.Difficulty),
			rewardDescription(mission.Reward),
			status,
		})
	}

	tableData := templates.GetTable(tableTitle, headers, rows, formatting)
	tplTxt, _ := templates.Process("tables/generic", tableData, user.UserId)
	user.SendText(tplTxt)

	if atBoard {
		user.SendText(`Type <ansi fg="command">mission accept #</ansi> to accept a mission.`)
	}

	return true, nil
}

// cmdAccept accepts a mission by index or name fragment.
// Must be executed at a board room; the board room ID is stored with the acceptance.
func (m *AutoMissionModule) cmdAccept(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {
	if !isBoardRoom(room) {
		user.SendText(`You must be at a mission board to accept missions.`)
		return true, nil
	}

	rest = strings.TrimSpace(rest)
	if rest == "" {
		user.SendText(`Accept which mission? Try <ansi fg="command">mission list</ansi> to see available missions.`)
		return true, nil
	}

	boardMissions := m.missionsForBoard(room.RoomId)
	mission := resolveMissionFrom(boardMissions, rest)
	if mission == nil {
		user.SendText(fmt.Sprintf(`No mission found matching '%s'.`, rest))
		return true, nil
	}

	if slices.Contains(mission.AcceptedBy, user.UserId) {
		user.SendText(`You have already accepted that mission.`)
		return true, nil
	}

	if slices.Contains(mission.CompletedBy, user.UserId) {
		user.SendText(`You have already completed that mission.`)
		return true, nil
	}

	// Check whether the user already has an active mission of this type.
	for _, active := range m.allMissions() {
		if active.Type != mission.Type {
			continue
		}
		if slices.Contains(active.AcceptedBy, user.UserId) || slices.Contains(active.ReadyToTurnIn, user.UserId) {
			user.SendText(fmt.Sprintf(`You already have an active <ansi fg="yellow">%s</ansi> mission. Turn it in or wait for the board to restock.`, string(mission.Type)))
			return true, nil
		}
	}

	mission.AcceptedBy = append(mission.AcceptedBy, user.UserId)
	addAcceptedEntry(user, mission.Id, room.RoomId)

	if mission.Type == MissionTypeEscort {
		m.spawnEscortMob(mission, user, room)
	}

	user.SendText(fmt.Sprintf(`You have accepted the mission: <ansi fg="yellow">%s</ansi>.`, mission.Title))
	return true, nil
}

// cmdTurnIn delivers rewards for all missions that are ready to turn in at this board.
// Only missions accepted at this specific board room can be turned in here.
func (m *AutoMissionModule) cmdTurnIn(user *users.UserRecord, room *rooms.Room) (bool, error) {
	if !isBoardRoom(room) {
		user.SendText(`You must be at a mission board to turn in missions.`)
		return true, nil
	}

	var readyHere []*Mission
	var readyElsewhere []*Mission

	for _, mission := range m.allMissions() {
		if !slices.Contains(mission.ReadyToTurnIn, user.UserId) {
			continue
		}
		originRoomId := boardRoomForAcceptedMission(user, mission.Id)
		if originRoomId == room.RoomId || (originRoomId == 0 && mission.BoardRoomId == room.RoomId) {
			readyHere = append(readyHere, mission)
		} else {
			readyElsewhere = append(readyElsewhere, mission)
		}
	}

	if len(readyHere) == 0 {
		if len(readyElsewhere) > 0 {
			// Give a helpful hint about where they need to go.
			var hints []string
			seen := make(map[int]bool)
			for _, mission := range readyElsewhere {
				originId := boardRoomForAcceptedMission(user, mission.Id)
				if originId == 0 {
					originId = mission.BoardRoomId
				}
				if seen[originId] {
					continue
				}
				seen[originId] = true
				if originRoom := rooms.LoadRoom(originId); originRoom != nil {
					hints = append(hints, fmt.Sprintf(`<ansi fg="yellow">%s</ansi>`, originRoom.Title))
				}
			}
			if len(hints) > 0 {
				user.SendText(fmt.Sprintf(`You have no missions to turn in here. Return to: %s.`, strings.Join(hints, ", ")))
			} else {
				user.SendText(`You have no missions to turn in here. Return to the board where you accepted them.`)
			}
		} else {
			user.SendText(`You have no missions ready to turn in.`)
		}
		return true, nil
	}

	for _, mission := range readyHere {
		// For find_item missions, verify the player has the item and take it.
		if mission.Type == MissionTypeFindItem {
			if !takeItemFromUser(user, mission.TargetItemId) {
				spec := items.GetItemSpec(mission.TargetItemId)
				itemName := "the required item"
				if spec != nil {
					itemName = fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, spec.Name)
				}
				user.SendText(fmt.Sprintf(`You need to have %s on you to turn in this mission.`, itemName))
				continue
			}
		}

		deliverReward(mission.Reward, user)

		mission.CompletedBy = append(mission.CompletedBy, user.UserId)
		mission.ReadyToTurnIn = removeInt(mission.ReadyToTurnIn, user.UserId)
		mission.AcceptedBy = removeInt(mission.AcceptedBy, user.UserId)
		m.removeAcceptedEntry(user, mission.Id)

		banner := buildBanner("REWARD CLAIMED!", mission.Title, "green") +
			"\n" + fmt.Sprintf(`<ansi fg="green">Reward: </ansi>%s`, rewardDescription(mission.Reward))
		user.SendText(banner)
	}

	return true, nil
}

// resolveMissionFrom finds a mission by 1-based index or title substring within a given slice.
func resolveMissionFrom(missions []*Mission, input string) *Mission {
	if idx, err := strconv.Atoi(input); err == nil {
		idx--
		if idx >= 0 && idx < len(missions) {
			return missions[idx]
		}
		return nil
	}

	lower := strings.ToLower(input)
	for _, mission := range missions {
		if strings.Contains(strings.ToLower(mission.Title), lower) {
			return mission
		}
	}
	return nil
}

// takeItemFromUser finds an item with the given spec ID in the user's inventory
// or equipped slots and removes it. Returns true if the item was found and taken.
func takeItemFromUser(user *users.UserRecord, itemId int) bool {
	// Check inventory first.
	for _, item := range user.Character.GetAllBackpackItems() {
		if item.ItemId == itemId {
			user.Character.RemoveItem(item)
			return true
		}
	}
	// Check equipped slots.
	for _, item := range user.Character.GetAllWornItems() {
		if item.ItemId == itemId {
			user.Character.RemoveFromBody(item)
			return true
		}
	}
	return false
}
