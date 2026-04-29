package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/conversations"
)

// GET /admin/api/v1/conversations
func apiV1GetConversations(w http.ResponseWriter, r *http.Request) {
	files, err := conversations.ListConversationFiles()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[[]conversations.ConversationFile]{
		Success: true,
		Data:    files,
	})
}

// GET /admin/api/v1/conversations/{zone}/{mobId}
func apiV1GetConversation(w http.ResponseWriter, r *http.Request) {
	zone, mobId, ok := resolveConversationPath(w, r)
	if !ok {
		return
	}

	contents, err := conversations.GetConversationFile(zone, mobId)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[conversations.ConversationFileContents]{
		Success: true,
		Data:    contents,
	})
}

// PUT /admin/api/v1/conversations/{zone}/{mobId}
func apiV1PutConversation(w http.ResponseWriter, r *http.Request) {
	zone, mobId, ok := resolveConversationPath(w, r)
	if !ok {
		return
	}

	var body []conversations.ConversationData
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	_, lookupErr := conversations.GetConversationFile(zone, mobId)
	isNew := lookupErr != nil

	if err := conversations.SaveConversationFile(zone, mobId, body); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	contents, err := conversations.GetConversationFile(zone, mobId)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}
	writeJSON(w, status, APIResponse[conversations.ConversationFileContents]{
		Success: true,
		Data:    contents,
	})
}

// DELETE /admin/api/v1/conversations/{zone}/{mobId}
func apiV1DeleteConversation(w http.ResponseWriter, r *http.Request) {
	zone, mobId, ok := resolveConversationPath(w, r)
	if !ok {
		return
	}

	if err := conversations.DeleteConversationFile(zone, mobId); err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// resolveConversationPath extracts and validates the {zone} and {mobId} path
// values, writing an error response and returning false on any failure.
func resolveConversationPath(w http.ResponseWriter, r *http.Request) (string, int, bool) {
	zone := strings.TrimSpace(r.PathValue("zone"))
	if zone == "" {
		writeAPIError(w, http.StatusBadRequest, "zone is required")
		return "", 0, false
	}

	mobIdStr := strings.TrimSpace(r.PathValue("mobId"))
	mobId, err := strconv.Atoi(mobIdStr)
	if err != nil || mobId <= 0 {
		writeAPIError(w, http.StatusBadRequest, "mobId must be a positive integer")
		return "", 0, false
	}

	return zone, mobId, true
}
