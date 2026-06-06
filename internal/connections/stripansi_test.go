package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripAnsiRemovesColorCodes(t *testing.T) {
	got := StripAnsi([]byte("\x1b[31mred\x1b[0m text"))
	assert.Equal(t, []byte("red text"), got)
}

func TestStripAnsiLeavesPlainText(t *testing.T) {
	in := []byte("no codes here")
	assert.Equal(t, in, StripAnsi(in))
}
