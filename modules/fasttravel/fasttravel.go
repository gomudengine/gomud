package fasttravel

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	//go:embed files/*
	files embed.FS
)

const (
	fastTravelTag = "fast travel"
	dataKeyFmt    = "fasttravel-user-%d"
)

func init() {
	m := &FastTravelModule{
		plug:    plugins.New(`fasttravel`, `1.0`),
		players: make(map[int]FastTravelData),
	}

	if err := m.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	m.plug.AddUserCommand(`fasttravel`, m.fastTravelCommand, false, false)

	m.plug.Web.AdminPage("Config", "fasttravel-config", "html/admin/fasttravel-config.html", true, "Modules", "Fast Travel", nil)

	m.plug.ReserveTags(fastTravelTag)

	m.plug.Callbacks.SetOnSave(m.onSave)

	events.RegisterListener(events.PlayerSpawn{}, m.onPlayerSpawn)
	events.RegisterListener(events.PlayerDespawn{}, m.onPlayerDespawn)

	rooms.OnRoomLook.Register(m.onRoomLook)
}

// FastTravelData holds the unlocked fast travel room IDs for a single user.
type FastTravelData struct {
	UnlockedRoomIds []int `yaml:"unlockedroomids,omitempty"`
}

// FastTravelModule owns all fast travel state.
type FastTravelModule struct {
	plug    *plugins.Plugin
	players map[int]FastTravelData // keyed by userId; loaded on PlayerSpawn
}

func dataKey(userId int) string {
	return fmt.Sprintf(dataKeyFmt, userId)
}

func (m *FastTravelModule) load(userId int) FastTravelData {
	var data FastTravelData
	m.plug.ReadIntoStruct(dataKey(userId), &data)
	return data
}

func (m *FastTravelModule) save(userId int, data FastTravelData) {
	m.plug.WriteStruct(dataKey(userId), data)
}

func (m *FastTravelModule) onSave() {
	for userId, data := range m.players {
		m.save(userId, data)
	}
}

func (m *FastTravelModule) onPlayerSpawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerSpawn)
	if !ok {
		return events.Continue
	}
	m.players[evt.UserId] = m.load(evt.UserId)
	return events.Continue
}

func (m *FastTravelModule) onPlayerDespawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerDespawn)
	if !ok {
		return events.Continue
	}
	if data, exists := m.players[evt.UserId]; exists {
		m.save(evt.UserId, data)
		delete(m.players, evt.UserId)
	}
	return events.Continue
}

// onRoomLook injects a fast travel alert when the room has the fast travel tag.
func (m *FastTravelModule) onRoomLook(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	for _, t := range d.Tags {
		if strings.EqualFold(t, fastTravelTag) {
			d.RoomAlerts = append(d.RoomAlerts,
				` <ansi fg="yellow-bold">This is a fast travel station!</ansi> <ansi fg="command">fasttravel</ansi> lists destinations.`,
			)
			return d
		}
	}
	return d
}

// roomIsFastTravel returns true if the room has the fast travel tag.
func roomIsFastTravel(room *rooms.Room) bool {
	return room.HasTag(fastTravelTag)
}

// hasUnlocked returns true if the user has unlocked the given room.
func (m *FastTravelModule) hasUnlocked(userId, roomId int) bool {
	data := m.players[userId]
	for _, id := range data.UnlockedRoomIds {
		if id == roomId {
			return true
		}
	}
	return false
}

// unlock adds the roomId to the user's unlocked list if not already present.
// Returns true if the room was newly unlocked, false if it was already known.
func (m *FastTravelModule) unlock(userId, roomId int) bool {
	if m.hasUnlocked(userId, roomId) {
		return false
	}
	data := m.players[userId]
	data.UnlockedRoomIds = append(data.UnlockedRoomIds, roomId)
	m.players[userId] = data
	return true
}

// travelCost reads the configured gold cost and required item ID from the module config.
// A cost of 0 means no gold is required. An itemId of 0 means no item is required.
func (m *FastTravelModule) travelCost() (goldCost int, itemId int) {
	if v, ok := m.plug.Config.Get(`GoldCost`).(int); ok && v > 0 {
		goldCost = v
	}
	if v, ok := m.plug.Config.Get(`RequiredItemId`).(int); ok && v > 0 {
		itemId = v
	}
	return goldCost, itemId
}

// destEntry holds a resolved fast travel destination for display and matching.
type destEntry struct {
	roomId int
	title  string
}

