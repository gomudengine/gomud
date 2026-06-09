package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      Skill
		wantErr bool
		wantId  string
		wantMax int
	}{
		{
			name:    "normalizes id and defaults maxlevel",
			in:      Skill{SkillId: "  Brawling ", Name: "Brawling", Description: "desc"},
			wantId:  "brawling",
			wantMax: DefaultMaxLevel,
		},
		{
			name:    "keeps explicit maxlevel",
			in:      Skill{SkillId: "cast", Name: "Cast", Description: "desc", MaxLevel: 6},
			wantId:  "cast",
			wantMax: 6,
		},
		{
			name:    "allows hyphenated id",
			in:      Skill{SkillId: "dual-wield", Name: "Dual Wield", Description: "desc"},
			wantId:  "dual-wield",
			wantMax: DefaultMaxLevel,
		},
		{
			name:    "rejects id with space",
			in:      Skill{SkillId: "dual wield", Name: "Dual Wield", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "rejects id starting with digit",
			in:      Skill{SkillId: "1skill", Name: "Skill", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "rejects empty name",
			in:      Skill{SkillId: "cast", Name: "", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "rejects empty description",
			in:      Skill{SkillId: "cast", Name: "Cast", Description: ""},
			wantErr: true,
		},
		{
			name:    "negative maxlevel defaults",
			in:      Skill{SkillId: "cast", Name: "Cast", Description: "desc", MaxLevel: -3},
			wantId:  "cast",
			wantMax: DefaultMaxLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.in
			err := s.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantId, s.SkillId)
			assert.Equal(t, tt.wantMax, s.MaxLevel)
		})
	}
}

func TestProfessionValidate(t *testing.T) {
	t.Run("lowercases and dedupes skills", func(t *testing.T) {
		p := Profession{ProfessionId: "Treasure Hunter", Name: "Treasure Hunter", Skills: []string{"Map", "map", " search "}}
		assert.NoError(t, p.Validate())
		assert.Equal(t, "treasure hunter", p.ProfessionId)
		assert.Equal(t, []string{"map", "search"}, p.Skills)
	})
	t.Run("rejects no skills", func(t *testing.T) {
		p := Profession{ProfessionId: "empty", Name: "Empty", Skills: nil}
		assert.Error(t, p.Validate())
	})
	t.Run("rejects no name", func(t *testing.T) {
		p := Profession{ProfessionId: "x", Name: "", Skills: []string{"map"}}
		assert.Error(t, p.Validate())
	})
}

func TestFilename(t *testing.T) {
	s := &Skill{SkillId: "dual-wield"}
	assert.Equal(t, "dual-wield.yaml", s.Filename())

	p := &Profession{ProfessionId: "treasure hunter"}
	assert.Equal(t, "treasure_hunter.yaml", p.Filename())
}

func seedStandard(t *testing.T) {
	t.Helper()
	SetTestData(
		[]*Skill{
			{SkillId: "brawling", Name: "Brawling", Description: "d", MaxLevel: 4},
			{SkillId: "dual-wield", Name: "Dual Wield", Description: "d", MaxLevel: 4},
			{SkillId: "cast", Name: "Cast", Description: "d", MaxLevel: 4},
			{SkillId: "enchant", Name: "Enchant", Description: "d", MaxLevel: 4},
			{SkillId: "deepskill", Name: "Deep Skill", Description: "d", MaxLevel: 6},
		},
		[]*Profession{
			{ProfessionId: "warrior", Name: "Warrior", Skills: []string{"brawling", "dual-wield"}},
			{ProfessionId: "sorcerer", Name: "Sorcerer", Skills: []string{"cast", "enchant"}},
			{ProfessionId: "deepclass", Name: "Deep Class", Skills: []string{"deepskill"}},
		},
	)
}

func TestSkillExistsAndMaxLevel(t *testing.T) {
	seedStandard(t)
	assert.True(t, SkillExists("brawling"))
	assert.True(t, SkillExists("BRAWLING"))
	assert.False(t, SkillExists("nonexistent"))

	assert.Equal(t, 4, MaxSkillLevel("brawling"))
	assert.Equal(t, 6, MaxSkillLevel("deepskill"))
	// Unknown skill falls back to default.
	assert.Equal(t, DefaultMaxLevel, MaxSkillLevel("nonexistent"))
}

func TestGetProfessionRanksMath(t *testing.T) {
	seedStandard(t)

	// Warrior: brawling 4 (max) + dual-wield 0.
	// PointsToMax = 10 + 10 = 20; spent = 10 + 0 = 10; completion 0.5.
	ranks := GetProfessionRanks(map[string]int{"brawling": 4})
	var warrior *ProfessionRank
	for i := range ranks {
		if ranks[i].Profession == "warrior" {
			warrior = &ranks[i]
		}
	}
	assert.NotNil(t, warrior)
	assert.Equal(t, 20.0, warrior.PointsToMax)
	assert.Equal(t, 10.0, warrior.TotalPointsSpent)
	assert.Equal(t, 0.5, warrior.Completion)
}

func TestGetProfessionRanksMaxlevel6(t *testing.T) {
	seedStandard(t)

	// deepskill maxlevel 6: PointsToMax = 6*7/2 = 21.
	// At level 6: spent = 6*7/2 = 21; completion 1.0.
	ranks := GetProfessionRanks(map[string]int{"deepskill": 6})
	var deep *ProfessionRank
	for i := range ranks {
		if ranks[i].Profession == "deepclass" {
			deep = &ranks[i]
		}
	}
	assert.NotNil(t, deep)
	assert.Equal(t, 21.0, deep.PointsToMax)
	assert.Equal(t, 21.0, deep.TotalPointsSpent)
	assert.Equal(t, 1.0, deep.Completion)

	// Saved above the cap is clamped on read.
	ranksOver := GetProfessionRanks(map[string]int{"deepskill": 99})
	for i := range ranksOver {
		if ranksOver[i].Profession == "deepclass" {
			assert.Equal(t, 21.0, ranksOver[i].TotalPointsSpent)
		}
	}
}

func TestGetProfessionTitles(t *testing.T) {
	seedStandard(t)

	// No skills trained -> scrub.
	assert.Equal(t, "scrub", GetProfession(map[string]int{}))

	// A bit of brawling -> novice warrior (completion 1/20 = 0.05 is below novice 0.1;
	// use level 2 brawling: spent 3/20 = 0.15 -> novice).
	assert.Equal(t, "novice warrior", GetProfession(map[string]int{"brawling": 2}))

	// Max everything -> demigod (all professions tied at completion 1.0).
	all := map[string]int{"brawling": 4, "dual-wield": 4, "cast": 4, "enchant": 4, "deepskill": 6}
	assert.Equal(t, "expert demigod", GetProfession(all))
}

func TestGetExperienceLevel(t *testing.T) {
	assert.Equal(t, "expert", GetExperienceLevel(0.95))
	assert.Equal(t, "journeyman", GetExperienceLevel(0.7))
	assert.Equal(t, "apprentice", GetExperienceLevel(0.4))
	assert.Equal(t, "novice", GetExperienceLevel(0.15))
	assert.Equal(t, "scrub", GetExperienceLevel(0.05))
}
