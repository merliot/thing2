//go:build !tinygo

package thing2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"runtime"
)

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

// install installs /device/{id} pattern for device in default ServeMux
func (d *Device) deviceInstall() {
	prefix := "/device/" + d.Id
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	fmt.Println("deviceInstall", prefix)
}

func devicesInstall() {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	for _, device := range devices {
		device.deviceInstall()
	}
}

func (d *Device) API() {
	d.RHandle("/", http.FileServer(http.FS(d.layeredFS)))
	d.RHandle("/{$}", d.showIndex())
	d.RHandle("/keepalive", d.keepAlive())
	d.RHandle("/full", d.showFull())
	d.RHandle("/tile", d.showTile())
	d.RHandle("/detail", d.showDetail())
	d.RHandle("/info", d.showInfo())
	d.RHandle("/state", d.showState())
	d.RHandle("/code", d.showCode())
	d.RHandle("/download", d.showDownload())
	d.RHandle("/download-target", d.showDownloadTarget())
	d.RHandle("/download-instructions", d.showDownloadInstructions())
	// d.RHandleFunc("/deploy", d.deploy)
}

func (d *Device) renderTemplate(w io.Writer, name string, data any) error {
	tmpl := d.templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", name)
	}
	return tmpl.Execute(w, data)
}

type pageVars map[string]any

type pageData struct {
	Vars   pageVars
	Device *Device
	State  any
}

func (d *Device) renderPage(w io.Writer, name string, pageVars pageVars) error {
	return d.renderTemplate(w, name, &pageData{
		Vars:   pageVars,
		Device: d,
		State:  d.GetState(),
	})
}

func dumpStackTrace() {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	log.Printf("Stack trace:\n%s", buf)
}

func (d *Device) RenderChildHTML(sessionId, childId, view string) (template.HTML, error) {

	println("RenderChildHTML", sessionId, childId, view)
	//dumpStackTrace()

	child, ok := devices[childId]
	if !ok {
		return template.HTML(""), deviceNotFound(childId)
	}

	_sessionDeviceSaveView(sessionId, childId, view)

	child.RLock()
	defer child.RUnlock()

	var buf bytes.Buffer
	if err := child._render(sessionId, &buf, view); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) showIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := newSession()
		sessionDeviceSaveView(sessionId, d.Id, "full")
		d.renderPage(w, "index.tmpl", pageVars{
			"sessionId": sessionId,
			"view":      "full",
			"models":    makers,
		})
	})
}

func (d *Device) keepAlive() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session-id")
		sessionUpdate(sessionId)
	})
}

func (d *Device) _showFull(sessionId string, w io.Writer) error {
	return d.renderPage(w, "device-full.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "full",
	})
}

func (d *Device) showFull() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session-id")
		sessionDeviceSaveView(sessionId, d.Id, "full")
		if err := d._showFull(sessionId, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
}

func (d *Device) _showTile(sessionId string, w io.Writer) error {
	return d.renderPage(w, "device-tile.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "tile",
	})
}

func (d *Device) showTile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session-id")
		sessionDeviceSaveView(sessionId, d.Id, "tile")
		if err := d._showTile(sessionId, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
}

func (d *Device) _showDetail(sessionId, childId string, w io.Writer) error {
	return d.renderPage(w, "device-full-detail.tmpl", pageVars{
		"sessionId": sessionId,
		"childId":   childId,
		"view":      "full",
	})
}

func (d *Device) showDetail() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		childId := r.URL.Query().Get("childId")
		sessionId := r.Header.Get("session-id")
		sessionDeviceSaveView(sessionId, childId, "full")
		if err := d._showDetail(sessionId, childId, w); err != nil {
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
		state, _ := json.MarshalIndent(d.GetState(), "", "\t")
		d.renderTemplate(w, "state.tmpl", string(state))
	})
}

func (d *Device) showCode() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve top-level entries
		entries, _ := fs.ReadDir(d.layeredFS, ".")
		// Collect entry names
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		d.renderTemplate(w, "code.tmpl", names)
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
			"missingWifi": len(wifiAuths) == 0,
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
