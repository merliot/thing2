package hub

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed images *.go template
var fs embed.FS

type Hub struct {
}

func NewModel() thing2.Devicer {
	return &Hub{}
}

func (h *Hub) GetConfig() thing2.Config {
	return thing2.Config{
		Model:   "hub",
		Flags:   thing2.FlagProgenitive,
		State:   h,
		FS:      &fs,
		Targets: []string{"demo", "x86-64", "rpi"},
	}
}

func (h *Hub) GetHandlers() thing2.Handlers {
	return thing2.Handlers{}
}

func (h *Hub) Setup() error            { return nil }
func (h *Hub) Poll(pkt *thing2.Packet) {}
