package stats

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatInfo_GainsForLevel(t *testing.T) {
	tests := []struct {
		name     string
		base     int
		level    int
		expected int
	}{
		{name: "level 1 base 10", base: 10, level: 1, expected: 0},
		{name: "level 1 base 0", base: 0, level: 1, expected: 0},
		{name: "level 0 clamps to 1", base: 10, level: 0, expected: 0},
		{name: "negative level clamps to 1", base: 10, level: -5, expected: 0},
		// level 2, base 10: levelScale=(2-1)*BaseModFactor=0.333, basePoints=int(0.333*10)=3, free=int(2*0.5)=1 => 4
		{name: "level 2 base 10", base: 10, level: 2, expected: 4},
		// level 10, base 10: levelScale=9*0.333=3.0, basePoints=int(3.0*10)=30, free=int(10*0.5)=5 => 35
		{name: "level 10 base 10", base: 10, level: 10, expected: 35},
		// level 10, base 0: basePoints=0, free=5 => 5
		{name: "level 10 base 0", base: 0, level: 10, expected: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := StatInfo{Base: tt.base}
			assert.Equal(t, tt.expected, si.GainsForLevel(tt.level))
		})
	}
}

func TestStatInfo_Recalculate_BelowCap(t *testing.T) {
	si := StatInfo{Base: 10, Training: 5}
	si.Recalculate(10)

	expectedRacial := si.GainsForLevel(10)
	assert.Equal(t, expectedRacial, si.Racial)
	assert.Equal(t, expectedRacial+5, si.Value)
	assert.Equal(t, si.Value, si.ValueAdj, "ValueAdj should equal Value when below cap")
}

func TestStatInfo_Recalculate_WithMods(t *testing.T) {
	si := StatInfo{Base: 10, Training: 3}
	si.SetMod(7)
	si.Recalculate(5)

	assert.Equal(t, si.Racial+3+7, si.Value)
}

func TestStatInfo_Recalculate_AboveCap(t *testing.T) {
	// Force a Value above 105 by using high training
	si := StatInfo{Base: 10, Training: 200}
	si.Recalculate(1)

	// Value = Racial + 200; Racial at level 1 = 0, so Value = 200
	assert.Equal(t, 200, si.Value)

	// ValueAdj formula: 100 + round(sqrt(overage) * 2), overage = 200 - 100 = 100
	expectedAdj := 100 + int(math.Round(math.Sqrt(100)*2))
	assert.Equal(t, expectedAdj, si.ValueAdj)
	assert.True(t, si.ValueAdj < si.Value, "ValueAdj should be less than Value when above cap")
}

func TestStatInfo_Recalculate_ExactlyAtCap(t *testing.T) {
	// Value == 105 triggers the cap branch (>= 105)
	si := StatInfo{Base: 0, Training: 105}
	si.Recalculate(1)

	assert.Equal(t, 105, si.Value)
	overage := 105 - 100
	expectedAdj := 100 + int(math.Round(math.Sqrt(float64(overage))*2))
	assert.Equal(t, expectedAdj, si.ValueAdj)
}

func TestStatInfo_SetMod(t *testing.T) {
	t.Run("single mod", func(t *testing.T) {
		si := StatInfo{}
		si.SetMod(5)
		assert.Equal(t, 5, si.Mods)
	})

	t.Run("multiple mods are summed", func(t *testing.T) {
		si := StatInfo{}
		si.SetMod(3, 7, -2)
		assert.Equal(t, 8, si.Mods)
	})

	t.Run("no args resets to zero", func(t *testing.T) {
		si := StatInfo{Mods: 99}
		si.SetMod()
		assert.Equal(t, 0, si.Mods)
	})
}
