package zombiemode

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func (m *ZombieModule) zombieCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(strings.TrimSpace(rest)))

	cfg, hasCfg := m.configs[user.UserId]
	if !hasCfg {
		cfg = ZombieConfig{Profiles: make(map[string]ZombieConfig)}
		m.configs[user.UserId] = cfg
	}

	if len(args) == 0 {
		m.showConfig(user, cfg)
		return true, nil
	}

	switch args[0] {

	case `status`:
		rt, isActive := m.active[user.UserId]
		if !isActive {
			user.SendText(`Zombie mode is not active. Use <ansi fg="command">zombie start</ansi> to begin.`)
			return true, nil
		}
		m.sendSummary(user, rt.Stats)
		return true, nil

	case `start`:
		if enabled, ok := m.plug.Config.Get(`Enabled`).(bool); ok && !enabled {
			user.SendText(`Zombie mode is disabled on this server.`)
			return true, nil
		}
		if _, alreadyActive := m.active[user.UserId]; alreadyActive {
			user.SendText(`Zombie mode is already active.`)
			return true, nil
		}
		m.active[user.UserId] = zombieRuntime{HomeRoom: user.Character.RoomId, Stats: newZombieStats()}
		user.Character.SetAdjective(`zombie`, true)
		user.SendText(`<ansi fg="yellow">Zombie mode activated. Send any input to wake up.</ansi>`)
		if cfg.RoamRadius > 0 && mapper.GetMapper(user.Character.RoomId) == nil {
			user.SendText(`<ansi fg="yellow">Warning: no map data for this area. Roaming is disabled.</ansi>`)
		}
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi>'s eyes glaze over...`, user.Character.Name), user.UserId)
		return true, nil

	case `set`:
		if len(args) < 3 {
			user.SendText(`Usage: zombie set <combat|roam|rest|loot> <value>`)
			return true, nil
		}
		return m.handleSet(args[1], args[2:], user, cfg)

	case `unset`:
		if len(args) < 2 {
			user.SendText(`Usage: zombie unset <combat|roam|rest|loot> [name]`)
			return true, nil
		}
		return m.handleUnset(args[1], args[2:], user, cfg)

	case `save`:
		if len(args) < 2 {
			user.SendText(`Usage: zombie save <profile-name>`)
			return true, nil
		}
		profileName := args[1]
		if cfg.Profiles == nil {
			cfg.Profiles = make(map[string]ZombieConfig)
		}
		if _, exists := cfg.Profiles[profileName]; !exists && len(cfg.Profiles) >= 5 {
			user.SendText(`You cannot save more than 5 profiles. Delete one with <ansi fg="command">zombie delete <name></ansi>.`)
			return true, nil
		}
		saved := ZombieConfig{
			CombatTargets: append([]string{}, cfg.CombatTargets...),
			RoamRadius:    cfg.RoamRadius,
			RestThreshold: cfg.RestThreshold,
			LootTargets:   append([]string{}, cfg.LootTargets...),
			Waypoints:     append([]int{}, cfg.Waypoints...),
		}
		cfg.Profiles[profileName] = saved
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Profile <ansi fg="yellow">%s</ansi> saved.`, profileName))
		return true, nil

	case `load`:
		if len(args) < 2 {
			user.SendText(`Usage: zombie load <profile-name>`)
			return true, nil
		}
		profileName := args[1]
		if cfg.Profiles == nil {
			user.SendText(`No profiles saved.`)
			return true, nil
		}
		p, ok := cfg.Profiles[profileName]
		if !ok {
			user.SendText(fmt.Sprintf(`Profile <ansi fg="yellow">%s</ansi> not found.`, profileName))
			return true, nil
		}
		cfg.CombatTargets = append([]string{}, p.CombatTargets...)
		cfg.RoamRadius = p.RoamRadius
		cfg.RestThreshold = p.RestThreshold
		cfg.LootTargets = append([]string{}, p.LootTargets...)
		cfg.Waypoints = append([]int{}, p.Waypoints...)
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Profile <ansi fg="yellow">%s</ansi> loaded.`, profileName))
		return true, nil

	case `list`:
		if len(cfg.Profiles) == 0 {
			user.SendText(`No saved profiles.`)
			return true, nil
		}
		names := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		user.SendText(fmt.Sprintf(`<ansi fg="black-bold">.:.</ansi> <ansi fg="magenta">Zombie Profiles</ansi> [<ansi fg="yellow">%d/5</ansi>]`, len(cfg.Profiles)))
		for _, name := range names {
			p := cfg.Profiles[name]
			user.SendText(``)
			user.SendText(fmt.Sprintf(`  <ansi fg="yellow-bold">%s</ansi>`, name))
			combatStr := `<ansi fg="red">none</ansi>`
			if len(p.CombatTargets) > 0 {
				combatStr = `<ansi fg="yellow">` + strings.Join(p.CombatTargets, `, `) + `</ansi>`
			}
			user.SendText(fmt.Sprintf(`    <ansi fg="white">Combat targets:</ansi> %s`, combatStr))
			roamStr := `<ansi fg="red">disabled</ansi>`
			if p.RoamRadius > 0 {
				roamStr = fmt.Sprintf(`<ansi fg="yellow">%d</ansi>`, p.RoamRadius)
			}
			user.SendText(fmt.Sprintf(`    <ansi fg="white">Roam radius:   </ansi> %s`, roamStr))
			restStr := `<ansi fg="red">disabled</ansi>`
			if p.RestThreshold > 0 {
				restStr = fmt.Sprintf(`<ansi fg="yellow">%d%%</ansi>`, p.RestThreshold)
			}
			user.SendText(fmt.Sprintf(`    <ansi fg="white">Rest threshold:</ansi> %s`, restStr))
			lootStr := `<ansi fg="red">none</ansi>`
			if len(p.LootTargets) > 0 {
				lootStr = `<ansi fg="yellow">` + strings.Join(p.LootTargets, `, `) + `</ansi>`
			}
			user.SendText(fmt.Sprintf(`    <ansi fg="white">Loot targets:  </ansi> %s`, lootStr))

			wpStr := `<ansi fg="red">none</ansi>`
			if len(p.Waypoints) > 0 {
				wpStr = fmt.Sprintf(`<ansi fg="yellow">%d defined</ansi>`, len(p.Waypoints))
			}
			user.SendText(fmt.Sprintf(`    <ansi fg="white">Waypoints:     </ansi> %s`, wpStr))
		}
		return true, nil

	case `delete`:
		if len(args) < 2 {
			user.SendText(`Usage: zombie delete <profile-name>`)
			return true, nil
		}
		profileName := args[1]
		if _, ok := cfg.Profiles[profileName]; !ok {
			user.SendText(fmt.Sprintf(`Profile <ansi fg="yellow">%s</ansi> not found.`, profileName))
			return true, nil
		}
		delete(cfg.Profiles, profileName)
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Profile <ansi fg="yellow">%s</ansi> deleted.`, profileName))
		return true, nil

	case `waypoints`:
		return m.handleWaypoints(args[1:], user, cfg)
	}

	user.SendText(`Unknown zombie subcommand. Type <ansi fg="command">help zombie</ansi> for usage.`)
	return true, nil
}

func (m *ZombieModule) handleSet(field string, valueArgs []string, user *users.UserRecord, cfg ZombieConfig) (bool, error) {
	value := strings.Join(valueArgs, ` `)

	switch field {
	case `combat`:
		name := value
		if name == `all` {
			name = `*`
		}
		if !containsString(cfg.CombatTargets, name) {
			cfg.CombatTargets = append(cfg.CombatTargets, name)
		}
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Combat target <ansi fg="yellow">%s</ansi> added.`, name))

	case `roam`:
		radius, err := strconv.Atoi(value)
		if err != nil || radius < 0 {
			user.SendText(`Roam radius must be a non-negative integer.`)
			return true, nil
		}
		if maxRoam, ok := m.plug.Config.Get(`MaximumRoam`).(int); ok && maxRoam > 0 && radius > maxRoam {
			user.SendText(fmt.Sprintf(`Roam radius cannot exceed <ansi fg="yellow">%d</ansi> on this server.`, maxRoam))
			return true, nil
		}
		cfg.RoamRadius = radius
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Roam radius set to <ansi fg="yellow">%d</ansi>.`, radius))

	case `rest`:
		pct, err := strconv.Atoi(value)
		if err != nil || pct < 1 || pct > 99 {
			user.SendText(`Rest threshold must be an integer between 1 and 99.`)
			return true, nil
		}
		cfg.RestThreshold = pct
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Rest threshold set to <ansi fg="yellow">%d%%</ansi>.`, pct))

	case `loot`:
		name := value
		if name == `all` {
			name = `*`
		}
		if !containsString(cfg.LootTargets, name) {
			cfg.LootTargets = append(cfg.LootTargets, name)
		}
		m.configs[user.UserId] = cfg
		user.SendText(fmt.Sprintf(`Loot target <ansi fg="yellow">%s</ansi> added.`, name))

	default:
		user.SendText(`Unknown setting. Valid settings: combat, roam, rest, loot`)
	}

	return true, nil
}

