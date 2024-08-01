//go:build !tinygo

package thing2

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"sync"
)

func (d *Device) Handle(pattern string, handler http.Handler) {
	d.ServeMux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.Lock()
		defer d.Unlock()
		handler.ServeHTTP(w, r)
	}))
}

func (d *Device) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	d.ServeMux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.Lock()
		defer d.Unlock()
		handlerFunc(w, r)
	}))
}

func (d *Device) RHandle(pattern string, handler http.Handler) {
	d.ServeMux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		handler.ServeHTTP(w, r)
	}))
}

func (d *Device) RHandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	d.ServeMux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.RLock()
		defer d.RUnlock()
		handlerFunc(w, r)
	}))
}

// Install /device/{id} pattern for device in default ServeMux
func (d *Device) InstallDevicePattern() {
	prefix := "/device/" + d.Id
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	fmt.Printf("InstallDevicePattern %s\n", prefix)
}

var modelPatterns = make(map[string]string)
var modelPatternsMu sync.RWMutex

// Install /model/{model} pattern for device in default ServeMux
func (d *Device) InstallModelPattern() {
	modelPatternsMu.Lock()
	defer modelPatternsMu.Unlock()
	// But only if it doesn't already exist
	if _, exists := modelPatterns[d.Model]; exists {
		return
	}
	prefix := "/model/" + d.Model
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	modelPatterns[d.Model] = prefix
	fmt.Printf("InstallModelPattern %s\n", prefix)
}

func (d *Device) API() {
	d.RHandle("/", http.FileServer(http.FS(d.LayeredFS)))
	d.RHandle("/{$}", d.showIndex())
	d.RHandle("/keepalive", d.keepAlive())
	d.RHandle("/full", d.showFull())
	d.RHandle("/info", d.showInfo())
	d.RHandle("/state", d.showState())
	d.RHandle("/code", d.showCode())
	d.RHandle("/download", d.showDownload())
	d.RHandle("/download-target", d.showDownloadTarget())
	d.RHandle("/download-instructions", d.showDownloadInstructions())
	// d.RHandleFunc("/deploy", d.deploy)
}

func (d *Device) render(w io.Writer, name string, data any) error {
	tmpl := d.templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", name)
	}
	return tmpl.Execute(w, data)
}

type pageVars map[string]any

type pageData struct {
	Vars   pageVars
	Device any
}

func (d *Device) renderPage(w io.Writer, name string, pageVars pageVars) error {
	return d.render(w, name, &pageData{
		Vars:   pageVars,
		Device: d.data,
	})
}

func (d *Device) Render(w io.Writer, view string) error {
	//d.RLock()
	//defer d.RUnlock()
	switch view {
	case "full":
		return d._showFull(w)
	}
	return nil
}

func (d *Device) TemplateShow(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.renderPage(w, name, pageVars{})
	})
}

func (d *Device) showIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.renderPage(w, "index.tmpl", pageVars{
			"view":      "full",
			"sessionId": newSession(),
		})
	})
}

func (d *Device) keepAlive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session-id")
		sessionUpdate(sessionId)
	})
}

func (d *Device) _showFull(w io.Writer) error {
	return d.renderPage(w, "device-full.tmpl", pageVars{
		"view": "full",
	})
}

func (d *Device) showFull() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session-id")
		sessionDeviceSaveView(sessionId, d, "full")
		if err := d._showFull(w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
}

func (d *Device) showInfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.renderPage(w, "device-info.tmpl", pageVars{
			"view":       "info",
			"modulePath": d.modulePath(),
		})
	})
}

func (d *Device) showState() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, _ := json.MarshalIndent(d.data, "", "\t")
		d.render(w, "state.tmpl", string(state))
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
		d.render(w, "code.tmpl", names)
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
		target := d.currentTarget(r.URL.Query())
		template := "instructions-" + target + ".tmpl"
		d.renderPage(w, template, pageVars{})
	})
}
