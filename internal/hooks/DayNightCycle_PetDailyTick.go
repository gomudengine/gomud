package hooks

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func PetDailyTick(e events.Event) events.ListenerReturn {

	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}

		for _, uId := range room.GetPlayers() {
			user := users.GetByUserId(uId)
			if user == nil || !user.Character.Pet.Exists() {
				continue
			}

			oldAbility := user.Character.Pet.GetCurrentAbilityDisplay()
			levelChange, ticked := user.Character.Pet.CheckDailyTick()
			if !ticked {
				continue
			}

			user.Character.Validate(true)
			newAbility := user.Character.Pet.GetCurrentAbilityDisplay()
			oldLevel := user.Character.Pet.Level - levelChange
			events.AddToQueue(events.PetLevelChange{
				UserId:     uId,
				PetName:    user.Character.Pet.DisplayName(),
				OldLevel:   oldLevel,
				NewLevel:   user.Character.Pet.Level,
				OldAbility: oldAbility,
				NewAbility: newAbility,
			})

			if user.Character.Pet.Food <= 1 {
				user.SendText(fmt.Sprintf(`%s is %s!`, user.Character.Pet.DisplayName(), user.Character.Pet.Food.String()))
			}
		}
	}

	return events.Continue
}
