package web

import (
	"text/template"

	"net/http"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func httpsIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("https.html").Funcs(funcMap).ParseFiles(
		configs.GetFilePathsConfig().AdminHtml.String()+"/_header.html",
		configs.GetFilePathsConfig().AdminHtml.String()+"/https.html",
		configs.GetFilePathsConfig().AdminHtml.String()+"/_footer.html",
	)
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
		http.Error(w, "Error parsing template files", http.StatusInternalServerError)
		return
	}

	tplData := map[string]any{
		"NAV":         buildAdminNav(),
		"httpsStatus": GetHTTPSStatus(),
	}

	w.Header().Set("Cache-Control", "no-store")
	if err := tmpl.Execute(w, tplData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}
}
