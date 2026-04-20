package gambling

import (
	"fmt"
	"strings"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// slotSymbol represents one reel symbol with a display glyph and relative weight.
type slotSymbol struct {
	glyph  string
	weight int
}

// slotOutcome maps a result category to a payout multiplier (applied to the cost).
// A multiplier of 0 means no payout (loss).
type slotOutcome struct {
	label      string
	multiplier int // payout = cost * multiplier  (0 = lose, 1 = break-even, 2+ = win)
}

var (
	reelSymbols = []slotSymbol{
		{`cherry`, 30},
		{`lemon`, 25},
		{`orange`, 20},
		{`plum`, 15},
		{`bell`, 7},
		{`bar`, 2},
		{`seven`, 1},
	}

	// symbolColors maps each glyph to an ANSI fg color for reel display.
	symbolColors = map[string]string{
		`cherry`: `red-bold`,
		`lemon`:  `yellow-bold`,
		`orange`: `214`,
		`plum`:   `magenta-bold`,
		`bell`:   `cyan-bold`,
		`bar`:    `white-bold`,
		`seven`:  `220`,
	}

	// slotMu guards the jackpot state.
	slotMu sync.Mutex
)

// SlotState holds the persistent jackpot pool.
type SlotState struct {
	Jackpot int `yaml:"Jackpot"`
}

// roomHasSlots returns true when the room carries a "slots" or "slot machine" tag.
func roomHasSlots(r *rooms.Room) bool {
	for _, t := range r.Tags {
		tl := strings.ToLower(t)
		if tl == `slots` || tl == `slot machine` {
			return true
		}
	}
	return false
}

// coloredGlyph wraps a symbol glyph in its designated ANSI color tag.
func coloredGlyph(s slotSymbol) string {
	color, ok := symbolColors[s.glyph]
	if !ok {
		color = `white`
	}
	return fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, color, s.glyph)
}

// spinReel picks one symbol according to weighted random selection.
func spinReel() slotSymbol {
	total := 0
	for _, s := range reelSymbols {
		total += s.weight
	}
	roll := util.Rand(total)
	cumulative := 0
	for _, s := range reelSymbols {
		cumulative += s.weight
		if roll < cumulative {
			return s
		}
	}
	return reelSymbols[len(reelSymbols)-1]
}

// evaluate returns the outcome for a three-reel spin.
func evaluate(a, b, c slotSymbol) slotOutcome {
	if a.glyph == b.glyph && b.glyph == c.glyph {
		switch a.glyph {
		case `seven`:
			return slotOutcome{`JACKPOT`, 0} // special: wins entire jackpot
		case `bar`:
			return slotOutcome{`TRIPLE BAR`, 20}
		case `bell`:
			return slotOutcome{`TRIPLE BELL`, 10}
		default:
			return slotOutcome{`TRIPLE ` + strings.ToUpper(a.glyph), 5}
		}
	}
	if a.glyph == b.glyph || b.glyph == c.glyph || a.glyph == c.glyph {
		return slotOutcome{`PAIR`, 2}
	}
	cherryCount := 0
	for _, s := range []slotSymbol{a, b, c} {
		if s.glyph == `cherry` {
			cherryCount++
		}
	}
	if cherryCount >= 2 {
		return slotOutcome{`CHERRIES`, 2}
	}
	return slotOutcome{``, 0}
}

// jackpotBanner returns the festive multi-color JACKPOT banner line.
func jackpotBanner() string {
	// Each letter of JACKPOT cycles through festive bold colors.
	colors := []string{`220`, `red-bold`, `green-bold`, `cyan-bold`, `magenta-bold`, `214`, `yellow-bold`}
	letters := []string{`J`, `A`, `C`, `K`, `P`, `O`, `T`}
	out := `<ansi fg="220">*** </ansi>`
	for i, l := range letters {
		out += fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, colors[i%len(colors)], l)
	}
	out += `<ansi fg="220"> ***</ansi>`
	return out
}

