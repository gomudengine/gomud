package skills

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Profession is a data-driven profession definition loaded from a YAML datafile.
// A profession groups a set of skills; a character's mastery of those skills
// determines their profession title.
type Profession struct {
	ProfessionId string   `yaml:"professionid"` // lowercase, spaces OK: "treasure hunter"
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description,omitempty"`
	Skills       []string `yaml:"skills"` // skill ids
}

var allProfessions = map[string]*Profession{}

func (p *Profession) Id() string {
	return p.ProfessionId
}

func (p *Profession) Validate() error {
	p.ProfessionId = strings.ToLower(strings.TrimSpace(p.ProfessionId))
	if p.ProfessionId == "" {
		return fmt.Errorf("profession has no id")
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("profession %q has no name", p.ProfessionId)
	}

	// lowercase + dedupe skill refs
	seen := map[string]struct{}{}
	skills := make([]string, 0, len(p.Skills))
	for _, sk := range p.Skills {
		sk = strings.ToLower(strings.TrimSpace(sk))
		if sk == "" {
			continue
		}
		if _, ok := seen[sk]; ok {
			continue
		}
		seen[sk] = struct{}{}
		skills = append(skills, sk)
	}
	if len(skills) < 1 {
		return fmt.Errorf("profession %q has no skills", p.ProfessionId)
	}
	p.Skills = skills

	// NOTE: cannot cross-check skill refs here — fileloader calls Validate per-file,
	// load order not guaranteed. Cross-ref warnings happen in LoadProfessionDataFiles.
	return nil
}

// Filename uses ConvertForFilename so spaced ids like "treasure hunter" become treasure_hunter.yaml.
func (p *Profession) Filename() string {
	return util.ConvertForFilename(p.ProfessionId) + `.yaml`
}

func (p *Profession) Filepath() string {
	return p.Filename()
}

// LoadProfessionDataFiles loads all profession definitions from disk into the in-memory cache.
// Must run AFTER LoadDataFiles() so unknown-skill cross-references can be warned about.
func LoadProfessionDataFiles() {

	start := time.Now()

	tmpProfessions, err := fileloader.LoadAllFlatFiles[string, *Profession](configs.GetFilePathsConfig().DataFiles.String() + `/professions`)
	if err != nil {
		panic(err)
	}

	allProfessions = tmpProfessions

	// Warn (but don't fail) on references to unknown skills — boot stays up, refs inert.
	for _, p := range allProfessions {
		for _, sk := range p.Skills {
			if !SkillExists(sk) {
				mudlog.Warn("skills.LoadProfessionDataFiles()", "profession", p.ProfessionId, "unknownSkill", sk)
			}
		}
	}

	mudlog.Info("skills.LoadProfessionDataFiles()", "loadedCount", len(allProfessions), "Time Taken", time.Since(start))
}

// GetProfessionSpec returns the profession definition for the given id, or nil if unknown.
func GetProfessionSpec(id string) *Profession {
	return allProfessions[strings.ToLower(strings.TrimSpace(id))]
}

// GetAllProfessions returns all loaded professions sorted by ProfessionId.
func GetAllProfessions() []Profession {
	ret := make([]Profession, 0, len(allProfessions))
	for _, p := range allProfessions {
		ret = append(ret, *p)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ProfessionId < ret[j].ProfessionId
	})
	return ret
}
