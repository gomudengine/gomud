package mobcommands

import (
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Pathto(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	toRoomIds := []int{}

	for _, roomIdStr := range util.SplitButRespectQuotes(strings.ToLower(rest)) {

		if roomIdStr == `home` {
			toRoomIds = append(toRoomIds, mob.HomeRoomId)
			continue
		}

		if roomIdInt, err := strconv.Atoi(roomIdStr); err == nil {
			toRoomIds = append(toRoomIds, roomIdInt)
		}
	}

	if len(toRoomIds) < 1 || toRoomIds[0] == mob.Character.RoomId {
		return true, nil
	}

	path, err := mapper.GetPath(mob.Character.RoomId, toRoomIds...)
	if err != nil {
		return false, err
	}

	newPath := []mobs.PathRoom{}

	// Copy everything over
	for _, p := range path {
		newPath = append(newPath, p)
	}

	mob.Path.SetPath(newPath)

	return true, nil
}
