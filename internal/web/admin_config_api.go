package web

import "net/http"

func adminConfigAPI(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "config-api.html", nil)
}
