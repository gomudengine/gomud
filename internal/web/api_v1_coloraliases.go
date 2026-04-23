package web

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/GoMudEngine/ansitags"
	"gopkg.in/yaml.v2"
)

// GET /admin/api/v1/color-aliases
// Returns all currently loaded ANSI color aliases.
func apiV1GetColorAliases(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[map[string]any]{
		Success: true,
		Data:    ansitags.GetAliases(),
	})
}

// PATCH /admin/api/v1/color-aliases
// Body: {"alias": "username", "value": 195}
// Sets a single alias in memory and persists it to ansi-aliases.yaml.
func apiV1PatchColorAlias(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Alias string `json:"alias"`
		Value int    `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if strings.TrimSpace(body.Alias) == "" {
		writeAPIError(w, http.StatusBadRequest, "alias is required")
		return
	}
	if body.Value < 0 || body.Value > 255 {
		writeAPIError(w, http.StatusBadRequest, "value must be between 0 and 255")
		return
	}

	if err := ansitags.SetAlias(body.Alias, body.Value); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := writeColorAliasYAML(ansitags.GetAliases()); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "alias set in memory but failed to persist: "+err.Error())
		return
	}

	templates.LoadAliases()

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// DELETE /admin/api/v1/color-aliases/{alias}
// Removes the named alias from ansi-aliases.yaml and reloads.
func apiV1DeleteColorAlias(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if alias == "" {
		writeAPIError(w, http.StatusBadRequest, "alias is required")
		return
	}

	current := ansitags.GetAliases()
	if _, exists := current[alias]; !exists {
		writeAPIError(w, http.StatusNotFound, "alias not found: "+alias)
		return
	}

	delete(current, alias)

	if err := writeColorAliasYAML(current); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "failed to persist: "+err.Error())
		return
	}

	templates.LoadAliases()

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

// writeColorAliasYAML serialises aliases back to the ansi-aliases.yaml file.
// The file format wraps aliases under a top-level "colors:" key to match the
// existing file structure that ansitags.LoadAliases expects.
func writeColorAliasYAML(aliases map[string]any) error {
	path := util.FilePath(string(configs.GetFilePathsConfig().DataFiles) + `/ansi-aliases.yaml`)

	// Normalise: the in-memory map may contain int or float64 values depending
	// on how they were loaded. Convert everything to int for clean YAML output.
	clean := make(map[string]int, len(aliases))
	for k, v := range aliases {
		switch val := v.(type) {
		case int:
			clean[k] = val
		case float64:
			clean[k] = int(val)
		}
	}

	doc := map[string]map[string]int{"colors": clean}

	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
