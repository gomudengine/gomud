package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
)

// GET /admin/api/v1/buffs/flags
func apiV1GetBuffFlags(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[[]buffs.FlagSpec]{
		Success: true,
		Data:    buffs.GetAllFlagSpecsSorted(),
	})
}

// GET /admin/api/v1/buffs/flags/{flag}
func apiV1GetBuffFlag(w http.ResponseWriter, r *http.Request) {
	flag := strings.ToLower(strings.TrimSpace(r.PathValue("flag")))
	spec := buffs.GetFlagSpec(flag)
	if spec == nil {
		writeAPIError(w, http.StatusNotFound, "flag not found: "+flag)
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*buffs.FlagSpec]{
		Success: true,
		Data:    spec,
	})
}

// POST /admin/api/v1/buffs/flags
func apiV1CreateBuffFlag(w http.ResponseWriter, r *http.Request) {
	var spec buffs.FlagSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	// New flags are always created unlocked.
	spec.Locked = false
	if err := buffs.CreateFlagSpec(&spec); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, APIResponse[buffs.FlagSpec]{
		Success: true,
		Data:    spec,
	})
}

// PATCH /admin/api/v1/buffs/flags/{flag}
func apiV1PatchBuffFlag(w http.ResponseWriter, r *http.Request) {
	flag := strings.ToLower(strings.TrimSpace(r.PathValue("flag")))

	existing := buffs.GetFlagSpec(flag)
	if existing == nil {
		writeAPIError(w, http.StatusNotFound, "flag not found: "+flag)
		return
	}

	updated := *existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	// The flag identifier cannot be changed via PATCH.
	updated.Flag = flag

	if err := buffs.SaveFlagSpec(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[buffs.FlagSpec]{
		Success: true,
		Data:    updated,
	})
}

// DELETE /admin/api/v1/buffs/flags/{flag}
func apiV1DeleteBuffFlag(w http.ResponseWriter, r *http.Request) {
	flag := strings.ToLower(strings.TrimSpace(r.PathValue("flag")))
	if err := buffs.DeleteFlagSpec(flag); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// GET /admin/api/v1/buffs/flags/{flag}/yaml
func apiV1GetBuffFlagYAML(w http.ResponseWriter, r *http.Request) {
	flag := strings.ToLower(strings.TrimSpace(r.PathValue("flag")))
	spec := buffs.GetFlagSpec(flag)
	if spec == nil {
		writeAPIError(w, http.StatusNotFound, "flag not found: "+flag)
		return
	}
	path := dataFiles() + `/buffs-flags/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResponse{YAML: content, Path: relPath(path)},
	})
}
