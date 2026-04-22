package keywords

import (
	"fmt"
	"os"
	"strings"

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

	path := configs.GetFilePathsConfig().DataFiles.String() + `/keywords.yaml`

	sectionComments := extractSectionComments(path)

	bytes, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("marshaling keywords: %w", err)
	}

	bytes = insertSectionComments(bytes, sectionComments)

	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return fmt.Errorf("writing keywords file: %w", err)
	}

	loadedKeywords = a
	if err := loadedKeywords.Validate(); err != nil {
		return fmt.Errorf("validating keywords after save: %w", err)
	}

	return nil
}

// extractSectionComments reads the file at path and returns comment blocks
// (including preceding blank lines) keyed by the top-level YAML key they precede.
func extractSectionComments(path string) map[string][]string {
	comments := make(map[string][]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return comments
	}

	lines := strings.Split(string(data), "\n")
	var pending []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			pending = append(pending, line)
			continue
		}

		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && strings.Contains(line, ":") {
			key := strings.TrimSpace(line[:strings.Index(line, ":")])
			if len(pending) > 0 {
				comments[key] = append([]string{}, pending...)
				pending = nil
			}
		} else {
			pending = nil
		}
	}

	return comments
}

// insertSectionComments splices saved comment blocks back into marshaled YAML
// output, placing each block immediately before its associated top-level key.
func insertSectionComments(data []byte, comments map[string][]string) []byte {
	if len(comments) == 0 {
		return data
	}

	lines := strings.Split(string(data), "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && trimmed != "" && strings.Contains(line, ":") {
			key := strings.TrimSpace(line[:strings.Index(line, ":")])
			if commentLines, ok := comments[key]; ok {
				result = append(result, commentLines...)
			}
		}
		result = append(result, line)
	}

	return []byte(strings.Join(result, "\n"))
}
