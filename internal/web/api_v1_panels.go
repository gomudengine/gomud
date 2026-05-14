package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/templates"
)

type panelEntry struct {
	Name string `json:"name"`
	YAML string `json:"yaml"`
}

// GET /admin/api/v1/panels
// Returns an array of all panel layout files and their YAML content.
func apiV1GetPanels(w http.ResponseWriter, r *http.Request) {
	layouts, err := templates.ListPanelLayouts()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]panelEntry, len(layouts))
	for i, l := range layouts {
		result[i] = panelEntry{Name: l.Name, YAML: l.YAML}
	}

	writeJSON(w, http.StatusOK, APIResponse[[]panelEntry]{
		Success: true,
		Data:    result,
	})
}

// POST /admin/api/v1/panels/validate
// Accepts a panel layout YAML body and returns a list of structural issues.
// An empty issues array means the layout is valid.
func apiV1PostPanelValidate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		YAML string `json:"yaml"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if strings.TrimSpace(body.YAML) == "" {
		writeAPIError(w, http.StatusBadRequest, "yaml is required")
		return
	}

	issues, err := templates.ValidatePanelLayout(body.YAML)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[map[string]any]{
		Success: true,
		Data: map[string]any{
			"valid":  len(issues) == 0,
			"issues": issues,
		},
	})
}

// POST /admin/api/v1/panels/preview
// Accepts a panel layout YAML body and returns two plain-text previews:
// "preview" has ANSI tags stripped; "preview_ansi" retains them verbatim.
func apiV1PostPanelPreview(w http.ResponseWriter, r *http.Request) {
	var body struct {
		YAML string `json:"yaml"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if strings.TrimSpace(body.YAML) == "" {
		writeAPIError(w, http.StatusBadRequest, "yaml is required")
		return
	}

	plain, err := templates.PreviewPanelLayout(body.YAML, true)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	ansi, err := templates.PreviewPanelLayout(body.YAML, false)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[map[string]string]{
		Success: true,
		Data: map[string]string{
			"preview":      plain,
			"preview_ansi": ansi,
		},
	})
}

// PUT /admin/api/v1/panels/{panel-name}
// Overwrites an existing panel layout file.
// panel-name uses slashes encoded as %2F (or the path segment may contain
// a wildcard catch-all so the full sub-path is preserved).
func apiV1PutPanel(w http.ResponseWriter, r *http.Request) {
	// The route is registered as /admin/api/v1/panels/{panel-name...} so the
	// wildcard value already contains the full relative path.
	name := r.PathValue("panelname")
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "panel name is required")
		return
	}

	var body struct {
		YAML string `json:"yaml"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if strings.TrimSpace(body.YAML) == "" {
		writeAPIError(w, http.StatusBadRequest, "yaml is required")
		return
	}

	if err := templates.SavePanelLayout(name, body.YAML); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
