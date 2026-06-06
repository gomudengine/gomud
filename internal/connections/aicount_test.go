package connections

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, base+2, ActiveAIConnectionCount())

	// cleanup
	Remove(human.ConnectionId())
	Remove(ai1.ConnectionId())
	Remove(ai2.ConnectionId())

	assert.Equal(t, base, ActiveAIConnectionCount())
}
