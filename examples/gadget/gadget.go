package gadget

import (
	"embed"
	"time"

	"github.com/merliot/thing2"
)

//go:embed *.go template
var fs embed.FS

type Gadget struct {
	Bottles int
	Restock int
}

func NewModel() thing2.Devicer {
	return &Gadget{Bottles: 99, Restock: 70}
}

func (g *Gadget) GetConfig() thing2.Config {
	return thing2.Config{
		Model:    "gadget",
		State:    g,
		FS:       &fs,
		Targets:  []string{"demo", "x86-64", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:  "sunflower",
	}
}

func (g *Gadget) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/state":   &thing2.Handler[Gadget]{g.state},
		"/takeone": &thing2.Handler[thing2.NoMsg]{g.takeone},
		"/update":  &thing2.Handler[Gadget]{g.state},
	}
}

func (g *Gadget) Setup() error { return nil }

func (g *Gadget) Poll(pkt *thing2.Packet) {
	if g.Bottles < 99 {
		if g.Restock == 1 {
			g.Bottles = 99
			g.Restock = 70
		} else {
			g.Restock--
		}
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *Gadget) state(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *Gadget) takeone(pkt *thing2.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}
