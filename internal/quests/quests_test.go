package quests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// seed the package-level quests map directly so tests have no file I/O dependency.
func setupQuests() {
	quests = map[int]*Quest{
		1: {
			QuestId: 1,
			Name:    "Test Quest",
			Steps: []QuestStep{
				{Id: "start"},
				{Id: "middle"},
				{Id: "end"},
			},
		},
		2: {
			QuestId: 2,
			Name:    "Single Step Quest",
			Steps: []QuestStep{
				{Id: "end"},
			},
		},
		3: {
			QuestId: 3,
			Name:    "Secret Quest",
			Secret:  true,
			Steps: []QuestStep{
				{Id: "start"},
			},
		},
	}
}

func TestTokenToParts(t *testing.T) {
	tests := []struct {
		token        string
		expectedId   int
		expectedStep string
	}{
		{"1-start", 1, "start"},
		{"1-middle", 1, "middle"},
		{"2-end", 2, "end"},
		{"1", 1, "start"},
		{"0-start", 0, "start"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			id, step := TokenToParts(tt.token)
			assert.Equal(t, tt.expectedId, id)
			assert.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestPartsToToken(t *testing.T) {
	assert.Equal(t, "1-start", PartsToToken(1, "start"))
	assert.Equal(t, "42-end", PartsToToken(42, "end"))
}

func TestPartsToToken_RoundTrip(t *testing.T) {
	token := PartsToToken(7, "middle")
	id, step := TokenToParts(token)
	assert.Equal(t, 7, id)
	assert.Equal(t, "middle", step)
}

func TestGetQuest(t *testing.T) {
	setupQuests()

	t.Run("valid token returns quest", func(t *testing.T) {
		q := GetQuest("1-start")
		assert.NotNil(t, q)
		assert.Equal(t, 1, q.QuestId)
	})

	t.Run("all+ token returns quest", func(t *testing.T) {
		q := GetQuest("1-all+")
		assert.NotNil(t, q)
	})

	t.Run("invalid step returns nil", func(t *testing.T) {
		q := GetQuest("1-nonexistent")
		assert.Nil(t, q)
	})

	t.Run("unknown quest id returns nil", func(t *testing.T) {
		q := GetQuest("999-start")
		assert.Nil(t, q)
	})
}

func TestIsTokenAfter(t *testing.T) {
	setupQuests()

	tests := []struct {
		name     string
		current  string
		next     string
		expected bool
	}{
		{
			name:     "no progress can start multi-step quest",
			current:  "1-",
			next:     "1-start",
			expected: true,
		},
		{
			name:     "no progress cannot jump to middle",
			current:  "1-",
			next:     "1-middle",
			expected: false,
		},
		{
			// quest 2 has a single step "end"; with no prior progress, "end" is reachable
			name:     "no progress can end single-step quest",
			current:  "2-",
			next:     "2-end",
			expected: true,
		},
		{
			// quest 1 has 3 steps; "end" is a valid step name but not reachable from no progress
			name:     "no progress cannot end multi-step quest via end branch",
			current:  "1-",
			next:     "1-end",
			expected: false,
		},
		{
			name:     "at start, middle is next",
			current:  "1-start",
			next:     "1-middle",
			expected: true,
		},
		{
			name:     "at start, end is not immediately next",
			current:  "1-start",
			next:     "1-end",
			expected: true, // end appears after start in the steps list
		},
		{
			name:     "at middle, end is next",
			current:  "1-middle",
			next:     "1-end",
			expected: true,
		},
		{
			name:     "same step is not after itself",
			current:  "1-start",
			next:     "1-start",
			expected: false,
		},
		{
			name:     "different quest ids are not related",
			current:  "1-start",
			next:     "2-middle",
			expected: false,
		},
		{
			name:     "unknown current quest returns false",
			current:  "999-start",
			next:     "999-middle",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTokenAfter(tt.current, tt.next)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetQuestCt(t *testing.T) {
	setupQuests()

	t.Run("exclude secret quests", func(t *testing.T) {
		assert.Equal(t, 2, GetQuestCt(false))
	})

	t.Run("include secret quests", func(t *testing.T) {
		assert.Equal(t, 3, GetQuestCt(true))
	})
}
