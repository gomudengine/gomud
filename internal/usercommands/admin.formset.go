package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func FormSet(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		infoOutput, _ := templates.Process("admincommands/help/command.formset", nil, user.UserId, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	targetName := args[0]

	if strings.ToLower(args[len(args)-1]) == "revert" && len(args) == 2 {

		targetUserId, targetMobInstanceId := room.FindByName(targetName)

		if targetUserId > 0 {
			if targetUser := users.GetByUserId(targetUserId); targetUser != nil {
				if !targetUser.Character.IsFormChanged() {
					user.SendText(fmt.Sprintf(`%s is not form-changed.`, targetUser.Character.Name))
					return true, nil
				}
				targetUser.Character.RevertFormChange()
				targetUser.Character.RemoveBuff(41)
				targetUser.Character.RemoveBuff(42)
				user.SendText(fmt.Sprintf(`Reverted %s to their true form.`, targetUser.Character.Name))
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> reverts to their true form!`, targetUser.Character.Name),
					user.UserId,
				)
				return true, nil
			}
		}

		if targetMobInstanceId > 0 {
			if targetMob := mobs.GetInstance(targetMobInstanceId); targetMob != nil {
				if !targetMob.Character.IsFormChanged() {
					user.SendText(fmt.Sprintf(`%s is not form-changed.`, targetMob.Character.Name))
					return true, nil
				}
				targetMob.Character.RevertFormChange()
				targetMob.Character.RemoveBuff(41)
				targetMob.Character.RemoveBuff(42)
				user.SendText(fmt.Sprintf(`Reverted %s to their true form.`, targetMob.Character.Name))
				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> reverts to their true form!`, targetMob.Character.Name),
					user.UserId,
				)
				return true, nil
			}
		}

		user.SendText(fmt.Sprintf(`Could not find "%s" in this room.`, targetName))
		return true, nil
	}

	durationKey := "short"

	lastArg := strings.ToLower(args[len(args)-1])
	if lastArg == "short" || lastArg == "medium" || lastArg == "long" {
		durationKey = lastArg
		args = args[:len(args)-1]
	}

	if len(args) < 2 {
		user.SendText("You must specify a race.")
		return true, nil
	}

	raceName := strings.Join(args[1:], " ")
	raceInfo, found := races.FindRace(raceName)
	if !found {
		user.SendText(fmt.Sprintf(`"%s" is not a valid race.`, raceName))
		return true, nil
	}

	minutes := 5
	switch durationKey {
	case "medium":
		minutes = 15
	case "long":
		minutes = 30
	}
	duration := configs.GetTimingConfig().MinutesToRounds(minutes)

	targetUserId, targetMobInstanceId := room.FindByName(targetName)

	if targetUserId > 0 {
		targetUser := users.GetByUserId(targetUserId)
		if targetUser == nil {
			user.SendText("Could not find that user.")
			return true, nil
		}

		targetUser.Character.ApplyFormChange(raceInfo.RaceId)
		targetUser.Character.AddBuff(41, false, duration)

		user.SendText(fmt.Sprintf(`Changed %s into a <ansi fg="race">%s</ansi> for %d rounds (~%d min).`, targetUser.Character.Name, raceInfo.Name, duration, minutes))
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> transforms into a <ansi fg="race">%s</ansi>!`, targetUser.Character.Name, raceInfo.Name),
			user.UserId,
		)
		return true, nil
	}

	if targetMobInstanceId > 0 {
		targetMob := mobs.GetInstance(targetMobInstanceId)
		if targetMob == nil {
			user.SendText("Could not find that mob.")
			return true, nil
		}

		targetMob.Character.ApplyFormChange(raceInfo.RaceId)
		targetMob.Character.AddBuff(41, false, duration)

		user.SendText(fmt.Sprintf(`Changed %s into a <ansi fg="race">%s</ansi> for %d rounds (~%d min).`, targetMob.Character.Name, raceInfo.Name, duration, minutes))
		room.SendText(
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> transforms into a <ansi fg="race">%s</ansi>!`, targetMob.Character.Name, raceInfo.Name),
			user.UserId,
		)
		return true, nil
	}

	user.SendText(fmt.Sprintf(`Could not find "%s" in this room.`, targetName))
	return true, nil
}
