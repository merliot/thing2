package hub

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed template
var fs embed.FS

type Hub struct {
	*thing2.Device
}

func New(id, name string) thing2.Devicer {
	println("NEW HUB")
	h := &Hub{
		Device: thing2.NewDevice(id, "hub", name, fs),
	}
	h.SetData(h)
	return h
}
