package web

import (
	"net/http"
)

func registerAdminRoutes(mux *http.ServeMux) {
	// Static assets (non-HTML files) under /admin/ require authentication.
	mux.HandleFunc("GET /admin/{file}", doBasicAuth(serveAdminStaticFile))

	mux.HandleFunc("GET /admin/", RunWithMUDLocked(
		doBasicAuth(adminIndex),
	))

	mux.HandleFunc("GET /admin/config", RunWithMUDLocked(
		doBasicAuth(adminConfig),
	))

	mux.HandleFunc("GET /admin/config-api", RunWithMUDLocked(
		doBasicAuth(adminConfigAPI),
	))

	registerAdminAPIRoutes(mux)
}
