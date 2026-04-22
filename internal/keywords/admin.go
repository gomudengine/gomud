package keywords

import (
	"fmt"
	"os"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"gopkg.in/yaml.v2"
)

func GetKeywords() *Aliases {
	if loadedKeywords == nil {
		return &Aliases{}
	}
	cp := *loadedKeywords

	cp.Help = make(map[string]map[string][]string, len(loadedKeywords.Help))
	for k, v := range loadedKeywords.Help {
		inner := make(map[string][]string, len(v))
		for ik, iv := range v {
			sl := make([]string, len(iv))
			copy(sl, iv)
			inner[ik] = sl
		}
		cp.Help[k] = inner
	}

	cp.HelpAliases = make(map[string][]string, len(loadedKeywords.HelpAliases))
	for k, v := range loadedKeywords.HelpAliases {
		sl := make([]string, len(v))
		copy(sl, v)
		cp.HelpAliases[k] = sl
	}

	cp.CommandAliases = make(map[string][]string, len(loadedKeywords.CommandAliases))
	for k, v := range loadedKeywords.CommandAliases {
		sl := make([]string, len(v))
		copy(sl, v)
		cp.CommandAliases[k] = sl
	}

	cp.DirectionAliases = make(map[string]string, len(loadedKeywords.DirectionAliases))
	for k, v := range loadedKeywords.DirectionAliases {
		cp.DirectionAliases[k] = v
	}

	cp.MapLegendOverrides = make(map[string]map[string]string, len(loadedKeywords.MapLegendOverrides))
	for k, v := range loadedKeywords.MapLegendOverrides {
		inner := make(map[string]string, len(v))
		for ik, iv := range v {
			inner[ik] = iv
		}
		cp.MapLegendOverrides[k] = inner
	}

	return &cp
}

func SaveKeywords(a *Aliases) error {
	if a == nil {
		return fmt.Errorf("keywords data cannot be nil")
	}

	bytes, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("marshaling keywords: %w", err)
	}

	path := configs.GetFilePathsConfig().DataFiles.String() + `/keywords.yaml`
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return fmt.Errorf("writing keywords file: %w", err)
	}

	loadedKeywords = a
	if err := loadedKeywords.Validate(); err != nil {
		return fmt.Errorf("validating keywords after save: %w", err)
	}

	return nil
}
