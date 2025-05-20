package characters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlignmentToString(t *testing.T) {
	tests := []struct {
		name      string
		alignment int8
		expected  string
	}{
		// Unholy
		{"Unholy lower bound", -100, "unholy"},
		{"Unholy upper bound", -80, "unholy"},
		{"Just above Unholy", -79, "evil"},

		// Evil
		{"Evil lower bound", -79, "evil"},
		{"Evil upper bound", -60, "evil"},
		{"Just above Evil", -59, "corrupt"},

		// Corrupt
		{"Corrupt lower bound", -59, "corrupt"},
		{"Corrupt upper bound", -40, "corrupt"},
		{"Just above Corrupt", -39, "misguided"},

		// Misguided
		{"Misguided lower bound", -39, "misguided"},
		{"Misguided upper bound", -20, "misguided"},
		{"Just above Misguided", -19, "neutral"},

		// Neutral
		{"Neutral low", -19, "neutral"},
		{"Neutral", 0, "neutral"},
		{"Neutral high", 19, "neutral"},
		{"Just above Neutral high", 20, "lawful"},

		// Lawful
		{"Lawful lower bound", 20, "lawful"},
		{"Lawful upper bound", 39, "lawful"},
		{"Just above Lawful", 40, "virtuous"},

		// Virtuous
		{"Virtuous lower bound", 40, "virtuous"},
		{"Virtuous upper bound", 59, "virtuous"},
		{"Just above Virtuous", 60, "good"},

		// Good
		{"Good lower bound", 60, "good"},
		{"Good upper bound", 79, "good"},
		{"Just above Good", 80, "holy"},

		// Holy
		{"Holy lower bound", 80, "holy"},
		{"Holy upper bound", 100, "holy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AlignmentToString(tt.alignment)
			assert.Equal(t, tt.expected, result)
		})
	}
}
