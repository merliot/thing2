package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed css *.go template
var fs embed.FS

type Gadget struct {
	Bottles int
}

type Update struct {
	Bottles int
}

func (g *Gadget) GetModel() string     { return "gadget" }
func (g *Gadget) GetFS() *embed.FS     { return &fs }
func (g *Gadget) GetTargets() []string { return []string{"demo", "x86-64", "nano-rp2040"} }
func (g *Gadget) GetData() any         { return g }

func NewModel() thing2.Modeler {
	return &Gadget{}
}

/*
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

	g.Handle("/takeone", thing2.RouteDown(g.Id, nil))

	return g
}
*/

func (g *Gadget) takeone(pkt *thing2.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		msg := Update{g.Bottles}
		pkt.SetPath("/tookone").Marshal(msg).RouteUp()
	}
}

func (g *Gadget) tookone(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}
