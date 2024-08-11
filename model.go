package thing2

import (
	"embed"
	"fmt"
	"net/http"
)

type ModelSpec struct {
	Package string
	Maker   string
}

type ModelSpecs []ModelSpec

type Modeler interface {
	GetModel() string
	GetState() any
	GetFS() *embed.FS
	GetTargets() []string
	GetHandlers() Handlers
}

// modelInstall installs /model/{model} pattern for device in default ServeMux
func (d *Device) modelInstall() {
	prefix := "/model/" + d.Model
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	fmt.Println("modelInstall", prefix)
}

func modelsInstall() {
	for model, maker := range makers {
		var proto = &Device{Model: model}
		proto.build(maker)
		proto.modelInstall()
	}
}
