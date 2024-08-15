package relays

import (
	"embed"

	"github.com/merliot/thing2"
	"github.com/merliot/thing2/io/relay"
)

//go:embed css images *.go template
var fs embed.FS

type Relays struct {
	Relays [4]relay.Relay
}

type MsgClick struct {
	Relay int
	State bool
}

func NewModel() thing2.Modeler {
	return &Relays{}
}

func (r *Relays) GetModel() string     { return "relays" }
func (r *Relays) GetState() any        { return r }
func (r *Relays) GetFS() *embed.FS     { return &fs }
func (r *Relays) GetTargets() []string { return []string{"demo", "rpi", "nano-rp2040", "wioterminal"} }

func (r *Relays) GetHandlers() thing2.Handlers {
	return thing2.Handlers{
		"/state":   &thing2.Handler[Relays]{r.state},
		"/click":   &thing2.Handler[MsgClick]{r.click},
		"/clicked": &thing2.Handler[MsgClick]{r.clicked},
	}
}

func (r *Relays) state(pkt *thing2.Packet) {
	pkt.Unmarshal(r).RouteUp()
}

func (r *Relays) click(pkt *thing2.Packet) {
}

func (r *Relays) clicked(pkt *thing2.Packet) {
}
