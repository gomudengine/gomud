package character

import (
	"embed"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
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
	characterTag = "character"
)

func init() {
	m := &CharacterModule{
		plug: plugins.New(`character`, `1.0`),
	}

	if err := m.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	m.plug.AddUserCommand(`character`, m.characterCommand, true, false)

	m.plug.ReserveTags(characterTag)

	rooms.OnRoomLook.Register(m.onRoomLook)
}

type CharacterModule struct {
	plug *plugins.Plugin
}

func (m *CharacterModule) onRoomLook(d rooms.RoomTemplateDetails) rooms.RoomTemplateDetails {
	for _, t := range d.Tags {
		if strings.EqualFold(t, characterTag) {
			d.RoomAlerts = append(d.RoomAlerts,
				`      <ansi fg="yellow-bold">This is a character room!</ansi> Type <ansi fg="command">character</ansi> to interact.`,
			)
			return d
		}
	}
	return d
}

func roomIsCharacter(room *rooms.Room) bool {
	for _, t := range room.Tags {
		if strings.EqualFold(t, characterTag) {
			return true
		}
	}
	return false
}

func (m *CharacterModule) characterCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if !roomIsCharacter(room) {
		return false, fmt.Errorf(`not in a character room`)
	}

	altNames := []string{}
	nameToAlt := map[string]characters.Character{}

	for _, char := range characters.LoadAlts(user.UserId) {
		altNames = append(altNames, char.Name)
		nameToAlt[char.Name] = char
	}

	c := configs.GetGamePlayConfig()

	if c.MaxAltCharacters == 0 {
		user.SendText(`<ansi fg="203">Alt character are disabled on this server.</ansi>`)
		return true, errors.New(`alt characters disabled`)
	}

	if user.Character.Level < 5 && len(nameToAlt) < 1 {
		user.SendText(`<ansi fg="203">You must reach level 5 with this character to access character alts.</ansi>`)
		return true, errors.New(`level 5 minimum`)
	}

	hiredOutChars := map[string]characters.Character{}
	for _, mobInstanceId := range user.Character.GetCharmIds() {
		mob := mobs.GetInstance(mobInstanceId)
		if mob == nil {
			continue
		}
		hiredOutChars[mob.Character.Name] = mob.Character
	}

	menuOptions := []string{`new`}

	cmdPrompt, isNew := user.StartPrompt(`character`, rest)

	if isNew {

		if len(altNames) > 0 {
			menuOptions = append(menuOptions, `view`)
			menuOptions = append(menuOptions, `change`)
			menuOptions = append(menuOptions, `delete`)
			menuOptions = append(menuOptions, `hire`)
		}

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt, hiredOutChars, user.UserId)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

	}

	menuOptions = append(menuOptions, `quit`)

	question := cmdPrompt.Ask(`What would you like to do?`, menuOptions, `quit`)
	if !question.Done {
		return true, nil
	}

	/////////////////////////
	// Leave menu
	/////////////////////////
	if question.Response == `quit` {
		user.ClearPrompt()
		return true, nil
	}

	/////////////////////////
	// Create a new alt
	/////////////////////////
	if question.Response == `new` {

		if len(altNames) >= int(c.MaxAltCharacters) {
			user.SendText(`<ansi fg="203">You already have too many alts.</ansi>`)
			user.SendText(`<ansi fg="203">You'll need to delete one to create a new one.</ansi>`)

			question.RejectResponse()
			return true, nil
		}

		question := cmdPrompt.Ask(`Are you SURE? (Your current character will be saved here to change back to later)`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `no` {
			user.ClearPrompt()
			return true, nil
		}

		newAlts := []characters.Character{}
		for _, char := range nameToAlt {
			newAlts = append(newAlts, char)
		}
		newAlts = append(newAlts, *user.Character)
		characters.SaveAlts(user.UserId, newAlts)

		user.Character = characters.New()
		user.Character.Name = user.TempName()

		room.RemovePlayer(user.UserId)
		rooms.MoveToRoom(user.UserId, -1)

	}

	/////////////////////////
	// Delete an existing alt
	/////////////////////////
	if question.Response == `delete` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt, hiredOutChars, user.UserId)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to delete:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			delChar := nameToAlt[match]

			if friend, ok := hiredOutChars[delChar.Name]; ok && friend.Description == delChar.Description {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is currently hired out.`, delChar.Name))
				user.ClearPrompt()
				return true, nil
			}

			question := cmdPrompt.Ask(`<ansi fg="red">Are you SURE you want to delete <ansi fg="username">`+delChar.Name+`</ansi>?</ansi>`, []string{`yes`, `no`}, `no`)
			if !question.Done {
				return true, nil
			}

			if question.Response == `no` {
				user.SendText(`<ansi fg="203">Okay. Aborting.</ansi>`)
				user.ClearPrompt()
				return true, nil
			}

			newAlts := []characters.Character{}
			for _, char := range nameToAlt {
				if char.Name != match {
					newAlts = append(newAlts, char)
				}
			}
			characters.SaveAlts(user.UserId, newAlts)

			user.EventLog.Add(`char`, `Deleted alt character: <ansi fg="username">`+match+`</ansi>`)

			user.SendText(`<ansi fg="username">` + match + `</ansi> <ansi fg="red">is deleted.</ansi>`)
			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// Swap characters
	/////////////////////////
	if question.Response == `change` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt, hiredOutChars, user.UserId)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to change to:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			char := nameToAlt[match]

			if friend, ok := hiredOutChars[char.Name]; ok && friend.Description == char.Description {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is currently hired out.`, char.Name))
				user.ClearPrompt()
				return true, nil
			}

			question := cmdPrompt.Ask(`<ansi fg="51">Are you SURE you want to change to <ansi fg="username">`+char.Name+`</ansi>?</ansi>`, []string{`yes`, `no`}, `no`)
			if !question.Done {
				return true, nil
			}

			if question.Response == `no` {
				user.SendText(`<ansi fg="203">Okay. Aborting.</ansi>`)
				user.ClearPrompt()
				return true, nil
			}

			oldName := user.Character.Name

			succes := user.SwapToAlt(match)
			if !succes {
				user.SendText(`<ansi fg="203">Something went wrong.</ansi>`)
				user.ClearPrompt()
				return true, nil
			}

			newRoom := rooms.LoadRoom(user.Character.RoomId)
			if newRoom == nil {
				user.Character.RoomId = 0
				newRoom = rooms.LoadRoom(user.Character.RoomId)
			}

			room.RemovePlayer(user.UserId)
			newRoom.AddPlayer(user.UserId)

			users.SaveUser(*user)

			user.EventLog.Add(`char`, `Changed from <ansi fg="username">`+oldName+`</ansi> to alt character: <ansi fg="username">`+char.Name+`</ansi>`)

			user.SendText(term.CRLFStr + `You dematerialize as <ansi fg="username">` + oldName + `</ansi>. and rematerialize as <ansi fg="username">` + char.Name + `</ansi>!` + term.CRLFStr)
			room.SendText(`<ansi fg="username">`+oldName+`</ansi> vanishes, and <ansi fg="username">`+char.Name+`</ansi> appears in a shower of sparks!`, user.UserId)

			user.ClearPrompt()

			events.AddToQueue(events.PlayerChanged{UserId: user.UserId})

			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// View characters
	/////////////////////////
	if question.Response == `view` {

		if len(nameToAlt) > 0 {
			altTblTxt := getAltTable(nameToAlt, hiredOutChars, user.UserId)
			user.SendText(``)
			user.SendText(altTblTxt)
		}

		question := cmdPrompt.Ask(`Enter the name of the character you wish to view:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			char := nameToAlt[match]

			if friend, ok := hiredOutChars[char.Name]; ok && friend.Description == char.Description {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is currently hired out.`, char.Name))
				user.ClearPrompt()
				return true, nil
			}

			char.Validate()

			tmpChar := user.Character
			user.Character = &char

			usercommands.TryCommand(`status`, ``, user.UserId, flags)

			user.Character = tmpChar

			mob := mobs.NewMobById(59, user.Character.RoomId)
			mob.Character = char
			room.AddMob(mob.InstanceId)
			mob.Character.Charm(user.UserId, -1, `suicide vanish`)

			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	/////////////////////////
	// Spawn a helper clone - experimental
	/////////////////////////
	if question.Response == `hire` {

		question := cmdPrompt.Ask(`Enter the name of the character you wish to hire:`, []string{})
		if !question.Done {
			return true, nil
		}

		match, closeMatch := util.FindMatchIn(question.Response, altNames...)
		if match == `` {
			match = closeMatch
		}

		if match != `` {

			char := nameToAlt[match]

			if friend, ok := hiredOutChars[char.Name]; ok && friend.Description == char.Description {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is already hired out.`, char.Name))
				user.ClearPrompt()
				return true, nil
			}

			char.Validate()

			gearValue := char.GetGearValue()

			charValue := gearValue + (250 * char.Level)

			mudlog.Debug(`Hire Alt`, `UserId`, user.UserId, `alt-name`, char.Name, `gear-value`, gearValue, `level`, char.Level, `total`, charValue)

			question := cmdPrompt.Ask(fmt.Sprintf(`<ansi fg="51">The price to hire <ansi fg="username">%s</ansi> is <ansi fg="gold">%d gold</ansi>. Are you sure?</ansi>`, char.Name, charValue), []string{`yes`, `no`}, `no`)
			if !question.Done {
				return true, nil
			}

			if question.Response != `yes` {
				user.ClearPrompt()
				return true, nil
			}

			if user.Character.Gold < charValue {
				user.SendText(fmt.Sprintf(`You only have <ansi fg="gold">%d gold</ansi> and it would cost <ansi fg="gold">%d gold</ansi> to hire <ansi fg="username">%s</ansi>.`, charValue, charValue, char.Name))
				user.ClearPrompt()
				return true, nil
			}

			maxCharmed := user.Character.GetSkillLevel(skills.Tame) + 1
			if len(hiredOutChars) >= maxCharmed {
				user.SendText(fmt.Sprintf(`You can only have %d mobs following you at a time.`, maxCharmed))
				user.ClearPrompt()
				return true, nil
			}

			user.Character.Gold -= charValue

			mob := mobs.NewMobById(59, user.Character.RoomId)
			mob.Character = char

			mob.Character.Items = []items.Item{}
			mob.Character.Gold = 0
			mob.Character.Bank = 0
			mob.Character.Shop = characters.Shop{}

			mob.Character.AddBuff(36, true)

			room.AddMob(mob.InstanceId)

			mob.Character.Charm(user.UserId, -1, `suicide vanish`)
			user.Character.TrackCharmed(mob.InstanceId, true)

			user.EventLog.Add(`char`, `Hired an alt character to help you out: <ansi fg="username">`+mob.Character.Name+`</ansi>`)

			user.SendText(`<ansi fg="username">` + mob.Character.Name + `</ansi> appears to help you out!`)
			room.SendText(`<ansi fg="username">`+mob.Character.Name+`</ansi> appears to help <ansi fg="username">`+user.Character.Name+`</ansi>!`, user.UserId)

			mob.Command(`emote waves sheepishly.`, 2)

			user.ClearPrompt()
			return true, nil

		}

		user.SendText(`<ansi fg="203">No character with the name <ansi fg="username">` + question.Response + `</ansi> found.</ansi>`)

		user.ClearPrompt()
		return true, nil

	}

	return true, nil
}

func getAltTable(nameToAlt map[string]characters.Character, charmedChars map[string]characters.Character, viewingUserId int) string {

	headers := []string{"Name", "Level", "Race", "Profession", "Alignment", "Status"}
	rows := [][]string{}

	for _, char := range nameToAlt {

		allRanks := char.GetAllSkillRanks()
		raceName := `Unknown`
		if raceInfo := races.GetRace(char.RaceId); raceInfo != nil {
			raceName = raceInfo.Name
		}

		mobBusy := ``
		if c, ok := charmedChars[char.Name]; ok {
			if c.Description == char.Description {
				mobBusy = `<ansi fg="210">busy</ansi>`
			}
		}

		rows = append(rows, []string{
			fmt.Sprintf(`<ansi fg="username">%s</ansi>`, char.Name),
			strconv.Itoa(char.Level),
			raceName,
			skills.GetProfession(allRanks),
			fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, char.AlignmentName(), char.AlignmentName()),
			mobBusy,
		})

	}

	sort.Slice(rows, func(i, j int) bool {
		num1, _ := strconv.Atoi(rows[i][1])
		num2, _ := strconv.Atoi(rows[j][1])
		return num1 < num2
	})

	altTableData := templates.GetTable(fmt.Sprintf(`Your alt characters (%d/%d)`, len(nameToAlt), configs.GetGamePlayConfig().MaxAltCharacters), headers, rows)
	tplTxt, _ := templates.Process("tables/generic", altTableData, viewingUserId)

	return tplTxt
}
