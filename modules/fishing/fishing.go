package fishing

import (
	"embed"
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/statmods"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	//go:embed files/*
	files embed.FS
)

const (
	maxFishingRounds = 15
	catchStartRound  = 3
	configKey        = `fishing-config`

	fishingStatMod statmods.StatName = `fishing`
)

var flavorMessages = []string{
	`Something stirs on the water's surface...`,
	`Ripples appear near your line.`,
	`Your line bobs slightly.`,
	`A fish jumps in the distance.`,
	`The water is still and quiet.`,
	`You feel a faint tug... but nothing.`,
	`A small wave laps at the bank.`,
	`The current gently pulls your line.`,
	`You hear a splash nearby.`,
	`Bubbles rise near your hook.`,
	`A shadow passes beneath the surface.`,
	`The water shimmers in the light.`,
	`Your line drifts lazily in the current.`,
	`A dragonfly lands on your line briefly.`,
	`The water is calm and peaceful.`,
	`You notice movement beneath the surface.`,
	`A frog croaks somewhere nearby.`,
	`The water ripples gently around your hook.`,
	`You feel the weight of the line in your hands.`,
	`Something brushes past your hook...`,
	`The surface of the water glitters like glass.`,
	`A cool breeze drifts across the water.`,
	`You watch a leaf drift past your line.`,
	`The water is deep and dark beneath you.`,
	`A heron stands motionless at the water's edge.`,
	`The smell of mud and algae drifts up from the bank.`,
	`A turtle surfaces briefly, then disappears.`,
	`You hear the distant call of a bird across the water.`,
	`The water is murky today - hard to see what lurks below.`,
	`A gentle mist hangs just above the surface.`,
	`Your reflection stares back at you from the still water.`,
	`A water strider skates across the surface near your line.`,
	`Something large moves beneath you, then is gone.`,
	`The current shifts, pulling your line to one side.`,
	`A cloud passes overhead, casting the water in shadow.`,
	`You hear a plop as something drops into the water nearby.`,
	`The line goes taut for just a moment, then relaxes.`,
	`A school of tiny fish darts beneath the surface.`,
	`The water here smells faintly of moss and earth.`,
	`You adjust your grip and wait patiently.`,
	`A kingfisher dives into the water downstream and comes up empty.`,
	`The bobber dips slightly, then steadies itself.`,
	`An eddy forms around your line and slowly unwinds.`,
	`The light plays tricks on the water - is that movement?`,
	`You hear the soft burble of water moving over stones.`,
	`A water snake glides silently past, ignoring you.`,
	`The surface breaks for just a moment, then seals over.`,
	`Your arms are getting tired. You shift your stance.`,
	`A faint ring spreads outward from somewhere near your hook.`,
	`The water feels alive today.`,
	`Clouds of silt drift up from the bottom, disturbed by something.`,
	`A large bubble rises and pops at the surface.`,
	`You catch a glimpse of silver beneath the water.`,
	`The wind picks up briefly, sending tiny waves across the surface.`,
	`A log drifts slowly past your position.`,
	`You hear the creak of reeds bending in the breeze.`,
	`The water grows darker near the far bank.`,
	`Something nudges your line from below, then retreats.`,
	`A splash erupts a few feet away - something feeding.`,
	`The sun warms your back as you wait.`,
	`You notice the water is clearer here than it looked.`,
	`A crayfish scuttles across the bottom near your hook.`,
	`The line trembles almost imperceptibly.`,
	`Insects buzz lazily above the water's surface.`,
	`A fallen branch bobs gently in the shallows.`,
	`The water makes a soft lapping sound against the bank.`,
	`You feel a subtle vibration travel up the line.`,
	`A pair of ducks paddle past without a second glance.`,
	`The light beneath the surface shifts as something moves.`,
	`You hold your breath for a moment, then let it out slowly.`,
	`A carp rolls lazily near the surface a short distance away.`,
	`The bottom is barely visible through the green-tinted water.`,
	`Something investigates your bait, then thinks better of it.`,
}

