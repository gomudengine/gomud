package web

import "net/http"

func adminStats(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "stats.html", nil)
}

func adminStatsAPI(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "stats-api.html", nil)
}
