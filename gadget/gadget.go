package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed css *.go template
var fs embed.FS

type Bottles int

func (b *Bottles) Htmx() {
}

type Gadget struct {
	*thing2.Device
	Bottles
}

type Takeone struct {
	Foo int
}

var targets = []string{"demo", "x86-64", "nano-rp2040"}

func New(id, name string) thing2.Devicer {
	println("NEW GADGET")

	g := &Gadget{
		Bottles: 99,
	}

	handlers := thing2.PacketHandlers{
		"/takeone": g.takeone,
		"/bottles": g.bottles,
	}

	g.Device = thing2.NewDevice(id, "gadget", name, fs, targets, handlers)
	g.SetData(g)

	g.Handle("/takeone", thing2.SendTo(g, &Takeone{}))
	g.RHandle("/bottles", g.TemplateShow("bottles.tmpl"))

	return g
}

func (g *Gadget) takeone(pkt *thing2.Packet) {
	println("/takeone")
	if g.Bottles > 0 {
		g.Bottles--
		//BcastUp("/bottles", g)
	}
}

func (g *Gadget) bottles(pkt *thing2.Packet) {
	println("/bottles")
}
