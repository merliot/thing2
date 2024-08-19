package hub

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed css images template
var fs embed.FS

type Hub struct {
	Demo bool `json:"-"`
}

func NewModel() thing2.Devicer {
	return &Hub{}
}

func (h *Hub) GetConfig() thing2.Config {
	return thing2.Config{
		Model:   "hub",
		State:   h,
		FS:      &fs,
		Targets: []string{"demo", "x86-64", "rpi"},
	}
}

func (h *Hub) GetHandlers() thing2.Handlers {
	return thing2.Handlers{}
}
