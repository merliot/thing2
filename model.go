package thing2

import (
	"embed"
	"fmt"
	"net/http"
)

type Modeler interface {
	GetModel() string
	GetFS() *embed.FS
	GetTargets() []string
}

// modelInstall installs /model/{model} pattern for device in default ServeMux
func (d *Device) modelInstall() {
	prefix := "/model/" + d.Model
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	fmt.Println("modelInstall", prefix)
}

// modelInstall installs /model/{model} patterns in default ServeMux for all
// models
func modelsInstall() {
	for model, maker := range makers {
		var proto = &Device{
			Model:   model,
			Modeler: maker(),
		}
		proto.modelInstall()
	}
}
