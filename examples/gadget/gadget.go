package gadget

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed *.go template
var fs embed.FS

type Gadget struct {
	Bottles int
}

type MsgUpdate struct {
	Bottles int
}

func NewModel() thing2.Devicer {
	return &Gadget{Bottles: 99}
}

func (g *Gadget) GetConfig() thing2.Config {
	return thing2.Config{
		Model:   "gadget",
		State:   g,
		FS:      &fs,
		Targets: []string{"demo", "x86-64", "nano-rp2040"},
	}
}

func (g *Gadget) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/state":   &thing2.Handler[Gadget]{g.state},
		"/takeone": &thing2.Handler[thing2.NoMsgType]{g.takeone},
		"/tookone": &thing2.Handler[MsgUpdate]{g.state},
	}
}

func (g *Gadget) Setup() error { return nil }
func (g *Gadget) Poll()        {}

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
