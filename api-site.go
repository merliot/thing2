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

func (d *Device) showSitePage(w http.ResponseWriter, r *http.Request,
	page string, tabs siteTabs, createSession bool) {
	var sessionId string
	var ok bool

	if createSession {
		sessionId, ok = newSession()
		if !ok {
			http.Error(w, "no more sessions", http.StatusTooManyRequests)
			return
		}
	}

	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"sessionId": sessionId,
		"page":      page,
		"tabs":      tabs,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showSiteHome(w http.ResponseWriter, r *http.Request) {
	d.showSitePage(w, r, "home", tabsHome, false)
}

func (d *Device) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	d.showSitePage(w, r, "demo", tabsDemo, true)
}

func (d *Device) showSiteStatus(w http.ResponseWriter, r *http.Request) {
	d.showSitePage(w, r, "status", tabsStatus, true)
}

func (d *Device) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	d.showSitePage(w, r, "docs", tabsDocs, false)
}
