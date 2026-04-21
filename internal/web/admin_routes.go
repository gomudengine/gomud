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
	mux.HandleFunc("GET /admin/https/", RunWithMUDLocked(
		doBasicAuth(httpsIndex),
	))

	mux.HandleFunc("GET /admin/config", RunWithMUDLocked(
		doBasicAuth(adminConfig),
	))
	mux.HandleFunc("GET /admin/config-api", RunWithMUDLocked(
		doBasicAuth(adminConfigAPI),
	))

	mux.HandleFunc("GET /admin/items", RunWithMUDLocked(
		doBasicAuth(adminItems),
	))
	mux.HandleFunc("GET /admin/items-api", RunWithMUDLocked(
		doBasicAuth(adminItemsAPI),
	))

	mux.HandleFunc("GET /admin/buffs", RunWithMUDLocked(
		doBasicAuth(adminBuffs),
	))
	mux.HandleFunc("GET /admin/buffs-api", RunWithMUDLocked(
		doBasicAuth(adminBuffsAPI),
	))

	mux.HandleFunc("GET /admin/quests", RunWithMUDLocked(
		doBasicAuth(adminQuests),
	))
	mux.HandleFunc("GET /admin/quests-api", RunWithMUDLocked(
		doBasicAuth(adminQuestsAPI),
	))

	registerAdminAPIRoutes(mux)
}
