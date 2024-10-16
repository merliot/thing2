//go:build !tinygo

package thing2

import "net/http"

type siteTab struct {
	Name string
	Href string
}

type siteTabs []siteTab

var (
	tabHome    = siteTab{"HOME", "/home"}
	tabDemo    = siteTab{"DEMO", "/demo"}
	tabStatus  = siteTab{"STATUS", "/status"}
	tabDocs    = siteTab{"DOCS", "/docs"}
	tabSource  = siteTab{"SOURCE", "https://github.com/merliot/thing2"}
	tabsHome   = siteTabs{tabHome, tabDemo, tabStatus, tabDocs, tabSource}
	tabsDemo   = siteTabs{tabDemo, tabHome, tabStatus, tabDocs, tabSource}
	tabsStatus = siteTabs{tabStatus, tabHome, tabDemo, tabDocs, tabSource}
	tabsDocs   = siteTabs{tabDocs, tabHome, tabDemo, tabStatus, tabSource}
)

func (d *device) showSiteHome(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "home",
		"tabs":    tabsHome,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	sessionId, ok := newSession()
	if !ok {
		d.noSessions(w, r)
		return
	}
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"sessionId": sessionId,
		"section":   "demo",
		"tabs":      tabsDemo,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteStatus(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "status",
		"tabs":    tabsStatus,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" {
		page = "intro"
	}
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "docs",
		"tabs":    tabsDocs,
		"pages":   docPages,
		"page":    page,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