func (m *ZombieModule) handleUnset(field string, valueArgs []string, user *users.UserRecord, cfg ZombieConfig) (bool, error) {
	value := strings.Join(valueArgs, ` `)

	switch field {
	case `combat`:
		if value == `` || value == `all` {
			cfg.CombatTargets = nil
			user.SendText(`All combat targets cleared.`)
		} else {
			cfg.CombatTargets = removeString(cfg.CombatTargets, value)
			user.SendText(fmt.Sprintf(`Combat target <ansi fg="yellow">%s</ansi> removed.`, value))
		}
		m.configs[user.UserId] = cfg

	case `roam`:
		cfg.RoamRadius = 0
		m.configs[user.UserId] = cfg
		user.SendText(`Roaming disabled.`)

	case `rest`:
		cfg.RestThreshold = 0
		m.configs[user.UserId] = cfg
		user.SendText(`Rest threshold disabled.`)

	case `loot`:
		if value == `` || value == `all` {
			cfg.LootTargets = nil
			user.SendText(`All loot targets cleared.`)
		} else {
			cfg.LootTargets = removeString(cfg.LootTargets, value)
			user.SendText(fmt.Sprintf(`Loot target <ansi fg="yellow">%s</ansi> removed.`, value))
		}
		m.configs[user.UserId] = cfg

	default:
		user.SendText(`Unknown setting. Valid settings: combat, roam, rest, loot`)
	}

	return true, nil
}