func init() {
	m := &FishingModule{
		plug:     plugins.New(`fishing`, `1.0`),
		sessions: make(map[int]*FishingSession),
	}

	if err := m.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	statmods.RegisterStatMod(fishingStatMod, `Modifies the percentage chance to catch something while fishing. Can be negative.`)

	m.plug.ReserveTags(`fishing`)

	rooms.OnRoomLook.Register(m.onRoomLook)

	m.plug.Web.AdminPage("Config", "fishing-config", "html/admin/fishing-config.html", true, "Modules", "Fishing", "Configure fishing spawn tables and catch rates.", "Fishing mini-game letting players catch fish and creatures from water rooms.", nil)
	m.plug.Web.AdminPage("Catchables", "fishing-catchables", "html/admin/fishing-catchables.html", true, "Modules", "Fishing", "Browse and edit catchable fish definitions.", "", nil)
	m.plug.Web.AdminPage("About", "fishing-about", "html/admin/fishing-about.html", true, "Modules", "Fishing", "Information and version details for the Fishing module.", "", nil)

	m.plug.Web.RegisterPermissions(plugins.ModulePermission{
		Key:         `fishing.write`,
		Description: `Manage fishing module configuration and catchable items.`,
		Category:    `Modules`,
	})

	m.plug.Web.AdminAPIEndpoint("GET", "fishing-config", m.apiGetConfig)
	m.plug.Web.AdminAPIEndpoint("PATCH", "fishing-config", m.apiPatchConfig, `fishing.write`)
	m.plug.Web.AdminAPIEndpoint("GET", "fishing-catchables", m.apiGetCatchables)
	m.plug.Web.AdminAPIEndpoint("POST", "fishing-catchables", m.apiAddCatchable, `fishing.write`)
	m.plug.Web.AdminAPIEndpoint("PATCH", "fishing-catchables/{id}", m.apiUpdateCatchable, `fishing.write`)
	m.plug.Web.AdminAPIEndpoint("DELETE", "fishing-catchables/{id}", m.apiDeleteCatchable, `fishing.write`)

	m.plug.AddUserCommand(`fishing`, m.fishingCommand, false, false)

	m.plug.Callbacks.SetOnLoad(m.load)
	m.plug.Callbacks.SetOnSave(m.save)

	events.RegisterListener(events.NewRound{}, m.onNewRound)
	events.RegisterListener(events.RoomChange{}, m.onRoomChange)
	events.RegisterListener(events.AggroChanged{}, m.onAggroChanged)
	events.RegisterListener(events.PlayerDrop{}, m.onPlayerDrop)
	events.RegisterListener(events.PlayerDespawn{}, m.onPlayerDespawn)
}

type CatchableItem struct {
	ID             int   `yaml:"id" json:"id"`
	ItemId         int   `yaml:"itemid" json:"itemid"`
	ChanceToCatch  int   `yaml:"chancetocatch" json:"chancetocatch"`
	RequiredBaitId int   `yaml:"requiredbaitid,omitempty" json:"requiredbaitid"`
	RoomIds        []int `yaml:"roomids,omitempty" json:"roomids"`
}

type FishingConfig struct {
	FishingRodItemIds []int           `yaml:"fishingroditems"`
	Catchables        []CatchableItem `yaml:"catchables"`
}

type FishingSession struct {
	UserId     int
	BaitItem   items.Item
	StartRound uint64
}

type FishingModule struct {
	plug     *plugins.Plugin
	cfg      FishingConfig
	sessions map[int]*FishingSession
}

func (m *FishingModule) load() {
	m.plug.ReadIntoStruct(configKey, &m.cfg)
}

func (m *FishingModule) save() {
	m.plug.WriteStruct(configKey, m.cfg)
}

func (m *FishingModule) nextCatchableId() int {
	maxId := 0
	for _, c := range m.cfg.Catchables {
		if c.ID > maxId {
			maxId = c.ID
		}
	}
	return maxId + 1
}

func (m *FishingModule) cancelFishing(userId int, msg string) {
	session, ok := m.sessions[userId]
	if !ok {
		return
	}
	delete(m.sessions, userId)

	user := users.GetByUserId(userId)
	if user == nil {
		return
	}
	user.Character.SetAdjective(`fishing`, false)

	if msg != `` {
		user.SendText(msg)
	}
	_ = session
}

