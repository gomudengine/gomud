package skills

import (
	"strings"
)

// Skills are identified by lowercase string ids (e.g. "cast", "dual-wield").
// The YAML datafiles in _datafiles/world/default/skills are the source of truth
// for which skills exist at runtime. Code that gates behavior on a specific
// skill references the id as a plain string literal.

type ProfessionRank struct {
	Profession       string
	ExperienceTitle  string
	TotalPointsSpent float64
	PointsToMax      float64
	Completion       float64
	Skills           []string
}

func GetProfessionRanks(allRanks map[string]int) []ProfessionRank {

	professionList := []ProfessionRank{}

	for _, profession := range allProfessions {

		if len(profession.Skills) == 0 {
			continue // divide-by-zero guard
		}

		ranking := ProfessionRank{Profession: profession.ProfessionId}

		for _, skillName := range profession.Skills {

			max := MaxSkillLevel(skillName)

			skillLevel := 0
			if rankVal, ok := allRanks[skillName]; ok {
				skillLevel = rankVal
			}
			if skillLevel > max {
				skillLevel = max
			}
			totalSkill := (skillLevel * (skillLevel + 1)) / 2

			// Max points spendable on a skill: 1+2+...+max == max*(max+1)/2 (== 10 for max 4).
			ranking.PointsToMax += float64(max*(max+1)) / 2
			ranking.TotalPointsSpent += float64(totalSkill)
			ranking.Skills = append(ranking.Skills, skillName)
		}

		ranking.Completion = ranking.TotalPointsSpent / ranking.PointsToMax
		ranking.ExperienceTitle = GetExperienceLevel(ranking.Completion)

		professionList = append(professionList, ranking)
	}

	return professionList
}

func GetProfession(allRanks map[string]int) string {

	rankData := GetProfessionRanks(allRanks)

	var highestCompletion float64 = 0
	//var highestSpend float64 = 0
	chosenProfessions := []string{}
	experienceName := ``

	for _, pRank := range rankData {

		if pRank.Completion == 0 {
			continue
		}

		if pRank.Completion > highestCompletion {
			highestCompletion = pRank.Completion
			//highestSpend = pRank.TotalPointsSpent
			chosenProfessions = []string{}
		}

		if pRank.Completion == highestCompletion {
			experienceName = pRank.ExperienceTitle
			chosenProfessions = append(chosenProfessions, pRank.Profession)
		}
	}

	if len(chosenProfessions) < 1 {
		return `scrub`
	}

	if len(experienceName) > 0 {
		experienceName = experienceName + ` `
	}

	if len(chosenProfessions) == len(allProfessions) {
		return experienceName + `demigod`
	}

	extra := ``
	if len(chosenProfessions) > 3 {
		chosenProfessions = chosenProfessions[0:3]
		extra = ` (and more)`
	}

	return experienceName + strings.Join(chosenProfessions, `/`) + extra
}

// Possible value is something like 1-10
func GetExperienceLevel(percentage float64) string {

	if percentage >= .9 { // avg level ~4
		return `expert`
	}

	if percentage >= .6 { // avg level 3
		return `journeyman`
	}

	if percentage >= .3 { // avg level 2
		return `apprentice`
	}

	if percentage >= .1 { // avg level 1
		return `novice`
	}

	return `scrub`
}
