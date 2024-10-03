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
	"path/filepath"
	"sort"
	"strings"
)

func (d *Device) api() {
	d.HandleFunc("GET /", d.serveStaticFile)
	d.HandleFunc("GET /{$}", d.showIndex)
	d.HandleFunc("PUT /keep-alive", d.keepAlive)
	d.HandleFunc("GET /show-view", d.showView)
	d.HandleFunc("GET /state", d.showState)
	d.HandleFunc("GET /code", d.showCode)

	//d.HandleFunc("GET /save", d.saveDevices)
	//d.HandleFunc("GET /devices", d.showDevices)
	//d.HandleFunc("GET /download", d.showDownload)

	d.HandleFunc("GET /download-target", d.showDownloadTarget)
	d.HandleFunc("GET /instructions", d.showInstructions)
	d.HandleFunc("GET /download-image", d.downloadImage)
	d.HandleFunc("GET /create", d.createChild)
	d.HandleFunc("DELETE /destroy", d.destroyChild)
	d.HandleFunc("GET /new-modal", d.showNewModal)
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

func (d *Device) _renderTmpl(w io.Writer, template string, data any) error {
	tmpl := d.templates.Lookup(template)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", template)
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (d *Device) renderTmpl(w io.Writer, template string, data any) error {
	d.RLock()
	defer d.RUnlock()
	return d._renderTmpl(w, template, data)
}

func (d *Device) _renderSession(w io.Writer, template, sessionId string, level int) error {
	return d._renderTmpl(w, template, map[string]any{
		"level":     level,
		"sessionId": sessionId,
	})
}

func (d *Device) renderSession(w io.Writer, template, sessionId string, level int) error {
	d.RLock()
	defer d.RUnlock()
	return d._renderSession(w, template, sessionId, level)
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

	// TODO allow other sort methods?

	// Sort the collected child devices by ToLower(Device.Name)
	sort.Slice(children, func(i, j int) bool {
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})

	// Render the child devices in sorted order
	for _, child := range children {

		view, _, err := _sessionLastView(sessionId, child.Id)
		if err != nil {
			// If there was no view saved, default to overview
			view = "overview"
		}

		if err := child._render(w, sessionId, "/device", view, level); err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) _render(w io.Writer, sessionId, path, view string, level int) error {

	path = strings.TrimPrefix(path, "/")
	template := path + "-" + view + ".tmpl"

	fmt.Println("_render", d.Id, sessionId, path, level, template)

	if err := d._renderSession(w, template, sessionId, level); err != nil {
		return err
	}

	_sessionSave(sessionId, d.Id, view, level)

	return nil
}

func (d *Device) render(w io.Writer, sessionId, path, view string, level int) error {
	d.RLock()
	defer d.RUnlock()
	return d._render(w, sessionId, path, view, level)
}

func (d *Device) _renderPkt(w io.Writer, sessionId string, pkt *Packet) error {

	view, level, err := _sessionLastView(sessionId, d.Id)
	if err != nil {
		return err
	}

	path := pkt.Path
	switch path {
	case "/state", "/offline":
		path = "/device"
	}

	fmt.Println("_renderPkt", d.Id, view, level, pkt)
	return d._render(w, sessionId, path, view, level)
}

func (d *Device) renderView(sessionId, path, view string, level int) (template.HTML, error) {
	var buf bytes.Buffer

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if err := d._render(&buf, sessionId, path, view, level); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) renderChildren(sessionId string, level int) (template.HTML, error) {
	var buf bytes.Buffer

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if err := d._renderChildren(&buf, sessionId, level); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *Device) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	fileExtension := filepath.Ext(r.URL.Path)
	switch fileExtension {
	case ".go", ".tmpl", ".css":
		w.Header().Set("Content-Type", "text/plain")
	case ".gz":
		w.Header().Set("Content-Encoding", "gzip")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *Device) keepAlive(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session-id")
	if !sessionUpdate(sessionId) {
		// Session expired, force full page refresh to start new session
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
	if err := d.renderSession(w, "index.tmpl", sessionId, 0); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showView(w http.ResponseWriter, r *http.Request) {
	println("show", r.Host, r.URL.String())

	view := r.URL.Query().Get("view")
	sessionId := r.Header.Get("session-id")

	_, level, err := sessionLastView(sessionId, d.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := d.render(w, sessionId, "/device", view, level); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showState(w http.ResponseWriter, r *http.Request) {
	d.renderTmpl(w, "device-state-state.tmpl", nil)
}

func (d *Device) showCode(w http.ResponseWriter, r *http.Request) {
	// Retrieve top-level entries
	entries, _ := fs.ReadDir(d.layeredFS, ".")
	// Collect entry names
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	d.renderTmpl(w, "code.tmpl", names)
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

func (d *Device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	selectedTarget := d.selectedTarget(r.URL.Query())
	err := d.renderTmpl(w, "device-download-target.tmpl", map[string]any{
		"selectedTarget": selectedTarget,
		"linuxTarget":    linuxTarget(selectedTarget),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) showInstructions(w http.ResponseWriter, r *http.Request) {
	target := d.selectedTarget(r.URL.Query())
	template := "instructions-" + target + ".tmpl"
	if err := d.renderTmpl(w, template, nil); err != nil {
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

	if err := addChild(d, child.Id, child.Model, child.Name); err != nil {
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

	if err := removeChild(msg.ChildId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rebuild routing table
	routesBuild(root)

	// Mark parent(s) dirty
	deviceDirty(parentId)

	// send /destroy msg up
	pkt.SetDst(parentId).RouteUp()
}

func (d *Device) showNewModal(w http.ResponseWriter, r *http.Request) {
	err := d.renderTmpl(w, "modal-new.tmpl", map[string]any{
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
