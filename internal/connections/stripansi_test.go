package connections

import (
	"bytes"
	"testing"
)

func TestStripAnsiRemovesColorCodes(t *testing.T) {
	in := []byte("\x1b[31mred\x1b[0m text")
	got := StripAnsi(in)
	if !bytes.Equal(got, []byte("red text")) {
		t.Errorf("expected %q, got %q", "red text", got)
	}
}

func TestStripAnsiLeavesPlainText(t *testing.T) {
	in := []byte("no codes here")
	if got := StripAnsi(in); !bytes.Equal(got, in) {
		t.Errorf("plain text should be unchanged, got %q", got)
	}
}
