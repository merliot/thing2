package thing2

import (
	"embed"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"sync"

	"github.com/merliot/thing2/target"
	"golang.org/x/exp/slices"
)

//go:embed css images js template favicon.ico
var deviceFs embed.FS

type Device struct {
	Id             string
	Model          string
	Name           string
	Online         bool
	Children       []string
	Modeler        `json:"-"`
	Handlers       `json:"-"`
	templates      *template.Template
	*http.ServeMux `json:"-"`
	sync.RWMutex   `json:"-"`

	layeredFS
	// DeployParams is device deploy configuration in an html param format
	DeployParams string
	// Administratively locked
	Locked bool `json:"-"`
	// Targets supported by device
	target.Targets `json:"-"`
}

var devices = make(map[string]*Device)
var devicesMu sync.RWMutex

func (d *Device) build(maker Maker) {

	d.Modeler = maker()
	d.Online = false
	d.ServeMux = http.NewServeMux()
	d.Targets = target.MakeTargets(d.GetTargets())

	// Build device's layered FS.  fs is stacked on top of
	// deviceFs, so fs:foo.tmpl will override deviceFs:foo.tmpl,
	// when searching for file foo.tmpl.
	d.layeredFS.stack(deviceFs)
	d.layeredFS.stack(d.GetFS())

	// Build the device templates
	d.templates = d.layeredFS.parseFS("template/*.tmpl")

	// All devices have a base device API
	d.API()

	// Install the device-specific API handlers
	d.Handlers = d.GetHandlers()
	d.handlersInstall()
}

func devicesMake() {
	for id, device := range devices {
		if id != device.Id {
			fmt.Println("Id", id, "mismatch, skipping device Id", device.Id)
			delete(devices, id)
			continue
		}
		maker, ok := makers[device.Model]
		if !ok {
			fmt.Println("Model", device.Model,
				"not registered, skipping device id", id)
			delete(devices, id)
			continue
		}
		device.build(maker)
	}
}

// devicesFindRoot returns the root *Device if there is exactly one tree
// defined by the devices map, otherwise nil.
func devicesFindRoot() (*Device, error) {

	// Create a map to track all devices that are children
	childSet := make(map[string]bool)

	// Populate the childSet with the Ids of all children
	for _, device := range devices {
		for _, child := range device.Children {
			if _, ok := devices[child]; !ok {
				return nil, fmt.Errorf("Child Id %s not found in devices", child)
			}
			childSet[child] = true
		}
	}

	// Find all root devices
	var roots []*Device
	for id, device := range devices {
		if _, isChild := childSet[id]; !isChild {
			roots = append(roots, device)
		}
	}

	// Return the root if there is exactly one tree
	switch {
	case len(roots) == 1:
		root := roots[0]
		root.Online = true
		return root, nil
	case len(roots) > 1:
		return nil, fmt.Errorf("More than one tree found in devices, aborting")
	}

	return nil, fmt.Errorf("No tree found in devices")
}

func (d *Device) addChild(child *Device) error {

	d.Lock()
	defer d.Unlock()

	if slices.Contains(d.Children, child.Id) {
		return fmt.Errorf("child '%s' already exists", child.Id)
	}

	d.Children = append(d.Children, child.Id)
	child.deviceInstall()

	return nil
}

func (d *Device) SetDeployParams(params string) {
	d.DeployParams = html.UnescapeString(params)
}

func deviceNotFound(id string) error {
	return fmt.Errorf("Device '%s' not found", id)
}

func (d *Device) handle(pkt *Packet) {
	d.Lock()
	defer d.Unlock()
	if handler, ok := d.Handlers[pkt.Path]; ok {
		fmt.Println("Handling", pkt.String())
		handler.Callback(pkt)
	}
}

func (d *Device) routeDown(pkt *Packet) {
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

func (d *Device) _render(sessionId string, w io.Writer, view string) error {
	switch view {
	case "full":
		return d._showFull(sessionId, w)
	case "tile":
		return d._showTile(sessionId, w)
	}
	return nil
}

func _deviceRender(sessionId string, id, view string, w io.Writer) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		return d._render(sessionId, w, view)
	}
	return deviceNotFound(id)
}

func deviceRender(sessionId, id, view string, w io.Writer) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		d.RLock()
		defer d.RUnlock()
		return d._render(sessionId, w, view)
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
	pkt.Unmarshal(d.GetState()).RouteUp()
}
