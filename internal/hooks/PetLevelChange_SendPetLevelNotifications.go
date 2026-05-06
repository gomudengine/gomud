package hooks

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/pets"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

type abilityDelta struct {
	Label  string
	OldVal string
	NewVal string
}

func SendPetLevelNotifications(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PetLevelChange)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PetLevelChange", "Actual Type", e.Type())
		return events.Cancel
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	if evt.NewLevel == evt.OldLevel {
		return events.Continue
	}

	levelUp := evt.NewLevel > evt.OldLevel

	var oldAb, newAb *pets.AbilityDisplay
	if evt.OldAbility != nil {
		oldAb = evt.OldAbility.(*pets.AbilityDisplay)
	}
	if evt.NewAbility != nil {
		newAb = evt.NewAbility.(*pets.AbilityDisplay)
	}

	changes := buildAbilityChanges(oldAb, newAb)

	petLevelData := map[string]interface{}{
		"petName":  evt.PetName,
		"oldLevel": evt.OldLevel,
		"newLevel": evt.NewLevel,
		"levelUp":  levelUp,
		"changes":  changes,
	}
	petLevelStr, _ := templates.Process("character/petlevelup", petLevelData, user.UserId)

	user.SendText(petLevelStr)

	return events.Continue
}

func paddedLabel(name string) string {
	return fmt.Sprintf("%-15s", name+":")
}

func buildAbilityChanges(oldAb, newAb *pets.AbilityDisplay) []abilityDelta {

	var changes []abilityDelta

	oldCombat, newCombat := 0, 0
	oldDice, newDice := "none", "none"
	if oldAb != nil && oldAb.CombatChance > 0 {
		oldCombat = oldAb.CombatChance
		oldDice = fmt.Sprintf("%dd%d", oldAb.DiceCount, oldAb.SideCount)
	}
	if newAb != nil && newAb.CombatChance > 0 {
		newCombat = newAb.CombatChance
		newDice = fmt.Sprintf("%dd%d", newAb.DiceCount, newAb.SideCount)
	}
	if oldCombat > 0 || newCombat > 0 {
		if oldCombat != newCombat {
			changes = append(changes, abilityDelta{
				Label:  paddedLabel("Combat Chance"),
				OldVal: fmt.Sprintf("%d%%", oldCombat),
				NewVal: fmt.Sprintf("%d%%", newCombat),
			})
		}
		if oldDice != newDice {
			changes = append(changes, abilityDelta{
				Label:  paddedLabel("Combat Damage"),
				OldVal: oldDice,
				NewVal: newDice,
			})
		}
	}

	statKeys := map[string]bool{}
	oldStats := map[string]int{}
	newStats := map[string]int{}
	if oldAb != nil {
		for k, v := range oldAb.StatMods {
			statKeys[k] = true
			oldStats[k] = v
		}
	}
	if newAb != nil {
		for k, v := range newAb.StatMods {
			statKeys[k] = true
			newStats[k] = v
		}
	}
	for k := range statKeys {
		if oldStats[k] != newStats[k] {
			changes = append(changes, abilityDelta{
				Label:  paddedLabel(k),
				OldVal: fmt.Sprintf("%d", oldStats[k]),
				NewVal: fmt.Sprintf("%d", newStats[k]),
			})
		}
	}

	oldBuffs := ""
	newBuffs := ""
	if oldAb != nil && len(oldAb.BuffNames) > 0 {
		for i, b := range oldAb.BuffNames {
			if i > 0 {
				oldBuffs += ", "
			}
			oldBuffs += b
		}
	} else {
		oldBuffs = "none"
	}
	if newAb != nil && len(newAb.BuffNames) > 0 {
		for i, b := range newAb.BuffNames {
			if i > 0 {
				newBuffs += ", "
			}
			newBuffs += b
		}
	} else {
		newBuffs = "none"
	}
	if (oldBuffs != "none" || newBuffs != "none") && oldBuffs != newBuffs {
		changes = append(changes, abilityDelta{
			Label:  paddedLabel("Buffs"),
			OldVal: oldBuffs,
			NewVal: newBuffs,
		})
	}

	oldCap, newCap := 0, 0
	if oldAb != nil {
		oldCap = oldAb.Capacity
	}
	if newAb != nil {
		newCap = newAb.Capacity
	}
	if (oldCap > 0 || newCap > 0) && oldCap != newCap {
		changes = append(changes, abilityDelta{
			Label:  paddedLabel("Carry Capacity"),
			OldVal: fmt.Sprintf("%d", oldCap),
			NewVal: fmt.Sprintf("%d", newCap),
		})
	}

	return changes
}
