package thing2

import (
	"embed"
	"errors"
	"html/template"
	"net/http"
	"sync"
)

//go:embed css images js template
var deviceFs embed.FS

type Devicer interface {
	GetId() string
	GetModel() string
	AddChild(child Devicer) error
	SetParent(parent Devicer)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	InstallDevicePattern()
	InstallModelPattern()
}

type Children map[string]Devicer
type Maker func(id, name string) Devicer
type Models []string

type Device struct {
	Id             string
	Model          string
	Name           string
	*http.ServeMux `json:"-"`
	LayeredFS      `json:"-"`
	// Data passed to render templates
	data      any
	templates *template.Template
	sync.RWMutex
	parent   Devicer
	Children `json:"-"`
	Models   `json:"-"`
	// DeployParams are device deploy configuration in an html param format
	DeployParams string
	// Administratively locked
	Locked bool `json:"-"`
}

func NewDevice(id, model, name string, fs embed.FS) *Device {
	println("NEW DEVICE", id, model, name)

	d := &Device{
		Id:       id,
		Model:    model,
		Name:     name,
		ServeMux: http.NewServeMux(),
		Children: make(Children),
	}
	d.data = d

	// Build device's layered FS.  fs is stacked on top of deviceFs, so
	// fs:foo.tmpl will override deviceFs:foo.tmpl, when searching for file
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

func (d Device) GetId() string             { return d.Id }
func (d Device) GetModel() string          { return d.Model }
func (d *Device) SetData(data any)         { d.data = data }
func (d *Device) SetParent(parent Devicer) { d.parent = parent }

// Install /device/{id} pattern for device in default ServeMux
func (d Device) InstallDevicePattern() {
	prefix := "/device/" + d.Id
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	println("InstallDevicePattern", prefix)
}

var modelPatterns = make(map[string]string)

// Install /model/{model} pattern for device in default ServeMux
func (d Device) InstallModelPattern() {
	// But only if it doesn't already exist
	if _, exists := modelPatterns[d.Model]; !exists {
		prefix := "/model/" + d.Model
		handler := basicAuthHandler(http.StripPrefix(prefix, d))
		http.Handle(prefix+"/", handler)
		modelPatterns[d.Model] = prefix
		println("InstallModelPattern", prefix)
	}
}

func (d *Device) AddChild(child Devicer) error {

	d.Lock()
	defer d.Unlock()

	if _, exists := d.Children[child.GetId()]; exists {
		return errors.New("child already exists")
	}

	d.Children[child.GetId()] = child
	child.SetParent(d)

	// Install the /device/{id} pattern for child
	child.InstallDevicePattern()

	// Install the /model/{model} pattern, using child as proto (but only
	// if we haven't seen this model before)
	child.InstallModelPattern()

	return nil
}
