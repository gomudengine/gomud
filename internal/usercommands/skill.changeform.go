package usercommands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
ChangeForm Skill
Level 1 - Transform into selectable races. Duration 10 rounds.
Level 2 - Transform into selectable races. Duration 20 rounds.
Level 3 - Transform into selectable races. Duration 40 rounds.
Level 4 - Transform into all races. Duration 80 rounds.
*/
func ChangeForm(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.ChangeForm)
	if skillLevel == 0 {
		user.SendText("You don't know how to change form.")
		return true, errors.New(`you don't know how to change form`)
	}

	if rest == `` {
		user.SendText(`Type <ansi fg="command">help changeform</ansi> for more information on the changeform skill.`)
		return true, nil
	}

	if user.Character.IsFormChanged() {
		if strings.ToLower(rest) == `revert` {
			user.Character.RevertFormChange()

			user.Character.RemoveBuff(41)
			user.Character.RemoveBuff(42)

			user.SendText("Your body shifts back to its original form.")
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> reverts to their true form!`, user.Character.Name),
				user.UserId,
			)
			return true, nil
		}
		user.SendText(`You are already transformed. Type <ansi fg="command">changeform revert</ansi> to change back first.`)
		return true, nil
	}

	if !user.Character.TryCooldown(skills.ChangeForm.String(), "20 rounds") {
		user.SendText(
			fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.ChangeForm.String())),
		)
		return true, errors.New(`you're doing that too often`)
	}

	targetRaceName := rest
	raceInfo, found := races.FindRace(targetRaceName)
	if !found {
		user.SendText(fmt.Sprintf(`"%s" is not a valid race.`, targetRaceName))
		return true, errors.New(`invalid race`)
	}

	if raceInfo.RaceId == user.Character.GetRaceId() {
		user.SendText("You are already that race!")
		return true, nil
	}

	if skillLevel < 4 && !raceInfo.Selectable {
		user.SendText("You don't have enough mastery to take that form.")
		return true, nil
	}

	duration := 0
	switch skillLevel {
	case 1:
		duration = 10
	case 2:
		duration = 20
	case 3:
		duration = 40
	case 4:
		duration = 80
	}

	user.Character.ApplyFormChange(raceInfo.RaceId)
	user.Character.AddBuff(41, false, duration)

	events.AddToQueue(events.SkillUsed{UserId: user.UserId, Skill: skills.ChangeForm})

	user.SendText(
		fmt.Sprintf(`Your body twists and reshapes — you are now a <ansi fg="race">%s</ansi>!`, raceInfo.Name),
	)
	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> transforms before your eyes into a <ansi fg="race">%s</ansi>!`, user.Character.Name, raceInfo.Name),
		user.UserId,
	)

	return true, nil
}
