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
	"sort"
	"strings"

	"github.com/merliot/thing2/target"
	"golang.org/x/exp/maps"
)

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
	pageVars["root"] = (d == root)
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

func (d *Device) renderChildren(w io.Writer, sessionId, path string, level int) error {

	fmt.Println("renderChildren", d.Id, path, level)
	// Collect child devices from d.Children
	var childDevices []*Device
	for _, childId := range d.Children {
		if child, exists := devices[childId]; exists {
			childDevices = append(childDevices, child)
		}
	}

	// Sort the collected child devices by ToLower(Device.Name)
	sort.Slice(childDevices, func(i, j int) bool {
		return strings.ToLower(childDevices[i].Name) < strings.ToLower(childDevices[j].Name)
	})

	// Render the child devices in sorted order
	for _, child := range childDevices {
		lastView, err := _sessionLastView(sessionId, child.Id)
		if err != nil {
			lastView.view = "overview"
			lastView.showChildren = false
		}
		lastView.level = level
		fmt.Println("renderChildren", child.Id, lastView.view, lastView.level, lastView.showChildren)
		fmt.Println()

		if err := child.render(w, sessionId, path, lastView.view,
			lastView.level, lastView.showChildren); err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) render(w io.Writer, sessionId, path, view string, level int, showChildren bool) error {

	path = strings.TrimPrefix(path, "/")
	template := path + "-" + view + ".tmpl"
	var pageVars = pageVars{
		"sessionId": sessionId,
		"level":     level * 10,
	}

	fmt.Println("render", d.Id, path, showChildren, template, &pageVars)

	if err := d.renderPage(w, template, pageVars); err != nil {
		return err
	}

	_sessionSave(sessionId, d.Id, view, level, showChildren)

	if !showChildren || len(d.Children) == 0 || path != "device" {
		return nil
	}

	return d.renderChildren(w, sessionId, path, level+1)
}

func (d *Device) renderPkt(w io.Writer, s *session, pkt *Packet) error {

	lastView, err := _sessionLastView(s.sessionId, d.Id)
	if err != nil {
		return err
	}

	fmt.Println("renderPkt", d.Id, lastView)
	return d.render(w, s.sessionId, pkt.Path, lastView.view, lastView.level, lastView.showChildren)
}

func (d *Device) Render(sessionId, path, view string, level int, showChildren bool) (template.HTML, error) {
	var buf bytes.Buffer

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if err := d.render(&buf, sessionId, path, view, level, showChildren); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

/*
func (d *Device) RenderChildren(sessionId string) (template.HTML, error) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	var buf bytes.Buffer
	var pkt = Packet{Path: "/state"}

	// Collect devices from d.Children
	var childDevices []*Device
	for _, childId := range d.Children {
		if child, exists := devices[childId]; exists {
			childDevices = append(childDevices, child)
		}
	}

	// Sort the collected devices by ToLower(Device.Name)
	sort.Slice(childDevices, func(i, j int) bool {
		return strings.ToLower(childDevices[i].Name) < strings.ToLower(childDevices[j].Name)
	})

	// Render the devices in sorted order
	for _, child := range childDevices {
		if err := child.renderPkt(&buf, sessionId, "tile", &pkt); err != nil {
			return template.HTML(""), err
		}
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) RenderChildHTML(sessionId, childId, view string) (template.HTML, error) {

	//println("RenderChildHTML", sessionId, childId, view)
	//dumpStackTrace()

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	child, ok := devices[childId]
	if !ok {
		return template.HTML(""), deviceNotFound(childId)
	}

	//_sessionDeviceSave(sessionId, childId, view)

	child.RLock()
	defer child.RUnlock()

	var buf bytes.Buffer
	var pkt = Packet{Path: "/state"}
	if err := child.renderPkt(&buf, sessionId, view, &pkt); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}
*/

func (d *Device) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *Device) showIndex(w http.ResponseWriter, r *http.Request) {
	println("showIndex", r.Host, r.URL.String())
	sessionId, ok := newSession()
	if !ok {
		http.Error(w, "no more sessions", http.StatusTooManyRequests)
		return
	}
	err := d.renderPage(w, "index.tmpl", pageVars{
		"sessionId": sessionId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) keepAlive(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	if !sessionUpdate(sessionId) {
		// Session expired, force full page refresh to start new
		// session
		w.Header().Set("HX-Refresh", "true")
	}
}

func (d *Device) showFull(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	//sessionDeviceSave(sessionId, d.Id, "full")
	err := d.renderPage(w, "device-full.tmpl", pageVars{
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
	err := d.renderPage(w, "device-list.tmpl", pageVars{
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
	err := d.renderPage(w, "device-tile.tmpl", pageVars{
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
	err := d.renderPage(w, "device-detail.tmpl", pageVars{
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
	err := d.renderPage(w, "device-info.tmpl", pageVars{
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
	err := d.renderPage(w, "device-download.tmpl", pageVars{
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
	err := d.renderPage(w, "device-download-target.tmpl", pageVars{
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
	if err := d.renderPage(w, template, pageVars{}); err != nil {
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
	err := d.renderPage(w, "modal-new.tmpl", pageVars{
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
