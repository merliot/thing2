package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed template
var fs embed.FS

type Gadget struct {
	*thing2.Device
}

func NewGadget(id, model, name string) *Gadget {
	println("NEW GADGET")
	return &Gadget{
		Device: thing2.NewDevice(id, model, name, fs),
	}
}
