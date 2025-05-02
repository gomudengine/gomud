package rooms

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
	"gopkg.in/yaml.v2"
)

func LoadRoomTemplate(roomId int) *Room {

	filename := ``

	if cachedPath, ok := roomManager.roomIdToFileCache[roomId]; ok {
		filename = cachedPath
	} else {
		filename = findRoomFile(roomId)
	}

	if len(filename) == 0 {
		return nil
	}

	filepath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `rooms`, `/`, filename)

	retRoom, _ := loadRoomFromFile(filepath)

	return retRoom
}

// Load room grabs the room from memory and returns a pointer to it.
// If the room hasn't been loaded yet, it:
// 1. loads it from the template file
// 2. loads any instance data and overlays it
// 3. adds it to memory
func LoadRoom(roomId int) *Room {

	// Room 0 aliases to start room
	if roomId == StartRoomIdAlias {
		if roomId = int(configs.GetSpecialRoomsConfig().StartRoom); roomId == 0 {
			roomId = 1
		}
	}

	room, ok := roomManager.rooms[roomId]

	if ok {
		return room
	}

	roomTpl := LoadRoomTemplate(roomId)

	if roomTpl != nil {

		filename := ``
		if cachedPath, ok := roomManager.roomIdToFileCache[roomId]; ok {
			filename = cachedPath
		} else {
			filename = findRoomFile(roomId)
		}

		filepath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `rooms.instances`, `/`, filename)
		if _, err := os.Stat(filepath); err == nil {
			if bytes, err := os.ReadFile(filepath); err == nil {
				yaml.Unmarshal(bytes, roomTpl)
			}
		}

		addRoomToMemory(roomTpl)
	}

	return roomTpl
}

// Saves whatever room file is passed to it as the template.
// Use with caution, accidentally passing an object that has items etc.
// will mean this room's blank state will always have these items.
func SaveRoomTemplate(r Room) error {

	data, err := yaml.Marshal(&r)
	if err != nil {
		return err
	}

	zone := ZoneToFolder(r.Zone)

	roomFilePath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `rooms`, `/`, fmt.Sprintf("%s%d.yaml", zone, r.RoomId))

	if err = os.WriteFile(roomFilePath, data, 0777); err != nil {
		return err
	}

	copyFromRoom := roomManager.rooms[r.RoomId]
	room := LoadRoomTemplate(r.RoomId)

	// Copy container contents (if new vs. old room container names match)
	for containerName, container := range copyFromRoom.Containers {

		if newContainer, ok := room.Containers[containerName]; ok {

			if newContainer.Gold == 0 {
				newContainer.Gold = container.Gold
			}

			if len(newContainer.Items) == 0 && len(container.Items) > 0 {
				newContainer.Items = make([]items.Item, len(container.Items))
				copy(newContainer.Items, container.Items)
			}

			room.Containers[containerName] = newContainer
		}

	}

	// Copy items and stashed items
	for _, itm := range copyFromRoom.GetAllFloorItems(true) {
		if itm.StashedBy > 0 {
			room.AddItem(itm, true)
		} else {
			room.AddItem(itm, false)
		}
	}

	// Copy gol don floor
	room.Gold = copyFromRoom.Gold

	// Copy signs
	room.Signs = make([]Sign, len(copyFromRoom.Signs))
	copy(room.Signs, copyFromRoom.Signs)

	// Copy mobs in room
	room.mobs = make([]int, len(copyFromRoom.mobs))
	copy(room.mobs, copyFromRoom.mobs)

	// Copy players in room
	room.players = make([]int, len(copyFromRoom.players))
	copy(room.players, copyFromRoom.players)

	roomManager.rooms[room.RoomId] = room

	SaveRoom(*room)

	return nil
}

// SaveRoom loads the original template for the room
// It then compares for changes and saves all elligible data
func SaveRoom(r Room) error {

	rTpl := LoadRoomTemplate(r.RoomId)

	rVal := reflect.ValueOf(r)
	tplVal := reflect.ValueOf(*rTpl)
	t := reflect.TypeOf(r)

	instanceSaveData := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == `-` {
			continue
		}

		if field.Tag.Get("instance") == "skip" {
			continue
		}

		rVal2 := rVal.Field(i)
		tplVal2 := tplVal.Field(i)

		// Compare using DeepEqual
		if !reflect.DeepEqual(rVal2.Interface(), tplVal2.Interface()) {

			marshal1, _ := yaml.Marshal(rVal2.Interface())
			marshal2, _ := yaml.Marshal(tplVal2.Interface())

			if string(marshal1) == string(marshal2) {
				continue
			}

			tagParts := strings.Split(yamlTag, ",")
			fieldName := tagParts[0]
			if fieldName == `` || fieldName == `omitempty` || fieldName == `flow` {
				fieldName = field.Name
			}

			instanceSaveData[fieldName] = rVal2.Interface()

		}
	}

	data, err := yaml.Marshal(instanceSaveData)
	if err != nil {
		return err
	}

	zone := ZoneToFolder(r.Zone)

	folderPath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `rooms.instances`, `/`, zone)
	instanceFilePath := fmt.Sprintf("%s%d.yaml", folderPath, r.RoomId)

	if string(data[0:2]) == `{}` {
		os.Remove(instanceFilePath)
		return nil
	}

	if err = os.WriteFile(instanceFilePath, data, 0777); err != nil {
		return err
	}

	return nil
}

