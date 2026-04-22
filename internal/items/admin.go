package items

import (
	"fmt"
	"os"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
)

// GetAllAttackMessages returns a copy of the loaded attack message groups keyed
// by weapon subtype.
func GetAllAttackMessages() map[ItemSubType]*WeaponAttackMessageGroup {
	result := make(map[ItemSubType]*WeaponAttackMessageGroup, len(attackMessages))
	for k, v := range attackMessages {
		result[k] = v
	}
	return result
}

// SaveItemSpec persists an existing ItemSpec back to its data file and updates
// the in-memory cache.  The ItemId must already be set.
func SaveItemSpec(spec *ItemSpec) error {
	if spec.ItemId < 1 {
		return fmt.Errorf("cannot save item spec with invalid ItemId %d", spec.ItemId)
	}

	if err := spec.Validate(); err != nil {
		return err
	}

	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*ItemSpec](configs.GetFilePathsConfig().DataFiles.String()+`/items`, spec, saveModes...); err != nil {
		return err
	}

	items[spec.ItemId] = spec
	return nil
}

// DeleteItemSpec removes an item spec from disk (YAML + JS script) and from the
// in-memory cache.
func DeleteItemSpec(itemId int) error {
	spec := GetItemSpec(itemId)
	if spec == nil {
		return fmt.Errorf("item %d not found", itemId)
	}

	basePath := configs.GetFilePathsConfig().DataFiles.String() + `/items/`

	yamlPath := basePath + spec.Filepath()
	if err := os.Remove(yamlPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing item yaml: %w", err)
	}

	scriptPath := spec.GetScriptPath()
	if err := os.Remove(scriptPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing item script: %w", err)
	}

	delete(items, itemId)
	return nil
}

// SaveItemScript writes (or overwrites) the JavaScript file for an item.  If
// content is empty the script file is deleted instead.
func SaveItemScript(itemId int, content string) error {
	spec := GetItemSpec(itemId)
	if spec == nil {
		return fmt.Errorf("item %d not found", itemId)
	}

	scriptPath := spec.GetScriptPath()

	if content == "" {
		if err := os.Remove(scriptPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing item script: %w", err)
		}
		return nil
	}

	return os.WriteFile(scriptPath, []byte(content), 0644)
}
