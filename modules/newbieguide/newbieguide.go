package newbieguide

import (
	"embed"
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

//go:embed files/*
var files embed.FS

const defaultGuideMobId = 38

type newbieGuideModule struct {
	plug *plugins.Plugin
}

var mod newbieGuideModule

func init() {
	mod = newbieGuideModule{
		plug: plugins.New(`newbieguide`, `1.0`),
	}

	if err := mod.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	events.RegisterListener(events.RoomChange{}, mod.spawnGuide)
	events.RegisterListener(events.LevelUp{}, mod.checkGuide)
}

func (m *newbieGuideModule) guideMobId() int {
	if v, ok := m.plug.Config.Get(`GuideMobId`).(int); ok && v > 0 {
		return v
	}
	return defaultGuideMobId
}

func (m *newbieGuideModule) spawnGuide(e events.Event) events.ListenerReturn {

	evt := e.(events.RoomChange)

	if evt.UserId == 0 {
		return events.Continue
	}

	if evt.ToRoomId < 1 {
		return events.Continue
	}

	user := users.GetByUserId(evt.UserId)
	if user.Character.Level > 5 {
		return events.Continue
	}

	fromRoomOriginal := rooms.GetOriginalRoom(evt.FromRoomId)
	if fromRoomOriginal >= 900 && fromRoomOriginal <= 999 {
		return events.Continue
	}

	toRoomOriginal := rooms.GetOriginalRoom(evt.ToRoomId)
	if toRoomOriginal >= 900 && toRoomOriginal <= 999 {
		return events.Continue
	}

	roundNow := util.GetRoundCount()

	var lastGuideRound uint64 = 0
	tmpLGR := user.GetTempData(`lastGuideRound`)
	if tmpLGRUint, ok := tmpLGR.(uint64); ok {
		lastGuideRound = tmpLGRUint
	}

	if (roundNow - lastGuideRound) < uint64(configs.GetTimingConfig().SecondsToRounds(300)) {
		return events.Continue
	}

	guideMobId := m.guideMobId()

	for _, miid := range user.Character.GetCharmIds() {
		if testMob := mobs.GetInstance(miid); testMob != nil && testMob.MobId == mobs.MobId(guideMobId) {
			return events.Continue
		}
	}

	room := rooms.LoadRoom(evt.ToRoomId)

	guideMob := mobs.NewMobById(mobs.MobId(guideMobId), 1)

	guideMob.Character.Name = fmt.Sprintf(`%s's Guide`, user.Character.Name)

	room.AddMob(guideMob.InstanceId)

	guideMob.Character.Charm(evt.UserId, characters.CharmPermanent, characters.CharmExpiredDespawn)

	user.Character.TrackCharmed(guideMob.InstanceId, true)

	room.SendText(`<ansi fg="mobname">` + guideMob.Character.Name + `</ansi> appears in a shower of sparks!`)

	guideMob.Command(`sayto ` + user.ShorthandId() + ` I'll be here to help protect you while you learn the ropes.`)
	guideMob.Command(`sayto ` + user.ShorthandId() + ` I can create a portal to take us back to Town Square any time. Just <ansi fg="command">ask</ansi> me about it.`)

	user.SendText(`<ansi fg="alert-3">Your guide will try and stick around until you reach level 5.</ansi>`)

	user.SetTempData(`lastGuideRound`, roundNow)

	return events.Continue
}

func (m *newbieGuideModule) checkGuide(e events.Event) events.ListenerReturn {

	evt := e.(events.LevelUp)

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	guideMobId := m.guideMobId()

	if user.Character.Level >= 5 {
		for _, mobInstanceId := range user.Character.CharmedMobs {
			if mob := mobs.GetInstance(mobInstanceId); mob != nil {
				if mob.MobId == mobs.MobId(guideMobId) {
					mob.Command(`say I see you have grown much stronger and more experienced. My assistance is now needed elsewhere. I wish you good luck!`)
					mob.Command(`emote clicks their heels together and disappears in a cloud of smoke.`, 10)
					mob.Command(`suicide vanish`, 10)
				}
			}
		}
	}

	return events.Continue
}
