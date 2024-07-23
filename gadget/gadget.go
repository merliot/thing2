package gadget

import (
	"embed"
	"net/http"

	"github.com/merliot/thing2"
)

//go:embed css *.go template
var fs embed.FS

type Gadget struct {
	*thing2.Device
	Bottles int
}

var targets = []string{"demo", "x86-64", "nano-rp2040"}

func New(id, name string) thing2.Devicer {
	println("NEW GADGET")
	g := &Gadget{
		Device: thing2.NewDevice(id, "gadget", name, fs, targets),
	}
	g.SetData(g)
	g.HandleFunc("/takeone", g.takeone)
	g.Handle("/bottles", g.TemplateShow("bottles.tmpl"))
	return g
}

func (g *Gadget) takeone(w http.ResponseWriter, r *http.Request) {
	g.Lock()
	defer g.Unlock()
	if g.Bottles > 0 {
		g.Bottles--
		//g.Inject(pkt.SetPath("tookone"))
	}
}
