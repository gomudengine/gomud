package automission

import (
	"encoding/json"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/configs"
)

// rewardEntry is the JSON shape for one entry in a reward pool.
type rewardEntry struct {
	Type   string `json:"Type" yaml:"Type"`
	Amount int    `json:"Amount" yaml:"Amount"`
}

// complexConfig holds all the config keys that cannot go through SetVal.
type complexConfig struct {
	EscortMobIds []int `yaml:"EscortMobIds"`

	KillMobEasyRewards  []rewardEntry `yaml:"KillMobEasyRewards"`
	KillMobHardRewards  []rewardEntry `yaml:"KillMobHardRewards"`
	FindItemEasyRewards []rewardEntry `yaml:"FindItemEasyRewards"`
	FindItemHardRewards []rewardEntry `yaml:"FindItemHardRewards"`
	ExploreEasyRewards  []rewardEntry `yaml:"ExploreEasyRewards"`
	ExploreHardRewards  []rewardEntry `yaml:"ExploreHardRewards"`
	EscortEasyRewards   []rewardEntry `yaml:"EscortEasyRewards"`
	EscortHardRewards   []rewardEntry `yaml:"EscortHardRewards"`

	KillMobEasyRewardItems  []int `yaml:"KillMobEasyRewardItems"`
	KillMobHardRewardItems  []int `yaml:"KillMobHardRewardItems"`
	FindItemEasyRewardItems []int `yaml:"FindItemEasyRewardItems"`
	FindItemHardRewardItems []int `yaml:"FindItemHardRewardItems"`
	ExploreEasyRewardItems  []int `yaml:"ExploreEasyRewardItems"`
	ExploreHardRewardItems  []int `yaml:"ExploreHardRewardItems"`
	EscortEasyRewardItems   []int `yaml:"EscortEasyRewardItems"`
	EscortHardRewardItems   []int `yaml:"EscortHardRewardItems"`
}

const complexConfigKey = "complex-config"

// loadComplexConfig reads the persisted complex config from plugin storage.
func (m *AutoMissionModule) loadComplexConfig() complexConfig {
	var cc complexConfig
	m.plug.ReadIntoStruct(complexConfigKey, &cc)
	return cc
}

// saveComplexConfig persists the complex config to plugin storage.
func (m *AutoMissionModule) saveComplexConfig(cc complexConfig) error {
	return m.plug.WriteStruct(complexConfigKey, cc)
}

