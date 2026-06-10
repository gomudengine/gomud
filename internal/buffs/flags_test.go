package buffs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// setTestFlags replaces the in-memory flag specs for the duration of a test
// and restores the originals afterward.
func setTestFlags(t *testing.T, specs map[string]*FlagSpec) {
	t.Helper()
	orig := flagSpecs
	flagSpecs = specs
	t.Cleanup(func() { flagSpecs = orig })
}

func TestIsValidFlag(t *testing.T) {
	setTestFlags(t, map[string]*FlagSpec{
		"hidden":     {Flag: "hidden", Name: "Hidden"},
		"perma-gear": {Flag: "perma-gear", Name: "Unremovable Gear", Locked: true},
	})

	assert.True(t, IsValidFlag("hidden"))
	assert.True(t, IsValidFlag("HIDDEN"), "lookup should be case-insensitive")
	assert.True(t, IsValidFlag("perma-gear"))
	assert.True(t, IsValidFlag(All), "the All sentinel is always valid")
	assert.False(t, IsValidFlag("does-not-exist"))
}

func TestGetFlagSpec(t *testing.T) {
	setTestFlags(t, map[string]*FlagSpec{
		"hidden": {Flag: "hidden", Name: "Hidden"},
	})

	assert.NotNil(t, GetFlagSpec("hidden"))
	assert.NotNil(t, GetFlagSpec("HIDDEN"))
	assert.Nil(t, GetFlagSpec("missing"))
}

func TestGetAllFlagSpecsSorted(t *testing.T) {
	setTestFlags(t, map[string]*FlagSpec{
		"zebra": {Flag: "zebra"},
		"alpha": {Flag: "alpha"},
		"mid":   {Flag: "mid"},
	})

	sorted := GetAllFlagSpecsSorted()
	assert.Equal(t, []string{"alpha", "mid", "zebra"}, []string{sorted[0].Flag, sorted[1].Flag, sorted[2].Flag})
}

func TestValidateFlagId(t *testing.T) {
	assert.NoError(t, ValidateFlagId("perma-gear"))
	assert.NoError(t, ValidateFlagId("see_hidden"))
	assert.Error(t, ValidateFlagId(""))
	assert.Error(t, ValidateFlagId("1starts-with-number"))
	assert.Error(t, ValidateFlagId("Has-Uppercase"))
	assert.Error(t, ValidateFlagId("has spaces"))
}

func TestSaveFlagSpec_LockedRejected(t *testing.T) {
	setTestFlags(t, map[string]*FlagSpec{
		"perma-gear": {Flag: "perma-gear", Name: "Unremovable Gear", Locked: true},
	})

	err := SaveFlagSpec(&FlagSpec{Flag: "perma-gear", Name: "Changed"})
	assert.Error(t, err, "locked flags cannot be edited")
}

func TestDeleteFlagSpec_LockedRejected(t *testing.T) {
	setTestFlags(t, map[string]*FlagSpec{
		"perma-gear": {Flag: "perma-gear", Name: "Unremovable Gear", Locked: true},
	})

	err := DeleteFlagSpec("perma-gear")
	assert.Error(t, err, "locked flags cannot be deleted")
	assert.NotNil(t, GetFlagSpec("perma-gear"), "flag should still exist after rejected delete")
}

func TestBuffs_HasFlag_AllSentinel(t *testing.T) {
	setTestFlags(t, map[string]*FlagSpec{
		"hidden": {Flag: "hidden", Name: "Hidden"},
	})

	bs := New()
	// No buffs present: All should report false (nothing to match).
	assert.False(t, bs.HasFlag(All, false))

	// An unknown flag is lenient: returns false without panicking.
	assert.False(t, bs.HasFlag("not-a-real-flag", false))
}
