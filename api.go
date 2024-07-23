package thing2

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
)

func (d *Device) API() {
	d.Handle("/", http.FileServer(http.FS(d.LayeredFS)))
	d.Handle("/{$}", d.showIndex())
	d.Handle("/full", d.showView("full", "device-full.tmpl"))
	d.Handle("/info", d.showInfo())
	d.Handle("/state", d.showState())
	d.Handle("/code", d.showCode())
	d.Handle("/download", d.showDownload())
	d.Handle("/download-target", d.showDownloadTarget())
	d.Handle("/download-instructions", d.showDownloadInstructions())
	// d.HandleFunc("/deploy", d.deploy)
}

func (d *Device) render(w io.Writer, name string, data any) error {
	tmpl := d.templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", name)
	}
	return tmpl.Execute(w, data)
}

func (d *Device) renderTemplateData(w http.ResponseWriter, name string, data any) {
	fmt.Printf("renderTemplateData %#v\n", data)
	if err := d.render(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

type pageVars map[string]any

type pageData struct {
	Vars   pageVars
	Device any
}

func (d *Device) renderPage(w http.ResponseWriter, name string, pageVars pageVars) {
	d.renderTemplateData(w, name, &pageData{
		Vars:   pageVars,
		Device: d.data,
	})
}

func (d *Device) TemplateShow(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		d.renderPage(w, name, pageVars{})
	})
}

func (d *Device) showIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		d.renderPage(w, "index.tmpl", pageVars{
			"view": "full",
		})
	})
}

func (d *Device) showView(view, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		d.renderPage(w, name, pageVars{
			"view": view,
		})
	})
}

func (d *Device) showInfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		d.renderPage(w, "device-info.tmpl", pageVars{
			"view":       "info",
			"modulePath": d.modulePath(),
		})
	})
}

func (d *Device) showState() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		state, _ := json.MarshalIndent(d.data, "", "\t")
		d.renderTemplateData(w, "state.tmpl", string(state))
	})
}

func (d *Device) showCode() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve top-level entries
		entries, _ := fs.ReadDir(d.LayeredFS, ".")
		// Collect entry names
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		d.renderTemplateData(w, "code.tmpl", names)
	})
}

func linuxTarget(target string) bool {
	return target == "demo" || target == "x86-64" || target == "rpi"
}

func (d *Device) DeployValues() url.Values {
	values, err := url.ParseQuery(d.DeployParams)
	if err != nil {
		panic(err.Error())
	}
	return values
}

func firstValue(values url.Values, key string) string {
	if v, ok := values[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func (d *Device) currentTarget(params url.Values) string {
	target := firstValue(params, "target")
	if target == "" {
		target = firstValue(d.DeployValues(), "target")
	}
	return target
}

func (d *Device) showDownload() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		values := d.DeployValues()
		target := firstValue(values, "target")
		d.renderPage(w, "device-download.tmpl", pageVars{
			"view":        "download",
			"target":      target,
			"linuxTarget": linuxTarget(target),
			"port":        firstValue(values, "port"),
		})
	})
}

func (d *Device) showDownloadTarget() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		values := d.DeployValues()
		target := d.currentTarget(r.URL.Query())
		d.renderPage(w, "device-download-target.tmpl", pageVars{
			"linuxTarget": linuxTarget(target),
			"missingWifi": len(d.WifiAuth) == 0,
			"ssid":        firstValue(values, "ssid"),
			"port":        firstValue(values, "port"),
		})
	})
}

func (d *Device) showDownloadInstructions() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		target := d.currentTarget(r.URL.Query())
		template := "instructions-" + target + ".tmpl"
		d.renderPage(w, template, pageVars{})
	})
}
