package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/colorpatterns"
)

// GET /admin/api/v1/colorpatterns
func apiV1GetColorPatterns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[map[string][]int]{
		Success: true,
		Data:    colorpatterns.GetAllColorPatterns(),
	})
}

// POST /admin/api/v1/colorpatterns
// Body: {"name": "flame", "colors": [1, 2, 3]}
func apiV1CreateColorPattern(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		Colors []int  `json:"colors"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if body.Name == "" {
		writeAPIError(w, http.StatusBadRequest, "name is required")
		return
	}
	for i, c := range body.Colors {
		if c < 0 || c > 255 {
			writeAPIError(w, http.StatusBadRequest, fmt.Sprintf("color value at index %d out of range 0-255: %d", i, c))
			return
		}
	}
	existing := colorpatterns.GetAllColorPatterns()
	if _, ok := existing[body.Name]; ok {
		writeAPIError(w, http.StatusConflict, "color pattern already exists: "+body.Name)
		return
	}
	if err := colorpatterns.SaveColorPattern(body.Name, body.Colors); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// PATCH /admin/api/v1/colorpatterns
// Body: {"patternName": [1,2,3], ...}  — updates or creates each named pattern.
func apiV1PatchColorPatterns(w http.ResponseWriter, r *http.Request) {
	var patches map[string][]int
	if err := json.NewDecoder(r.Body).Decode(&patches); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	for name, colors := range patches {
		for i, c := range colors {
			if c < 0 || c > 255 {
				writeAPIError(w, http.StatusBadRequest, fmt.Sprintf("color value at index %d of pattern %q out of range 0-255: %d", i, name, c))
				return
			}
		}
		if err := colorpatterns.SaveColorPattern(name, colors); err != nil {
			writeAPIError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// DELETE /admin/api/v1/colorpatterns
// Body: {"name": "patternName"}
func apiV1DeleteColorPattern(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if body.Name == "" {
		writeAPIError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := colorpatterns.DeleteColorPattern(body.Name); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
