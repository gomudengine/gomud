package zombiemode

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// zombieCommand enqueues a command on behalf of the zombie AI.
// ReadyTurn is set to 0 so it bypasses the per-turn input tracker and
// executes immediately in the same event loop pass.
func zombieCommand(user *users.UserRecord, inputTxt string) {
	events.AddToQueue(events.Input{
		UserId:    user.UserId,
		InputText: inputTxt,
		ReadyTurn: 0,
		Flags:     cmdZombieAI,
	})
}

// zombieActCommand is the zombieact user command registered by this module.
// It overrides the core stub for voluntary zombies and falls back to idle
// flavor for disconnection-triggered zombies.
func (m *ZombieModule) zombieActCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if !user.Character.HasAdjective(`zombie`) {
		return false, nil
	}

	// Only apply AI behavior for voluntary zombies (those in m.active).
	rt, isVoluntary := m.active[user.UserId]
	if !isVoluntary {
		// Disconnection-triggered zombie: emit idle flavor (original behavior).
		if util.Rand(5) == 0 {
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> moans, groans and sways a bit...`, user.Character.Name), user.UserId)
		}
		return true, nil
	}

	// If the feature has been disabled server-side, exit any active zombies.
	if enabled, ok := m.plug.Config.Get(`Enabled`).(bool); ok && !enabled {
		m.exitZombieMode(user.UserId)
		user.SendText(`Zombie mode has been disabled on this server.`)
		return true, nil
	}

	cfg := m.configs[user.UserId]

	// 1. Rest check: if HP% is below threshold, do nothing (or flee if in combat).
	if cfg.RestThreshold > 0 && user.Character.HealthMax.Value > 0 {
		hpPct := (user.Character.Health * 100) / user.Character.HealthMax.Value
		if hpPct < cfg.RestThreshold {
			if user.Character.Aggro != nil {
				zombieCommand(user, `flee`)
			}
			return true, nil
		}
	}

	// 2. Combat: if already in combat, keep attacking.
	if user.Character.Aggro != nil {
		if user.Character.Aggro.MobInstanceId > 0 {
			zombieCommand(user, fmt.Sprintf(`attack #%d`, user.Character.Aggro.MobInstanceId))
			return true, nil
		}
		if user.Character.Aggro.UserId > 0 {
			zombieCommand(user, fmt.Sprintf(`attack @%d`, user.Character.Aggro.UserId))
			return true, nil
		}
	}

	// 3. Combat: scan room for matching mobs and attack.
	if len(cfg.CombatTargets) > 0 {
		for _, mobInstId := range room.GetMobs(rooms.FindNeutral, rooms.FindHostile) {
			mob := mobs.GetInstance(mobInstId)
			if mob == nil {
				continue
			}
			if matchesCombatTarget(mob.Character.Name, cfg.CombatTargets) {
				zombieCommand(user, fmt.Sprintf(`attack #%d`, mobInstId))
				return true, nil
			}
		}
	}

	// 4. Loot: pick up matching items or gold from the floor.
	if len(cfg.LootTargets) > 0 {
		if room.Gold > 0 && matchesLootTarget(`gold`, cfg.LootTargets) {
			zombieCommand(user, `get gold`)
			return true, nil
		}
		if len(user.Character.Items) < user.Character.CarryCapacity() {
			for _, item := range room.GetAllFloorItems(false) {
				if matchesLootTarget(item.DisplayName(), cfg.LootTargets) {
					zombieCommand(user, fmt.Sprintf(`get %s`, item.DisplayName()))
					return true, nil
				}
			}
		}
	}

	// 5. Roam: pick a random exit within radius of the home room.
	if cfg.RoamRadius > 0 && user.Character.Aggro == nil {
		if exitName := m.pickRoamExit(user.Character.RoomId, rt.HomeRoom, cfg.RoamRadius); exitName != `` {
			zombieCommand(user, fmt.Sprintf(`go %s`, exitName))
			return true, nil
		}
	}

	// 6. Idle: emit a flavor message occasionally.
	if util.Rand(5) == 0 {
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> moans, groans and sways a bit...`, user.Character.Name), user.UserId)
	}

	return true, nil
}

// pickRoamExit returns a random exit name from the current room that keeps
// the player within roamRadius rooms of homeRoomId, or "" if none found.
func (m *ZombieModule) pickRoamExit(currentRoomId, homeRoomId, roamRadius int) string {

	currentRoom := rooms.LoadRoom(currentRoomId)
	if currentRoom == nil {
		return ``
	}

	zMapper := mapper.GetMapper(currentRoomId)
	if zMapper == nil {
		return ``
	}

	// Collect exits whose destination is within radius of the home room.
	validExits := []string{}
	for exitName, exitInfo := range currentRoom.Exits {
		if exitInfo.Secret || exitInfo.Lock.IsLocked() {
			continue
		}
		destRooms := zMapper.FindRoomsInDistance(homeRoomId, roamRadius)
		for _, rId := range destRooms {
			if rId == exitInfo.RoomId {
				validExits = append(validExits, exitName)
				break
			}
		}
		// Also allow staying within radius by checking the destination directly.
		if exitInfo.RoomId == homeRoomId {
			if !containsString(validExits, exitName) {
				validExits = append(validExits, exitName)
			}
		}
	}

	if len(validExits) == 0 {
		// Nowhere in radius; head back toward home.
		return m.exitTowardHome(currentRoomId, homeRoomId)
	}

	return validExits[util.Rand(len(validExits))]
}

// exitTowardHome returns an exit name that moves toward homeRoomId, or "".
func (m *ZombieModule) exitTowardHome(currentRoomId, homeRoomId int) string {
	if currentRoomId == homeRoomId {
		return ``
	}

	currentRoom := rooms.LoadRoom(currentRoomId)
	if currentRoom == nil {
		return ``
	}

	// Direct exit to home?
	for exitName, exitInfo := range currentRoom.Exits {
		if exitInfo.RoomId == homeRoomId {
			return exitName
		}
	}

	// Use mapper path if available.
	zMapper := mapper.GetMapper(currentRoomId)
	if zMapper == nil {
		return ``
	}

	homeRooms := zMapper.FindRoomsInDistance(homeRoomId, 1)
	for exitName, exitInfo := range currentRoom.Exits {
		if exitInfo.Secret || exitInfo.Lock.IsLocked() {
			continue
		}
		for _, rId := range homeRooms {
			if rId == exitInfo.RoomId {
				return exitName
			}
		}
	}

	return ``
}

// matchesCombatTarget returns true if the mob name matches any combat target.
func matchesCombatTarget(mobName string, targets []string) bool {
	mobLower := strings.ToLower(mobName)
	for _, t := range targets {
		if t == `*` {
			return true
		}
		if strings.Contains(mobLower, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

// matchesLootTarget returns true if the item name matches any loot target.
func matchesLootTarget(itemName string, targets []string) bool {
	itemLower := strings.ToLower(itemName)
	for _, t := range targets {
		if t == `*` {
			return true
		}
		if strings.Contains(itemLower, strings.ToLower(t)) {
			return true
		}
	}
	return false
}
