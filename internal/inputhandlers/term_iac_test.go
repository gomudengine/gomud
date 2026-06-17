package inputhandlers

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/term"
)

// newEnvironPayload builds the bytes that appear between `IAC SB NEW-ENVIRON IS`
// and the trailing `IAC SE` for a sequence of VAR/VALUE pairs.
func newEnvironPayload(pairs ...[2]string) []byte {
	out := []byte{}
	for _, p := range pairs {
		out = append(out, term.TELNET_NEWENV_VAR)
		out = append(out, []byte(p[0])...)
		out = append(out, term.TELNET_NEWENV_VALUE)
		out = append(out, []byte(p[1])...)
	}
	return out
}

func TestNewEnvironIsMudlet(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    bool
	}{
		{
			name:    "mudlet with version",
			payload: newEnvironPayload([2]string{"CLIENT_NAME", "Mudlet"}, [2]string{"CLIENT_VERSION", "4.17.2"}),
			want:    true,
		},
		{
			name:    "mudlet uppercase value",
			payload: newEnvironPayload([2]string{"CLIENT_NAME", "MUDLET"}),
			want:    true,
		},
		{
			name:    "client name not first",
			payload: newEnvironPayload([2]string{"CLIENT_VERSION", "4.17.2"}, [2]string{"CLIENT_NAME", "Mudlet"}),
			want:    true,
		},
		{
			name:    "different client",
			payload: newEnvironPayload([2]string{"CLIENT_NAME", "TinTin++"}),
			want:    false,
		},
		{
			name:    "client name without value",
			payload: append([]byte{term.TELNET_NEWENV_VAR}, []byte("CLIENT_NAME")...),
			want:    false,
		},
		{
			name:    "empty payload",
			payload: []byte{},
			want:    false,
		},
		{
			name:    "uservar segment ignored, var matched",
			payload: append(append([]byte{term.TELNET_NEWENV_USERVAR}, []byte("FOO")...), newEnvironPayload([2]string{"CLIENT_NAME", "Mudlet"})...),
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := newEnvironIsMudlet(tc.payload); got != tc.want {
				t.Errorf("newEnvironIsMudlet() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestNewEnvironResponseMatcher verifies the term matcher extracts the payload
// that newEnvironIsMudlet expects from a full client response frame.
func TestNewEnvironResponseMatcher(t *testing.T) {
	frame := term.TelnetNewEnvironResponse.BytesWithPayload(
		newEnvironPayload([2]string{"CLIENT_NAME", "Mudlet"}),
	)

	ok, payload := term.Matches(frame, term.TelnetNewEnvironResponse)
	if !ok {
		t.Fatalf("expected frame to match TelnetNewEnvironResponse")
	}
	if !newEnvironIsMudlet(payload) {
		t.Errorf("expected extracted payload to be detected as Mudlet, payload=%v", payload)
	}
}
