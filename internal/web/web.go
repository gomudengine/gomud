package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/gorilla/websocket"
)

var (
	httpServer  *http.Server
	httpsServer *http.Server

	// internalMux is the single ServeMux used by both the live HTTP servers and
	// the in-process InternalRequest dispatcher. All routes must be registered
	// on this mux rather than http.DefaultServeMux.
	internalMux = http.NewServeMux()

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	httpRoot = ``

	// Used to interface with plugins and request web stuff
	webPlugins WebPlugin = nil

	// defaultRegistrar is the singleton that implements ModuleAdminRegistrar.
	defaultRegistrar = &moduleAdminRegistrarImpl{}
)

// WebNav is used for the public-facing navigation.
type WebNav struct {
	Name   string
	Target string
}

// WebNavItem represents a top-level admin nav entry, optionally with sub-items.
type WebNavItem struct {
	Name     string
	Target   string // primary href; empty if dropdown-only
	SubItems []WebNavSub
}

// WebNavSub is a single item inside a dropdown.
type WebNavSub struct {
	Label  string
	Target string
}

// ModuleAdminRegistrar is implemented by internal/web and provided to plugins
// via plugins.SetAdminRegistrar. This breaks the import cycle.
type ModuleAdminRegistrar interface {
	// RegisterAdminPage registers a module admin page.
	// htmlContent is the raw HTML read from the plugin's embedded FS.
	RegisterAdminPage(name, slug, htmlContent string, addToNav bool, navParent string, dataFunc func(*http.Request) map[string]any)
	// RegisterAdminAPIEndpoint registers a module API handler.
	// handler receives the request and returns (statusCode, success, data).
	RegisterAdminAPIEndpoint(method, slug string, handler func(*http.Request) (int, bool, any))
}

// moduleAdminRegistrarImpl holds module-contributed nav and API routes.
type moduleAdminRegistrarImpl struct {
	navItems []WebNavItem
}

// GetAdminRegistrar returns the ModuleAdminRegistrar that main.go passes to
// plugins.SetAdminRegistrar.
func GetAdminRegistrar() ModuleAdminRegistrar {
	return defaultRegistrar
}

// RegisterAdminPage registers a module admin page on internalMux and records
// its nav contribution.
func (reg *moduleAdminRegistrarImpl) RegisterAdminPage(
	name, slug, htmlContent string,
	addToNav bool,
	navParent string,
	dataFunc func(*http.Request) map[string]any,
) {
	path := "/admin/" + slug

	handler := func(w http.ResponseWriter, r *http.Request) {
		adminHtml := configs.GetFilePathsConfig().AdminHtml.String()

		tmpl, err := template.New(slug+".html").Funcs(funcMap).ParseFiles(
			adminHtml+"/_header.html",
			adminHtml+"/_footer.html",
		)
		if err != nil {
			mudlog.Error("HTML ERROR", "error", err)
			http.Error(w, "Error parsing template files", http.StatusInternalServerError)
			return
		}
		tmpl, err = tmpl.Parse(htmlContent)
		if err != nil {
			mudlog.Error("HTML ERROR", "error", err)
			http.Error(w, "Error parsing module html", http.StatusInternalServerError)
			return
		}

		templateData := map[string]any{
			"CONFIG": configs.GetConfig(),
			"STATS":  GetStats(),
			"NAV":    buildAdminNav(),
		}
		if dataFunc != nil {
			for k, v := range dataFunc(r) {
				if _, exists := templateData[k]; !exists {
					templateData[k] = v
				}
			}
		}

		w.Header().Set("Cache-Control", "no-store")
		if err := tmpl.Execute(w, templateData); err != nil {
			mudlog.Error("HTML ERROR", "action", "Execute", "error", err)
		}
	}

	internalMux.HandleFunc("GET "+path, RunWithMUDLocked(doBasicAuth(handler)))

	if !addToNav {
		return
	}

	if navParent == "" {
		// Top-level nav item with a single sub-item pointing to itself.
		reg.navItems = append(reg.navItems, WebNavItem{
			Name:   name,
			Target: path,
			SubItems: []WebNavSub{
				{Label: "View", Target: path},
			},
		})
		return
	}

	// Attach as sub-item to an existing nav entry.
	for i, item := range reg.navItems {
		if item.Name == navParent {
			reg.navItems[i].SubItems = append(reg.navItems[i].SubItems, WebNavSub{
				Label:  name,
				Target: path,
			})
			return
		}
	}
	// Parent not found yet — add as top-level with the sub-item.
	reg.navItems = append(reg.navItems, WebNavItem{
		Name:   navParent,
		Target: "",
		SubItems: []WebNavSub{
			{Label: name, Target: path},
		},
	})
}

