package web

import (
	"net/http"
	"sort"
)

// roomTagProvider is set by main at startup via SetRoomTagProvider.
// It returns a map of module name to the room tags that module has reserved.
var roomTagProvider func() map[string][]string = func() map[string][]string { return nil }

// SetRoomTagProvider registers the function that returns registered room tags.
func SetRoomTagProvider(f func() map[string][]string) {
	roomTagProvider = f
}

type tagEntry struct {
	Tag    string `json:"tag"`
	Module string `json:"module"`
}

// GET /admin/api/v1/tags
func apiV1GetTags(w http.ResponseWriter, r *http.Request) {
	registered := roomTagProvider()

	modules := make([]string, 0, len(registered))
	for name := range registered {
		modules = append(modules, name)
	}
	sort.Strings(modules)

	result := make([]tagEntry, 0)
	for _, name := range modules {
		for _, tag := range registered[name] {
			result = append(result, tagEntry{Tag: tag, Module: name})
		}
	}

	writeJSON(w, http.StatusOK, APIResponse[[]tagEntry]{
		Success: true,
		Data:    result,
	})
}
