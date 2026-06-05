package connections

import (
	"net"
	"testing"
)

func TestActiveAIConnectionCount(t *testing.T) {
	base := ActiveAIConnectionCount()

	// net.Pipe gives a real in-memory net.Conn whose Close() is safe to call.
	hConn, _ := net.Pipe()
	a1Conn, _ := net.Pipe()
	a2Conn, _ := net.Pipe()

	human := Add(hConn, nil) // defaults to ConnHuman
	ai1 := Add(a1Conn, nil, ConnAI)
	ai2 := Add(a2Conn, nil, ConnAI)

	if got := ActiveAIConnectionCount(); got != base+2 {
		t.Errorf("expected %d AI connections, got %d", base+2, got)
	}

	// cleanup
	Remove(human.ConnectionId())
	Remove(ai1.ConnectionId())
	Remove(ai2.ConnectionId())

	if got := ActiveAIConnectionCount(); got != base {
		t.Errorf("expected count to return to %d after removal, got %d", base, got)
	}
}
