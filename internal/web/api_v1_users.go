package web

import (
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/users"
)

// GET /admin/api/v1/users/search?name={search-name}
func apiV1SearchUsers(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "name query parameter is required")
		return
	}

	results := users.SearchUsers(name)
	if results == nil {
		results = []users.UserSearchResult{}
	}

	writeJSON(w, http.StatusOK, APIResponse[[]users.UserSearchResult]{
		Success: true,
		Data:    results,
	})
}
