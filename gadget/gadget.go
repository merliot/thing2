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

func NewModel() thing2.Modeler {
	return &Gadget{Bottles: 99}
}

func (g *Gadget) GetModel() string     { return "gadget" }
func (g *Gadget) GetState() any        { return g }
func (g *Gadget) GetFS() *embed.FS     { return &fs }
func (g *Gadget) GetTargets() []string { return []string{"demo", "x86-64", "nano-rp2040"} }

func (g *Gadget) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/takeone": {nil, g.takeone},
		"/tookone": {&Update{}, g.tookone},
	}
}

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
