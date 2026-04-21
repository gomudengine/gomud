package gambling

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

const defaultClawCost = 10
const defaultClawWinChance = 10 // percent

// clawPrize describes one prize slot in the claw machine.
type clawPrize struct {
	itemId    int
	name      string
	configKey string // config key for this prize's weight
	defaultWt int    // weight used when config key is absent
}

// clawPrizes is the ordered list of prizes the claw machine can award.
// Weights are relative: a prize with weight 20 is twice as likely as one with weight 10.
// Set a prize's weight to 0 in config to remove it from the pool entirely.
var clawPrizes = []clawPrize{
	{dieItemId, `6-sided die`, `ClawPrizeDie`, 20},
	{coinItemId, `lucky coin`, `ClawPrizeCoin`, 20},
	{bottleItemId, `empty bottle`, `ClawPrizeBottle`, 20},
	{cardsItemId, `deck of cards`, `ClawPrizeCards`, 15},
	{eightItemId, `magic 8-ball`, `ClawPrize8Ball`, 15},
	{tarotItemId, `tarot deck`, `ClawPrizeTarot`, 10},
}

// roomHasClaw returns true when the room carries a "claw machine" tag.
func roomHasClaw(r *rooms.Room) bool {
	for _, t := range r.Tags {
		if strings.ToLower(t) == `claw machine` {
			return true
		}
	}
	return false
}

// clawPrizeWeight returns the configured weight for a prize, falling back to the default.
func (g *GamblingModule) clawPrizeWeight(p clawPrize) int {
	if v, ok := g.plug.Config.Get(p.configKey).(int); ok {
		return v
	}
	return p.defaultWt
}

// pickClawPrize selects a prize using weighted random selection over the active pool.
// Returns nil if the pool is empty (all weights zero).
func (g *GamblingModule) pickClawPrize() *clawPrize {
	total := 0
	for i := range clawPrizes {
		total += g.clawPrizeWeight(clawPrizes[i])
	}
	if total <= 0 {
		return nil
	}
	roll := util.Rand(total)
	cumulative := 0
	for i := range clawPrizes {
		cumulative += g.clawPrizeWeight(clawPrizes[i])
		if roll < cumulative {
			return &clawPrizes[i]
		}
	}
	return &clawPrizes[len(clawPrizes)-1]
}

// onRoomLookClaw injects a claw machine alert when the room has the claw machine tag.
func (g *GamblingModule) onRoomLookClaw(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	for _, t := range d.Tags {
		if strings.ToLower(t) == `claw machine` {
			d.RoomAlerts = append(d.RoomAlerts,
				`There is a <ansi fg="cyan-bold">claw machine</ansi> here! You can <ansi fg="command">look</ansi> at or <ansi fg="command">play</ansi> it.`,
			)
			return d
		}
	}
	return d
}

