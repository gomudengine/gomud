package web

import "net/http"

func adminAudio(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "audio.html", nil)
}

func adminAudioAPI(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "audio-api.html", nil)
}
