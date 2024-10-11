package relays

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/io/relay"
)

type Relays struct {
	Relays [4]relay.Relay
}

type MsgClick struct {
	Relay int
}

type MsgClicked struct {
	Relay int
	State bool
}

func NewModel() thing2.Devicer {
	return &Relays{}
}

func (r *Relays) GetConfig() thing2.Config {
	return thing2.Config{
		Model:   "relays",
		State:   r,
		FS:      &fs,
		Targets: []string{"rpi", "nano-rp2040", "wioterminal"},
		BgColor: "orange",
	}
}

func (r *Relays) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/state":   &thing2.Handler[Relays]{r.state},
		"/click":   &thing2.Handler[MsgClick]{r.click},
		"/clicked": &thing2.Handler[MsgClicked]{r.clicked},
	}
}

func (r *Relays) Setup() error {
	for i := range r.Relays {
		relay := &r.Relays[i]
		if err := relay.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relays) Poll(pkt *thing2.Packet) {
}

func (r *Relays) state(pkt *thing2.Packet) {
	pkt.Unmarshal(r).RouteUp()
}

func (r *Relays) click(pkt *thing2.Packet) {
	var click MsgClick
	pkt.Unmarshal(&click)
	relay := &r.Relays[click.Relay]
	relay.Set(!relay.State)
	var clicked = MsgClicked{click.Relay, relay.State}
	pkt.SetPath("/clicked").Marshal(&clicked).RouteUp()
}

func (r *Relays) clicked(pkt *thing2.Packet) {
	var clicked MsgClicked
	pkt.Unmarshal(&clicked)
	relay := &r.Relays[clicked.Relay]
	relay.Set(clicked.State)
	pkt.RouteUp()
}

func (r *Relays) DemoSetup() error            { return r.Setup() }
func (r *Relays) DemoPoll(pkt *thing2.Packet) { r.Poll(pkt) }