// playClaw executes one claw machine attempt for the user.
func (g *GamblingModule) playClaw(user *users.UserRecord, room *rooms.Room) {

	user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

	cost := defaultClawCost
	if v, ok := g.plug.Config.Get(`ClawCost`).(int); ok && v > 0 {
		cost = v
	}

	winChance := defaultClawWinChance
	if v, ok := g.plug.Config.Get(`ClawWinChance`).(int); ok && v > 0 {
		winChance = v
	}

	if user.Character.Gold < cost {
		user.SendText(fmt.Sprintf(
			`You need at least <ansi fg="gold">%d gold</ansi> to play the claw machine.`,
			cost,
		))
		return
	}

	user.Character.Gold -= cost
	events.AddToQueue(events.EquipmentChange{
		UserId:     user.UserId,
		GoldChange: -cost,
	})

	user.SendText(term.CRLFStr)
	user.SendText("You put your money in and grip the joystick...")

	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> feeds coins into the claw machine and grips the joystick...`,
			user.Character.Name),
		user.UserId,
	)

	user.SendText(term.CRLFStr)

	if util.Rand(100) >= winChance {
		user.SendText(`<ansi fg="cyan">The claw descends...</ansi> hovers tantalizingly over a prize...`)
		user.SendText(term.CRLFStr)
		user.SendText(`<ansi fg="8">...and drops it at the last moment. Better luck next time!</ansi>`)
		user.SendText(term.CRLFStr)
		room.SendText(
			fmt.Sprintf(`<ansi fg="8">The claw drops its prize just before the chute. <ansi fg="username">%s</ansi> walks away empty-handed.</ansi>`,
				user.Character.Name),
			user.UserId,
		)
		return
	}

	selected := g.pickClawPrize()
	if selected == nil {
		user.SendText(`<ansi fg="8">The claw machine whirs but the prize pool is empty. (All prize weights are zero.)</ansi>`)
		user.SendText(term.CRLFStr)
		return
	}

	prize := items.New(selected.itemId)
	if !prize.IsValid() {
		user.SendText(`<ansi fg="8">The claw machine whirs but produces nothing. (Something went wrong internally.)</ansi>`)
		user.SendText(term.CRLFStr)
		return
	}

	if !user.Character.StoreItem(prize) {
		user.SendText(fmt.Sprintf(
			`<ansi fg="cyan">The claw snatches up a <ansi fg="item">%s</ansi>!</ansi> <ansi fg="214">But your backpack is full — it tumbles to the floor.</ansi>`,
			selected.name,
		))

		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> wins a <ansi fg="item">%s</ansi> but it falls to the floor!`,
				user.Character.Name,
				selected.name),
			user.UserId,
		)

		room.AddItem(prize, false)
		user.SendText(term.CRLFStr)
		return
	}

	events.AddToQueue(events.ItemOwnership{
		UserId: user.UserId,
		Item:   prize,
		Gained: true,
	})

	user.SendText(`<ansi fg="cyan-bold">The claw descends with purpose...</ansi>`)

	user.SendText(term.CRLFStr)

	user.SendText(fmt.Sprintf(
		`<ansi fg="green-bold">...and snatches up a <ansi fg="item">%s</ansi>!</ansi> It drops neatly into the prize chute. <ansi fg="yellow-bold">Congratulations!</ansi>`,
		selected.name,
	))

	user.SendText(term.CRLFStr)

	room.SendText(
		fmt.Sprintf(`<ansi fg="cyan-bold">The claw machine rattles!</ansi> <ansi fg="username">%s</ansi> <ansi fg="green">wins a prize!</ansi>`,
			user.Character.Name),
		user.UserId,
	)
}

// clawMachineNounDesc returns the noun description shown when a player types
// "look claw machine" in a room with the claw machine tag.
func (g *GamblingModule) clawMachineNounDesc(room *rooms.Room) string {

	cost := defaultClawCost
	if v, ok := g.plug.Config.Get(`ClawCost`).(int); ok && v > 0 {
		cost = v
	}
	winChance := defaultClawWinChance
	if v, ok := g.plug.Config.Get(`ClawWinChance`).(int); ok && v > 0 {
		winChance = v
	}

	totalWeight := 0
	for i := range clawPrizes {
		totalWeight += g.clawPrizeWeight(clawPrizes[i])
	}

	type weightedPrize struct {
		prize clawPrize
		wt    int
	}
	sorted := make([]weightedPrize, 0, len(clawPrizes))
	for i := range clawPrizes {
		wt := g.clawPrizeWeight(clawPrizes[i])
		if wt > 0 {
			sorted = append(sorted, weightedPrize{clawPrizes[i], wt})
		}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].wt > sorted[j].wt })

	var sb strings.Builder
	sb.WriteString(`<ansi fg="cyan">╔════════════════════════════════╗</ansi>` + "\n")
	sb.WriteString(`<ansi fg="cyan">║</ansi>     <ansi fg="cyan-bold">C L A W  M A C H I N E</ansi>     <ansi fg="cyan">║</ansi>` + "\n")
	sb.WriteString(`<ansi fg="cyan">╚════════════════════════════════╝</ansi>` + "\n")
	sb.WriteString("\n")
	sb.WriteString("A tall glass cabinet filled with small prizes, lit from within by a warm glow.\n")
	sb.WriteString("A mechanical claw hangs from a gantry inside, waiting to be guided by a brave soul.\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("    Cost to play:  <ansi fg=\"gold\">%d gold</ansi>\n", cost))
	sb.WriteString(fmt.Sprintf("    Chance to win: <ansi fg=\"cyan-bold\">%d%%</ansi>\n", winChance))
	sb.WriteString("\n")
	sb.WriteString(`<ansi fg="cyan">Prizes</ansi> <ansi fg="8">(chance on win):</ansi>` + "\n")
	for _, wp := range sorted {
		pct := 0
		if totalWeight > 0 {
			pct = wp.wt * 100 / totalWeight
		}
		sb.WriteString(fmt.Sprintf("    <ansi fg=\"item\">%-16s</ansi>  <ansi fg=\"cyan\">%d%%</ansi>\n", wp.prize.name, pct))
	}
	sb.WriteString("\n")
	sb.WriteString(`Type <ansi fg="command">play claw machine</ansi> to try your luck.`)
	return sb.String()
}
