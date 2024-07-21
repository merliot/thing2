package gadget

import (
	"embed"

	"github.com/merliot/thing2/device"
)

//go:embed template
var fs embed.FS

type Gadget struct {
	*device.Device
}

func NewGadget(id, model, name string) *Gadget {
	println("NEW GADGET")
	return &Gadget{
		Device: device.NewDevice(id, model, name, fs),
	}
}
