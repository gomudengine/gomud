package web

import "net/http"

func adminPanels(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "panels.html", nil)
}

func adminPanelsAPI(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "panels-api.html", nil)
}
