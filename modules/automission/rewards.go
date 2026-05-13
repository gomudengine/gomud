package automission

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/GoMudEngine/ansitags"
)

// RewardConfig describes what a mission pays out on completion.
type RewardConfig struct {
	Type   string // "gold", "item", "experience"
	Amount int
	ItemId int // only set when Type == "item"
}

// rewardKeyPrefix returns the config key prefix for a given mission type and difficulty.
// e.g. MissionTypeKillMob + DifficultyEasy -> "KillMobEasy"
func rewardKeyPrefix(mType MissionType, diff MissionDifficulty) string {
	var typePart string
	switch mType {
	case MissionTypeKillMob:
		typePart = "KillMob"
	case MissionTypeFindItem:
		typePart = "FindItem"
	case MissionTypeExplore:
		typePart = "Explore"
	case MissionTypeEscort:
		typePart = "Escort"
	default:
		typePart = "KillMob"
	}
	var diffPart string
	if diff == DifficultyHard {
		diffPart = "Hard"
	} else {
		diffPart = "Easy"
	}
	return typePart + diffPart
}

// deliverReward grants the reward to the user.
func deliverReward(reward RewardConfig, user *userRecord) {
	switch reward.Type {
	case "gold":
		user.Character.Gold += reward.Amount
		events.AddToQueue(events.EquipmentChange{
			UserId:     user.UserId,
			GoldChange: reward.Amount,
		})
	case "experience":
		user.GrantXP(reward.Amount, "mission")
	case "item":
		spec := items.GetItemSpec(reward.ItemId)
		if spec != nil {
			item := items.New(reward.ItemId)
			if user.Character.StoreItem(item) {
				events.AddToQueue(events.ItemOwnership{
					UserId: user.UserId,
					Item:   item,
					Gained: true,
				})
			}
		}
	}
}

// selectReward picks a reward for the given mission type and difficulty.
// It looks up the per-type pool first, then falls back to a generic default.
func (m *AutoMissionModule) selectReward(mType MissionType, diff MissionDifficulty) RewardConfig {
	prefix := rewardKeyPrefix(mType, diff)
	pool := m.configSlice(prefix + "Rewards")
	itemPool := m.configSlice(prefix + "RewardItems")

	// If the specific pool is empty, fall back to the easy pool for that type
	// (scaled up for hard), then to a bare default.
	if len(pool) == 0 && diff == DifficultyHard {
		easyPrefix := rewardKeyPrefix(mType, DifficultyEasy)
		pool = m.configSlice(easyPrefix + "Rewards")
		itemPool = m.configSlice(easyPrefix + "RewardItems")
		if len(pool) > 0 {
			idx := util.Rand(len(pool))
			entry := asStringMap(pool[idx])
			rc := RewardConfig{
				Type:   stringVal(entry, "Type", "gold"),
				Amount: intVal(entry, "Amount", 100) * 3,
			}
			if rc.Type == "item" {
				if id := pickItemId(itemPool); id > 0 {
					rc.ItemId = id
				} else {
					rc.Type = "gold"
				}
			}
			return rc
		}
		return defaultReward(mType, diff)
	}

	if len(pool) == 0 {
		return defaultReward(mType, diff)
	}

	idx := util.Rand(len(pool))
	entry := asStringMap(pool[idx])
	rc := RewardConfig{
		Type:   stringVal(entry, "Type", "gold"),
		Amount: intVal(entry, "Amount", 100),
	}

	if rc.Type == "item" {
		if id := pickItemId(itemPool); id > 0 {
			rc.ItemId = id
		} else {
			rc.Type = "gold"
		}
	}

	return rc
}

// defaultReward returns a hardcoded fallback reward when no pool is configured.
func defaultReward(mType MissionType, diff MissionDifficulty) RewardConfig {
	amount := 100
	if diff == DifficultyHard {
		amount = 300
	}
	return RewardConfig{Type: "gold", Amount: amount}
}

// rewardDescription returns a colored, human-readable reward string.
func rewardDescription(r RewardConfig) string {
	switch r.Type {
	case "gold":
		return fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, r.Amount)
	case "experience":
		return fmt.Sprintf(`<ansi fg="experience">%d experience</ansi>`, r.Amount)
	case "item":
		spec := items.GetItemSpec(r.ItemId)
		if spec != nil {
			return fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, spec.Name)
		}
		return `<ansi fg="itemname">an item</ansi>`
	}
	return "unknown reward"
}

// buildBanner formats a fixed-width mission notification banner.
// ANSI tags in title and subtitle are stripped before display so markup
// does not corrupt the border alignment.
func buildBanner(title, subtitle, color string) string {
	const width = 44
	border := strings.Repeat("=", width-2)

	titlePlain := ansitags.Parse(title, ansitags.StripTags)
	titleLine := padCenter(titlePlain, width-4)

	var subtitleLine string
	if subtitle != "" {
		subtitlePlain := ansitags.Parse(subtitle, ansitags.StripTags)
		subtitleLine = padCenter(subtitlePlain, width-4)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<ansi fg=\"%s\">", color))
	sb.WriteString(fmt.Sprintf("╔%s╗\n", border))
	sb.WriteString(fmt.Sprintf("║  %-*s  ║\n", width-6, titleLine))
	if subtitle != "" {
		sb.WriteString(fmt.Sprintf("║  %-*s  ║\n", width-6, subtitleLine))
	}
	sb.WriteString(fmt.Sprintf("╚%s╝", border))
	sb.WriteString("</ansi>")

	return sb.String()
}

// padCenter centers s within fieldWidth, truncating with "..." if too long.
func padCenter(s string, fieldWidth int) string {
	if len(s) > fieldWidth {
		if fieldWidth > 3 {
			return s[:fieldWidth-3] + "..."
		}
		return s[:fieldWidth]
	}
	return s
}

// pickItemId selects a random item ID from a config slice.
func pickItemId(pool []any) int {
	if len(pool) == 0 {
		return 0
	}
	return toInt(pool[util.Rand(len(pool))])
}

// configSlice is defined in admin_api.go.

// asStringMap converts map[any]any or map[string]any to map[string]any.
func asStringMap(v any) map[string]any {
	if mm, ok := v.(map[any]any); ok {
		out := make(map[string]any, len(mm))
		for k, val := range mm {
			out[fmt.Sprint(k)] = val
		}
		return out
	}
	if mm, ok := v.(map[string]any); ok {
		return mm
	}
	return nil
}

func stringVal(m map[string]any, key, def string) string {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func intVal(m map[string]any, key string, def int) int {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		return toInt(v)
	}
	return def
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	}
	return 0
}
