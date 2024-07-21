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

func NewHub(id, model, name string) *Hub {
	println("NEW HUB")
	return &Hub{
		Device: thing2.NewDevice(id, model, name, fs),
	}
}
