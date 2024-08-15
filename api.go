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
	"strings"
)

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

func (d *Device) serveFile(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
		//w.Header().Set("Content-Type", "application/javascript")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *Device) API() {
	d.HandleFunc("GET /", d.serveFile)
	d.HandleFunc("GET /{$}", d.showIndex)
	d.HandleFunc("PUT /keepalive", d.keepAlive)
	d.HandleFunc("GET /full", d.showFull)
	d.HandleFunc("GET /tile", d.showTile)
	d.HandleFunc("GET /list", d.showList)
	d.HandleFunc("GET /detail", d.showDetail)
	d.HandleFunc("GET /info", d.showInfo)
	d.HandleFunc("GET /code", d.showCode)
	d.HandleFunc("GET /download", d.showDownload)
	d.HandleFunc("GET /download-target", d.showDownloadTarget)
	d.HandleFunc("GET /download-instructions", d.showDownloadInstructions)
	// d.RHandleFunc("/deploy", d.deploy)
	d.HandleFunc("GET /create", d.createChild)
	d.HandleFunc("DELETE /destroy", d.destroyChild)
	d.HandleFunc("GET /newModal", d.showNewModal)
	d.HandleFunc("GET /selectModel", d.selectModel)
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

func (d *Device) showIndex(w http.ResponseWriter, r *http.Request) {
	sessionId := newSession()
	sessionDeviceSaveView(sessionId, d.Id, "full")
	d.renderPage(w, "index.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "full",
		"models":    makers,
	})
}

func (d *Device) keepAlive(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionUpdate(sessionId)
}

func (d *Device) _showFull(sessionId string, w io.Writer) error {
	return d.renderPage(w, "device-full.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "full",
	})
}

func (d *Device) showFull(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSaveView(sessionId, d.Id, "full")
	if err := d._showFull(sessionId, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) _showList(sessionId string, w io.Writer) error {
	return d.renderPage(w, "device-list.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "list",
	})
}

func (d *Device) showList(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSaveView(sessionId, d.Id, "list")
	if err := d._showList(sessionId, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) _showTile(sessionId string, w io.Writer) error {
	return d.renderPage(w, "device-tile.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "tile",
	})
}

func (d *Device) showTile(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSaveView(sessionId, d.Id, "tile")
	if err := d._showTile(sessionId, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) _showDetail(sessionId, childId, prevView string, w io.Writer) error {
	return d.renderPage(w, "device-detail.tmpl", pageVars{
		"sessionId": sessionId,
		"childId":   childId,
		"view":      "detail",
		"prevView":  prevView,
	})
}

func (d *Device) showDetail(w http.ResponseWriter, r *http.Request) {
	childId := r.URL.Query().Get("childId")
	prevView := r.URL.Query().Get("prevView")
	sessionId := r.Header.Get("session-id")
	sessionDeviceSaveView(sessionId, childId, "detail")
	if err := d._showDetail(sessionId, childId, prevView, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showInfo(w http.ResponseWriter, r *http.Request) {
	d.renderPage(w, "device-info.tmpl", pageVars{
		"view":       "info",
		"modulePath": d.modulePath(),
	})
}

func (d *Device) showState(w http.ResponseWriter, r *http.Request) {
	state, _ := json.MarshalIndent(d.GetState(), "", "\t")
	d.renderTemplate(w, "state.tmpl", string(state))
}

func (d *Device) showCode(w http.ResponseWriter, r *http.Request) {
	// Retrieve top-level entries
	entries, _ := fs.ReadDir(d.layeredFS, ".")
	// Collect entry names
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	d.renderTemplate(w, "code.tmpl", names)
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

func (d *Device) showDownload(w http.ResponseWriter, r *http.Request) {
	values := d.DeployValues()
	target := firstValue(values, "target")
	d.renderPage(w, "device-download.tmpl", pageVars{
		"view":        "download",
		"target":      target,
		"linuxTarget": linuxTarget(target),
		"port":        firstValue(values, "port"),
	})
}

func (d *Device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	values := d.DeployValues()
	target := d.currentTarget(r.URL.Query())
	d.renderPage(w, "device-download-target.tmpl", pageVars{
		"linuxTarget": linuxTarget(target),
		"missingWifi": len(wifiAuths) == 0,
		"ssid":        firstValue(values, "ssid"),
		"port":        firstValue(values, "port"),
	})
}

func (d *Device) showDownloadInstructions(w http.ResponseWriter, r *http.Request) {
	target := d.currentTarget(r.URL.Query())
	template := "instructions-" + target + ".tmpl"
	d.renderPage(w, template, pageVars{})
}

func (d *Device) createChild(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Missing device name")
}

type MsgDestroy struct {
	ChildId string
}

func (d *Device) destroyChild(w http.ResponseWriter, r *http.Request) {

	// Remove the child from devices
	childId := r.URL.Query().Get("ChildId")
	if err := d.removeChild(childId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rebuild routing table
	routesBuild()

	// Send /destroy msg up
	var msg MsgDestroy
	pkt, err := NewPacketFromURL(r.URL, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pkt.SetDst(d.Id).RouteUp()
}

func (d *Device) showNewModal(w http.ResponseWriter, r *http.Request) {
	d.renderPage(w, "new-modal.tmpl", pageVars{
		"models": makers,
	})
}

func (d *Device) selectModel(w http.ResponseWriter, r *http.Request) {
	value := r.FormValue("value")
	fmt.Fprintf(w, `<input type="hidden" id="model" name="Model" value="%s">`, value)
}
