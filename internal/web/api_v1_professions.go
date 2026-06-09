package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/skills"
)

// resolveProfessionId lowercases a path segment and writes a 404 response if the
// profession is not loaded. Profession ids may contain spaces; r.PathValue has
// already decoded any percent-encoding.
func resolveProfessionId(w http.ResponseWriter, idStr string) (string, bool) {
	professionId := strings.ToLower(strings.TrimSpace(idStr))
	if skills.GetProfessionSpec(professionId) == nil {
		writeAPIError(w, http.StatusNotFound, "profession not found: "+idStr)
		return "", false
	}
	return professionId, true
}

// GET /admin/api/v1/professions
func apiV1GetProfessions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[map[string]*skills.Profession]{
		Success: true,
		Data:    skills.GetProfessionsMap(),
	})
}

// POST /admin/api/v1/professions
func apiV1CreateProfession(w http.ResponseWriter, r *http.Request) {
	var newProfession skills.Profession
	if err := json.NewDecoder(r.Body).Decode(&newProfession); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if err := skills.CreateProfession(&newProfession); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*skills.Profession]{Success: true, Data: &newProfession})
}

// PATCH /admin/api/v1/professions/{professionId}
func apiV1PatchProfession(w http.ResponseWriter, r *http.Request) {
	professionId, ok := resolveProfessionId(w, r.PathValue("professionId"))
	if !ok {
		return
	}
	existing := skills.GetProfessionSpec(professionId)
	updated := *existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	updated.ProfessionId = professionId // id is immutable
	if err := skills.SaveProfession(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*skills.Profession]{Success: true, Data: &updated})
}

// DELETE /admin/api/v1/professions/{professionId}
func apiV1DeleteProfession(w http.ResponseWriter, r *http.Request) {
	professionId, ok := resolveProfessionId(w, r.PathValue("professionId"))
	if !ok {
		return
	}
	if err := skills.DeleteProfession(professionId); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
