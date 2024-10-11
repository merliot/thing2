package gps

import (
	"time"

	"github.com/merliot/thing2"
	"github.com/merliot/thing2/io/gps"
)

type Gps struct {
	Lat        float64
	Long       float64
	Radius     float64 // units: meters
	PollPeriod uint    // units: seconds
	gps.Gps
}

type updateMsg struct {
	Lat  float64
	Long float64
}

func NewModel() thing2.Devicer {
	return &Gps{Radius: 50, PollPeriod: 30}
}

func (g *Gps) GetConfig() thing2.Config {
	return thing2.Config{
		Model:      "gps",
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi", "nano-rp2040", "wioterminal"},
		BgColor:    "green",
		PollPeriod: time.Second * time.Duration(g.PollPeriod),
	}
}

func (g *Gps) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/state":  &thing2.Handler[Gps]{g.state},
		"/update": &thing2.Handler[updateMsg]{g.update},
	}
}

func (g *Gps) Poll(pkt *thing2.Packet) {
	lat, long, _ := g.Location()
	dist := gps.Distance(lat, long, g.Lat, g.Long)
	if dist >= g.Radius {
		var up = updateMsg{lat, long}
		g.Lat, g.Long = lat, long
		pkt.SetPath("/update").Marshal(&up).RouteUp()
	}
}

func (g *Gps) state(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *Gps) update(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *Gps) DemoSetup() error            { return nil }
func (g *Gps) DemoPoll(pkt *thing2.Packet) {}
