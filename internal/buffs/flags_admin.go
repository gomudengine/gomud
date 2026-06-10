package buffs

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var validFlagIdPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// ValidateFlagId checks that a flag identifier is well-formed.
func ValidateFlagId(flag string) error {
	if flag == `` {
		return fmt.Errorf("flag identifier is required")
	}
	if !validFlagIdPattern.MatchString(flag) {
		return fmt.Errorf("flag must be lowercase alphanumeric with hyphens/underscores, starting with a letter: %q", flag)
	}
	return nil
}

func flagSaveModes() []fileloader.SaveOption {
	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}
	return saveModes
}

// CreateFlagSpec registers a new flag. Returns an error if the id already
// exists or is invalid. New flags default to unlocked.
func CreateFlagSpec(spec *FlagSpec) error {
	spec.Flag = strings.ToLower(strings.TrimSpace(spec.Flag))

	if err := ValidateFlagId(spec.Flag); err != nil {
		return err
	}

	if _, exists := flagSpecs[spec.Flag]; exists {
		return fmt.Errorf("flag already exists: %s", spec.Flag)
	}

	if err := spec.Validate(); err != nil {
		return err
	}

	if err := fileloader.SaveFlatFile[*FlagSpec](configs.GetFilePathsConfig().DataFiles.String()+`/buffs-flags`, spec, flagSaveModes()...); err != nil {
		return err
	}

	flagSpecs[spec.Flag] = spec
	return nil
}

// SaveFlagSpec persists changes to an existing flag. Locked flags cannot be
// edited and the flag identifier cannot be changed.
func SaveFlagSpec(spec *FlagSpec) error {
	spec.Flag = strings.ToLower(strings.TrimSpace(spec.Flag))

	if err := ValidateFlagId(spec.Flag); err != nil {
		return err
	}

	existing := GetFlagSpec(spec.Flag)
	if existing == nil {
		return fmt.Errorf("flag not found: %s", spec.Flag)
	}

	if existing.Locked {
		return fmt.Errorf("flag %q is locked and cannot be edited", spec.Flag)
	}

	if err := spec.Validate(); err != nil {
		return err
	}

	if err := fileloader.SaveFlatFile[*FlagSpec](configs.GetFilePathsConfig().DataFiles.String()+`/buffs-flags`, spec, flagSaveModes()...); err != nil {
		return err
	}

	flagSpecs[spec.Flag] = spec
	return nil
}

// DeleteFlagSpec removes a flag definition from disk and memory. Locked flags
// cannot be deleted.
func DeleteFlagSpec(flag string) error {
	flag = strings.ToLower(strings.TrimSpace(flag))

	spec := GetFlagSpec(flag)
	if spec == nil {
		return fmt.Errorf("flag not found: %s", flag)
	}

	if spec.Locked {
		return fmt.Errorf("flag %q is locked and cannot be deleted", flag)
	}

	basePath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String() + `/buffs-flags/`)

	yamlPath := basePath + spec.Filepath()
	if err := os.Remove(yamlPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing flag yaml: %w", err)
	}

	delete(flagSpecs, flag)
	return nil
}
