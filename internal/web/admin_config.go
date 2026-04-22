package web

import (
	"net/http"
	"text/template"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func adminConfig(w http.ResponseWriter, r *http.Request) {

	adminHtml := configs.GetFilePathsConfig().AdminHtml.String()

	tmpl, err := template.New("config.html").Funcs(funcMap).ParseFiles(
		adminHtml+"/_header.html",
		adminHtml+"/config.html",
		adminHtml+"/_footer.html",
	)
	if err != nil {
		mudlog.Error("HTML ERROR", "error", err)
		http.Error(w, "Error parsing template files", http.StatusInternalServerError)
		return
	}

	templateData := map[string]any{
		"CONFIG": configs.GetConfig(),
		"STATS":  GetStats(),
		"NAV":    buildAdminNav(),
	}

	w.Header().Set("Cache-Control", "no-store")
	if err := tmpl.Execute(w, templateData); err != nil {
		mudlog.Error("HTML ERROR", "action", "Execute", "error", err)
	}
}
