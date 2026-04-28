package combat

import (
	"fmt"
	"math"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/races"
)

func TestAlignmentChange(t *testing.T) {
	tests := []struct {
		killerAlignment int8
		killedAlignment int8
		expectedChange  int
	}{

		{0, 0, 0},
		{0, 5, 0},
		{5, 0, 0},

		{0, 15, 0},
		{0, -15, 0},

		{15, -15, 0},
		{-15, 15, 0},

		{15, 25, 0},
		{25, 15, 0},

		{-20, -25, 2},
		{-25, -20, 2},

		{50, -10, -1},
		{-50, 10, -1},

		{50, -50, 2},
		{-50, 50, -2},

		{90, 0, -2},
		{-90, 0, -2},

		{100, 20, -4},
		{-100, -20, 4},

		{90, -90, 4},
		{-90, 90, -4},
	}

	for _, test := range tests {
		desc := fmt.Sprintf(`%s kills %s`, characters.AlignmentToString(test.killerAlignment), characters.AlignmentToString(test.killedAlignment))
		delta := int(math.Abs(math.Max(float64(test.killerAlignment), float64(test.killedAlignment))-math.Min(float64(test.killerAlignment), float64(test.killedAlignment))) * 0.5)
		result := AlignmentChange(test.killerAlignment, test.killedAlignment)
		if result != test.expectedChange {
			t.Errorf("%s [Delta: %d]: AlignmentChange(%d, %d) = %d; want %d",
				desc, delta, test.killerAlignment, test.killedAlignment, result, test.expectedChange)
		}
	}
}

func TestAttackCount(t *testing.T) {
	tests := []struct {
		atkSpd     int
		defSpd     int
		attacksMod int
		want       int
	}{
		{50, 50, 0, 1},  // equal speed -> ceil(0/25)=0, floor to 1
		{75, 50, 0, 1},  // diff=25 -> ceil(25/25)=1
		{100, 50, 0, 2}, // diff=50 -> ceil(50/25)=2
		{50, 100, 0, 1}, // negative diff -> floor to 1
		{100, 50, 1, 3}, // 2 + attacksMod=1
		{50, 50, -5, 1}, // 1 + (-5) = -4 -> floor to 1
	}
	for _, tt := range tests {
		got := attackCount(tt.atkSpd, tt.defSpd, tt.attacksMod)
		if got != tt.want {
			t.Errorf("attackCount(%d, %d, %d) = %d; want %d", tt.atkSpd, tt.defSpd, tt.attacksMod, got, tt.want)
		}
	}
}

