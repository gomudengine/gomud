package elections

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/colorpatterns"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	//go:embed files/*
	files embed.FS
)

const (
	pollTag         = "election poll"
	cofferTag       = "coffer"
	officialOnlyTag = "elected officials only"
	maxCoffer       = 100000000
)

func init() {
	m := &ElectionsModule{
		plug: plugins.New(`elections`, `1.0`),
	}

	if err := m.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	m.plug.ReserveTags(pollTag, cofferTag, officialOnlyTag)

	m.plug.Web.AdminPage("Config", "elections-config", "html/admin/elections-config.html", true, "Modules", "Elections", nil)
	m.plug.Web.AdminPage("About", "elections-about", "html/admin/elections-about.html", true, "Modules", "Elections", nil)
	m.plug.Web.AdminPage("Audit", "elections-audit", "html/admin/elections-audit.html", true, "Modules", "Elections", nil)
	m.plug.Web.AdminPage("API Docs", "elections-api", "html/admin/elections-api.html", true, "Modules", "Elections", nil)

	m.plug.Web.AdminAPIEndpoint("GET", "elections/zones", m.apiGetZones)
	m.plug.Web.AdminAPIEndpoint("PATCH", "elections/zones/{zone}", m.apiPatchZone, "elections.write")
	m.plug.Web.AdminAPIEndpoint("DELETE", "elections/zones/{zone}/official", m.apiDeleteZoneOfficial, "elections.write")
	m.plug.Web.RegisterPermissions(plugins.ModulePermission{
		Key:         "elections.write",
		Description: "Edit zone tax rates, coffers, and officials",
		Category:    "Modules",
	})

	m.plug.AddUserCommand(`election`, m.electionAdminCommand, true, true)
	m.plug.AddUserCommand(`coffer`, m.cofferCommand, false, false)

	m.plug.Callbacks.SetOnLoad(m.load)
	m.plug.Callbacks.SetOnSave(m.save)

	events.RegisterListener(events.Input{}, m.onInput, events.First)
	events.RegisterListener(events.Input{}, m.onMovementGate, events.First)
	events.RegisterListener(events.PlayerSpawn{}, m.onPlayerSpawn)
	events.RegisterListener(events.NewRound{}, m.onNewRound)
	events.RegisterListener(events.Purchase{}, m.onPurchase)

	rooms.OnRoomLook.Register(m.onRoomLook)
	characters.OnGetFormattedName.Register(m.onGetFormattedName)
	usercommands.OnShopList.Register(m.onShopList)
	usercommands.OnShopListRendered.Register(m.onShopListRendered)
	usercommands.OnInsufficientFunds.Register(m.onInsufficientFunds)
}

// ElectionsState is the persisted state for the elections module.
type ElectionsState struct {
	ActiveElection *Election         `yaml:"activeelection,omitempty"`
	Winners        map[string]Winner `yaml:"winners,omitempty"`
	Coffers        map[string]int    `yaml:"coffers,omitempty"`
	TaxRates       map[string]int    `yaml:"taxrates,omitempty"`
}

// Election represents a running election.
type Election struct {
	Title      string      `yaml:"title"`
	Zone       string      `yaml:"zone"`
	StartRound uint64      `yaml:"startround"`
	Nominees   []string    `yaml:"nominees,omitempty"`
	NomineeIds []int       `yaml:"nomineeids,omitempty"`
	Votes      map[int]int `yaml:"votes,omitempty"`
}

// Winner records the outcome of a completed election.
type Winner struct {
	CharacterName     string `yaml:"charactername"`
	UserId            int    `yaml:"userid"`
	Title             string `yaml:"title"`
	LastElectionRound uint64 `yaml:"lastelectionround,omitempty"`
}

// ElectionsModule owns all elections state.
type ElectionsModule struct {
	plug  *plugins.Plugin
	state ElectionsState
}

func (m *ElectionsModule) load() {
	m.plug.ReadIntoStruct(`elections-state`, &m.state)
	if m.state.Winners == nil {
		m.state.Winners = make(map[string]Winner)
	}
	if m.state.Coffers == nil {
		m.state.Coffers = make(map[string]int)
	}
	if m.state.TaxRates == nil {
		m.state.TaxRates = make(map[string]int)
	}
	if m.state.ActiveElection != nil && m.state.ActiveElection.Votes == nil {
		m.state.ActiveElection.Votes = make(map[int]int)
	}
}

