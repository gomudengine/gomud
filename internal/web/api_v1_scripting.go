package web

import (
	"encoding/json"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/scripting"
)

// GET /admin/api/v1/scripting/functions
func apiV1GetScriptFunctions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[*scripting.ScriptFunctionsSchema]{
		Success: true,
		Data:    scripting.GetScriptFunctionsSchema(),
	})
}

// POST /admin/api/v1/scripting/validate
func apiV1ValidateScript(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Script string `json:"script"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	if len(req.Script) == 0 {
		writeAPIError(w, http.StatusBadRequest, "Script body is empty.")
		return
	}

	result := scripting.ValidateScript("script", req.Script)

	status := http.StatusOK
	if !result.Valid {
		status = http.StatusUnprocessableEntity
	}

	writeJSON(w, status, APIResponse[scripting.ValidationResult]{
		Success: result.Valid,
		Data:    result,
	})
}
