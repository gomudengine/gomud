package web

import (
	"net/http"
)

func registerAdminAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/api/v1/config", RunWithMUDLocked(
		doBasicAuth(apiV1GetConfig),
	))
	mux.HandleFunc("PATCH /admin/api/v1/config", RunWithMUDLocked(
		doBasicAuth(apiV1PatchConfig),
	))
}
