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
	GetChildren() Children
	GetParent() Devicer
	SetParent(parent Devicer)
	SetDeployParams(params string)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	InstallDevicePattern()
	InstallModelPattern()
	Route(pkt *Packet)
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
	parent         Devicer
	children       Children
	handlers       PacketHandlers
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

	// Build device's layered FS.  fs is stacked on top of deviceFs, so
	// fs:foo.tmpl will override deviceFs:foo.tmpl, when searching for file
	// foo.tmpl.
	d.LayeredFS.Stack(deviceFs)
	d.LayeredFS.Stack(fs)

	// Build the device templates
	d.templates = d.LayeredFS.ParseFS("template/*.tmpl")

	// All devices inherit this base device API
	d.API()

	return d
}

func (d Device) GetId() string             { return d.Id }
func (d Device) GetModel() string          { return d.Model }
func (d Device) GetParent() Devicer        { return d.parent }
func (d *Device) SetParent(parent Devicer) { d.parent = parent }
func (d *Device) GetChildren() Children    { return d.children }
func (d *Device) SetData(data any)         { d.data = data }
func (d *Device) Dispatch(pkt *Packet)     {}
func (d *Device) String() string           { return d.Id + ":" + d.Model + ":" + d.Name }

func (d *Device) AddChild(child Devicer) error {

	d.Lock()
	defer d.Unlock()

	if _, exists := d.children[child.GetId()]; exists {
		return errors.New("child already exists")
	}

	d.children[child.GetId()] = child
	child.SetParent(d)

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

func (d *Device) Route(pkt *Packet) {
	fmt.Println(d, pkt)
	if pkt.Dst == d.Id {
		if h, ok := d.handlers[pkt.Path]; ok {
			h(pkt)
		}
		return
	}
}
