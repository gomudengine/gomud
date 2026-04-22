package web

import (
	"encoding/json"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/keywords"
)

// GET /admin/api/v1/keywords
func apiV1GetKeywords(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[*keywords.Aliases]{
		Success: true,
		Data:    keywords.GetKeywords(),
	})
}

// PATCH /admin/api/v1/keywords
func apiV1PatchKeywords(w http.ResponseWriter, r *http.Request) {
	var incoming keywords.Aliases
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if err := keywords.SaveKeywords(&incoming); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*keywords.Aliases]{
		Success: true,
		Data:    keywords.GetKeywords(),
	})
}
