package zombiemode

import (
	"embed"
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/users"
)

var (
	//go:embed files/*
	files embed.FS
)

func init() {
	m := &ZombieModule{
		plug:    plugins.New(`zombiemode`, `1.0`),
		configs: make(map[int]ZombieConfig),
		active:  make(map[int]zombieRuntime),
	}

	if err := m.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	m.plug.AddUserCommand(`zombie`, m.zombieCommand, false, false)
	m.plug.AddUserCommand(`zombieact`, m.zombieActCommand, true, false)

	m.plug.Callbacks.SetOnSave(m.onSave)

	events.RegisterListener(events.PlayerSpawn{}, m.onPlayerSpawn)
	events.RegisterListener(events.PlayerDespawn{}, m.onPlayerDespawn)
	events.RegisterListener(events.PlayerDrop{}, m.onPlayerDrop)
	events.RegisterListener(events.AggroChanged{}, m.onAggroChanged)
	events.RegisterListener(events.Input{}, m.onInput, events.First)
}

// ZombieConfig holds per-user zombie behavior configuration.
// It is persisted to disk via the plugin file system.
type ZombieConfig struct {
	CombatTargets []string                `yaml:"combattargets,omitempty"`
	RoamRadius    int                     `yaml:"roamradius,omitempty"`
	RestThreshold int                     `yaml:"restthreshold,omitempty"`
	LootTargets   []string                `yaml:"loottargets,omitempty"`
	Profiles      map[string]ZombieConfig `yaml:"profiles,omitempty"`
}

// zombieRuntime holds transient state that is only valid while zombie mode is active.
type zombieRuntime struct {
	HomeRoom int
}

// cmdZombieAI is an EventFlag bit used to tag input events issued by the zombie
// AI itself, so that onInput can distinguish them from real player input.
const cmdZombieAI events.EventFlag = 0b00100000

// ZombieModule is the module struct.
type ZombieModule struct {
	plug    *plugins.Plugin
	configs map[int]ZombieConfig  // keyed by userId; loaded on PlayerSpawn, flushed on PlayerDespawn
	active  map[int]zombieRuntime // keyed by userId; present only while voluntarily active
}

// exitZombieMode clears the zombie adjective and removes the active entry.
func (m *ZombieModule) exitZombieMode(userId int) {
	delete(m.active, userId)
	if user := users.GetByUserId(userId); user != nil {
		user.Character.SetAdjective(`zombie`, false)
	}
}

// configKey returns the plugin storage identifier for a user's config.
func configKey(userId int) string {
	return fmt.Sprintf(`user-%d`, userId)
}

// onSave persists configs for all currently online players.
func (m *ZombieModule) onSave() {
	for userId, cfg := range m.configs {
		m.plug.WriteStruct(configKey(userId), cfg)
	}
}

// onPlayerSpawn loads the user's config and clears any stale active entry.
func (m *ZombieModule) onPlayerSpawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerSpawn)
	if !ok {
		return events.Continue
	}

	// Safety guard: clear any stale active entry from a previous session.
	if _, stale := m.active[evt.UserId]; stale {
		m.exitZombieMode(evt.UserId)
	}

	var cfg ZombieConfig
	m.plug.ReadIntoStruct(configKey(evt.UserId), &cfg)
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]ZombieConfig)
	}
	m.configs[evt.UserId] = cfg

	return events.Continue
}

// onPlayerDespawn persists the user's config and exits zombie mode.
func (m *ZombieModule) onPlayerDespawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerDespawn)
	if !ok {
		return events.Continue
	}

	if cfg, exists := m.configs[evt.UserId]; exists {
		m.plug.WriteStruct(configKey(evt.UserId), cfg)
		delete(m.configs, evt.UserId)
	}

	m.exitZombieMode(evt.UserId)

	return events.Continue
}

// onPlayerDrop exits zombie mode on player death (config survives for the respawn).
func (m *ZombieModule) onPlayerDrop(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerDrop)
	if !ok {
		return events.Continue
	}

	if _, isActive := m.active[evt.UserId]; isActive {
		m.exitZombieMode(evt.UserId)
		if user := users.GetByUserId(evt.UserId); user != nil {
			user.SendText(`Zombie mode deactivated.`)
		}
	}

	return events.Continue
}

// onAggroChanged fires whenever a mob or player enters/leaves aggro state.
// If a mob has just targeted a voluntary zombie, issue an attack command so
// the zombie fights back.
func (m *ZombieModule) onAggroChanged(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.AggroChanged)
	if !ok {
		return events.Continue
	}

	if evt.MobInstanceId == 0 {
		return events.Continue
	}

	mob := mobs.GetInstance(evt.MobInstanceId)
	if mob == nil || mob.Character.Aggro == nil || mob.Character.Aggro.UserId == 0 {
		return events.Continue
	}

	targetUserId := mob.Character.Aggro.UserId
	if _, isActive := m.active[targetUserId]; !isActive {
		return events.Continue
	}

	user := users.GetByUserId(targetUserId)
	if user == nil || user.Character.Aggro != nil {
		return events.Continue
	}

	zombieCommand(user, fmt.Sprintf(`attack #%d`, evt.MobInstanceId))

	return events.Continue
}

// onInput intercepts player input while zombie mode is active.
// Any input that did not originate from the zombie AI itself wakes the player.
func (m *ZombieModule) onInput(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.Input)
	if !ok {
		return events.Continue
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	if _, isActive := m.active[evt.UserId]; !isActive {
		return events.Continue
	}

	// Ignore the zombieact tick and any commands the AI issued (tagged with cmdZombieAI).
	if evt.InputText == `zombieact` || evt.Flags.Has(cmdZombieAI) {
		return events.Continue
	}

	m.exitZombieMode(evt.UserId)
	if user := users.GetByUserId(evt.UserId); user != nil {
		user.SendText(`You snap out of zombie mode.`)
	}

	return events.Continue
}
