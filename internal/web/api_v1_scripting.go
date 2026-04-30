package web

import (
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