// playSlots executes one spin for the user, charging the cost and paying out
// any winnings. It writes all output directly to the user and room.
func (g *GamblingModule) playSlots(user *users.UserRecord, room *rooms.Room) {

	cost := defaultCost
	if v, ok := g.plug.Config.Get(`SlotCost`).(int); ok && v > 0 {
		cost = v
	}

	if user.Character.Gold < cost {
		user.SendText(fmt.Sprintf(
			`You need at least <ansi fg="gold">%d gold</ansi> to play the slot machine.`,
			cost,
		))
		return
	}

	// Deduct cost and add to jackpot pool.
	user.Character.Gold -= cost

	slotMu.Lock()
	g.state.Jackpot += cost / 2 // half of each play feeds the jackpot
	slotMu.Unlock()

	events.AddToQueue(events.EquipmentChange{
		UserId:     user.UserId,
		GoldChange: -cost,
	})

	a, b, c := spinReel(), spinReel(), spinReel()

	reelLine := fmt.Sprintf(
		`<ansi fg="yellow">[ </ansi>%s <ansi fg="yellow">|</ansi> %s <ansi fg="yellow">|</ansi> %s<ansi fg="yellow"> ]</ansi>`,
		coloredGlyph(a), coloredGlyph(b), coloredGlyph(c),
	)

	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> pulls the lever on the slot machine...`,
			user.Character.Name),
		user.UserId,
	)

	outcome := evaluate(a, b, c)

	if outcome.label == `JACKPOT` {
		slotMu.Lock()
		prize := g.state.Jackpot
		g.state.Jackpot = 0
		slotMu.Unlock()

		user.Character.Gold += prize
		events.AddToQueue(events.EquipmentChange{
			UserId:     user.UserId,
			GoldChange: prize,
		})

		banner := jackpotBanner()
		user.SendText(" ")
		user.SendText(reelLine)
		user.SendText(" ")
		user.SendText(fmt.Sprintf(`%s <ansi fg="gold">You win %d gold!</ansi>`, banner, prize))
		user.SendText(" ")
		room.SendText(
			fmt.Sprintf(`%s <ansi fg="username">%s</ansi> <ansi fg="yellow-bold">has hit the JACKPOT!!!</ansi>`,
				banner, user.Character.Name),
			user.UserId,
		)
		return
	}

	if outcome.multiplier > 0 {
		prize := cost * outcome.multiplier

		// Color the outcome label by tier.
		var labelColor string
		switch {
		case outcome.multiplier >= 20:
			labelColor = `gold`
		case outcome.multiplier >= 10:
			labelColor = `cyan-bold`
		default:
			labelColor = `green-bold`
		}

		user.Character.Gold += prize
		events.AddToQueue(events.EquipmentChange{
			UserId:     user.UserId,
			GoldChange: prize,
		})

		user.SendText(reelLine)
		user.SendText(" ")
		user.SendText(fmt.Sprintf(
			`<ansi fg="%s">%s!</ansi> You win <ansi fg="gold">%d gold</ansi>!`,
			labelColor, outcome.label, prize,
		))
		user.SendText(" ")
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="green">wins</ansi> on the slot machine!`, user.Character.Name),
			user.UserId,
		)
		return
	}

	user.SendText(reelLine)
	user.SendText(fmt.Sprintf(`<ansi fg="8">No luck this time. You lost <ansi fg="gold">%d gold</ansi>.</ansi>`, cost))
	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="8">loses on the slot machine.</ansi>`, user.Character.Name),
		user.UserId,
	)
}

// slotMachineNounDesc returns the noun description shown when a player types
// "look slot machine" in a room with the slots tag.
func (g *GamblingModule) slotMachineNounDesc(room *rooms.Room) string {
	cost := defaultCost
	if v, ok := g.plug.Config.Get(`SlotCost`).(int); ok && v > 0 {
		cost = v
	}
	slotMu.Lock()
	jackpot := g.state.Jackpot
	slotMu.Unlock()
	return fmt.Sprintf(
		`A gleaming mechanical contraption adorned with spinning reels and flashing lights. A worn lever protrudes from its side. Cost to play: <ansi fg="gold">%d gold</ansi>. Current jackpot: <ansi fg="220">%d gold</ansi>. Type <ansi fg="command">play slots</ansi> to try your luck.`,
		cost, jackpot,
	)
}


func (g *GamblingModule) lookSlotMachine(user *users.UserRecord, room *rooms.Room) {

	cost := defaultCost
	if v, ok := g.plug.Config.Get(`SlotCost`).(int); ok && v > 0 {
		cost = v
	}

	slotMu.Lock()
	jackpot := g.state.Jackpot
	slotMu.Unlock()

	user.SendText(``)
	user.SendText(`<ansi fg="220">╔════════════════════════════════╗</ansi>`)
	user.SendText(`<ansi fg="220">║</ansi>     <ansi fg="yellow-bold">S L O T  M A C H I N E</ansi>     <ansi fg="220">║</ansi>`)
	user.SendText(`<ansi fg="220">╚════════════════════════════════╝</ansi>`)
	user.SendText(``)
	user.SendText(`A gleaming mechanical contraption adorned with spinning reels and flashing lights.`)
	user.SendText(`A worn lever protrudes from its side, inviting the bold and foolish alike.`)
	user.SendText(`A placard on the front reads:`)
	user.SendText(``)
	user.SendText(fmt.Sprintf(
		`    Cost to play:    <ansi fg="gold">%d gold</ansi>`,
		cost,
	))
	user.SendText(fmt.Sprintf(
		`    Current jackpot: <ansi fg="gold">%d gold</ansi>`,
		jackpot,
	))
	user.SendText(``)
	user.SendText(`Type <ansi fg="command">play slots</ansi> to try your luck.`)
	user.SendText(``)

	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> examines the slot machine.`, user.Character.Name),
		user.UserId,
	)
}