func (m *ElectionsModule) save() {
	m.plug.WriteStruct(`elections-state`, m.state)
}

// zoneTaxRate returns the tax rate (0-100) for the given zone key.
// Defaults to 1 if no rate has been explicitly set.
func (m *ElectionsModule) zoneTaxRate(zoneKey string) int {
	if rate, ok := m.state.TaxRates[zoneKey]; ok {
		return rate
	}
	return 1
}

// electionAlert formats a bordered election update message.
func electionAlert(lines ...string) string {
	border := `<ansi fg="yellow-bold">********************</ansi> <ansi fg="white-bold">Election Update</ansi> <ansi fg="yellow-bold">********************</ansi>`
	close := `<ansi fg="yellow-bold">*********************************************************</ansi>`
	parts := []string{border}
	parts = append(parts, lines...)
	parts = append(parts, close)
	return strings.Join(parts, "\n")
}

// broadcastZoneAlert sends a message to all players currently in any room
// belonging to the named zone.
func broadcastZoneAlert(zoneName string, msg string) {
	for _, roomId := range rooms.GetAllZoneRoomsIds(zoneName) {
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}
		for _, uid := range room.GetPlayers() {
			if u := users.GetByUserId(uid); u != nil {
				u.SendText(msg)
			}
		}
	}
}

// broadcastAlert sends an alert to all online players.
func broadcastAlert(msg string) {
	for _, uid := range users.GetOnlineUserIds() {
		if u := users.GetByUserId(uid); u != nil {
			u.SendText(msg)
		}
	}
}

func (m *ElectionsModule) electionDuration() string {
	if v, ok := m.plug.Config.Get(`ElectionDuration`).(string); ok && v != `` {
		return v
	}
	return `2 days`
}

func (m *ElectionsModule) electionCyclePeriod() string {
	if v, ok := m.plug.Config.Get(`ElectionCyclePeriod`).(string); ok {
		return v
	}
	return `1 year`
}

func (m *ElectionsModule) titleColor() string {
	if v, ok := m.plug.Config.Get(`TitleColor`).(string); ok && v != `` {
		return v
	}
	return `yellow`
}

// daysRemaining returns a human-readable days-remaining string for an election.
func (m *ElectionsModule) daysRemaining() string {
	if m.state.ActiveElection == nil {
		return `0 days`
	}
	gd := gametime.GetDate(m.state.ActiveElection.StartRound)
	endRound := gd.AddPeriod(m.electionDuration())
	now := util.GetRoundCount()
	if endRound <= now {
		return `0 days`
	}
	remaining := endRound - now
	// Use config rounds-per-day for conversion.
	roundsPerDay := gd.RoundsPerDay
	if roundsPerDay < 1 {
		roundsPerDay = 480
	}
	days := (int(remaining) + roundsPerDay - 1) / roundsPerDay
	if days == 1 {
		return `1 day`
	}
	return fmt.Sprintf(`%d days`, days)
}

// endElection tallies votes, saves the winner, and broadcasts the result.
// adminUser may be nil for auto-end.
func (m *ElectionsModule) endElection(adminUser *users.UserRecord) {
	el := m.state.ActiveElection
	if el == nil {
		return
	}

	// Tally votes: count votes per nominee userId.
	voteCounts := make(map[int]int, len(el.NomineeIds))
	for _, nomineeId := range el.NomineeIds {
		voteCounts[nomineeId] = 0
	}
	for _, nomineeId := range el.Votes {
		voteCounts[nomineeId]++
	}

	m.state.ActiveElection = nil

	// No nominees or no votes cast — election ends with no winner.
	if len(el.NomineeIds) == 0 || len(el.Votes) == 0 {
		msg := electionAlert(
			fmt.Sprintf(`The election for <ansi fg="white-bold">%s</ansi> has ended with no winner.`, el.Title),
		)
		broadcastAlert(msg)
		return
	}

	// Find the highest vote count, collect all nominees tied at that count.
	topVotes := 0
	for _, count := range voteCounts {
		if count > topVotes {
			topVotes = count
		}
	}
	type candidate struct {
		userId int
		name   string
	}
	var tied []candidate
	for i, nomineeId := range el.NomineeIds {
		if voteCounts[nomineeId] == topVotes {
			tied = append(tied, candidate{userId: nomineeId, name: el.Nominees[i]})
		}
	}

	// Pick randomly among tied candidates.
	winner := tied[util.Rand(len(tied))]
	winnerUserId := winner.userId
	winnerName := winner.name

	zoneKey := strings.ToLower(el.Zone)
	m.state.Winners[zoneKey] = Winner{
		CharacterName:     winnerName,
		UserId:            winnerUserId,
		Title:             el.Title,
		LastElectionRound: util.GetRoundCount(),
	}

	msg := electionAlert(
		fmt.Sprintf(`The election has ended for <ansi fg="white-bold">%s</ansi>, and the winner is <ansi fg="username">%s</ansi>!`, el.Title, winnerName),
		fmt.Sprintf(`Congratulations <ansi fg="username">%s</ansi>, <ansi fg="white-bold">%s</ansi>!`, winnerName, el.Title),
	)
	broadcastAlert(msg)
}