func (m *FishingModule) fishingCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	rest = strings.TrimSpace(rest)
	restLower := strings.ToLower(rest)

	if restLower == `stop` || restLower == `quit` || restLower == `end` {
		if _, ok := m.sessions[user.UserId]; ok {
			m.cancelFishing(user.UserId, `You reel in your line.`)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> reels in their fishing line.`, user.Character.Name),
				user.UserId,
			)
		} else {
			user.SendText(`You are not currently fishing.`)
		}
		return true, nil
	}

	if !room.HasTag(`fishing`) {
		user.SendText(`There is no suitable fishing spot here.`)
		return true, nil
	}

	if _, alreadyFishing := m.sessions[user.UserId]; alreadyFishing {
		user.SendText(`You are already fishing. Use <ansi fg="command">fishing stop</ansi> to reel in.`)
		return true, nil
	}

	if len(m.cfg.FishingRodItemIds) > 0 {
		equippedWeaponId := user.Character.Equipment.Weapon.ItemId
		hasRod := false
		for _, rodId := range m.cfg.FishingRodItemIds {
			if equippedWeaponId == rodId {
				hasRod = true
				break
			}
		}
		if !hasRod {
			user.SendText(`You need a fishing rod equipped to fish.`)
			return true, nil
		}
	}

	if rest == `` {
		m.listBait(user)
		return true, nil
	}

	baitItem, found := user.Character.FindInBackpack(rest)
	if !found {
		user.SendText(fmt.Sprintf(`You don't have any <ansi fg="itemname">%s</ansi> to use as bait.`, rest))
		m.listBait(user)
		return true, nil
	}

	baitSpec := baitItem.GetSpec()
	if baitSpec.Type != items.Food || baitSpec.Subtype != items.Edible {
		user.SendText(fmt.Sprintf(`<ansi fg="itemname">%s</ansi> doesn't make good bait.`, baitItem.DisplayName()))
		return true, nil
	}

	if usesLeft := user.Character.UseItem(baitItem); usesLeft < 1 {
		events.AddToQueue(events.ItemOwnership{
			UserId: user.UserId,
			Item:   baitItem,
			Gained: false,
		})
	}

	user.Character.SetAdjective(`fishing`, true)

	m.sessions[user.UserId] = &FishingSession{
		UserId:     user.UserId,
		BaitItem:   baitItem,
		StartRound: util.GetRoundCount(),
	}

	user.SendText(`You bait your hook and cast your line into the water...`)
	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> casts a fishing line into the water.`, user.Character.Name),
		user.UserId,
	)

	return true, nil
}

func (m *FishingModule) listBait(user *users.UserRecord) {
	baitItems := []string{}
	for _, itm := range user.Character.GetAllBackpackItems() {
		spec := itm.GetSpec()
		if spec.Type == items.Food && spec.Subtype == items.Edible {
			baitItems = append(baitItems, itm.DisplayName())
		}
	}

	if len(baitItems) == 0 {
		user.SendText(`You have no bait in your backpack. Find some food items to use as bait.`)
		user.SendText(`Usage: <ansi fg="command">fishing <baitname></ansi>`)
		return
	}

	user.SendText(`You have the following bait available:`)
	for _, name := range baitItems {
		user.SendText(fmt.Sprintf(`  <ansi fg="itemname">%s</ansi>`, name))
	}
	user.SendText(`Usage: <ansi fg="command">fishing <baitname></ansi>`)
}

func (m *FishingModule) onNewRound(e events.Event) events.ListenerReturn {
	evt := e.(events.NewRound)
	currentRound := evt.RoundNumber

	for userId, session := range m.sessions {
		roundNumber := int(currentRound-session.StartRound) + 1

		if roundNumber > maxFishingRounds {
			user := users.GetByUserId(userId)
			if user != nil {
				room := rooms.LoadRoom(user.Character.RoomId)
				if room != nil {
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> reels in their line, empty-handed.`, user.Character.Name),
						userId,
					)
				}
			}
			m.cancelFishing(userId, `Your line sits motionless. Nothing is biting today.`)
			continue
		}

		user := users.GetByUserId(userId)
		if user == nil {
			delete(m.sessions, userId)
			continue
		}

		if util.Rand(2) == 0 {
			flavor := flavorMessages[util.Rand(len(flavorMessages))]
			user.SendText(fmt.Sprintf(`<ansi fg="cyan">%s</ansi>`, flavor))
		}

		if roundNumber < catchStartRound {
			continue
		}

		room := rooms.LoadRoom(user.Character.RoomId)
		if room == nil {
			continue
		}

		roundDecay := roundNumber - catchStartRound
		statBonus := user.Character.StatMod(string(fishingStatMod))

		eligible := m.buildEligiblePool(session.BaitItem.ItemId, room.RoomId)
		if len(eligible) == 0 {
			continue
		}

		// Build adjusted weights, clamping each to [0, 100].
		type weightedEntry struct {
			item   CatchableItem
			weight int
		}
		var pool []weightedEntry
		totalWeight := 0
		for _, c := range eligible {
			w := c.ChanceToCatch - roundDecay + statBonus
			if w <= 0 {
				continue
			}
			if w > 100 {
				w = 100
			}
			pool = append(pool, weightedEntry{item: c, weight: w})
			totalWeight += w
		}

		if totalWeight == 0 {
			continue
		}

		// Single roll in [0, totalWeight). Walk the pool until the
		// cumulative weight exceeds the roll — each item gets exactly
		// its proportional share of outcomes.
		roll := util.Rand(totalWeight)
		var caught *CatchableItem
		running := 0
		for _, entry := range pool {
			running += entry.weight
			if roll < running {
				c := entry.item
				caught = &c
				break
			}
		}

		if caught == nil {
			continue
		}

		caughtItemSpec := items.GetItemSpec(caught.ItemId)
		if caughtItemSpec == nil {
			continue
		}

		caughtItem := items.New(caught.ItemId)
		if caughtItem.ItemId == 0 {
			continue
		}

		if user.Character.StoreItem(caughtItem) {
			events.AddToQueue(events.ItemOwnership{
				UserId: userId,
				Item:   caughtItem,
				Gained: true,
			})

			user.SendText(fmt.Sprintf(
				`<ansi fg="yellow-bold">You reel in a <ansi fg="itemname">%s</ansi>!</ansi>`,
				caughtItem.DisplayName(),
			))
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> reels in a <ansi fg="itemname">%s</ansi>!`,
					user.Character.Name, caughtItem.DisplayName()),
				userId,
			)
		}

		m.cancelFishing(userId, ``)
	}

	return events.Continue
}

func (m *FishingModule) buildEligiblePool(baitItemId int, roomId int) []CatchableItem {
	var pool []CatchableItem
	for _, c := range m.cfg.Catchables {
		if c.RequiredBaitId != 0 && c.RequiredBaitId != baitItemId {
			continue
		}
		if len(c.RoomIds) > 0 {
			inRoom := false
			for _, rid := range c.RoomIds {
				if rid == roomId {
					inRoom = true
					break
				}
			}
			if !inRoom {
				continue
			}
		}
		pool = append(pool, c)
	}
	return pool
}

func (m *FishingModule) onRoomChange(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.RoomChange)
	if !ok || evt.UserId == 0 {
		return events.Continue
	}
	if _, fishing := m.sessions[evt.UserId]; fishing {
		m.cancelFishing(evt.UserId, `You reel in your line as you move away.`)
	}
	return events.Continue
}

func (m *FishingModule) onAggroChanged(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.AggroChanged)
	if !ok || evt.UserId == 0 {
		return events.Continue
	}
	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}
	if user.Character.Aggro == nil {
		return events.Continue
	}
	if _, fishing := m.sessions[evt.UserId]; fishing {
		m.cancelFishing(evt.UserId, `Combat interrupts your fishing!`)
	}
	return events.Continue
}

func (m *FishingModule) onPlayerDrop(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerDrop)
	if !ok {
		return events.Continue
	}
	if _, fishing := m.sessions[evt.UserId]; fishing {
		m.cancelFishing(evt.UserId, ``)
	}
	return events.Continue
}

func (m *FishingModule) onPlayerDespawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerDespawn)
	if !ok {
		return events.Continue
	}
	if _, fishing := m.sessions[evt.UserId]; fishing {
		user := users.GetByUserId(evt.UserId)
		if user != nil {
			user.Character.SetAdjective(`fishing`, false)
		}
		delete(m.sessions, evt.UserId)
	}
	return events.Continue
}

func (m *FishingModule) onRoomLook(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	if len(m.cfg.FishingRodItemIds) == 0 || len(m.cfg.Catchables) == 0 {
		return d
	}
	for _, t := range d.Tags {
		if strings.EqualFold(t, `fishing`) {
			d.Alert(`This is a <ansi fg="cyan-bold">fishing spot</ansi>! Equip a rod, bring bait, and use <ansi fg="command">fishing <baitname></ansi> to cast your line.`)
			return d
		}
	}
	return d
}
