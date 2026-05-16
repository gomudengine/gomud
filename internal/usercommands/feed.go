package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Feed(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if !user.Character.Pet.Exists() {
		user.SendText(`You don't have a pet to feed.`)
		return true, nil
	}

	if user.Character.Pet.Food >= 3 {
		user.SendText(fmt.Sprintf(`%s is already full.`, user.Character.Pet.DisplayName()))
		return true, nil
	}

	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to feed your pet.`, rest))
		return true, nil
	}

	itemSpec := matchItem.GetSpec()

	if itemSpec.Subtype != items.Edible {
		user.SendText(
			fmt.Sprintf(`You can't feed <ansi fg="itemname">%s</ansi> to your pet.`, matchItem.DisplayName()),
		)
		return true, nil
	}

	user.Character.CancelBuffsWithFlag(buffs.Hidden)

	if usesLeft := user.Character.UseItem(matchItem); usesLeft < 1 {
		events.AddToQueue(events.ItemOwnership{
			UserId: user.UserId,
			Item:   matchItem,
			Gained: false,
		})
	}

	user.Character.Pet.Food.Add()

	events.AddToQueue(events.PetFed{
		UserId: user.UserId,
	})

	user.SendText(fmt.Sprintf(`You feed the <ansi fg="itemname">%s</ansi> to %s. They gobble it up!`, matchItem.DisplayName(), user.Character.Pet.DisplayName()))
	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> feeds some <ansi fg="itemname">%s</ansi> to %s.`, user.Character.Name, matchItem.DisplayName(), user.Character.Pet.DisplayName()), user.UserId)

	return true, nil
}