// electionAdminCommand handles the `election` admin command.
func (m *ElectionsModule) electionAdminCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	args := strings.Fields(rest)

	if len(args) == 0 {
		tplTxt, _ := templates.Process(`help/elections-admin`, nil, user.UserId)
		user.SendText(tplTxt)
		return true, nil
	}

	switch strings.ToLower(args[0]) {

	case `start`:
		if len(args) < 2 {
			user.SendText(`Usage: <ansi fg="command">election start <title></ansi>` + term.CRLFStr)
			return true, nil
		}

		if m.state.ActiveElection != nil {
			user.SendText(fmt.Sprintf(`An election for <ansi fg="white-bold">%s</ansi> is already in progress.`+term.CRLFStr, m.state.ActiveElection.Title))
			return true, nil
		}

		title := strings.Join(args[1:], ` `)
		zone := room.Zone

		m.state.ActiveElection = &Election{
			Title:      title,
			Zone:       zone,
			StartRound: util.GetRoundCount(),
			Votes:      make(map[int]int),
		}

		daysLeft := m.daysRemaining()
		msg := electionAlert(
			fmt.Sprintf(`An election has started for "<ansi fg="white-bold">%s</ansi>."`, title),
			`Type <ansi fg="command">help elections</ansi> to find out more about it.`,
			`Cast your vote at the nearest polling location.`,
			fmt.Sprintf(`The election ends in <ansi fg="white-bold">%s</ansi>!`, daysLeft),
		)
		broadcastAlert(msg)

	case `end`:
		if m.state.ActiveElection == nil {
			user.SendText(`There is no active election to end.` + term.CRLFStr)
			return true, nil
		}
		m.endElection(user)

	case `taxrate`:
		// Admin/mod only. Usage:
		//   election taxrate <0-100>            (uses current room's zone)
		//   election taxrate <zonename> <0-100> (explicit zone)
		if len(args) < 2 {
			user.SendText(`Usage: <ansi fg="command">election taxrate <0-100></ansi> or <ansi fg="command">election taxrate <zone> <0-100></ansi>` + term.CRLFStr)
			return true, nil
		}

		zoneKey := strings.ToLower(room.Zone)
		rateStr := args[1]

		if len(args) >= 3 {
			zoneKey = strings.ToLower(args[1])
			rateStr = args[2]
		}

		rate, err := strconv.Atoi(rateStr)
		if err != nil || rate < 0 || rate > 100 {
			user.SendText(`Tax rate must be a number between 0 and 100.` + term.CRLFStr)
			return true, nil
		}

		m.state.TaxRates[zoneKey] = rate
		user.SendText(fmt.Sprintf(`Tax rate for <ansi fg="white-bold">%s</ansi> set to <ansi fg="yellow">%d%%</ansi>.`+term.CRLFStr, zoneKey, rate))
		broadcastZoneAlert(zoneKey, fmt.Sprintf(
			`<ansi fg="username">%s</ansi> has set the tax rate in <ansi fg="white-bold">%s</ansi> to <ansi fg="yellow">%d%%</ansi>!`,
			user.Character.Name, zoneKey, rate,
		))

	default:
		tplTxt, _ := templates.Process(`help/elections-admin`, nil, user.UserId)
		user.SendText(tplTxt)
	}

	return true, nil
}

