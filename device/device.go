package device

import (
	"embed"
	"errors"
	"html/template"
	"net/http"
	"sync"
)

//go:embed template
var deviceFs embed.FS

var User string
var Passwd string

type Device struct {
	Id string
	Model string
	Name string
	*http.ServeMux `json:"-"`
	LayeredFS `json:"-"`
	templates *template.Template
	parent *Device
	children map[string]*Device
	sync.RWMutex
}

func NewDevice(id, model, name string, fs embed.FS) *Device {
	println("NEW DEVICE", id, model, name)

	d := &Device{
		Id:        id,
		Model:     model,
		Name:      name,
		ServeMux:  http.NewServeMux(),
		children:  make(map[string]*Device),
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

// BasicAuthMiddleware is a middleware function for HTTP Basic Authentication
func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != User || pass != Passwd {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		println("BasicAuth", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (d *Device) HandlePrefix() {
	prefix := "/device/" + d.Id
	handler := BasicAuth(http.StripPrefix(prefix, d))
	http.Handle(prefix + "/", handler)
}

func (d *Device) AddChild(child *Device) error {
	d.Lock()
	defer d.Unlock()
	if _, exists := d.children[child.Id]; exists {
		return errors.New("child already exists")
	}

	d.children[child.Id] = child
	child.parent = d

	child.HandlePrefix()

	return nil
}
