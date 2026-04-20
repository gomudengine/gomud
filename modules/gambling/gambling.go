package gambling

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	//go:embed files/*
	files embed.FS
)

const dieItemId = 1040000
const coinItemId = 1040001
const tarotItemId = 1040002
const eightItemId = 1040003
const bottleItemId = 1040004
const cardsItemId = 1040005

const defaultCost = 10

func init() {

	g := &GamblingModule{
		plug:  plugins.New(`gambling`, `1.0`),
		state: SlotState{Jackpot: 0},
	}
	g.roomCache, _ = lru.New[int, struct{}](256)

	if err := g.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	// Register item scripts from embedded filesystem.
	for itemId, path := range map[int]string{
		dieItemId:    `files/datafiles/items/1040000-6_sided_die.js`,
		coinItemId:   `files/datafiles/items/1040001-lucky_coin.js`,
		tarotItemId:  `files/datafiles/items/1040002-tarot_deck.js`,
		eightItemId:  `files/datafiles/items/1040003-magic_8_ball.js`,
		bottleItemId: `files/datafiles/items/1040004-empty_bottle.js`,
		cardsItemId:  `files/datafiles/items/1040005-deck_of_cards.js`,
	} {
		scriptBytes, err := fs.ReadFile(files, path)
		if err != nil {
			mudlog.Error("gambling: failed to read item script", "path", path, "error", err)
			continue
		}
		items.RegisterItemScript(itemId, string(scriptBytes))
	}

	// Register the "play" user command (handles both slots and claw machine).
	g.plug.AddUserCommand(`play`, g.playCommand, false, false)

	// Persist the jackpot across restarts.
	g.plug.Callbacks.SetOnLoad(g.load)
	g.plug.Callbacks.SetOnSave(g.save)

	// Hook into room look to inject slot machine and claw machine alerts.
	rooms.OnRoomLook.Register(g.onRoomLook)
	rooms.OnRoomLook.Register(g.onRoomLookClaw)

	// Inject gambling nouns when players enter or spawn in tagged rooms.
	// This ensures "look slot machine" and "look claw machine" are resolved
	// synchronously by the core look command rather than falling through to
	// "Look at what???".
	events.RegisterListener(events.RoomChange{}, g.onRoomChange)
	events.RegisterListener(events.PlayerSpawn{}, g.onPlayerSpawn)
}

// ensureRoomNouns injects gambling fixture nouns into the room's Nouns map so
// that "look slot machine", "look slots", and "look claw machine" are handled
// synchronously by the core look command rather than falling through to
// "Look at what???". Results are cached by room ID so the tag scan and map
// writes only happen once per room per server session.
func (g *GamblingModule) ensureRoomNouns(room *rooms.Room) {
	if _, already := g.roomCache.Get(room.RoomId); already {
		return
	}

	if room.Nouns == nil {
		room.Nouns = make(map[string]string)
	}

	if roomHasSlots(room) {
		room.Nouns[`slot machine`] = g.slotMachineNounDesc(room)
		room.Nouns[`slots`] = `:slot machine`
	}

	if roomHasClaw(room) {
		room.Nouns[`claw machine`] = g.clawMachineNounDesc(room)
		room.Nouns[`claw`] = `:claw machine`
	}

	g.roomCache.Add(room.RoomId, struct{}{})
}

// onRoomChange injects gambling nouns whenever a player enters a tagged room.
func (g *GamblingModule) onRoomChange(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.RoomChange)
	if !ok || evt.UserId == 0 {
		return events.Continue
	}
	if room := rooms.LoadRoom(evt.ToRoomId); room != nil {
		g.ensureRoomNouns(room)
	}
	return events.Continue
}

// onPlayerSpawn injects gambling nouns for players who log in inside a tagged room.
func (g *GamblingModule) onPlayerSpawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerSpawn)
	if !ok {
		return events.Continue
	}
	if room := rooms.LoadRoom(evt.RoomId); room != nil {
		g.ensureRoomNouns(room)
	}
	return events.Continue
}


// GamblingModule holds module-level state for the gambling plugin.
type GamblingModule struct {
	plug      *plugins.Plugin
	state     SlotState
	// roomCache tracks room IDs whose gambling nouns have already been injected,
	// so ensureRoomNouns skips work on every subsequent player movement through
	// rooms that were already processed.
	roomCache *lru.Cache[int, struct{}]
}

func (g *GamblingModule) load() {
	g.plug.ReadIntoStruct(`slotstate`, &g.state)
}

func (g *GamblingModule) save() {
	g.plug.WriteStruct(`slotstate`, g.state)
}

// onRoomLook injects a slot machine alert when the room has the slots tag.
func (g *GamblingModule) onRoomLook(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	for _, t := range d.Tags {
		tl := strings.ToLower(t)
		if tl == `slots` || tl == `slot machine` {
			d.RoomAlerts = append(d.RoomAlerts,
				`There is a <ansi fg="cyan-bold">slot machine</ansi> here! You can <ansi fg="command">look</ansi> at or <ansi fg="command">play</ansi> it.`,
			)
			return d
		}
	}
	return d
}

// playCommand handles "play slots" / "play slot machine" / "play claw machine" / "play claw".
func (g *GamblingModule) playCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	arg := strings.TrimSpace(rest)

	if m, c := util.FindMatchIn(arg, `slot machine`, `slots`, `slot`); m != `` || c != `` {
		if !roomHasSlots(room) {
			user.SendText(`There is no slot machine here.`)
			return true, nil
		}
		g.playSlots(user, room)
		return true, nil
	}

	if m, c := util.FindMatchIn(arg, `claw machine`, `claw`); m != `` || c != `` {
		if !roomHasClaw(room) {
			user.SendText(`There is no claw machine here.`)
			return true, nil
		}
		g.playClaw(user, room)
		return true, nil
	}

	user.SendText(`Play what? Try <ansi fg="command">play slots</ansi> or <ansi fg="command">play claw machine</ansi>.`)
	return true, nil
}
