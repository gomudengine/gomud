package usercommands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Status(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	//possibleStatuses := []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`}

	if rest != `` {

		if rest == `bonuses` {
			return statusBonuses(user)
		}

		if rest != `train` {
			user.SendText("status WHAT???")
			return true, nil
		}

		user.DidTip(`status train`, true)

		cmdPrompt, isNew := user.StartPrompt(`status`, rest)

		if isNew {
			tplTxt, _ := templates.Process("character/status-train", user, user.UserId)
			user.SendText(tplTxt)
		}

		question := cmdPrompt.Ask(`Increase which?`, []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`, `quit`}, `quit`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `quit` {
			user.ClearPrompt()
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`}...)

		question.RejectResponse() // Always reset this question, since we want to keep reusing it.

		if user.Character.StatPoints < 1 {
			user.SendText(`Oops! You have no stat points to spend!`)
			user.ClearPrompt()
			return true, nil
		}
		selection := match
		if match == `` {
			selection = closeMatch
		}

		before := 0
		after := 0
		spent := 0

		switch selection {
		case `strength`:
			before = user.Character.Stats.Strength.Value - user.Character.Stats.Strength.Mods
			user.Character.Stats.Strength.Training += 1
			spent = 1
		case `speed`:
			before = user.Character.Stats.Speed.Value - user.Character.Stats.Speed.Mods
			user.Character.Stats.Speed.Training += 1
			spent = 1
		case `smarts`:
			before = user.Character.Stats.Smarts.Value - user.Character.Stats.Smarts.Mods
			user.Character.Stats.Smarts.Training += 1
			spent = 1
		case `vitality`:
			before = user.Character.Stats.Vitality.Value - user.Character.Stats.Vitality.Mods
			user.Character.Stats.Vitality.Training += 1
			spent = 1
		case `mysticism`:
			before = user.Character.Stats.Mysticism.Value - user.Character.Stats.Mysticism.Mods
			user.Character.Stats.Mysticism.Training += 1
			spent = 1
		case `perception`:
			before = user.Character.Stats.Perception.Value - user.Character.Stats.Perception.Mods
			user.Character.Stats.Perception.Training += 1
			spent = 1
		}

		if spent > 0 {
			after = before + 1
			user.Character.StatPoints -= 1

			user.Character.Validate()

			user.SendText(fmt.Sprintf(term.CRLFStr+`<ansi fg="210">Your <ansi fg="yellow">%s</ansi> training improves from <ansi fg="201">%d</ansi> to <ansi fg="201">%d</ansi>!</ansi>`, selection, before, after))

			events.AddToQueue(events.CharacterTrained{UserId: user.UserId})
		}

		tplTxt, _ := templates.Process("character/status-train", user, user.UserId)

		if spent > 0 {
			tplTxt = strings.Replace(tplTxt, `fakeprop="`+selection+`"`, `bg="highlight"`, 1)
		}

		user.SendText(tplTxt)

		return true, nil
	}

	tplTxt, _ := templates.Process("character/status", user, user.UserId)
	user.SendText(tplTxt)

	Inventory(``, user, room, flags)

	return true, nil
}