// RegisterAdminAPIEndpoint registers a module API endpoint on internalMux.
func (reg *moduleAdminRegistrarImpl) RegisterAdminAPIEndpoint(
	method, slug string,
	handler func(*http.Request) (int, bool, any),
) {
	path := "/admin/api/v1/" + slug

	h := func(w http.ResponseWriter, r *http.Request) {
		status, success, data := handler(r)
		writeJSON(w, status, APIResponse[any]{Success: success, Data: data})
	}

	internalMux.HandleFunc(method+" "+path, RunWithMUDLocked(doBasicAuth(h)))
}

// buildAdminNav returns the full admin navigation, combining hardcoded core
// entries with module-contributed entries.
func buildAdminNav() []WebNavItem {
	nav := []WebNavItem{
		{
			Name:   "Dashboard",
			Target: "/admin/",
		},
		{
			Name:   "Config",
			Target: "/admin/config",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/config"},
				{Label: "API Docs", Target: "/admin/config-api"},
			},
		},
		{
			Name:   "Items",
			Target: "/admin/items",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/items"},
				{Label: "API Docs", Target: "/admin/items-api"},
			},
		},
		{
			Name:   "Buffs",
			Target: "/admin/buffs",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/buffs"},
				{Label: "API Docs", Target: "/admin/buffs-api"},
			},
		},
		{
			Name:   "Quests",
			Target: "/admin/quests",
			SubItems: []WebNavSub{
				{Label: "View / Edit", Target: "/admin/quests"},
				{Label: "API Docs", Target: "/admin/quests-api"},
			},
		},
	}
	nav = append(nav, defaultRegistrar.navItems...)
	return nav
}

type WebPlugin interface {
	NavLinks() map[string]string                                                    // Name=>Path pairs
	WebRequest(r *http.Request) (html string, templateData map[string]any, ok bool) // Get the first handler of a given request
}

func SetWebPlugin(wp WebPlugin) {
	webPlugins = wp
}

