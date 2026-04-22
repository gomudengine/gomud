package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/quests"
)

// GET /admin/api/v1/quests
func apiV1GetQuests(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[[]quests.Quest]{
		Success: true,
		Data:    quests.GetAllQuests(),
	})
}

// PATCH /admin/api/v1/quests
// Creates or replaces the quest identified by QuestId in the body.
func apiV1PatchQuest(w http.ResponseWriter, r *http.Request) {
	var incoming quests.Quest
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if incoming.QuestId < 1 {
		writeAPIError(w, http.StatusBadRequest, "questId must be set in the request body")
		return
	}

	if err := quests.SaveQuest(&incoming); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*quests.Quest]{
		Success: true,
		Data:    &incoming,
	})
}

// DELETE /admin/api/v1/quests/{questId}
func apiV1DeleteQuest(w http.ResponseWriter, r *http.Request) {
	questId, err := strconv.Atoi(r.PathValue("questId"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "questId must be an integer: "+r.PathValue("questId"))
		return
	}

	if err := quests.DeleteQuest(questId); err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