func findRoomFile(roomId int) string {

	foundFilePath := ``
	searchFileName := filepath.FromSlash(fmt.Sprintf(`/%d.yaml`, roomId))

	walkPath := filepath.FromSlash(configs.GetFilePathsConfig().DataFiles.String() + `/rooms`)

	filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if strings.HasSuffix(path, searchFileName) {
			foundFilePath = path
			return errors.New(`found`)
		}

		return nil
	})

	return strings.TrimPrefix(foundFilePath, walkPath)
}

func loadRoomFromFile(roomFilePath string) (*Room, error) {

	roomFilePath = util.FilePath(roomFilePath)

	roomPtr, err := fileloader.LoadFlatFile[*Room](roomFilePath)
	if err != nil {
		mudlog.Error("loadRoomFromFile()", "error", err.Error())
		return roomPtr, err
	}

	// Automatically set the last visitor to now (reset the timer)
	roomPtr.lastVisited = util.GetRoundCount()

	return roomPtr, err
}

func SaveAllRooms() error {

	start := time.Now()
	saveCt := 0
	errCt := 0
	for _, r := range roomManager.rooms {

		if SaveRoom(*r) != nil {
			errCt++
			continue
		}
		saveCt++

	}

	mudlog.Info("SaveAllRooms()", "savedCount", saveCt, "expectedCt", len(roomManager.rooms), "errorCount", errCt, "Time Taken", time.Since(start))

	return nil
}

// Goes through all of the rooms and caches key information
func loadAllRoomZones() error {
	start := time.Now()

	nextRoomId := GetNextRoomId()
	defer func() {
		if nextRoomId != GetNextRoomId() {
			SetNextRoomId(nextRoomId)
		}
	}()

	loadedRooms, err := fileloader.LoadAllFlatFiles[int, *Room](configs.GetFilePathsConfig().DataFiles.String() + `/rooms`)
	if err != nil {
		return err
	}

	roomsWithoutEntrances := map[int]string{}

	for _, loadedRoom := range loadedRooms {

		// configs.GetConfig().DeathRecoveryRoom is the death/shadow realm and gets a pass
		if loadedRoom.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) {
			continue
		}

		// If it has never been set, set it to the filepath
		if _, ok := roomsWithoutEntrances[loadedRoom.RoomId]; !ok {
			roomsWithoutEntrances[loadedRoom.RoomId] = loadedRoom.Filepath()
		}

		for _, exit := range loadedRoom.Exits {
			roomsWithoutEntrances[exit.RoomId] = ``
		}

	}

	for roomId, filePath := range roomsWithoutEntrances {

		if filePath == `` {
			delete(roomsWithoutEntrances, roomId)
			continue
		}

		mudlog.Warn("No Entrance", "roomId", roomId, "filePath", filePath)
	}

	for _, loadedRoom := range loadedRooms {
		// Keep track of the highest roomId

		if loadedRoom.RoomId >= nextRoomId {
			nextRoomId = loadedRoom.RoomId + 1
		}

		// Cache the file path for every roomId
		roomManager.roomIdToFileCache[loadedRoom.RoomId] = loadedRoom.Filepath()

		// Update the zone info cache
		if _, ok := roomManager.zones[loadedRoom.Zone]; !ok {
			roomManager.zones[loadedRoom.Zone] = ZoneInfo{
				RootRoomId: 0,
				RoomIds:    make(map[int]struct{}),
			}

			folderPath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `rooms.instances`, `/`, ZoneNameSanitize(loadedRoom.Zone))
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				os.MkdirAll(folderPath, 0755)
			}
		}

		// Update the zone info
		zoneInfo := roomManager.zones[loadedRoom.Zone]
		zoneInfo.RoomIds[loadedRoom.RoomId] = struct{}{}

		if loadedRoom.ZoneConfig.RoomId == loadedRoom.RoomId {
			zoneInfo.RootRoomId = loadedRoom.RoomId
			zoneInfo.DefaultBiome = loadedRoom.Biome

			if len(loadedRoom.ZoneConfig.Mutators) > 0 {
				zoneInfo.HasZoneMutators = true
			}
		}

		roomManager.zones[loadedRoom.Zone] = zoneInfo
	}

	mudlog.Info("rooms.loadAllRoomZones()", "loadedCount", len(loadedRooms), "Time Taken", time.Since(start))

	return nil
}
