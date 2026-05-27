package web

import "net/http"

func adminScriptingAPI(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "scripting-api.html", nil)
}