func statusBonuses(user *users.UserRecord) (bool, error) {

	var sb strings.Builder

	sb.WriteString(term.CRLFStr)
	sb.WriteString(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Stat Bonuses for </ansi><ansi fg="username">`)
	sb.WriteString(user.Character.Name)
	sb.WriteString(`</ansi>`)
	sb.WriteString(term.CRLFStr)

	type slotItem struct {
		SlotName string
		Item     items.Item
	}

	slots := []slotItem{
		{`Weapon`, user.Character.Equipment.Weapon},
		{`Offhand`, user.Character.Equipment.Offhand},
		{`Head`, user.Character.Equipment.Head},
		{`Neck`, user.Character.Equipment.Neck},
		{`Body`, user.Character.Equipment.Body},
		{`Belt`, user.Character.Equipment.Belt},
		{`Gloves`, user.Character.Equipment.Gloves},
		{`Ring`, user.Character.Equipment.Ring},
		{`Legs`, user.Character.Equipment.Legs},
		{`Feet`, user.Character.Equipment.Feet},
	}

	sb.WriteString(term.CRLFStr)
	sb.WriteString(` ┌─ <ansi fg="black-bold">.:</ansi><ansi fg="20">Equipment Bonuses</ansi> ──────────────────────────────────────────────────────┐`)

	equipFound := false
	for _, slot := range slots {
		if slot.Item.ItemId < 1 {
			continue
		}
		spec := slot.Item.GetSpec()
		if len(spec.StatMods) == 0 {
			continue
		}
		equipFound = true
		sb.WriteString(term.CRLFStr)
		sb.WriteString(fmt.Sprintf(`   <ansi fg="yellow">%-8s</ansi> %s`, slot.SlotName+`:`, slot.Item.DisplayName()))
		sb.WriteString(term.CRLFStr)
		sb.WriteString(`            `)
		sb.WriteString(formatStatMods(spec.StatMods))
	}

	if !equipFound {
		sb.WriteString(term.CRLFStr)
		sb.WriteString(`   <ansi fg="black-bold">None</ansi>`)
	}

	sb.WriteString(term.CRLFStr)
	sb.WriteString(` └────────────────────────────────────────────────────────────────────────────┘`)

	activeBuffs := user.Character.Buffs.GetBuffs()
	sb.WriteString(term.CRLFStr)
	sb.WriteString(` ┌─ <ansi fg="black-bold">.:</ansi><ansi fg="20">Buff Bonuses</ansi> ───────────────────────────────────────────────────────────┐`)

	buffFound := false
	for _, b := range activeBuffs {
		spec := buffs.GetBuffSpec(b.BuffId)
		if spec == nil || spec.Secret || len(spec.StatMods) == 0 {
			continue
		}
		buffFound = true
		sb.WriteString(term.CRLFStr)
		sb.WriteString(fmt.Sprintf(`   <ansi fg="yellow-bold">%s</ansi>`, spec.Name))
		sb.WriteString(term.CRLFStr)
		sb.WriteString(`            `)
		sb.WriteString(formatStatMods(spec.StatMods))
	}

	if !buffFound {
		sb.WriteString(term.CRLFStr)
		sb.WriteString(`   <ansi fg="black-bold">None</ansi>`)
	}

	sb.WriteString(term.CRLFStr)
	sb.WriteString(` └────────────────────────────────────────────────────────────────────────────┘`)

	sb.WriteString(term.CRLFStr)
	sb.WriteString(` ┌─ <ansi fg="black-bold">.:</ansi><ansi fg="20">Pet Bonuses</ansi> ────────────────────────────────────────────────────────────┐`)

	if user.Character.Pet.Exists() {
		mods := user.Character.Pet.GetEffectiveStatMods()
		if len(mods) > 0 {
			sb.WriteString(term.CRLFStr)
			sb.WriteString(fmt.Sprintf(`   %s`, user.Character.Pet.DisplayName()))
			sb.WriteString(term.CRLFStr)
			sb.WriteString(`            `)
			sb.WriteString(formatStatMods(mods))
		} else {
			sb.WriteString(term.CRLFStr)
			sb.WriteString(fmt.Sprintf(`   %s <ansi fg="black-bold">(no stat bonuses)</ansi>`, user.Character.Pet.DisplayName()))
		}
	} else {
		sb.WriteString(term.CRLFStr)
		sb.WriteString(`   <ansi fg="black-bold">No pet</ansi>`)
	}

	sb.WriteString(term.CRLFStr)
	sb.WriteString(` └────────────────────────────────────────────────────────────────────────────┘`)

	sb.WriteString(term.CRLFStr)
	sb.WriteString(term.CRLFStr)
	sb.WriteString(` <ansi fg="yellow">Total Stat Modifiers:</ansi>`)

	allStatNames := []string{`strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`, `healthmax`, `manamax`, `healthrecovery`, `manarecovery`, `attacks`, `damage`}
	displayNames := map[string]string{
		`strength`: `Strength`, `speed`: `Speed`, `smarts`: `Smarts`,
		`vitality`: `Vitality`, `mysticism`: `Mysticism`, `perception`: `Perception`,
		`healthmax`: `Health Max`, `manamax`: `Mana Max`,
		`healthrecovery`: `HP Recovery`, `manarecovery`: `MP Recovery`,
		`attacks`: `Attacks`, `damage`: `Damage`,
	}

	totalFound := false
	for _, stat := range allStatNames {
		total := user.Character.StatMod(stat)
		if total == 0 {
			continue
		}
		totalFound = true
		color := `green`
		sign := `+`
		if total < 0 {
			color = `red`
			sign = ``
		}
		sb.WriteString(term.CRLFStr)
		sb.WriteString(fmt.Sprintf(`   <ansi fg="yellow">%-12s</ansi> <ansi fg="%s">%s%d</ansi>`, displayNames[stat]+`:`, color, sign, total))
	}

	if !totalFound {
		sb.WriteString(term.CRLFStr)
		sb.WriteString(`   <ansi fg="black-bold">None</ansi>`)
	}

	sb.WriteString(term.CRLFStr)

	user.SendText(sb.String())

	return true, nil
}

func formatStatMods(mods map[string]int) string {

	displayNames := map[string]string{
		`strength`: `Str`, `speed`: `Spd`, `smarts`: `Smt`,
		`vitality`: `Vit`, `mysticism`: `Mys`, `perception`: `Per`,
		`healthmax`: `HP`, `manamax`: `MP`,
		`healthrecovery`: `HP Rec`, `manarecovery`: `MP Rec`,
		`attacks`: `Atk`, `damage`: `Dmg`,
		`casting`: `Cast`, `xpscale`: `XP%`,
		`picklock`: `Pick`, `tame`: `Tame`,
	}

	keys := make([]string, 0, len(mods))
	for k := range mods {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := []string{}
	for _, k := range keys {
		v := mods[k]
		if v == 0 {
			continue
		}
		name := displayNames[k]
		if name == `` {
			if strings.HasPrefix(k, `casting-`) {
				name = `Cast:` + strings.TrimPrefix(k, `casting-`)
			} else if strings.HasPrefix(k, `racial-bonus-`) {
				name = `vs ` + strings.TrimPrefix(k, `racial-bonus-`)
			} else {
				name = k
			}
		}
		color := `green`
		sign := `+`
		if v < 0 {
			color = `red`
			sign = ``
		}
		parts = append(parts, fmt.Sprintf(`<ansi fg="%s">%s%d %s</ansi>`, color, sign, v, name))
	}

	return strings.Join(parts, `  `)
}
