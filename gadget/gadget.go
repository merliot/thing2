package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed css *.go template
var fs embed.FS

type Gadget struct {
	*thing2.Device
	Bottles int
}

type Update struct {
	Bottles int
}

var targets = []string{"demo", "x86-64", "nano-rp2040"}

func New(id, name string) thing2.Devicer {
	println("NEW GADGET")

	g := &Gadget{
		Bottles: 99,
	}

	handlers := thing2.PacketHandlers{
		"/takeone": g.takeone,
		"/tookone": g.tookone,
	}

	g.Device = thing2.NewDevice(id, "gadget", name, fs, targets, handlers)
	g.SetData(g)

	g.Handle("/takeone", thing2.SendTo(g, nil))

	return g
}

func (g *Gadget) takeone(pkt *thing2.Packet) {
	println("/takeone")
	if g.Bottles > 0 {
		g.Bottles--
		msg := Update{g.Bottles}
		pkt.SetPath("/tookone").SetMsg(msg).RouteUp()
	}
}

func (g *Gadget) tookone(pkt *thing2.Packet) {
	println("/tookone")
	var update = pkt.GetMsg().(Update)
	g.Bottles = update.Bottles
}
