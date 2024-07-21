package hub

import (
	"embed"

	"github.com/merliot/thing2/device"
)

//go:embed template
var fs embed.FS

type Hub struct {
	*device.Device
}

func NewHub(id, model, name string) *Hub {
	println("NEW HUB")
	return &Hub{
		Device: device.NewDevice(id, model, name, fs),
	}
}
