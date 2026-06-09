package skills

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// GetSkillsMap returns a copy of all loaded skills keyed by SkillId.
func GetSkillsMap() map[string]*Skill {
	result := make(map[string]*Skill, len(allSkills))
	for k, v := range allSkills {
		cp := *v
		result[k] = &cp
	}
	return result
}

// GetProfessionsMap returns a copy of all loaded professions keyed by ProfessionId.
func GetProfessionsMap() map[string]*Profession {
	result := make(map[string]*Profession, len(allProfessions))
	for k, v := range allProfessions {
		cp := *v
		result[k] = &cp
	}
	return result
}

func skillSaveModes() []fileloader.SaveOption {
	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}
	return saveModes
}

// CreateSkill validates and persists a new skill, rejecting duplicate ids.
func CreateSkill(s *Skill) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if _, ok := allSkills[s.SkillId]; ok {
		return fmt.Errorf("skill %q already exists", s.SkillId)
	}
	return saveSkill(s)
}

// SaveSkill validates, persists a skill to disk, and updates the in-memory cache.
func SaveSkill(s *Skill) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return saveSkill(s)
}

func saveSkill(s *Skill) error {
	if err := fileloader.SaveFlatFile[*Skill](configs.GetFilePathsConfig().DataFiles.String()+`/skills`, s, skillSaveModes()...); err != nil {
		return err
	}
	allSkills[s.SkillId] = s
	return nil
}

// DeleteSkill removes a skill from disk and the in-memory cache.
// It is rejected if any profession references the skill.
func DeleteSkill(skillId string) error {
	skillId = strings.ToLower(strings.TrimSpace(skillId))
	s, ok := allSkills[skillId]
	if !ok {
		return fmt.Errorf("skill %q not found", skillId)
	}

	users := []string{}
	for _, p := range allProfessions {
		for _, sk := range p.Skills {
			if sk == skillId {
				users = append(users, p.ProfessionId)
				break
			}
		}
	}
	if len(users) > 0 {
		return fmt.Errorf("skill %q is used by professions: %s", skillId, strings.Join(users, ", "))
	}

	path := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `skills`, `/`, s.Filename())
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing skill file: %w", err)
	}
	delete(allSkills, skillId)
	return nil
}

// CreateProfession validates and persists a new profession, rejecting duplicate ids
// and unknown skill references.
func CreateProfession(p *Profession) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, ok := allProfessions[p.ProfessionId]; ok {
		return fmt.Errorf("profession %q already exists", p.ProfessionId)
	}
	if err := validateProfessionSkillRefs(p); err != nil {
		return err
	}
	return saveProfession(p)
}

// SaveProfession validates, persists a profession to disk, and updates the in-memory cache.
// Unknown skill references are hard-rejected.
func SaveProfession(p *Profession) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := validateProfessionSkillRefs(p); err != nil {
		return err
	}
	return saveProfession(p)
}

func validateProfessionSkillRefs(p *Profession) error {
	for _, sk := range p.Skills {
		if !SkillExists(sk) {
			return fmt.Errorf("profession references unknown skill: %s", sk)
		}
	}
	return nil
}

func saveProfession(p *Profession) error {
	if err := fileloader.SaveFlatFile[*Profession](configs.GetFilePathsConfig().DataFiles.String()+`/professions`, p, skillSaveModes()...); err != nil {
		return err
	}
	allProfessions[p.ProfessionId] = p
	return nil
}

// DeleteProfession removes a profession from disk and the in-memory cache.
func DeleteProfession(id string) error {
	id = strings.ToLower(strings.TrimSpace(id))
	p, ok := allProfessions[id]
	if !ok {
		return fmt.Errorf("profession %q not found", id)
	}
	path := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `professions`, `/`, p.Filename())
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing profession file: %w", err)
	}
	delete(allProfessions, id)
	return nil
}