// serveTemplate searches for the requested file in the HTTP_ROOT,
// parses it as a template, and serves it.
func serveTemplate(w http.ResponseWriter, r *http.Request) {

	if httpRoot == "" {
		httpRoot = filepath.Clean(configs.GetFilePathsConfig().PublicHtml.String())
	}

	// Clean the path to prevent directory traversal.
	reqPath := filepath.Clean(r.URL.Path) // Example: / or /info/faq

	// Build the full file path.
	fullPath := filepath.Join(httpRoot, reqPath)

	// If the path is a directory, look for an index.html.
	info, err := os.Stat(fullPath)
	if err != nil {
		if filepath.Ext(fullPath) != ".html" {
			fullPath += ".html"
		}
	} else if info.IsDir() {
		fullPath = filepath.Join(fullPath, "index.html")
	}

	fileExt := filepath.Ext(fullPath)
	fileBase := filepath.Base(fullPath)

	// All template files to load from the filesystem
	templateFiles := []string{}

	var pageFound bool = true

	var pluginHtml string = ``
	var pluginTplData map[string]any = nil
	var ok bool = false
	var fSize int64 = 0
	var source string = `PublicHtml folder`

	// Check if the file exists, else 404
	fInfo, err := os.Stat(fullPath)
	if err != nil {
		pageFound = false
	}

	// Allow plugin to override request
	if webPlugins != nil {
		pluginHtml, pluginTplData, ok = webPlugins.WebRequest(r)
		fSize = int64(len([]byte(pluginHtml)))
		if ok {
			source = `module`
			pageFound = true
		}
	}

	if !pageFound || len(fileBase) > 0 && fileBase[0] == '_' {
		mudlog.Info("Web", "ip", r.RemoteAddr, "ref", r.Header.Get("Referer"), "file path", fullPath, "file extension", fileExt, "error", "Not found")

		fullPath = filepath.Join(httpRoot, `404.html`)
		fInfo, err = os.Stat(fullPath)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		fSize = fInfo.Size()

		w.WriteHeader(http.StatusNotFound)
	}

	// Log the request
	mudlog.Info("Web", "ip", r.RemoteAddr, "ref", r.Header.Get("Referer"), "file path", fullPath, "file extension", fileExt, "file source", source, "size", fmt.Sprintf(`%.2fk`, float64(fSize)/1024))

	// For non-HTML files, serve them statically.
	if fileExt != ".html" {
		http.ServeFile(w, r, fullPath)
		return
	}

	templateData := map[string]any{
		"REQUEST": r,
		"PATH":    reqPath,
		"CONFIG":  configs.GetConfig(),
		"STATS":   GetStats(),
		"NAV": []WebNav{
			{`Home`, `/`},
			{`Who's Online`, `/online`},
			{`Web Client`, `/webclient`},
			{`See Configuration`, `/viewconfig`},
		},
	}

	// Copy any plugin navigation
	if webPlugins != nil {

		currentNav := templateData[`NAV`].([]WebNav)

		for name, path := range webPlugins.NavLinks() {

			found := false
			for i := len(currentNav) - 1; i >= 0; i-- {

				if currentNav[i].Name == name {
					found = true
					if path == `` {
						currentNav = append(currentNav[:i], currentNav[i+1:]...)
					} else {
						currentNav[i].Target = path
					}
					break
				}

			}

			if !found {
				currentNav = append(currentNav, WebNav{name, path})
			}
		}

		templateData[`NAV`] = currentNav
	}

	// Copy over any plugin data loaded.
	for name, value := range pluginTplData {
		// Don't allow overwriting defaults
		if _, ok := templateData[name]; !ok {
			templateData[name] = value
		}
	}

	// Parse special files intended to be used as template includes
	globFiles, err := filepath.Glob(filepath.Join(httpRoot, "_*.html"))
	if err == nil {
		templateFiles = append(templateFiles, globFiles...)
	}

	// Parse special files intended to be used as template includes (from the request folder)
	requestDir := filepath.Dir(fullPath)
	if httpRoot != requestDir {
		globFiles, err = filepath.Glob(filepath.Join(requestDir, "_*.html"))
		if err == nil {
			templateFiles = append(templateFiles, globFiles...)
		}
	}

	// Add the final (actual) file

	// Parse
	tmpl := template.New(filepath.Base(fullPath)).Funcs(funcMap)

	if pluginHtml == `` {
		templateFiles = append(templateFiles, fullPath)

	}

	tmpl, err = tmpl.ParseFiles(templateFiles...)
	if err != nil {
		mudlog.Error("HTML ERROR", "action", "ParseFiles", "error", err)
		http.Error(w, "Error parsing template files", http.StatusInternalServerError)
	}

	if pluginHtml != `` {
		tmpl, err = tmpl.Parse(pluginHtml)
		if err != nil {
			mudlog.Error("HTML ERROR", "action", "Parse", "error", err)
			http.Error(w, "Error parsing plugin html", http.StatusInternalServerError)
		}
	}

	// Execute the template and write it to the response.
	w.Header().Set("Cache-Control", "no-store")
	if err := tmpl.Execute(w, templateData); err != nil {
		mudlog.Error("HTML ERROR", "action", "Execute", "error", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func Listen(wg *sync.WaitGroup, webSocketHandler func(*websocket.Conn)) {

	networkConfig := configs.GetNetworkConfig()

	if networkConfig.HttpPort == 0 && networkConfig.HttpsPort == 0 {
		slog.Error(`Web`, "error", "No ports defined. No web server will be started.")
		return
	}

	// Routing
	// Basic homepage

	internalMux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = `/static/images/favicon.ico`
		serveTemplate(w, r)
	})

	internalMux.HandleFunc("/", serveTemplate)

	// websocket upgrade
	internalMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		webSocketHandler(conn)
	})

	registerAdminRoutes(internalMux)

	//
	// Https server start up
	//

	if networkConfig.HttpsPort > 0 {

		filePaths := configs.GetFilePathsConfig()

		if len(filePaths.HttpsCertFile) == 0 || len(filePaths.HttpsKeyFile) == 0 {

			mudlog.Info("HTTPS", "stage", "skipping", "error", "Undefined public/private key files", "Public Cert", filePaths.HttpsCertFile, "Private Key", filePaths.HttpsKeyFile)

		} else {

			if filePaths.HttpsCertFile != `` && filePaths.HttpsKeyFile != `` {

				mudlog.Info("HTTPS", "stage", "Validating public/private key pair", "Public Cert", filePaths.HttpsCertFile, "Private Key", filePaths.HttpsKeyFile)

				cert, err := tls.LoadX509KeyPair(string(filePaths.HttpsCertFile), string(filePaths.HttpsKeyFile))

				if err != nil {

					mudlog.Error("HTTPS", "error", fmt.Errorf("Error loading certificate and key: %w", err))

				} else {

					tlsConfig := &tls.Config{
						Certificates: []tls.Certificate{cert},
					}

					wg.Add(1)

					httpsServer = &http.Server{
						Addr:      fmt.Sprintf(`:%d`, networkConfig.HttpsPort),
						TLSConfig: tlsConfig,
						Handler:   internalMux,
					}

					mudlog.Info("HTTPS", "stage", "Starting https server", "port", networkConfig.HttpsPort)
					go func() {
						defer wg.Done()
						if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
							mudlog.Error("HTTPS", "error", fmt.Errorf("Error starting HTTPS web server: %w", err))
						}
					}()
				}
			}
		}
	}

	//
	// Http server start up
	//

	if networkConfig.HttpPort > 0 {

		httpServer = &http.Server{
			Addr:    fmt.Sprintf(`:%d`, networkConfig.HttpPort),
			Handler: internalMux,
		}

		if networkConfig.HttpsRedirect {

			if httpsServer == nil {

				mudlog.Error("HTTP", "error", "Cannot enable https redirect. There is no https server configured/running.")

			} else {

				var redirectHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {

					target := buildHTTPSRedirectTarget(r.Host, int(networkConfig.HttpsPort), r.RequestURI)

					http.Redirect(w, r, target, http.StatusMovedPermanently)
				}

				httpServer.Handler = redirectHandler

			}

		}

		// HTTP Server
		wg.Add(1)

		mudlog.Info("HTTP", "stage", "Starting http server", "port", networkConfig.HttpPort)
		go func() {
			defer wg.Done()

			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				mudlog.Error("HTTP", "error", fmt.Errorf("Error starting web server: %w", err))
			}
		}()
	}

}

