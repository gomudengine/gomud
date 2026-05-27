package web

import "net/http"

func adminConfig(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "config.html", nil)
}
