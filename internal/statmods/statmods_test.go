package statmods

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatMods_Get_Empty(t *testing.T) {
	var s StatMods
	assert.Equal(t, 0, s.Get("strength"))
}

func TestStatMods_Get_SingleKey(t *testing.T) {
	s := StatMods{"strength": 5}
	assert.Equal(t, 5, s.Get("strength"))
	assert.Equal(t, 0, s.Get("speed"))
}

func TestStatMods_Get_MultipleKeys(t *testing.T) {
	s := StatMods{"strength": 3, "speed": 7}
	assert.Equal(t, 10, s.Get("strength", "speed"))
}

func TestStatMods_Get_MissingKeysReturnZero(t *testing.T) {
	s := StatMods{"strength": 5}
	assert.Equal(t, 0, s.Get("smarts", "vitality"))
}

func TestStatMods_Get_NoArgs(t *testing.T) {
	s := StatMods{"strength": 5}
	assert.Equal(t, 0, s.Get())
}

func TestStatMods_Add_NewKey(t *testing.T) {
	s := make(StatMods)
	s.Add("strength", 10)
	assert.Equal(t, 10, s.Get("strength"))
}

func TestStatMods_Add_ExistingKey(t *testing.T) {
	s := StatMods{"strength": 5}
	s.Add("strength", 3)
	assert.Equal(t, 8, s.Get("strength"))
}

func TestStatMods_Add_NegativeValue(t *testing.T) {
	s := StatMods{"strength": 10}
	s.Add("strength", -4)
	assert.Equal(t, 6, s.Get("strength"))
}

func TestStatMods_Add_MultipleKeys(t *testing.T) {
	s := make(StatMods)
	s.Add("strength", 5)
	s.Add("speed", 3)
	assert.Equal(t, 5, s.Get("strength"))
	assert.Equal(t, 3, s.Get("speed"))
	assert.Equal(t, 8, s.Get("strength", "speed"))
}