func (m *ZombieModule) showConfig(user *users.UserRecord, cfg ZombieConfig) {
	_, isActive := m.active[user.UserId]

	activeStr := `<ansi fg="red">inactive</ansi>`
	if isActive {
		activeStr = `<ansi fg="green">ACTIVE</ansi>`
	}

	user.SendText(fmt.Sprintf(`<ansi fg="black-bold">.:.</ansi> <ansi fg="magenta">Zombie Mode</ansi> [%s]`, activeStr))
	user.SendText(``)

	combatStr := `<ansi fg="red">none</ansi>`
	if len(cfg.CombatTargets) > 0 {
		combatStr = `<ansi fg="yellow">` + strings.Join(cfg.CombatTargets, `, `) + `</ansi>`
	}
	user.SendText(fmt.Sprintf(`  <ansi fg="white">Combat targets:</ansi> %s`, combatStr))

	roamStr := `<ansi fg="red">disabled</ansi>`
	if cfg.RoamRadius > 0 {
		roamStr = fmt.Sprintf(`<ansi fg="yellow">%d</ansi>`, cfg.RoamRadius)
	}
	user.SendText(fmt.Sprintf(`  <ansi fg="white">Roam radius:   </ansi> %s`, roamStr))

	restStr := `<ansi fg="red">disabled</ansi>`
	if cfg.RestThreshold > 0 {
		restStr = fmt.Sprintf(`<ansi fg="yellow">%d%%</ansi>`, cfg.RestThreshold)
	}
	user.SendText(fmt.Sprintf(`  <ansi fg="white">Rest threshold:</ansi> %s`, restStr))

	lootStr := `<ansi fg="red">none</ansi>`
	if len(cfg.LootTargets) > 0 {
		lootStr = `<ansi fg="yellow">` + strings.Join(cfg.LootTargets, `, `) + `</ansi>`
	}
	user.SendText(fmt.Sprintf(`  <ansi fg="white">Loot targets:  </ansi> %s`, lootStr))

	wpStr := `<ansi fg="red">none</ansi>`
	if len(cfg.Waypoints) > 0 {
		wpStr = fmt.Sprintf(`<ansi fg="yellow">%d defined</ansi>`, len(cfg.Waypoints))
	}
	user.SendText(fmt.Sprintf(`  <ansi fg="white">Waypoints:     </ansi> %s`, wpStr))

	user.SendText(``)
	user.SendText(`Type <ansi fg="command">help zombie</ansi> for usage.`)
}