// cofferCommand handles the `coffer` user command.
func (m *ElectionsModule) cofferCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	if !room.HasTag(cofferTag) {
		user.SendText(`You are not at a coffer location.` + term.CRLFStr)
		return true, nil
	}

	zoneKey := strings.ToLower(room.Zone)
	balance := m.state.Coffers[zoneKey]

	if rest == `` {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`The coffer of <ansi fg="white-bold">%s</ansi> holds <ansi fg="gold">%d gold</ansi>.`, room.Zone, balance))
		user.SendText(`You can <ansi fg="command">coffer deposit</ansi> or <ansi fg="command">coffer withdraw</ansi> from the coffer.` + term.CRLFStr)
		return true, nil
	}

	// Only the zone's current title-holder may deposit or withdraw.
	winner, hasWinner := m.state.Winners[zoneKey]
	if !hasWinner || winner.UserId != user.UserId {
		user.SendText(fmt.Sprintf(`Only the elected <ansi fg="white-bold">%s</ansi> may manage this coffer.`+term.CRLFStr, func() string {
			if hasWinner {
				return winner.Title
			}
			return `leader`
		}()))
		return true, nil
	}

	if rest == `deposit` || rest == `withdraw` {
		user.SendText(fmt.Sprintf(`%s how much? Include the amount of gold or "all".%s`, rest, term.CRLFStr))
		return true, nil
	}

	parts := strings.Fields(strings.ToLower(rest))
	if len(parts) < 2 || (parts[0] != `deposit` && parts[0] != `withdraw`) {
		user.SendText(`Try <ansi fg="command">help elections</ansi> for more information about coffers.` + term.CRLFStr)
		return true, nil
	}

	action := parts[0]
	amountStr := parts[1]
	amount, _ := strconv.Atoi(amountStr)

	if amount < 1 && amountStr != `all` {
		user.SendText(fmt.Sprintf(`You must specify an amount greater than zero to %s.%s`, action, term.CRLFStr))
		return true, nil
	}

	if action == `deposit` {
		if amountStr == `all` {
			amount = user.Character.Gold
		}
		if amount > user.Character.Gold {
			amount = user.Character.Gold
			user.SendText(`You don't have that much gold on hand, but you deposit what you have.`)
		}
		newBalance := balance + amount
		if newBalance > maxCoffer {
			amount = maxCoffer - balance
			newBalance = maxCoffer
			user.SendText(`The coffer is nearly full; only a portion was deposited.`)
		}
		user.Character.Gold -= amount
		m.state.Coffers[zoneKey] = newBalance

		events.AddToQueue(events.EquipmentChange{
			UserId:     user.UserId,
			GoldChange: -amount,
		})

		user.SendText(fmt.Sprintf(`You deposit <ansi fg="gold">%d gold</ansi> into the coffer.`, amount))
		user.SendText(fmt.Sprintf(`The coffer now holds <ansi fg="gold">%d gold</ansi>.`, newBalance))

	} else if action == `withdraw` {
		if amountStr == `all` {
			amount = balance
		}
		if amount > balance {
			amount = balance
			user.SendText(`The coffer doesn't hold that much, but you withdraw what is there.`)
		}
		newBalance := balance - amount
		user.Character.Gold += amount
		m.state.Coffers[zoneKey] = newBalance

		events.AddToQueue(events.EquipmentChange{
			UserId:     user.UserId,
			GoldChange: amount,
		})

		user.SendText(fmt.Sprintf(`You withdraw <ansi fg="gold">%d gold</ansi> from the coffer.`, amount))
		user.SendText(fmt.Sprintf(`The coffer now holds <ansi fg="gold">%d gold</ansi>.`, newBalance))
	}

	user.SendText(``)
	return true, nil
}

// onInput intercepts `nominate`, `vote`, and `taxrate` commands.
func (m *ElectionsModule) onInput(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.Input)
	if !ok || evt.UserId == 0 {
		return events.Continue
	}

	inputLower := strings.TrimSpace(strings.ToLower(evt.InputText))
	cmd, rest, _ := strings.Cut(inputLower, ` `)

	switch cmd {
	case `nominate`:
		return m.handleNominate(evt.UserId, strings.TrimSpace(rest))
	case `vote`:
		return m.handleVote(evt.UserId, strings.TrimSpace(rest))
	case `taxrate`:
		return m.handleTaxRate(evt.UserId, strings.TrimSpace(rest))
	}

	return events.Continue
}

