package races

import (
	"fmt"
	"os"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// GetRacesMap returns a copy of all loaded races keyed by RaceId.
func GetRacesMap() map[int]*Race {
	result := make(map[int]*Race, len(races))
	for k, v := range races {
		cp := *v
		result[k] = &cp
	}
	return result
}

// SaveRace validates, persists a race to disk, and updates the in-memory cache.
func SaveRace(r *Race) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := r.Save(); err != nil {
		return err
	}
	races[r.RaceId] = r
	return nil
}

// DeleteRace removes a race from disk and the in-memory cache.
func DeleteRace(raceId int) error {
	r := GetRace(raceId)
	if r == nil {
		return fmt.Errorf("race %d not found", raceId)
	}
	path := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `races`, `/`, r.Filename())
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing race file: %w", err)
	}
	delete(races, raceId)
	return nil
}

// NewRace assigns the next available ID to r, saves it, and registers it in memory.
func NewRace(r *Race) error {
	nextId := 0
	for id := range races {
		if id >= nextId {
			nextId = id + 1
		}
	}
	r.RaceId = nextId
	return SaveRace(r)
}
