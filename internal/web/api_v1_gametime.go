package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/gametime"
)

// GET /admin/api/v1/gametime
// Returns all calendar names and their configurations.
func apiV1GetGameTime(w http.ResponseWriter, r *http.Request) {
	names := gametime.GetCalendars()
	result := make(map[string]gametime.CalendarConfig, len(names))
	for _, name := range names {
		cfg, _ := gametime.GetCalendar(name)
		result[name] = cfg
	}
	writeJSON(w, http.StatusOK, APIResponse[map[string]gametime.CalendarConfig]{
		Success: true,
		Data:    result,
	})
}

// GET /admin/api/v1/gametime/{calendar}
// Returns the configuration for a single named calendar.
func apiV1GetGameTimeCalendar(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("calendar"))
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "calendar name is required")
		return
	}
	cfg, ok := gametime.GetCalendar(name)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "calendar not found: "+name)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[gametime.CalendarConfig]{
		Success: true,
		Data:    cfg,
	})
}

// POST /admin/api/v1/gametime
// Body: {"name": "...", "config": {...}}
// Creates a new named calendar. Returns 400 if the name already exists.
func apiV1CreateGameTimeCalendar(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string                  `json:"name"`
		Config gametime.CalendarConfig `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		writeAPIError(w, http.StatusBadRequest, "name is required")
		return
	}
	if _, exists := gametime.GetCalendar(body.Name); exists {
		writeAPIError(w, http.StatusBadRequest, "calendar already exists: "+body.Name)
		return
	}
	if err := gametime.SaveCalendar(body.Name, body.Config); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	cfg, _ := gametime.GetCalendar(body.Name)
	writeJSON(w, http.StatusCreated, APIResponse[gametime.CalendarConfig]{
		Success: true,
		Data:    cfg,
	})
}

// PATCH /admin/api/v1/gametime/{calendar}
// Updates an existing named calendar. The name is taken from the URL path.
func apiV1PatchGameTimeCalendar(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("calendar"))
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "calendar name is required")
		return
	}
	existing, ok := gametime.GetCalendar(name)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "calendar not found: "+name)
		return
	}
	updated := existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if err := gametime.SaveCalendar(name, updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	cfg, _ := gametime.GetCalendar(name)
	writeJSON(w, http.StatusOK, APIResponse[gametime.CalendarConfig]{
		Success: true,
		Data:    cfg,
	})
}

// DELETE /admin/api/v1/gametime/{calendar}
// Deletes a named calendar. The "default" calendar cannot be deleted.
func apiV1DeleteGameTimeCalendar(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("calendar"))
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "calendar name is required")
		return
	}
	if err := gametime.DeleteCalendar(name); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
