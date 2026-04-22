package plugins

import "net/http"

// ModuleAPIHandler is the signature for module-provided admin API handlers.
// The handler receives the raw *http.Request and returns:
//
//	status  - HTTP status code (e.g. 200, 400, 500)
//	success - written into APIResponse.Success
//	data    - written into APIResponse.Data (must be JSON-serialisable)
type ModuleAPIHandler func(r *http.Request) (status int, success bool, data any)

// AdminWebPage describes an admin page contributed by a module.
type AdminWebPage struct {
	Name         string
	Slug         string // e.g. "mudmail" -> /admin/mudmail
	Path         string // "/admin/mudmail"
	HTMLFile     string // path inside plugin FS, relative to datafiles/html/admin/
	AddToNav     bool
	NavParent    string // if non-empty, nest under this parent nav entry name
	DataFunction func(*http.Request) map[string]any
}

// AdminAPIRoute describes an admin API endpoint contributed by a module.
type AdminAPIRoute struct {
	Method  string
	Slug    string
	Path    string // "/admin/api/v1/<slug>"
	Handler ModuleAPIHandler
}

type WebConfig struct {
	navLinks       map[string]string  // name=>path
	pages          map[string]WebPage // path=>WebPage
	adminPages     []AdminWebPage
	adminAPIRoutes []AdminAPIRoute
}

type WebPage struct {
	Name         string
	Path         string
	Filepath     string
	DataFunction func(r *http.Request) map[string]any
}

func newWebConfig() WebConfig {
	return WebConfig{
		navLinks: map[string]string{},
		pages:    map[string]WebPage{},
	}
}

func (w *WebConfig) NavLink(name string, path string) {
	w.navLinks[name] = path
}

func (w *WebConfig) WebPage(name string, path string, file string, addToNav bool, dataFunc func(r *http.Request) map[string]any) {
	if addToNav {
		w.NavLink(name, path)
	}
	w.pages[path] = WebPage{
		Name:         name,
		Path:         path,
		Filepath:     file,
		DataFunction: dataFunc,
	}
}

// AdminPage registers an admin-only page served under /admin/<slug>.
//
//   - name      - display label used in nav
//   - slug      - URL path segment, e.g. "mudmail" -> /admin/mudmail
//   - htmlFile  - path inside the plugin's embedded FS, relative to datafiles/html/admin/
//   - addToNav  - whether to add a nav entry
//   - navParent - if non-empty, adds this page as a sub-item under the named parent nav entry
//   - dataFunc  - optional function to supply extra template data; receives *http.Request
func (w *WebConfig) AdminPage(name, slug, htmlFile string, addToNav bool, navParent string, dataFunc func(*http.Request) map[string]any) {
	w.adminPages = append(w.adminPages, AdminWebPage{
		Name:         name,
		Slug:         slug,
		Path:         "/admin/" + slug,
		HTMLFile:     htmlFile,
		AddToNav:     addToNav,
		NavParent:    navParent,
		DataFunction: dataFunc,
	})
}

// AdminAPIEndpoint registers an HTTP handler under /admin/api/v1/<slug>.
//
//   - method  - HTTP method string: "GET", "POST", "PATCH", "DELETE", etc.
//   - slug    - path suffix, e.g. "mudmail" -> /admin/api/v1/mudmail
//   - handler - the handler function
func (w *WebConfig) AdminAPIEndpoint(method, slug string, handler ModuleAPIHandler) {
	w.adminAPIRoutes = append(w.adminAPIRoutes, AdminAPIRoute{
		Method:  method,
		Slug:    slug,
		Path:    "/admin/api/v1/" + slug,
		Handler: handler,
	})
}
