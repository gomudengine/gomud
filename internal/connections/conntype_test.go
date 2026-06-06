package connections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnTypeDefaultsToHuman(t *testing.T) {
	cd := NewConnectionDetails(1, nil, nil, nil)
	assert.Equal(t, ConnHuman, cd.ConnType())
}

func TestConnTypeSetGet(t *testing.T) {
	cd := NewConnectionDetails(2, nil, nil, nil)
	cd.SetConnType(ConnAI)
	assert.Equal(t, ConnAI, cd.ConnType())
}