func (m *FastTravelModule) fastTravelCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	if !roomIsFastTravel(room) {
		user.SendText(`You must be at a fast travel station to use fast travel.` + term.CRLFStr)
		return true, nil
	}

	// Always unlock the current room when the player uses the command here.
	if m.unlock(user.UserId, room.RoomId) {
		user.SendText(`<ansi fg="yellow-bold">You have discovered a new fast travel station: </ansi><ansi fg="room-title">` + room.Title + `</ansi><ansi fg="yellow-bold">!</ansi>` + term.CRLFStr)
	}

	// Build the list of unlocked destinations, excluding the current room.
	data := m.players[user.UserId]
	destinations := make([]destEntry, 0, len(data.UnlockedRoomIds))
	for _, id := range data.UnlockedRoomIds {
		if id == room.RoomId {
			continue
		}
		r := rooms.LoadRoom(id)
		if r == nil {
			continue
		}
		destinations = append(destinations, destEntry{roomId: id, title: r.Title})
	}
	sort.Slice(destinations, func(i, j int) bool {
		return destinations[i].title < destinations[j].title
	})

	// List mode.
	if rest == `` {
		if len(destinations) == 0 {
			user.SendText(`<ansi fg="yellow">You have not discovered any other fast travel stations yet.</ansi>` + term.CRLFStr)
			return true, nil
		}

		var sb strings.Builder
		sb.WriteString(`<ansi fg="yellow-bold">Fast Travel Destinations:</ansi>` + "\n")
		for i, dest := range destinations {
			sb.WriteString(fmt.Sprintf(`  <ansi fg="cyan">%d)</ansi> <ansi fg="room-title">%s</ansi>`+"\n", i+1, dest.title))
		}
		user.SendText(sb.String())
		return true, nil
	}

	// Travel mode: match the argument against destination titles.
	titles := make([]string, len(destinations))
	for i, dest := range destinations {
		titles[i] = dest.title
	}

	closeMatch, exactMatch := util.FindMatchIn(rest, titles...)

	matchTitle := exactMatch
	if matchTitle == `` {
		matchTitle = closeMatch
	}
	if matchTitle == `` {
		user.SendText(fmt.Sprintf(`No fast travel destination matching "<ansi fg="cyan">%s</ansi>" found.`+term.CRLFStr, rest))
		return true, nil
	}

	var destRoomId int
	for _, dest := range destinations {
		if strings.EqualFold(dest.title, matchTitle) {
			destRoomId = dest.roomId
			break
		}
	}
	if destRoomId == 0 {
		user.SendText(fmt.Sprintf(`No fast travel destination matching "<ansi fg="cyan">%s</ansi>" found.`+term.CRLFStr, rest))
		return true, nil
	}

	destRoom := rooms.LoadRoom(destRoomId)
	if destRoom == nil {
		user.SendText(`That fast travel destination is no longer available.` + term.CRLFStr)
		return true, nil
	}

	// Enforce travel cost before moving.
	goldCost, requiredItemId := m.travelCost()

	if goldCost > 0 {
		if user.Character.Gold < goldCost {
			user.SendText(fmt.Sprintf(`You need at least <ansi fg="gold">%d gold</ansi> to use the fast travel network.`+term.CRLFStr, goldCost))
			return true, nil
		}
	}

	var requiredItem items.Item
	if requiredItemId > 0 {
		itm, found := user.Character.FindInBackpack(fmt.Sprintf(`!%d`, requiredItemId))
		if !found {
			iSpec := items.GetItemSpec(requiredItemId)
			itemName := fmt.Sprintf(`item #%d`, requiredItemId)
			if iSpec != nil {
				itemName = iSpec.Name
			}
			user.SendText(fmt.Sprintf(`You need a <ansi fg="itemname">%s</ansi> to use the fast travel network.`+term.CRLFStr, itemName))
			return true, nil
		}
		requiredItem = itm
	}

	// Deduct gold cost.
	if goldCost > 0 {
		user.Character.Gold -= goldCost
		events.AddToQueue(events.EquipmentChange{
			UserId:     user.UserId,
			GoldChange: -goldCost,
		})
	}

	// Consume required item.
	if requiredItem.ItemId > 0 {
		user.Character.RemoveItem(requiredItem)
		events.AddToQueue(events.ItemOwnership{
			UserId: user.UserId,
			Item:   requiredItem,
			Gained: false,
		})
	}

	user.SendText(fmt.Sprintf(`You step through the fast travel network and arrive at <ansi fg="room-title">%s</ansi>.`+term.CRLFStr, destRoom.Title))
	rooms.MoveToRoom(user.UserId, destRoomId)

	return true, nil
}