func TestHitChance(t *testing.T) {
	tests := []struct {
		atkSpd  int
		defSpd  int
		wantMin int
		wantMax int
	}{
		{0, 0, 30, 100},   // zero speeds: atkPlusDef clamped to 1, result = 30 + 0 = 30
		{100, 0, 100, 100}, // attacker dominates: 30 + 70 = 100
		{0, 100, 30, 30},   // defender dominates: 30 + 0 = 30
		{50, 50, 65, 65},   // equal: 30 + 35 = 65
	}
	for _, tt := range tests {
		got := hitChance(tt.atkSpd, tt.defSpd)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("hitChance(%d, %d) = %d; want [%d, %d]", tt.atkSpd, tt.defSpd, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestCritChance(t *testing.T) {
	tests := []struct {
		atkStr         int
		atkSpd         int
		levelDiff      int
		hasAccuracy    bool
		targetHasBlink bool
		want           int
	}{
		// base: 5 + round((10+10)/1) = 25, no flags
		{10, 10, 1, false, false, 25},
		// accuracy doubles: 25*2 = 50
		{10, 10, 1, true, false, 50},
		// blink halves: 25/2 = 12
		{10, 10, 1, false, true, 12},
		// both: 25*2/2 = 25
		{10, 10, 1, true, true, 25},
		// clamp min: very low stats
		{0, 0, 100, false, false, 5},
		// clamp max: very high stats
		{1000, 1000, 1, false, false, 75},
		// levelDiff < 1 clamped to 1
		{10, 10, -5, false, false, 25},
	}
	for _, tt := range tests {
		got := critChance(tt.atkStr, tt.atkSpd, tt.levelDiff, tt.hasAccuracy, tt.targetHasBlink)
		if got != tt.want {
			t.Errorf("critChance(%d, %d, %d, %v, %v) = %d; want %d",
				tt.atkStr, tt.atkSpd, tt.levelDiff, tt.hasAccuracy, tt.targetHasBlink, got, tt.want)
		}
	}
}

func TestDualWieldHitPenalty(t *testing.T) {
	tests := []struct {
		dwLevel int
		want    int
	}{
		{0, -35},
		{1, -35},
		{2, -35},
		{3, -35},
		{4, -25},
		{5, -25},
	}
	for _, tt := range tests {
		got := dualWieldHitPenalty(tt.dwLevel)
		if got != tt.want {
			t.Errorf("dualWieldHitPenalty(%d) = %d; want %d", tt.dwLevel, got, tt.want)
		}
	}
}

func TestDualWieldActiveWeaponCount(t *testing.T) {
	// Deterministic cases
	tests := []struct {
		dwLevel   int
		bothClaws bool
		want      int
	}{
		{0, false, 1},
		{1, false, 1},
		{3, false, 2},
		{4, false, 2},
		{0, true, 2},
		{1, true, 2},
	}
	for _, tt := range tests {
		got := dualWieldActiveWeaponCount(tt.dwLevel, tt.bothClaws)
		if got != tt.want {
			t.Errorf("dualWieldActiveWeaponCount(%d, %v) = %d; want %d", tt.dwLevel, tt.bothClaws, got, tt.want)
		}
	}

	// Probabilistic case: dwLevel == 2, bothClaws == false should return 1 or 2
	// Run enough iterations to verify both outcomes occur.
	saw1, saw2 := false, false
	for i := 0; i < 200; i++ {
		v := dualWieldActiveWeaponCount(2, false)
		if v == 1 {
			saw1 = true
		} else if v == 2 {
			saw2 = true
		} else {
			t.Errorf("dualWieldActiveWeaponCount(2, false) returned unexpected value %d", v)
		}
		if saw1 && saw2 {
			break
		}
	}
	if !saw1 || !saw2 {
		t.Errorf("dualWieldActiveWeaponCount(2, false) did not produce both 1 and 2 over 200 iterations")
	}
}

func TestCritDamageBonus(t *testing.T) {
	tests := []struct {
		dCount, dSides, dBonus int
		want                   int
	}{
		{2, 6, 3, 15},  // 2*6+3
		{1, 8, 0, 8},   // 1*8+0
		{3, 4, -2, 10}, // 3*4-2
		{0, 0, 5, 5},   // 0*0+5
	}
	for _, tt := range tests {
		got := critDamageBonus(tt.dCount, tt.dSides, tt.dBonus)
		if got != tt.want {
			t.Errorf("critDamageBonus(%d, %d, %d) = %d; want %d", tt.dCount, tt.dSides, tt.dBonus, got, tt.want)
		}
	}
}

func TestApplyDefenseReduction(t *testing.T) {
	// Zero defense: no reduction ever
	for i := 0; i < 50; i++ {
		final, red := applyDefenseReduction(100, 0)
		if final != 100 || red != 0 {
			t.Errorf("applyDefenseReduction(100, 0) = (%d, %d); want (100, 0)", final, red)
		}
	}

	// Non-zero defense: final + reduction == original damage
	for i := 0; i < 100; i++ {
		final, red := applyDefenseReduction(100, 50)
		if final+red != 100 {
			t.Errorf("applyDefenseReduction(100, 50): final(%d) + reduction(%d) != 100", final, red)
		}
		if final < 0 || red < 0 {
			t.Errorf("applyDefenseReduction(100, 50): negative value final=%d red=%d", final, red)
		}
	}
}

func TestDamagePercentOfMax(t *testing.T) {
	tests := []struct {
		damage, dCount, dSides, dBonus int
		want                           int
	}{
		{0, 2, 6, 0, 0},    // 0% of max
		{12, 2, 6, 0, 100}, // max damage = 12, 100%
		{6, 2, 6, 0, 50},   // half max
		{1, 0, 0, 0, 100},  // maxDmg clamped to 1, 100%
		{5, 2, 6, 0, 42},   // ceil(5/12*100) = 42
	}
	for _, tt := range tests {
		got := damagePercentOfMax(tt.damage, tt.dCount, tt.dSides, tt.dBonus)
		if got != tt.want {
			t.Errorf("damagePercentOfMax(%d, %d, %d, %d) = %d; want %d",
				tt.damage, tt.dCount, tt.dSides, tt.dBonus, got, tt.want)
		}
	}
}

func TestTameSizeModifier(t *testing.T) {
	tests := []struct {
		size races.Size
		want int
	}{
		{races.Large, -25},
		{races.Small, 0},
		{races.Medium, -10},
		{races.Size("unknown"), -10},
	}
	for _, tt := range tests {
		got := tameSizeModifier(tt.size)
		if got != tt.want {
			t.Errorf("tameSizeModifier(%q) = %d; want %d", tt.size, got, tt.want)
		}
	}
}

func TestTameHealthBonus(t *testing.T) {
	correctTests := []struct {
		currentHP, maxHP int
		want             float64
	}{
		{100, 100, 0},
		{50, 100, 25},
		{1, 100, 49},
		{0, 100, 50}, // math.Ceil(0) = 0, so 50-0=50; but division by zero risk handled by caller
	}
	for _, tt := range correctTests {
		got := tameHealthBonus(tt.currentHP, tt.maxHP)
		if got != tt.want {
			t.Errorf("tameHealthBonus(%d, %d) = %g; want %g", tt.currentHP, tt.maxHP, got, tt.want)
		}
	}
}

func TestChanceToTame(t *testing.T) {
	tests := []struct {
		proficiency   int
		levelDiff     int
		currentHP     int
		maxHP         int
		tamerSize     races.Size
		targetIsAggro bool
		wantMin       int
		wantMax       int
	}{
		// High proficiency, level advantage, full HP, medium size, not aggro
		{100, 10, 100, 100, races.Medium, false, 100, 110},
		// Low proficiency, level disadvantage, full HP, medium size, not aggro
		{1, -25, 100, 100, races.Medium, false, -34, -33},
		// Aggro halves the result (round up via Ceil)
		{50, 0, 100, 100, races.Medium, true, 20, 21},
		// Large size penalty
		{50, 0, 100, 100, races.Large, false, 25, 25},
		// Small size no penalty
		{50, 0, 100, 100, races.Small, false, 50, 50},
		// Proficiency clamped at min=1
		{-10, 0, 100, 100, races.Medium, false, -9, -9},
		// Proficiency clamped at max=100
		{200, 0, 100, 100, races.Medium, false, 90, 90},
		// levelDiff clamped at max=25
		{50, 100, 100, 100, races.Medium, false, 65, 65},
		// levelDiff clamped at min=-25
		{50, -100, 100, 100, races.Medium, false, 15, 15},
	}
	for _, tt := range tests {
		got := chanceToTame(tt.proficiency, tt.levelDiff, tt.currentHP, tt.maxHP, tt.tamerSize, tt.targetIsAggro)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("chanceToTame(%d, %d, %d, %d, %q, %v) = %d; want [%d, %d]",
				tt.proficiency, tt.levelDiff, tt.currentHP, tt.maxHP, tt.tamerSize, tt.targetIsAggro,
				got, tt.wantMin, tt.wantMax)
		}
	}
}
