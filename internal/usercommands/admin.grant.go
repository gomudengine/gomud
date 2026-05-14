package usercommands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"

	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
* Role Permissions:
* grant 				(All)
 */
func Grant(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.grant", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	// args should look like one of the following:
	// [?target] 1000 experience - grant experience points to target, or self if unspecified target
	args := util.SplitButRespectQuotes(rest)

	targetUserId := 0
	targetMobInstanceId := 0

	lastWord := args[len(args)-1]

	if len(args) >= 2 && ((len(lastWord) >= 3 && lastWord[0:3] == `exp`) || lastWord == `xp`) {

		expAmt := 0

		if len(args) > 2 {

			targetUserId, targetMobInstanceId = room.FindByName(args[0])
			expAmt, _ = strconv.Atoi(args[1])

		} else {
			targetUserId = user.UserId
			expAmt, _ = strconv.Atoi(args[0])
		}

		if targetUserId > 0 {

			if u := users.GetByUserId(targetUserId); u != nil {
				u.GrantXP(expAmt, `admin grant`)
				user.SendText(fmt.Sprintf(`Granted <ansi fg="experience">%d experience</ansi> to <ansi fg="username">%s</ansi>.`, expAmt, u.Character.Name))
				return true, nil
			}

		} else if targetMobInstanceId > 0 {
			user.SendText(`Cannot grant experience to mobs.`)
			return true, nil
		}

		user.SendText(`Target not found.`)
		return true, errors.New(`target not found`)

	}

	if len(args) >= 1 && (lastWord == `petlevelup` || lastWord == `petleveldown`) {

		delta := 1
		if lastWord == `petleveldown` {
			delta = -1
		}

		if len(args) >= 2 {
			targetUserId, _ = room.FindByName(args[0])
		} else {
			targetUserId = user.UserId
		}

		if targetUserId > 0 {
			if u := users.GetByUserId(targetUserId); u != nil {

				if !u.Character.Pet.Exists() {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> does not have a pet.`, u.Character.Name))
					return true, nil
				}

				oldAbility := u.Character.Pet.GetCurrentAbilityDisplay()
				oldLevel, newLevel, changed := u.Character.Pet.LevelChange(delta)

				if !changed {
					user.SendText(fmt.Sprintf(`%s is already at %s level.`, u.Character.Pet.DisplayName(), map[bool]string{true: "maximum", false: "minimum"}[delta > 0]))
					return true, nil
				}

				u.Character.Validate(true)
				newAbility := u.Character.Pet.GetCurrentAbilityDisplay()

				events.AddToQueue(events.PetLevelChange{
					UserId:     u.UserId,
					PetName:    u.Character.Pet.DisplayName(),
					OldLevel:   oldLevel,
					NewLevel:   newLevel,
					OldAbility: oldAbility,
					NewAbility: newAbility,
				})

				user.SendText(fmt.Sprintf(`Granted %s to <ansi fg="username">%s</ansi>'s pet %s (level %d → %d).`, lastWord, u.Character.Name, u.Character.Pet.DisplayName(), oldLevel, newLevel))
				return true, nil
			}
		}

		user.SendText(`Target not found.`)
		return true, errors.New(`target not found`)
	}

	user.SendText(`Invalid command.`)

	return false, errors.New(`unrecognized command`)
}
