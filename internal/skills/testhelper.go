package skills

// SetTestData replaces the in-memory skill and profession caches. It is intended
// for use by tests in this and other packages that need a deterministic, on-disk-free
// skill set seeded before exercising skill-dependent behavior.
func SetTestData(skillList []*Skill, professionList []*Profession) {
	allSkills = map[string]*Skill{}
	for _, s := range skillList {
		cp := *s
		allSkills[cp.SkillId] = &cp
	}

	allProfessions = map[string]*Profession{}
	for _, p := range professionList {
		cp := *p
		allProfessions[cp.ProfessionId] = &cp
	}
}
