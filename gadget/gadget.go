package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed css template
var fs embed.FS

type Gadget struct {
	*thing2.Device
}

func New(id, name string) thing2.Devicer {
	println("NEW GADGET")
	g := &Gadget{
		Device: thing2.NewDevice(id, "gadget", name, fs),
	}
	g.SetData(g)
	return g
}