func buildHTTPSRedirectTarget(host string, httpsPort int, requestURI string) string {
	if strings.Contains(host, ":") {
		if splitHost, _, err := net.SplitHostPort(host); err == nil {
			host = splitHost
		}
	}

	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}

	if httpsPort == 443 {
		return fmt.Sprintf("https://%s%s", host, requestURI)
	}

	return fmt.Sprintf("https://%s:%d%s", host, httpsPort, requestURI)
}

// RunWithMUDLocked wraps a handler with the game mutex. Internal requests
// (dispatched via InternalRequest) skip locking because the caller is
// responsible for holding the lock when required.
func RunWithMUDLocked(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsInternalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}
		util.LockMud()
		defer util.UnlockMud()
		next.ServeHTTP(w, r)
	})
}

func Shutdown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		if err := httpServer.Shutdown(ctx); err != nil {
			mudlog.Error("HTTP", "error", fmt.Errorf("HTTP server shutdown failed: %w", err))
		} else {
			mudlog.Info("HTTP", "stage", "stopped")
		}
	}

	if httpsServer != nil {
		if err := httpsServer.Shutdown(ctx); err != nil {
			mudlog.Error("HTTPS", "error", fmt.Errorf("HTTPS server shutdown failed: %w", err))
		} else {
			mudlog.Info("HTTPS", "stage", "stopped")
		}
	}
}

// serveAdminStaticFile serves static assets from the admin HTML directory.
// The full URL path relative to /admin/ is preserved so subdirectories work.
func serveAdminStaticFile(w http.ResponseWriter, r *http.Request) {
	adminRoot := filepath.Clean(configs.GetFilePathsConfig().AdminHtml.String())
	rel := strings.TrimPrefix(r.URL.Path, "/admin")
	fullPath := filepath.Join(adminRoot, filepath.Clean(rel))
	http.ServeFile(w, r, fullPath)
}

func sendError(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
