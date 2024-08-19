package thing2

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/exp/slices"
)

//go:embed css docs images js template favicon.ico
var deviceFs embed.FS

type Devicer interface {
	GetConfig() Config
	GetHandlers() Handlers
}

type Config struct {
	Model   string
	State   any
	FS      *embed.FS
	Targets []string
}

type Device struct {
	Id             string
	Model          string
	Name           string
	Online         bool
	Children       []string
	Devicer        `json:"-"`
	Handlers       `json:"-"`
	*http.ServeMux `json:"-"`
	sync.RWMutex   `json:"-"`
	cfg            Config
	templates      *template.Template

	layeredFS
	// DeployParams is device deploy configuration in an html param format
	DeployParams string
	// Administratively locked
	Locked bool `json:"-"`
}

var devices = make(map[string]*Device)
var devicesMu sync.RWMutex

func (d *Device) build(maker Maker) {

	d.Devicer = maker()

	d.cfg = d.GetConfig()
	d.Online = false
	d.ServeMux = http.NewServeMux()

	// Build device's layered FS.  fs is stacked on top of
	// deviceFs, so fs:foo.tmpl will override deviceFs:foo.tmpl,
	// when searching for file foo.tmpl.
	d.layeredFS.stack(deviceFs)
	d.layeredFS.stack(d.cfg.FS)

	// Build the device templates
	d.templates = d.layeredFS.parseFS("template/*.tmpl")

	// All devices have a base device API
	d.api()

	// Install the device-specific API handlers
	d.Handlers = d.GetHandlers()
	d.handlersInstall()

	// Configure the device using DeployParams
	_, err := d.formConfig(d.DeployParams)
	if err != nil {
		fmt.Println("Error configuring device using DeployParams:", err, d)
	}
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

func (d *Device) removeChild(childId string) error {
	devicesMu.Lock()
	defer devicesMu.Unlock()
	if _, ok := devices[childId]; ok {
		delete(devices, childId)
		for _, device := range devices {
			device.Lock()
			if index := slices.Index(device.Children, childId); index != -1 {
				device.Children = slices.Delete(device.Children, index, index+1)
				// TODO remove everything below child
			}
			device.Unlock()
		}
		return nil
	}
	return deviceNotFound(childId)
}

func (d *Device) formConfig(rawQuery string) (changed bool, err error) {

	// rawQuery is the proposed new DeployParams
	proposedParams, err := url.QueryUnescape(rawQuery)
	if err != nil {
		return false, err
	}
	values, err := url.ParseQuery(proposedParams)
	if err != nil {
		return false, err
	}

	d.Lock()
	defer d.Unlock()

	fmt.Println("Proposed DeployParams:", proposedParams)

	// Form-decode these values into the device to configure the device
	if err := decoder.Decode(d.cfg.State, values); err != nil {
		return false, err
	}

	if proposedParams == d.DeployParams {
		// No change
		return false, nil
	}

	// Save changes.  Stored DeployParams unescaped.
	d.DeployParams = proposedParams
	return true, nil
}

func deviceNotFound(id string) error {
	return fmt.Errorf("Device '%s' not found", id)
}

func (d *Device) routeDown(pkt *Packet) {

	// If device is the root device, deliver packet to device.  The root
	// device is running on 'metal', so this is the packet's final
	// destination.
	if d == root {
		d.handle(pkt)
		return
	}

	// Otherwise, route the packet down
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

func (d *Device) _render(w io.Writer, sessionId string, url *url.URL) error {
	switch url.Path {
	case "/", "/full":
		return d._showFull(w, sessionId)
	case "/tile":
		return d._showTile(w, sessionId)
	case "/list":
		return d._showList(w, sessionId)
	}
	return nil
}

func _deviceRender(w io.Writer, sessionId, id string, url *url.URL) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		return d._render(w, sessionId, url)
	}
	return deviceNotFound(id)
}

func deviceRender(w io.Writer, sessionId, id string, url *url.URL) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		d.RLock()
		defer d.RUnlock()
		return d._render(w, sessionId, url)
	}
	return deviceNotFound(id)
}

func deviceOnline(ann announcement) error {
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	d, ok := devices[ann.Id]
	if !ok {
		return deviceNotFound(ann.Id)
	}

	if d.Model != ann.Model {
		return fmt.Errorf("Device model wrong.  Want %s; got %s",
			d.Model, ann.Model)
	}

	if d.Name != ann.Name {
		return fmt.Errorf("Device name wrong.  Want %s; got %s",
			d.Name, ann.Name)
	}

	if d.DeployParams != ann.DeployParams {
		return fmt.Errorf("Device DeployParams wrong.\nWant: %s\nGot: %s",
			d.DeployParams, ann.DeployParams)
	}

	d.Lock()
	d.Online = true
	d.Unlock()

	// We don't need to send a /online pkt up because /state is going to be
	// sent UP

	return nil
}

func deviceOffline(id string) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	d, ok := devices[id]
	if !ok {
		return
	}

	d.Lock()
	d.Online = false
	d.Unlock()

	pkt := &Packet{
		Dst:  id,
		Path: "/offline",
	}
	pkt.RouteUp()
}
