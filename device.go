package thing2

import (
	"embed"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/merliot/thing2/target"
)

//go:embed css images js template favicon.ico
var deviceFs embed.FS

type Devicer interface {
	GetId() string
	GetModel() string
	GetName() string
	AddChild(child Devicer) error
	GetChildren() Children
	SetDeployParams(params string)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	InstallDevicePattern()
	InstallModelPattern()
	String() string
}

type Children map[string]Devicer
type Maker func(id, name string) Devicer
type Models []string
type WifiAuth map[string]string // key: ssid; value: passphrase

type Device struct {
	*http.ServeMux `json:"-"`
	Id             string
	Model          string
	Name           string
	Online         bool
	children       Children
	handlers       PacketHandlers
	sessionsMu     sync.RWMutex
	LayeredFS      `json:"-"`
	Models         `json:"-"`
	// WifiAuth is a map of SSID:PASSPHRASE pairs
	WifiAuth `json:"-"`
	// DeployParams is device deploy configuration in an html param format
	DeployParams string
	// Administratively locked
	Locked bool `json:"-"`
	// Data passed to render templates
	data         any
	templates    *template.Template
	sync.RWMutex `json:"-"`
	// Targets supported by device
	target.Targets `json:"-"`
}

var devices = make(map[string]*Device)
var devicesMu sync.RWMutex

func NewDevice(id, model, name string, fs embed.FS, targets []string,
	handlers PacketHandlers) *Device {
	println("NEW DEVICE", id, model, name)

	d := &Device{
		Id:       id,
		Model:    model,
		Name:     name,
		ServeMux: http.NewServeMux(),
		children: make(Children),
		Targets:  target.MakeTargets(targets),
		WifiAuth: make(WifiAuth),
		handlers: handlers,
	}
	d.data = d

	// Add common device handlers
	d.handlers["/state"] = d.saveState

	// Build device's layered FS.  fs is stacked on top of deviceFs, so
	// fs:foo.tmpl will override deviceFs:foo.tmpl, when searching for file
	// foo.tmpl.
	d.LayeredFS.Stack(deviceFs)
	d.LayeredFS.Stack(fs)

	// Build the device templates
	d.templates = d.LayeredFS.ParseFS("template/*.tmpl")

	// All devices inherit this base device API
	d.API()

	// Add device to map of devices
	devicesMu.Lock()
	devices[id] = d
	devicesMu.Unlock()

	return d
}

func (d Device) GetId() string          { return d.Id }
func (d Device) GetModel() string       { return d.Model }
func (d Device) GetName() string        { return d.Name }
func (d *Device) GetChildren() Children { return d.children }
func (d *Device) SetData(data any)      { d.data = data }
func (d *Device) Dispatch(pkt *Packet)  {}
func (d *Device) String() string        { return d.Id + ":" + d.Model + ":" + d.Name }

func (d *Device) AddChild(child Devicer) error {

	d.Lock()
	defer d.Unlock()

	if _, exists := d.children[child.GetId()]; exists {
		return errors.New("child already exists")
	}

	d.children[child.GetId()] = child

	// Install the /device/{id} pattern for child
	child.InstallDevicePattern()

	// Install the /model/{model} pattern, using child as proto (but only
	// if we haven't seen this model before)
	child.InstallModelPattern()

	return nil
}

func (d *Device) SetDeployParams(params string) {
	d.DeployParams = html.UnescapeString(params)
}

func (d *Device) SetWifiAuth(ssids, passphrases string) {
	if ssids == "" {
		return
	}
	keys := strings.Split(ssids, ",")
	values := strings.Split(passphrases, ",")
	for i, key := range keys {
		if i < len(values) {
			d.WifiAuth[key] = values[i]
		}
	}
}

func deviceNotFound(id string) error {
	return fmt.Errorf("Device '%s' not found", id)
}

func (d *Device) handle(pkt *Packet) {
	d.Lock()
	defer d.Unlock()
	if h, ok := d.handlers[pkt.Path]; ok {
		println("handling", pkt.Path)
		h(pkt)
	}
}

func (d *Device) routeDown(pkt *Packet) {
	fmt.Println("routeDown", d, pkt)
	if pkt.Dst == d.Id {
		d.handle(pkt)
		return
	}
	downlinkRoute(pkt)
}

func deviceRouteDown(id string, pkt *Packet) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		d.routeDown(pkt)
	}
}

func deviceRouteUp(id string, pkt *Packet) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		d.handle(pkt)
	}
}

func (d *Device) render(w io.Writer, view string) error {
	switch view {
	case "full":
		return d._showFull(w)
	}
	return nil
}

func _deviceRender(id, view string, w io.Writer) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		return d.render(w, view)
	}
	return deviceNotFound(id)
}

func deviceRender(id, view string, w io.Writer) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		d.RLock()
		defer d.RUnlock()
		return d.render(w, view)
	}
	return deviceNotFound(id)
}

func deviceOnline(id string) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		if d.Online {
			return fmt.Errorf("Device '%s' already online", id)
		}
		d.Lock()
		d.Online = true
		d.Unlock()
		return nil
	}
	return deviceNotFound(id)
}

func deviceOffline(id string) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		d.Lock()
		d.Online = false
		d.Unlock()
	}
}

func deviceCheck(id, model, name string) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if d, ok := devices[id]; ok {
		if d.Model != model {
			return fmt.Errorf("Device model wrong.  Want %s; have %s",
				d.Model, model)
		}
		if d.Name != name {
			return fmt.Errorf("Device name wrong.  Want %s; have %s",
				d.Name, name)
		}
		return nil
	}

	return deviceNotFound(id)
}

func (d *Device) saveState(pkt *Packet) {
	// TODO check pkt Model and Name match device
	pkt.Unmarshal(d.data).RouteUp()
}
