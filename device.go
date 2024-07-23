package thing2

import (
	"embed"
	"errors"
	"fmt"
	"html"
	"html/template"
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
	AddChild(child Devicer) error
	SetParent(parent Devicer)
	SetDeployParams(params string)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Children map[string]Devicer
type Maker func(id, name string) Devicer
type Models []string
type WifiAuth map[string]string // key: ssid; value: passphrase

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
	// Targets supported by device
	target.Targets `json:"-"`
	// WifiAuth is a map of SSID:PASSPHRASE pairs
	WifiAuth `json:"-"`
	// DeployParams is device deploy configuration in an html param format
	DeployParams string
	// Administratively locked
	Locked bool `json:"-"`
}

func NewDevice(id, model, name string, fs embed.FS, targets []string) *Device {
	println("NEW DEVICE", id, model, name)

	d := &Device{
		Id:       id,
		Model:    model,
		Name:     name,
		ServeMux: http.NewServeMux(),
		Children: make(Children),
		Targets:  target.MakeTargets(targets),
		WifiAuth: make(WifiAuth),
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
	d.API()

	return d
}

func (d Device) GetId() string    { return d.Id }
func (d Device) GetModel() string { return d.Model }
func (d *Device) SetData(data any) {
	d.data = data
}
func (d *Device) SetParent(parent Devicer) { d.parent = parent }

// Install /device/{id} pattern for device in default ServeMux
func InstallDevicePattern(d Devicer) {
	prefix := "/device/" + d.GetId()
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	fmt.Printf("InstallDevicePattern %s %#v\n", prefix, d)
}

var modelPatterns = make(map[string]string)

// Install /model/{model} pattern for device in default ServeMux
func InstallModelPattern(d Devicer) {
	// But only if it doesn't already exist
	if _, exists := modelPatterns[d.GetModel()]; !exists {
		prefix := "/model/" + d.GetModel()
		handler := basicAuthHandler(http.StripPrefix(prefix, d))
		http.Handle(prefix+"/", handler)
		modelPatterns[d.GetModel()] = prefix
		fmt.Printf("InstallModelPattern %s %#v\n", prefix, d)
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
	InstallDevicePattern(child)

	// Install the /model/{model} pattern, using child as proto (but only
	// if we haven't seen this model before)
	InstallModelPattern(child)

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
