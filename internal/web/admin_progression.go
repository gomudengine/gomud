package web

import "net/http"

func adminProgression(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "progression.html", nil)
}
