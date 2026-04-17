package web

import (
	"net/http"
)

func registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/", RunWithMUDLocked(
		doBasicAuth(adminIndex),
	))
	mux.HandleFunc("GET /admin/https/", RunWithMUDLocked(
		doBasicAuth(httpsIndex),
	))

	registerAdminAPIRoutes(mux)
}
