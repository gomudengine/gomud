package web

import "net/http"

func adminConfigWizard(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "config-wizard.html", nil)
}
