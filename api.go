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
	"sort"
	"strings"

	"github.com/merliot/thing2/target"
	"golang.org/x/exp/maps"
)

func (d *Device) api() {
	d.HandleFunc("GET /", d.serveStaticFile)
	d.HandleFunc("GET /{$}", d.showIndex)
	d.HandleFunc("PUT /keepalive", d.keepAlive)
	d.HandleFunc("GET /overview", d.overview)
	d.HandleFunc("GET /detail", d.detail)
	d.HandleFunc("GET /expand", d.expand)
	d.HandleFunc("GET /collapse", d.collapse)

	//d.HandleFunc("GET /full", d.showFull)
	//d.HandleFunc("GET /tile", d.showTile)
	//d.HandleFunc("GET /list", d.showList)
	//d.HandleFunc("GET /detail", d.showDetail)

	d.HandleFunc("GET /info", d.showInfo)
	d.HandleFunc("GET /code", d.showCode)
	d.HandleFunc("GET /save", d.saveDevices)
	d.HandleFunc("GET /devices", d.showDevices)
	d.HandleFunc("GET /download", d.showDownload)
	d.HandleFunc("GET /download-target", d.showDownloadTarget)
	d.HandleFunc("GET /download-instructions", d.showDownloadInstructions)
	d.HandleFunc("GET /download-image", d.downloadImage)
	d.HandleFunc("GET /create", d.createChild)
	d.HandleFunc("DELETE /destroy", d.destroyChild)
	d.HandleFunc("GET /newModal", d.showNewModal)
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

type renderVars map[string]any

type renderData struct {
	Vars   renderVars
	Device *Device
	State  any
}

func (d *Device) _renderTmpl(w io.Writer, name string, renderVars renderVars) error {

	values := d.deployValues()
	target := values.Get("target")

	// Add common Vars for all pages
	renderVars["target"] = target
	renderVars["root"] = (d == root)
	renderVars["online"] = d.Flags.IsSet(flagOnline)
	renderVars["demo"] = d.Flags.IsSet(flagDemo)
	renderVars["dirty"] = d.Flags.IsSet(flagDirty)
	renderVars["locked"] = d.Flags.IsSet(flagLocked)

	//fmt.Println("_renderTmpl", name, renderVars)

	return d.renderTemplate(w, name, &renderData{
		Vars:   renderVars,
		Device: d,
		State:  d.State,
	})
}

func (d *Device) renderTmpl(w io.Writer, name string, renderVars renderVars) error {
	d.RLock()
	defer d.RUnlock()
	return d._renderTmpl(w, name, renderVars)
}

func (d *Device) _renderChildren(w io.Writer, sessionId string, level int) error {

	if len(d.Children) == 0 {
		return nil
	}

	// Collect child devices from d.Children
	var children []*Device
	for _, childId := range d.Children {
		if child, exists := devices[childId]; exists {
			children = append(children, child)
		}
	}

	// Sort the collected child devices by ToLower(Device.Name)
	sort.Slice(children, func(i, j int) bool {
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})

	// Render the child devices in sorted order
	for _, child := range children {

		view, _, showChildren, err := _sessionLastView(sessionId, child.Id)
		if err != nil {
			// If there was no view saved, default to overview, collapsed
			view = "overview"
			showChildren = false
		}

		if err := child._render(w, sessionId, "/device", view, level, showChildren); err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) _render(w io.Writer, sessionId, path, view string, level int, showChildren bool) error {

	path = strings.TrimPrefix(path, "/")
	template := path + "-" + view + ".tmpl"
	var renderVars = renderVars{
		"sessionId":    sessionId,
		"level":        level,
		"showChildren": showChildren,
	}

	fmt.Println("_render", d.Id, path, showChildren, template, &renderVars)

	if err := d._renderTmpl(w, template, renderVars); err != nil {
		return err
	}

	_sessionSave(sessionId, d.Id, view, level, showChildren)

	return nil
}

func (d *Device) render(w io.Writer, sessionId, path, view string, level int, showChildren bool) error {
	d.RLock()
	defer d.RUnlock()
	return d._render(w, sessionId, path, view, level, showChildren)
}

func (d *Device) _renderPkt(w io.Writer, sessionId string, pkt *Packet) error {

	view, level, showChildren, err := _sessionLastView(sessionId, d.Id)
	if err != nil {
		return err
	}

	fmt.Println("_renderPkt", d.Id, view, level, showChildren, pkt)
	return d._render(w, sessionId, pkt.Path, view, level, showChildren)
}

func (d *Device) Render(sessionId, path, view string, level int, showChildren bool) (template.HTML, error) {
	var buf bytes.Buffer

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if err := d._render(&buf, sessionId, path, view, level, showChildren); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) RenderChildren(sessionId string, level int) (template.HTML, error) {
	var buf bytes.Buffer

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if err := d._renderChildren(&buf, sessionId, level); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *Device) keepAlive(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	if !sessionUpdate(sessionId) {
		// Session expired, force full page refresh to start new
		// session
		w.Header().Set("HX-Refresh", "true")
	}
}

func (d *Device) showIndex(w http.ResponseWriter, r *http.Request) {
	println("showIndex", r.Host, r.URL.String())

	sessionId, ok := newSession()
	if !ok {
		http.Error(w, "no more sessions", http.StatusTooManyRequests)
		return
	}
	err := d.renderTmpl(w, "index.tmpl", renderVars{
		"sessionId": sessionId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) overview(w http.ResponseWriter, r *http.Request) {
	println("overview", r.Host, r.URL.String())

	sessionId := r.Header.Get("session-id")
	_, level, showChildren, err := sessionLastView(sessionId, d.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := d.render(w, sessionId, "/device", "overview", level, showChildren); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) detail(w http.ResponseWriter, r *http.Request) {
	println("detail", r.Host, r.URL.String())

	sessionId := r.Header.Get("session-id")
	_, level, showChildren, err := sessionLastView(sessionId, d.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := d.render(w, sessionId, "/device", "detail", level, showChildren); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) toggle(w http.ResponseWriter, r *http.Request, showChildren bool) {
	println("toggle", r.Host, r.URL.String(), showChildren)

	sessionId := r.Header.Get("session-id")
	view, level, _, err := sessionLastView(sessionId, d.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := d.render(w, sessionId, "/device", view, level, showChildren); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) expand(w http.ResponseWriter, r *http.Request) {
	d.toggle(w, r, true)
}

func (d *Device) collapse(w http.ResponseWriter, r *http.Request) {
	d.toggle(w, r, false)
}

func (d *Device) showFull(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	//sessionDeviceSave(sessionId, d.Id, "full")
	err := d.renderTmpl(w, "device-full.tmpl", renderVars{
		"sessionId": sessionId,
		"view":      "full",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showList(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	//sessionDeviceSave(sessionId, d.Id, "list")
	err := d.renderTmpl(w, "device-list.tmpl", renderVars{
		"sessionId": sessionId,
		"view":      "list",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showTile(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	//sessionDeviceSave(sessionId, d.Id, "tile")
	err := d.renderTmpl(w, "device-tile.tmpl", renderVars{
		"sessionId": sessionId,
		"view":      "tile",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showDetail(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	//sessionDeviceSave(sessionId, d.Id, "detail")
	childId := r.URL.Query().Get("childId")
	prevView := r.URL.Query().Get("prevView")
	err := d.renderTmpl(w, "device-detail.tmpl", renderVars{
		"sessionId": sessionId,
		"childId":   childId,
		"view":      "detail",
		"prevView":  prevView,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showInfo(w http.ResponseWriter, r *http.Request) {
	err := d.renderTmpl(w, "device-info.tmpl", renderVars{
		"view":    "info",
		"package": Models[d.Model].Package,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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

func (d *Device) saveDevices(w http.ResponseWriter, r *http.Request) {
	if d != root {
		http.Error(w, fmt.Sprintf("Only root device can save"), http.StatusBadRequest)
		return
	}
	if err := devicesSave(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	devicesClean()
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
	err := d.renderTmpl(w, "device-download.tmpl", renderVars{
		"view":        "download",
		"targets":     target.MakeTargets(d.Targets),
		"linuxTarget": linuxTarget(t),
		"port":        values.Get("port"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	values := d.deployValues()
	selectedTarget := d.selectedTarget(r.URL.Query())
	wifiAuths := wifiAuths()
	err := d.renderTmpl(w, "device-download-target.tmpl", renderVars{
		"targets":        target.MakeTargets(d.Targets),
		"selectedTarget": selectedTarget,
		"linuxTarget":    linuxTarget(selectedTarget),
		"missingWifi":    len(wifiAuths) == 0,
		"ssids":          maps.Keys(wifiAuths),
		"ssid":           values.Get("ssid"),
		"port":           values.Get("port"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showDownloadInstructions(w http.ResponseWriter, r *http.Request) {
	target := d.selectedTarget(r.URL.Query())
	template := "instructions-" + target + ".tmpl"
	if err := d.renderTmpl(w, template, renderVars{}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) createChild(w http.ResponseWriter, r *http.Request) {
	var child Device

	pkt, err := newPacketFromURL(r.URL, &child)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO validate msg.Id, msg.Model, msg.Name

	if err := d.addChild(child.Id, child.Model, child.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rebuild routing table
	routesBuild(root)

	// Mark child and parent(s) dirty
	deviceDirty(child.Id)

	// send /create msg up
	pkt.SetDst(d.Id).RouteUp()
}

type msgDestroy struct {
	ChildId string
}

func (d *Device) destroyChild(w http.ResponseWriter, r *http.Request) {
	var msg msgDestroy

	pkt, err := newPacketFromURL(r.URL, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parentId := deviceParent(msg.ChildId)

	// Remove the child from devices
	if err := d.removeChild(msg.ChildId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rebuild routing table
	routesBuild(root)

	// Mark parent(s) dirty
	deviceDirty(parentId)

	// send /destroy msg up
	pkt.SetDst(d.Id).RouteUp()
}

func (d *Device) showNewModal(w http.ResponseWriter, r *http.Request) {
	err := d.renderTmpl(w, "modal-new.tmpl", renderVars{
		"models": Models,
		"newid":  GenerateRandomId(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func templateShow(w http.ResponseWriter, temp string, data any) {
	tmpl, err := template.New("main").Parse(temp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
