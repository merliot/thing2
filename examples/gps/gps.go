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

func NewModel() thing2.Devicer {
	return &Gps{Radius: 50, PollPeriod: 30}
}

func (g *Gps) GetConfig() thing2.Config {
	return thing2.Config{
		Model:      "gps",
		State:      g,
		FS:         &fs,
		Targets:    []string{"demo", "rpi", "nano-rp2040", "wioterminal"},
		BgColor:    "green",
		PollPeriod: time.Second * time.Duration(g.PollPeriod),
	}
}

func (g *Gps) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/state":  &thing2.Handler[Gps]{g.state},
		"/update": &thing2.Handler[Gps]{g.update},
	}
}

func (g *Gps) Poll(pkt *thing2.Packet) {
	lat, long := g.Location()
	dist := gps.Distance(lat, long, g.Lat, g.Long)
	if dist >= g.Radius {
		g.Lat, g.Long = lat, long
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *Gps) state(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *Gps) update(pkt *thing2.Packet) {
	pkt.Unmarshal(g).RouteUp()
}