func (m *ElectionsModule) handleTaxRate(userId int, rateStr string) events.ListenerReturn {
	user := users.GetByUserId(userId)
	if user == nil {
		return events.Continue
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil || !room.HasTag(cofferTag) {
		return events.Continue
	}

	zoneKey := strings.ToLower(room.Zone)
	winner, hasWinner := m.state.Winners[zoneKey]
	if !hasWinner || winner.UserId != userId {
		return events.Continue
	}

	// At this point: elected official, coffer room, their zone.
	if rateStr == `` {
		user.SendText(fmt.Sprintf(
			`Current tax rate for <ansi fg="white-bold">%s</ansi> is <ansi fg="yellow">%d%%</ansi>. Use <ansi fg="command">taxrate <0-100></ansi> to change it.`,
			room.Zone, m.zoneTaxRate(zoneKey),
		) + term.CRLFStr)
		return events.Cancel
	}

	rate, err := strconv.Atoi(rateStr)
	if err != nil || rate < 0 || rate > 100 {
		user.SendText(`Tax rate must be a number between 0 and 100.` + term.CRLFStr)
		return events.Cancel
	}

	m.state.TaxRates[zoneKey] = rate
	user.SendText(fmt.Sprintf(
		`Tax rate for <ansi fg="white-bold">%s</ansi> set to <ansi fg="yellow">%d%%</ansi>.`+term.CRLFStr,
		room.Zone, rate,
	))
	broadcastZoneAlert(room.Zone, fmt.Sprintf(
		`<ansi fg="username">%s</ansi> has set the tax rate in <ansi fg="white-bold">%s</ansi> to <ansi fg="yellow">%d%%</ansi>!`,
		user.Character.Name, room.Zone, rate,
	))
	return events.Cancel
}

func (m *ElectionsModule) handleNominate(userId int, targetName string) events.ListenerReturn {
	user := users.GetByUserId(userId)
	if user == nil {
		return events.Continue
	}

	if m.state.ActiveElection == nil {
		return events.Continue
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil || !room.HasTag(pollTag) {
		user.SendText(`You must be at a polling location to nominate someone.`)
		return events.Cancel
	}

	if targetName == `` {
		user.SendText(`Nominate who? Try <ansi fg="command">nominate <playername></ansi>.`)
		return events.Cancel
	}

	// Find the target player online (exact match, case-insensitive).
	var targetUser *users.UserRecord
	for _, uid := range users.GetOnlineUserIds() {
		u := users.GetByUserId(uid)
		if u != nil && strings.EqualFold(u.Character.Name, targetName) {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		user.SendText(fmt.Sprintf(`No online player named <ansi fg="username">%s</ansi> was found.`, targetName))
		return events.Cancel
	}

	el := m.state.ActiveElection

	// Check for duplicate nomination.
	for _, nid := range el.NomineeIds {
		if nid == targetUser.UserId {
			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is already on the ballot.`, targetUser.Character.Name))
			return events.Cancel
		}
	}

	el.Nominees = append(el.Nominees, targetUser.Character.Name)
	el.NomineeIds = append(el.NomineeIds, targetUser.UserId)

	msg := electionAlert(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> was nominated for <ansi fg="white-bold">%s</ansi>!`, targetUser.Character.Name, el.Title),
	)
	broadcastAlert(msg)

	return events.Cancel
}

func (m *ElectionsModule) handleVote(userId int, targetName string) events.ListenerReturn {
	user := users.GetByUserId(userId)
	if user == nil {
		return events.Continue
	}

	if m.state.ActiveElection == nil {
		return events.Continue
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil || !room.HasTag(pollTag) {
		user.SendText(`You must be at a polling location to vote for someone.`)
		return events.Cancel
	}

	el := m.state.ActiveElection

	// `vote` with no argument: list candidates.
	if targetName == `` {
		if len(el.Nominees) == 0 {
			user.SendText(fmt.Sprintf(`No one has been nominated yet for <ansi fg="white-bold">%s</ansi>.`, el.Title))
			return events.Cancel
		}

		// Tally current votes.
		voteCounts := make(map[int]int, len(el.NomineeIds))
		for _, nomineeId := range el.Votes {
			voteCounts[nomineeId]++
		}
		totalVotes := len(el.Votes)

		// Build a sorted list of name+index pairs.
		type entry struct {
			name   string
			userId int
		}
		entries := make([]entry, len(el.Nominees))
		for i, name := range el.Nominees {
			entries[i] = entry{name: name, userId: el.NomineeIds[i]}
		}
		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
		})

		headers := []string{`Candidate`, `% of Vote`}
		rows := make([][]string, len(entries))
		for i, e := range entries {
			pct := 0
			if totalVotes > 0 {
				pct = voteCounts[e.userId] * 100 / totalVotes
			}
			rows[i] = []string{e.name, fmt.Sprintf(`%d%%`, pct)}
		}
		formatting := []string{`<ansi fg="username">%s</ansi>`, `<ansi fg="white">%s</ansi>`}
		tblData := templates.GetTable(fmt.Sprintf(`Candidates for %s`, el.Title), headers, rows, formatting)
		tplTxt, _ := templates.Process(`tables/generic`, tblData, user.UserId)
		user.SendText(tplTxt)
		user.SendText(`Type <ansi fg="command">vote <name></ansi> to cast your vote.`)
		return events.Cancel
	}

	// Check if already voted.
	if _, alreadyVoted := el.Votes[userId]; alreadyVoted {
		user.SendText(`You have already cast your vote in this election.`)
		return events.Cancel
	}

	// Find nominee by name.
	nomineeUserId := 0
	nomineeName := ``
	for i, name := range el.Nominees {
		if strings.EqualFold(name, targetName) {
			nomineeUserId = el.NomineeIds[i]
			nomineeName = name
			break
		}
	}

	if nomineeUserId == 0 {
		user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is not on the ballot. Type <ansi fg="command">vote</ansi> to see candidates.`, targetName))
		return events.Cancel
	}

	el.Votes[userId] = nomineeUserId
	user.SendText(fmt.Sprintf(`You have cast your vote for <ansi fg="username">%s</ansi>.`, nomineeName))

	return events.Cancel
}

// onPlayerSpawn sends the election reminder to a newly spawned player.
func (m *ElectionsModule) onPlayerSpawn(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.PlayerSpawn)
	if !ok {
		return events.Continue
	}

	if m.state.ActiveElection == nil {
		return events.Continue
	}

	el := m.state.ActiveElection
	daysLeft := m.daysRemaining()

	u := users.GetByUserId(evt.UserId)
	if u == nil {
		return events.Continue
	}

	msg := electionAlert(
		fmt.Sprintf(`An election is ongoing for <ansi fg="white-bold">"%s."</ansi>`, el.Title),
		`Type <ansi fg="command">help elections</ansi> to find out more about it.`,
		`Cast your vote at the nearest polling location.`,
		fmt.Sprintf(`The election ends in <ansi fg="white-bold">%s</ansi>!`, daysLeft),
	)
	u.SendText(msg)

	return events.Continue
}

// onNewRound checks whether the election duration has elapsed and auto-ends if so.
// When no election is active, it checks whether any zone's election cycle period
// has elapsed and starts a new election if so.
func (m *ElectionsModule) onNewRound(e events.Event) events.ListenerReturn {
	if m.state.ActiveElection != nil {
		gd := gametime.GetDate(m.state.ActiveElection.StartRound)
		endRound := gd.AddPeriod(m.electionDuration())

		if util.GetRoundCount() >= endRound {
			m.endElection(nil)
		}
		return events.Continue
	}

	cyclePeriod := m.electionCyclePeriod()
	if cyclePeriod == `` {
		return events.Continue
	}

	now := util.GetRoundCount()
	for zoneKey, winner := range m.state.Winners {
		if winner.LastElectionRound == 0 {
			continue
		}
		gd := gametime.GetDate(winner.LastElectionRound)
		nextElectionRound := gd.AddPeriod(cyclePeriod)
		if now < nextElectionRound {
			continue
		}
		m.state.ActiveElection = &Election{
			Title:      winner.Title,
			Zone:       zoneKey,
			StartRound: now,
			Votes:      make(map[int]int),
		}
		daysLeft := m.daysRemaining()
		msg := electionAlert(
			fmt.Sprintf(`A new election has started for "<ansi fg="white-bold">%s</ansi>."`, winner.Title),
			`Type <ansi fg="command">help elections</ansi> to find out more about it.`,
			`Cast your vote at the nearest polling location.`,
			fmt.Sprintf(`The election ends in <ansi fg="white-bold">%s</ansi>!`, daysLeft),
		)
		broadcastAlert(msg)
		break
	}

	return events.Continue
}

// onPurchase adds the tax portion of a purchase to the zone coffer and
// notifies the buyer. evt.Cost is the tax-inclusive total charged by buy.go
// (because onShopList inflated the price for the buy path). The tax amount
// is back-calculated as total - base = total * rate / (100 + rate).
func (m *ElectionsModule) onPurchase(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.Purchase)
	if !ok || evt.Cost <= 0 {
		return events.Continue
	}

	zoneName := rooms.GetZoneForRoom(evt.RoomId)
	if zoneName == `` {
		return events.Continue
	}
	zoneKey := strings.ToLower(zoneName)
	rate := m.zoneTaxRate(zoneKey)
	if rate == 0 {
		return events.Continue
	}

	// Back-calculate the tax from the inflated total:
	// total = base + base*rate/100 = base*(100+rate)/100
	// tax   = total - base = total*rate/(100+rate)
	tax := evt.Cost * rate / (100 + rate)
	if tax < 1 {
		return events.Continue
	}

	current := m.state.Coffers[zoneKey]
	current += tax
	if current > maxCoffer {
		current = maxCoffer
	}
	m.state.Coffers[zoneKey] = current

	if u := users.GetByUserId(evt.UserId); u != nil {
		u.SendText(fmt.Sprintf(`You paid <ansi fg="gold">%d gold</ansi> in taxes.`, tax))
	}

	return events.Continue
}

// onInsufficientFunds fires when a buyer can't afford a purchase. If the
// zone has a tax rate and the buyer could afford the base price but not the
// tax-inclusive total, it sends a specific tax message and marks the request
// handled to suppress the generic insufficient-funds message.
func (m *ElectionsModule) onInsufficientFunds(r usercommands.InsufficientFundsRequest) usercommands.InsufficientFundsRequest {
	if r.Room == nil || r.Buyer == nil {
		return r
	}
	zoneKey := strings.ToLower(r.Room.Zone)
	rate := m.zoneTaxRate(zoneKey)
	if rate == 0 {
		return r
	}
	// Back-calculate the base price from the inflated total.
	// inflated = base * (100 + rate) / 100  =>  base = inflated * 100 / (100 + rate)
	base := r.Price * 100 / (100 + rate)
	if r.Gold >= base && r.Gold < r.Price {
		tax := r.Price - base
		r.Buyer.SendText(fmt.Sprintf(
			`You have enough for the <ansi fg="gold">%d gold</ansi> purchase, but not enough to pay the <ansi fg="gold">%d gold</ansi> tax (<ansi fg="yellow">%d%%</ansi>).`,
			base, tax, rate,
		))
		r.Handled = true
	}
	return r
}

// onShopList mutates item prices to be tax-inclusive for the buy command only,
// so that the affordability check in buy.go accounts for the full cost.
// The list command sees base prices unchanged.
// Prices of 0 mean "use the item's default value" — we resolve that here
// so the tax is applied to the true base price in all cases.
func (m *ElectionsModule) onShopList(r usercommands.ShopListRequest) usercommands.ShopListRequest {
	if r.Room == nil || !r.IsBuy {
		return r
	}
	zoneKey := strings.ToLower(r.Room.Zone)
	rate := m.zoneTaxRate(zoneKey)
	if rate == 0 {
		return r
	}
	for i, si := range r.Stock {
		base := si.Price
		if base < 0 {
			continue // free item, no tax
		}
		if base == 0 {
			// Resolve the default price the same way tryPurchase does.
			if si.ItemId > 0 {
				item := items.New(si.ItemId)
				base = item.GetSpec().Value
			} else if si.MobId > 0 {
				if mobSpec := mobs.GetMobSpec(mobs.MobId(si.MobId)); mobSpec != nil {
					base = int(configs.GetGamePlayConfig().MercHirePricePerLevel) * mobSpec.Character.Level
				}
			} else if si.BuffId > 0 {
				base = 1000
			} else if si.PetType != `` {
				base = 10000
			}
		}
		if base > 0 {
			r.Stock[i].Price = base + (base*rate)/100
		}
	}
	return r
}

// onShopListRendered sends the zone tax rate notice to the buyer after all
// shop tables for a seller have been rendered. Fires for the list command only.
func (m *ElectionsModule) onShopListRendered(r usercommands.ShopListRequest) usercommands.ShopListRequest {
	if r.Room == nil || r.Buyer == nil {
		return r
	}
	zoneKey := strings.ToLower(r.Room.Zone)
	rate := m.zoneTaxRate(zoneKey)
	r.Buyer.SendText(fmt.Sprintf(
		`<ansi fg="yellow">Zone tax rate: <ansi fg="white-bold">%d%%</ansi> (charged on top of base prices)</ansi>`,
		rate,
	))
	return r
}

// onRoomLook injects poll/coffer alerts into the room look details.
func (m *ElectionsModule) onRoomLook(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	for _, t := range d.Tags {
		if strings.EqualFold(t, pollTag) {
			if m.state.ActiveElection != nil {
				d.Alert(`<ansi fg="yellow-bold">This is a polling location!</ansi> <ansi fg="command">vote</ansi> for or <ansi fg="command">nominate</ansi> someone.`)
			}
		} else if strings.EqualFold(t, cofferTag) {
			d.Alert(`<ansi fg="yellow-bold">This room holds the local coffer.</ansi> Type <ansi fg="command">coffer</ansi> to manage it.`)
			zoneKey := strings.ToLower(d.Zone)
			if winner, hasWinner := m.state.Winners[zoneKey]; hasWinner && winner.UserId == d.UserId {
				rate := m.zoneTaxRate(zoneKey)
				d.Alert(fmt.Sprintf(`<ansi fg="yellow-bold">Set the zone tax rate (currently <ansi fg="white-bold">%d%%</ansi>) with <ansi fg="command">taxrate <0-100></ansi>.</ansi></ansi>`, rate))
			}
		} else if strings.EqualFold(t, officialOnlyTag) {
			zoneKey := strings.ToLower(d.Zone)
			if winner, hasWinner := m.state.Winners[zoneKey]; hasWinner {
				d.Alert(fmt.Sprintf(`<ansi fg="yellow-bold">This area is restricted to the <ansi fg="white-bold">%s</ansi> and their party.</ansi>`, winner.Title))
			} else {
				d.Alert(`<ansi fg="yellow-bold">This area is restricted to elected officials only.</ansi>`)
			}
		}
	}

	return d
}

// isAllowedInOfficialRoom returns true when the user may enter a room tagged
// officialOnlyTag. Allowed: admins, the zone's elected official, and members
// of the elected official's active party.
func (m *ElectionsModule) isAllowedInOfficialRoom(user *users.UserRecord, zoneKey string) bool {
	if user.Role == users.RoleAdmin {
		return true
	}
	winner, hasWinner := m.state.Winners[zoneKey]
	if !hasWinner {
		return false
	}
	if winner.UserId == user.UserId {
		return true
	}
	if officialParty := parties.Get(winner.UserId); officialParty != nil {
		for _, memberId := range officialParty.UserIds {
			if memberId == user.UserId {
				return true
			}
		}
	}
	return false
}

// onMovementGate intercepts movement commands and blocks entry into rooms
// tagged officialOnlyTag when the player is not permitted.
func (m *ElectionsModule) onMovementGate(e events.Event) events.ListenerReturn {
	evt, ok := e.(events.Input)
	if !ok || evt.UserId == 0 {
		return events.Continue
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	sourceRoom := rooms.LoadRoom(user.Character.RoomId)
	if sourceRoom == nil {
		return events.Continue
	}

	// Already inside a restricted area — movement between tagged rooms is allowed.
	if sourceRoom.HasTag(officialOnlyTag) {
		return events.Continue
	}

	fields := strings.Fields(evt.InputText)
	if len(fields) == 0 {
		return events.Continue
	}
	cmd := strings.ToLower(fields[0])

	_, destRoomId := sourceRoom.FindExitByName(cmd)
	if destRoomId == 0 {
		return events.Continue
	}

	destRoom := rooms.LoadRoom(destRoomId)
	if destRoom == nil || !destRoom.HasTag(officialOnlyTag) {
		return events.Continue
	}

	zoneKey := strings.ToLower(destRoom.Zone)
	if m.isAllowedInOfficialRoom(user, zoneKey) {
		return events.Continue
	}

	winner, hasWinner := m.state.Winners[zoneKey]
	if hasWinner {
		user.SendText(fmt.Sprintf(`Only the <ansi fg="white-bold">%s</ansi> and their party may enter here.`, winner.Title))
	} else {
		user.SendText(`This area is restricted to elected officials only.`)
	}
	return events.Cancel
}

// onGetFormattedName appends the player's title (if any) to their formatted name.
func (m *ElectionsModule) onGetFormattedName(f characters.FormattedName) characters.FormattedName {
	if f.Type != `username` {
		return f
	}

	for _, w := range m.state.Winners {
		if strings.EqualFold(w.CharacterName, f.Name) {
			color := m.titleColor()
			if len(color) > 0 {
				if num, err := strconv.Atoi(color); err == nil {
					f.Title = fmt.Sprintf(`<ansi fg="%d">%s</ansi>`, num, w.Title)
				} else {
					f.Title = colorpatterns.ApplyColorPattern(w.Title, color)
				}
			}
			return f
		}
	}

	return f
}