// configSlice reads a complex config key, checking plugin storage first,
// then falling back to the overlay config (for defaults from config.yaml).
func (m *AutoMissionModule) configSlice(key string) []any {
	cc := m.loadComplexConfig()

	switch key {
	case "EscortMobIds":
		return intsToAny(cc.EscortMobIds)
	case "KillMobEasyRewards":
		return rewardsToAny(cc.KillMobEasyRewards)
	case "KillMobHardRewards":
		return rewardsToAny(cc.KillMobHardRewards)
	case "FindItemEasyRewards":
		return rewardsToAny(cc.FindItemEasyRewards)
	case "FindItemHardRewards":
		return rewardsToAny(cc.FindItemHardRewards)
	case "ExploreEasyRewards":
		return rewardsToAny(cc.ExploreEasyRewards)
	case "ExploreHardRewards":
		return rewardsToAny(cc.ExploreHardRewards)
	case "EscortEasyRewards":
		return rewardsToAny(cc.EscortEasyRewards)
	case "EscortHardRewards":
		return rewardsToAny(cc.EscortHardRewards)
	case "KillMobEasyRewardItems":
		return intsToAny(cc.KillMobEasyRewardItems)
	case "KillMobHardRewardItems":
		return intsToAny(cc.KillMobHardRewardItems)
	case "FindItemEasyRewardItems":
		return intsToAny(cc.FindItemEasyRewardItems)
	case "FindItemHardRewardItems":
		return intsToAny(cc.FindItemHardRewardItems)
	case "ExploreEasyRewardItems":
		return intsToAny(cc.ExploreEasyRewardItems)
	case "ExploreHardRewardItems":
		return intsToAny(cc.ExploreHardRewardItems)
	case "EscortEasyRewardItems":
		return intsToAny(cc.EscortEasyRewardItems)
	case "EscortHardRewardItems":
		return intsToAny(cc.EscortHardRewardItems)
	}

	// Fall back to overlay config for any key not managed here.
	v := m.plug.Config.Get(key)
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

// apiGetConfig handles GET /admin/api/v1/automission-config.
func (m *AutoMissionModule) apiGetConfig(r *http.Request) (int, bool, any) {
	cc := m.loadComplexConfig()

	// Merge scalar keys from the overlay config so the response is complete.
	type fullConfig struct {
		RestockPeriod   string `json:"RestockPeriod"`
		MaxMissions     any    `json:"MaxMissions"`
		EscortTimeLimit string `json:"EscortTimeLimit"`
		complexConfig
	}

	fc := fullConfig{
		RestockPeriod:   m.restockPeriod(),
		EscortTimeLimit: m.escortTimeLimit(),
		complexConfig:   cc,
	}

	if v := m.plug.Config.Get("MaxMissions"); v != nil {
		fc.MaxMissions = v
	} else {
		fc.MaxMissions = 6
	}

	// Seed reward pools from overlay defaults if plugin storage is empty.
	if len(cc.KillMobEasyRewards) == 0 {
		fc.KillMobEasyRewards = overlayRewards(m, "KillMobEasyRewards")
	}
	if len(cc.KillMobHardRewards) == 0 {
		fc.KillMobHardRewards = overlayRewards(m, "KillMobHardRewards")
	}
	if len(cc.FindItemEasyRewards) == 0 {
		fc.FindItemEasyRewards = overlayRewards(m, "FindItemEasyRewards")
	}
	if len(cc.FindItemHardRewards) == 0 {
		fc.FindItemHardRewards = overlayRewards(m, "FindItemHardRewards")
	}
	if len(cc.ExploreEasyRewards) == 0 {
		fc.ExploreEasyRewards = overlayRewards(m, "ExploreEasyRewards")
	}
	if len(cc.ExploreHardRewards) == 0 {
		fc.ExploreHardRewards = overlayRewards(m, "ExploreHardRewards")
	}
	if len(cc.EscortEasyRewards) == 0 {
		fc.EscortEasyRewards = overlayRewards(m, "EscortEasyRewards")
	}
	if len(cc.EscortHardRewards) == 0 {
		fc.EscortHardRewards = overlayRewards(m, "EscortHardRewards")
	}

	return http.StatusOK, true, fc
}

// apiPatchConfig handles PATCH /admin/api/v1/automission-config.
func (m *AutoMissionModule) apiPatchConfig(r *http.Request) (int, bool, any) {
	var body struct {
		RestockPeriod   string `json:"RestockPeriod"`
		MaxMissions     string `json:"MaxMissions"`
		EscortTimeLimit string `json:"EscortTimeLimit"`
		complexConfig
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, false, "malformed request body: " + err.Error()
	}

	// Save scalar keys through the normal config system.
	if body.RestockPeriod != "" {
		_ = configs.SetVal("Modules.automission.RestockPeriod", body.RestockPeriod)
	}
	if body.MaxMissions != "" {
		_ = configs.SetVal("Modules.automission.MaxMissions", body.MaxMissions)
	}
	if body.EscortTimeLimit != "" {
		_ = configs.SetVal("Modules.automission.EscortTimeLimit", body.EscortTimeLimit)
	}

	// Save complex keys via plugin storage.
	if err := m.saveComplexConfig(body.complexConfig); err != nil {
		return http.StatusInternalServerError, false, "failed to save config: " + err.Error()
	}

	return http.StatusOK, true, "saved"
}

// overlayRewards reads a reward pool from the overlay config (config.yaml defaults).
func overlayRewards(m *AutoMissionModule, key string) []rewardEntry {
	v := m.plug.Config.Get(key)
	if v == nil {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	var result []rewardEntry
	for _, item := range raw {
		entry := asStringMap(item)
		if entry == nil {
			continue
		}
		result = append(result, rewardEntry{
			Type:   stringVal(entry, "Type", "gold"),
			Amount: intVal(entry, "Amount", 0),
		})
	}
	return result
}

// intsToAny converts []int to []any for configSlice compatibility.
func intsToAny(ids []int) []any {
	if len(ids) == 0 {
		return nil
	}
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

// rewardsToAny converts []rewardEntry to []any (map[string]any per entry).
func rewardsToAny(entries []rewardEntry) []any {
	if len(entries) == 0 {
		return nil
	}
	out := make([]any, len(entries))
	for i, e := range entries {
		out[i] = map[string]any{
			"Type":   e.Type,
			"Amount": e.Amount,
		}
	}
	return out
}
