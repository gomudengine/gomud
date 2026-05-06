package rooms

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// GetAllBiomesSorted returns a copy of every loaded biome, sorted by BiomeId.
func GetAllBiomesSorted() []BiomeInfo {
	result := make([]BiomeInfo, 0, len(biomes))
	for _, b := range biomes {
		result = append(result, *b)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].BiomeId < result[j].BiomeId
	})
	return result
}

// SaveBiome validates, writes a biome YAML to disk, and updates the in-memory
// registry.
func SaveBiome(b *BiomeInfo) error {
	b.BiomeId = strings.ToLower(strings.TrimSpace(b.BiomeId))
	if err := b.Validate(); err != nil {
		return err
	}

	basePath := configs.GetFilePathsConfig().DataFiles.String() + `/biomes`

	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*BiomeInfo](basePath, b, saveModes...); err != nil {
		return fmt.Errorf("saving biome: %w", err)
	}

	cp := *b
	biomes[b.BiomeId] = &cp
	return nil
}

// CreateBiome registers a new biome. Returns an error if the id already exists.
func CreateBiome(b *BiomeInfo) error {
	b.BiomeId = strings.ToLower(strings.TrimSpace(b.BiomeId))
	if b.BiomeId == "" {
		return fmt.Errorf("biomeid cannot be empty")
	}
	if _, exists := biomes[b.BiomeId]; exists {
		return fmt.Errorf("biome %q already exists", b.BiomeId)
	}
	return SaveBiome(b)
}

// DeleteBiome removes a biome from disk and the in-memory registry.
// The "default" biome cannot be deleted.
func DeleteBiome(biomeId string) error {
	biomeId = strings.ToLower(strings.TrimSpace(biomeId))
	if biomeId == "default" {
		return fmt.Errorf("the default biome cannot be deleted")
	}
	b, ok := biomes[biomeId]
	if !ok {
		return fmt.Errorf("biome %q not found", biomeId)
	}
	path := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/biomes/`, b.Filepath())
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing biome file: %w", err)
	}
	delete(biomes, biomeId)
	return nil
}
