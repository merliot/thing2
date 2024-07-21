package thing2

import (
	"embed"
	"errors"
	"html/template"
	"net/http"
	"sync"
)

//go:embed template
var deviceFs embed.FS

type Device struct {
	Id             string
	Model          string
	Name           string
	*http.ServeMux `json:"-"`
	LayeredFS      `json:"-"`
	templates      *template.Template
	parent         *Device
	children       map[string]*Device
	sync.RWMutex
}

func NewDevice(id, model, name string, fs embed.FS) *Device {
	println("NEW DEVICE", id, model, name)

	d := &Device{
		Id:       id,
		Model:    model,
		Name:     name,
		ServeMux: http.NewServeMux(),
		children: make(map[string]*Device),
	}

	// Build device's layered FS.  fs is stacked on top of deviceFs, so
	// fs:foo.tmpl will override deviceFs:foo.tmpl, when seraching for file
	// foo.tmpl.
	d.LayeredFS.Stack(deviceFs)
	d.LayeredFS.Stack(fs)

	// Build the device templates
	d.templates = d.LayeredFS.ParseFS("template/*.tmpl")

	// All devices inherit this device API
	d.Handle("/", http.FileServer(http.FS(d.LayeredFS)))
	d.Handle("/{$}", d.showIndex())

	return d
}

// HandleDevice installs /device/{id} pattern for device in default ServeMux
func (d *Device) HandleDevice() {
	prefix := "/device/" + d.Id
	handler := BasicAuth(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
}

func (d *Device) AddChild(child *Device) error {
	d.Lock()
	defer d.Unlock()

	if _, exists := d.children[child.Id]; exists {
		return errors.New("child already exists")
	}

	d.children[child.Id] = child
	child.parent = d

	child.HandleDevice()
	return nil
}
