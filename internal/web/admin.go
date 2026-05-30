package web

import (
	"net/http"
	"strings"
	"text/template"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func adminIndex(w http.ResponseWriter, r *http.Request) {

	adminHtml := configs.GetFilePathsConfig().AdminHtml.String()

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(
		adminHtml+"/_header.html",
		adminHtml+"/_nav.html",
		adminHtml+"/index.html",
		adminHtml+"/_footer.html",
	)
	if err != nil {
		mudlog.Error("HTML ERROR", "error", err)
		http.Error(w, "Error parsing template files", http.StatusInternalServerError)
		return
	}

	writeKey := pageWritePermissions[strings.TrimRight(r.URL.Path, "/")]
	templateData := map[string]any{
		"CONFIG":           configs.GetConfig(),
		"STATS":            GetStats(),
		"NAV":              buildAdminNav(),
		"AUTHED_USER":      GetAuthedUser(r),
		"WRITE_PERMISSION": writeKey,
		"READ_ONLY":        pageReadOnly(r),
	}

	if r.URL.Query().Get(`login`) != `` {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		http.Redirect(w, r, scheme+"://"+r.Host+"/admin/", http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	if err := tmpl.Execute(w, templateData); err != nil {
		mudlog.Error("HTML ERROR", "action", "Execute", "error", err)
	}
}
