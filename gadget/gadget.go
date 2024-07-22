package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed css *.go template
var fs embed.FS

type Gadget struct {
	*thing2.Device
}

var targets = []string{"demo", "x86-64", "nano-rp2040"}

func New(id, name string) thing2.Devicer {
	println("NEW GADGET")
	g := &Gadget{
		Device: thing2.NewDevice(id, "gadget", name, fs, targets),
	}
	g.SetData(g)
	return g
}
