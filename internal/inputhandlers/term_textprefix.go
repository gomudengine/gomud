package inputhandlers

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
)

var (
	prefixHandlers = []PrefixHandler{}
)

type PrefixHandler interface {
	HandleTextPrefix(uint64, []byte) bool
}

func AddTextPrefixHandler(h PrefixHandler) {
	prefixHandlers = append(prefixHandlers, h)
}

func TextPrefixHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	if len(prefixHandlers) > 0 {

		for _, h := range prefixHandlers {
			if h.HandleTextPrefix(clientInput.ConnectionId, clientInput.DataIn) {
				return false
			}
		}

	}

	// Continue to next handler
	return true
}
