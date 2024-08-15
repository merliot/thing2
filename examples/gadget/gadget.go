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

type MsgUpdate struct {
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
		"/state":   &thing2.Handler[Gadget]{g.state},
		"/takeone": &thing2.Handler[thing2.NoMsgType]{g.takeone},
		"/tookone": &thing2.Handler[MsgUpdate]{g.state},
	}
}

func (g *Gadget) state(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *Gadget) takeone(pkt *thing2.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		msg := MsgUpdate{g.Bottles}
		pkt.SetPath("/tookone").Marshal(msg).RouteUp()
	}
}
