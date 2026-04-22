package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/races"
)

// resolveRaceId parses a path segment as an integer race ID and writes an
// error response if it is not a valid integer or the race does not exist.
func resolveRaceId(w http.ResponseWriter, idStr string) (int, bool) {
	raceId, err := strconv.Atoi(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "raceId must be an integer: "+idStr)
		return 0, false
	}
	if races.GetRace(raceId) == nil {
		writeAPIError(w, http.StatusNotFound, "race not found: "+idStr)
		return 0, false
	}
	return raceId, true
}

// GET /admin/api/v1/races
func apiV1GetRaces(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[map[int]*races.Race]{
		Success: true,
		Data:    races.GetRacesMap(),
	})
}

// POST /admin/api/v1/races
func apiV1CreateRace(w http.ResponseWriter, r *http.Request) {
	var newRace races.Race
	if err := json.NewDecoder(r.Body).Decode(&newRace); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if err := races.NewRace(&newRace); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*races.Race]{Success: true, Data: &newRace})
}

// PATCH /admin/api/v1/races/{raceId}
func apiV1PatchRace(w http.ResponseWriter, r *http.Request) {
	raceId, ok := resolveRaceId(w, r.PathValue("raceId"))
	if !ok {
		return
	}
	existing := races.GetRace(raceId)
	updated := *existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	updated.RaceId = raceId
	if err := races.SaveRace(&updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*races.Race]{Success: true, Data: &updated})
}

// DELETE /admin/api/v1/races/{raceId}
func apiV1DeleteRace(w http.ResponseWriter, r *http.Request) {
	raceId, ok := resolveRaceId(w, r.PathValue("raceId"))
	if !ok {
		return
	}
	if err := races.DeleteRace(raceId); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