func (m *ZombieModule) handleWaypoints(args []string, user *users.UserRecord, cfg ZombieConfig) (bool, error) {

	if len(args) == 0 {
		if len(cfg.Waypoints) == 0 {
			user.SendText(`No waypoints defined. Use <ansi fg="command">zombie waypoints add</ansi> to add your current room.`)
			return true, nil
		}
		user.SendText(fmt.Sprintf(`<ansi fg="black-bold">.:.</ansi> <ansi fg="magenta">Zombie Waypoints</ansi> [<ansi fg="yellow">%d</ansi>]`, len(cfg.Waypoints)))
		user.SendText(``)
		for i, roomId := range cfg.Waypoints {
			roomName := `unknown`
			if r := rooms.LoadRoom(roomId); r != nil {
				roomName = r.Title
			}
			marker := ``
			if roomId == user.Character.RoomId {
				marker = ` <ansi fg="green">(here)</ansi>`
			}
			user.SendText(fmt.Sprintf(`  <ansi fg="yellow">%d.</ansi> <ansi fg="white">Room %d</ansi> - %s%s`, i+1, roomId, roomName, marker))
		}
		return true, nil
	}

	switch args[0] {
	case `add`:
		maxWaypoints := 10
		if maxWP, ok := m.plug.Config.Get(`MaxWaypoints`).(int); ok && maxWP > 0 {
			maxWaypoints = maxWP
		}
		if len(cfg.Waypoints) >= maxWaypoints {
			user.SendText(fmt.Sprintf(`Cannot add more than <ansi fg="yellow">%d</ansi> waypoints.`, maxWaypoints))
			return true, nil
		}

		currentRoomId := user.Character.RoomId

		for _, wp := range cfg.Waypoints {
			if wp == currentRoomId {
				user.SendText(`This room is already a waypoint.`)
				return true, nil
			}
		}

		if len(cfg.Waypoints) > 0 {
			lastWP := cfg.Waypoints[len(cfg.Waypoints)-1]
			if _, err := mapper.GetPath(lastWP, currentRoomId); err != nil {
				user.SendText(fmt.Sprintf(`No valid path from the previous waypoint (room %d) to here. Cannot add.`, lastWP))
				return true, nil
			}
		}

		cfg.Waypoints = append(cfg.Waypoints, currentRoomId)
		m.configs[user.UserId] = cfg

		roomName := `unknown`
		if r := rooms.LoadRoom(currentRoomId); r != nil {
			roomName = r.Title
		}
		user.SendText(fmt.Sprintf(`Waypoint added: room %d (%s). Total: %d.`, currentRoomId, roomName, len(cfg.Waypoints)))

		if len(cfg.Waypoints) > 1 {
			if _, err := mapper.GetPath(currentRoomId, cfg.Waypoints[0]); err != nil {
				user.SendText(fmt.Sprintf(`<ansi fg="yellow">Warning:</ansi> No path from here back to the first waypoint (room %d). The cycle may not complete.`, cfg.Waypoints[0]))
			}
		}

		return true, nil

	case `reset`:
		cfg.Waypoints = nil
		m.configs[user.UserId] = cfg
		if rt, isActive := m.active[user.UserId]; isActive {
			rt.WaypointIndex = 0
			rt.WaypointPath = nil
			m.active[user.UserId] = rt
		}
		user.SendText(`All waypoints cleared.`)
		return true, nil
	}

	user.SendText(`Usage: <ansi fg="command">zombie waypoints</ansi>, <ansi fg="command">zombie waypoints add</ansi>, or <ansi fg="command">zombie waypoints reset</ansi>`)
	return true, nil
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	out := slice[:0]
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
