package gadget

import (
	"embed"
	"fmt"
	"strconv"

	"github.com/merliot/thing2"
)

//go:embed css *.go template
var fs embed.FS

type Gadget struct {
	*thing2.Device
	Bottles int
}

var targets = []string{"demo", "x86-64", "nano-rp2040"}

func New(id, name string) thing2.Devicer {
	println("NEW GADGET")
	g := &Gadget{
		Device:  thing2.NewDevice(id, "gadget", name, fs, targets),
		Bottles: 99,
	}
	g.SetData(g)
	g.Handle("/takeone", thing2.Sink(g))
	g.RHandle("/bottles", g.TemplateShow("bottles.tmpl"))
	return g
}

func (g *Gadget) Dispatch(msg *thing2.Msg) {
	fmt.Printf("Dispatch %#v\n", msg)
	switch msg.Path {
	case "/takeone":
		if g.Bottles > 0 {
			g.Bottles--
			msg.Path = "/tookone"
			msg.Set("Bottles", strconv.Itoa(g.Bottles))
			fmt.Printf("Bcast %#v\n", msg)
			//thing2.Bcast(msg)
		}
	}
}
