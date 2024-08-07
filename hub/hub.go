package hub

import (
	"embed"

	"github.com/merliot/thing2"
)

//go:embed template
var fs embed.FS

type Hub struct {
}

func NewModel() thing2.Modeler {
	return &Hub{}
}

func (h *Hub) GetModel() string     { return "hub" }
func (h *Hub) GetFS() *embed.FS     { return &fs }
func (h *Hub) GetTargets() []string { return []string{"demo", "x86-64", "rpi"} }
