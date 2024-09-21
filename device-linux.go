//go:build !tinygo

package thing2

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
)

//go:embed css docs images js template
var deviceFs embed.FS

type devicesMap map[string]*Device // key: device id

var devices = make(devicesMap)
var devicesMu sync.RWMutex

type deviceOS struct {
	*http.ServeMux
	templates *template.Template
	layeredFS
}

func (d *Device) buildOS() error {
	d.ServeMux = http.NewServeMux()

	// Build device's layered FS.  fs is stacked on top of
	// deviceFs, so fs:foo.tmpl will override deviceFs:foo.tmpl,
	// when searching for file foo.tmpl.
	d.layeredFS.stack(deviceFs)
	d.layeredFS.stack(d.FS)

	// Build the device templates
	d.templates = d.layeredFS.parseFS("template/*.tmpl", template.FuncMap{
		"title": func(s string) string { return strings.Title(s) },
		"add":   func(a, b int) int { return a + b },
		"mult":  func(a, b int) int { return a * b },
	})

	// All devices have a base device API
	d.api()

	// Install the device-specific API handlers
	d.handlersInstall()

	return nil
}

func devicesMake() {
	for id, device := range devices {
		if id != device.Id {
			fmt.Println("Id", id, "mismatch, skipping device Id", device.Id)
			delete(devices, id)
			continue
		}
		model, ok := Models[device.Model]
		if !ok {
			fmt.Println("Model", device.Model,
				"not registered, skipping device id", id)
			delete(devices, id)
			continue
		}
		if err := device.build(model.Maker); err != nil {
			fmt.Println("Device setup failed, skipping device id", id, err)
			delete(devices, id)
			continue
		}
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
		root.Flags.Set(flagOnline)
		root.Flags.Set(flagMetal)
		return root, nil
	case len(roots) > 1:
		return nil, fmt.Errorf("More than one tree found in devices, aborting")
	}

	return nil, fmt.Errorf("No tree found in devices")
}

func (d *Device) addChild(id, model, name string) error {
	var child = &Device{Id: id, Model: model, Name: name}

	maker, ok := Models[model]
	if !ok {
		return fmt.Errorf("Unknown model")
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()

	if _, ok := devices[id]; ok {
		return fmt.Errorf("Child device already exists")
	}

	d.Lock()
	defer d.Unlock()

	if err := child.build(maker.Maker); err != nil {
		return err
	}

	if slices.Contains(d.Children, id) {
		return fmt.Errorf("Device's children already includes child")
	}

	d.Children = append(d.Children, id)

	devices[id] = child
	child.deviceInstall()

	return nil
}

func (d *Device) removeChild(id string) error {

	devicesMu.Lock()
	defer devicesMu.Unlock()

	if _, ok := devices[id]; ok {
		delete(devices, id)
		for _, device := range devices {
			device.Lock()
			if index := slices.Index(device.Children, id); index != -1 {
				device.Children = slices.Delete(device.Children, index, index+1)
				// TODO remove everything below child
			}
			device.Unlock()
		}
		return nil
	}

	return deviceNotFound(id)
}

func deviceNotFound(id string) error {
	//dumpStackTrace()
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

func deviceRenderPkt(w io.Writer, sessionId string, pkt *Packet) error {
	//fmt.Println("deviceRenderPkt", pkt)
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[pkt.Dst]; ok {
		return d._renderPkt(w, sessionId, pkt)
	}
	return deviceNotFound(pkt.Dst)
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
	d.Flags.Set(flagOnline)
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
	d.Flags.Unset(flagOnline)
	d.Unlock()

	pkt := &Packet{Dst: id, Path: "/offline"}
	pkt.RouteUp()
}

func (d *Device) updateDirty(dirty bool) {
	println("d.update")

	d.Lock()
	if dirty {
		d.Flags.Set(flagDirty)
	} else {
		d.Flags.Unset(flagDirty)
	}
	d.Unlock()

	pkt := &Packet{Dst: d.Id, Path: "/dirty"}
	pkt.RouteUp()
}

func deviceDirty(id string) {
	println("deviceDirty")
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	for deviceId, device := range devices {
		if deviceId == id {
			device.updateDirty(true)
		}
		// Set parent dirty also
		if slices.Contains(device.Children, id) {
			devices[deviceId].updateDirty(true)
		}
	}
}

func devicesClean() {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	for _, device := range devices {
		device.updateDirty(false)
	}
}

func deviceParent(id string) string {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	for _, device := range devices {
		if slices.Contains(device.Children, id) {
			return device.Id
		}
	}
	return ""
}

func devicesLoad() error {
	var devicesJSON = Getenv("DEVICES", "")
	var devicesFile = Getenv("DEVICES_FILE", "devices.json")

	devicesMu.Lock()
	defer devicesMu.Unlock()

	// Give DEVICES priority over DEVICES_FILE

	if devicesJSON == "" {
		return fileReadJSON(devicesFile, &devices)
	}

	return json.Unmarshal([]byte(devicesJSON), &devices)
}

func devicesSave() error {
	var devicesJSON = Getenv("DEVICES", "")
	var devicesFile = Getenv("DEVICES_FILE", "devices.json")

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	if devicesJSON == "" {
		return fileWriteJSON(devicesFile, &devices)
	}

	return nil
}

func devicesSendState(l linker) {
	fmt.Println("Sending /state to all devices")
	devicesMu.RLock()
	for id, d := range devices {
		var pkt = &Packet{
			Dst:  id,
			Path: "/state",
		}
		d.RLock()
		pkt.Marshal(d.State)
		d.RUnlock()
		fmt.Println("Sending:", pkt)
		l.Send(pkt)
	}
	devicesMu.RUnlock()
}
