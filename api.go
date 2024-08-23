//go:build !tinygo

package thing2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/merliot/thing2/target"
	"golang.org/x/exp/maps"
)

// modelInstall installs /model/{model} pattern for device in default ServeMux
func (d *Device) modelInstall() {
	prefix := "/model/" + d.Model
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	fmt.Println("modelInstall", prefix)
}

func modelsInstall() {
	for name, model := range Models {
		var proto = &Device{Model: name}
		proto.build(model.Maker)
		proto.modelInstall()
	}
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

func (d *Device) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *Device) api() {
	d.HandleFunc("GET /", d.serveStaticFile)
	d.HandleFunc("GET /{$}", d.showIndex)
	d.HandleFunc("PUT /keepalive", d.keepAlive)
	d.HandleFunc("GET /full", d.showFull)
	d.HandleFunc("GET /tile", d.showTile)
	d.HandleFunc("GET /list", d.showList)
	d.HandleFunc("GET /detail", d.showDetail)
	d.HandleFunc("GET /info", d.showInfo)
	d.HandleFunc("GET /code", d.showCode)
	d.HandleFunc("GET /devices", d.showDevices)
	d.HandleFunc("GET /download", d.showDownload)
	d.HandleFunc("GET /download-target", d.showDownloadTarget)
	d.HandleFunc("GET /download-instructions", d.showDownloadInstructions)
	d.HandleFunc("GET /download-image", d.downloadImage)
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
	err := tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

type pageVars map[string]any

type pageData struct {
	Vars   pageVars
	Device *Device
	State  any
}

func (d *Device) renderPage(w io.Writer, name string, pageVars pageVars) error {

	values := d.deployValues()
	target := values.Get("target")

	// Add common Vars for all pages
	pageVars["target"] = target
	pageVars["online"] = d.Flags.IsSet(flagOnline)
	pageVars["demo"] = d.Flags.IsSet(flagDemo)
	pageVars["dirty"] = d.Flags.IsSet(flagDirty)
	pageVars["locked"] = d.Flags.IsSet(flagLocked)

	//fmt.Println("renderPage", name, pageVars)

	return d.renderTemplate(w, name, &pageData{
		Vars:   pageVars,
		Device: d,
		State:  d.State,
	})
}

/*
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
*/

func (d *Device) RenderChildHTML(sessionId, childId, rawUrl string) (template.HTML, error) {

	println("RenderChildHTML", sessionId, childId, rawUrl)
	//dumpStackTrace()

	child, ok := devices[childId]
	if !ok {
		return template.HTML(""), deviceNotFound(childId)
	}

	url, err := url.Parse(rawUrl)
	if err != nil {
		return template.HTML(""), err
	}

	_sessionDeviceSave(sessionId, childId, url)

	child.RLock()
	defer child.RUnlock()

	var buf bytes.Buffer
	if err := child._render(&buf, sessionId, url); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) showIndex(w http.ResponseWriter, r *http.Request) {
	println("showIndex", r.Host, r.URL.String())
	sessionId, ok := newSession()
	if !ok {
		http.Error(w, "no more sessions", http.StatusTooManyRequests)
		return
	}
	sessionDeviceSave(sessionId, d.Id, r.URL)
	d.renderPage(w, "index.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "full",
		"models":    Models,
	})
}

func (d *Device) keepAlive(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	if !sessionUpdate(sessionId) {
		// Session expired, force full page refresh to start new
		// session
		w.Header().Set("HX-Refresh", "true")
	}
}

func (d *Device) _showFull(w io.Writer, sessionId string) error {
	return d.renderPage(w, "device-full.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "full",
	})
}

func (d *Device) showFull(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSave(sessionId, d.Id, r.URL)
	if err := d._showFull(w, sessionId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) _showList(w io.Writer, sessionId string) error {
	return d.renderPage(w, "device-list.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "list",
	})
}

func (d *Device) showList(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSave(sessionId, d.Id, r.URL)
	if err := d._showList(w, sessionId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) _showTile(w io.Writer, sessionId string) error {
	return d.renderPage(w, "device-tile.tmpl", pageVars{
		"sessionId": sessionId,
		"view":      "tile",
	})
}

func (d *Device) showTile(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSave(sessionId, d.Id, r.URL)
	if err := d._showTile(w, sessionId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) _showDetail(w io.Writer, sessionId string, url *url.URL) error {
	childId := url.Query().Get("childId")
	prevView := url.Query().Get("prevView")
	return d.renderPage(w, "device-detail.tmpl", pageVars{
		"sessionId": sessionId,
		"childId":   childId,
		"view":      "detail",
		"prevView":  prevView,
	})
}

func (d *Device) showDetail(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	sessionDeviceSave(sessionId, d.Id, r.URL)
	if err := d._showDetail(w, sessionId, r.URL); err != nil {
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
	w.Header().Set("Content-Type", "application/json")
	state, _ := json.MarshalIndent(d.State, "", "\t")
	w.Write(state)
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

func (d *Device) showDevices(w http.ResponseWriter, r *http.Request) {
	var childDevices = make(devicesMap)

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	for _, childId := range d.Children {
		child := devices[childId]
		childDevices[childId] = child
	}
	childDevices[d.Id] = d // add self

	w.Header().Set("Content-Type", "application/json")
	state, _ := json.MarshalIndent(childDevices, "", "\t")
	w.Write(state)
}

func linuxTarget(target string) bool {
	return target == "demo" || target == "x86-64" || target == "rpi"
}

func (d *Device) deployValues() url.Values {
	values, err := url.ParseQuery(string(d.DeployParams))
	if err != nil {
		panic(err.Error())
	}
	return values
}

func (d *Device) selectedTarget(params url.Values) string {
	target := params.Get("target")
	if target == "" {
		target = d.deployValues().Get("target")
	}
	return target
}

func (d *Device) showDownload(w http.ResponseWriter, r *http.Request) {
	values := d.deployValues()
	t := values.Get("target")
	d.renderPage(w, "device-download.tmpl", pageVars{
		"targets":     target.MakeTargets(d.Targets),
		"linuxTarget": linuxTarget(t),
		"port":        values.Get("port"),
	})
}

func (d *Device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	values := d.deployValues()
	selectedTarget := d.selectedTarget(r.URL.Query())
	wifiAuths := wifiAuths()
	d.renderPage(w, "device-download-target.tmpl", pageVars{
		"targets":        target.MakeTargets(d.Targets),
		"selectedTarget": selectedTarget,
		"linuxTarget":    linuxTarget(selectedTarget),
		"missingWifi":    len(wifiAuths) == 0,
		"ssids":          maps.Keys(wifiAuths),
		"ssid":           values.Get("ssid"),
		"port":           values.Get("port"),
	})
}

func (d *Device) showDownloadInstructions(w http.ResponseWriter, r *http.Request) {
	target := d.selectedTarget(r.URL.Query())
	template := "instructions-" + target + ".tmpl"
	d.renderPage(w, template, pageVars{})
}

func (d *Device) createChild(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Missing device name")
}

type msgDestroy struct {
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
	var msg msgDestroy
	pkt, err := newPacketFromURL(r.URL, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pkt.SetDst(d.Id).RouteUp()
}

func (d *Device) showNewModal(w http.ResponseWriter, r *http.Request) {
	d.renderPage(w, "new-modal.tmpl", pageVars{
		"models": Models,
	})
}

func (d *Device) selectModel(w http.ResponseWriter, r *http.Request) {
	value := r.FormValue("value")
	fmt.Fprintf(w, `<input type="hidden" id="model" name="Model" value="%s">`, value)
}
