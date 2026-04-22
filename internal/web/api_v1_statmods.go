package web

import (
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/statmods"
)

// GET /admin/api/v1/statmods
func apiV1GetStatMods(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[map[statmods.StatName]string]{
		Success: true,
		Data:    statmods.GetStatMods(),
	})
}
