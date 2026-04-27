package web

import (
	"encoding/json"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/audio"
)

type audioAPIResponse struct {
	Sounds map[string]audio.AudioConfig `json:"sounds"`
	Music  []string                     `json:"music"`
}

// GET /admin/api/v1/audio
func apiV1GetAudio(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[audioAPIResponse]{
		Success: true,
		Data: audioAPIResponse{
			Sounds: audio.GetAllAudio(),
			Music:  audio.GetMusicFiles(),
		},
	})
}

// PATCH /admin/api/v1/audio
// Body: {"identifier": {"description":"...","filepath":"...","volume":80}, ...}
// Only the sounds map is writable; music is derived from disk.
func apiV1PatchAudio(w http.ResponseWriter, r *http.Request) {
	var patches map[string]audio.AudioConfig
	if err := json.NewDecoder(r.Body).Decode(&patches); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if err := audio.SaveAudio(patches); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[audioAPIResponse]{
		Success: true,
		Data: audioAPIResponse{
			Sounds: audio.GetAllAudio(),
			Music:  audio.GetMusicFiles(),
		},
	})
}
