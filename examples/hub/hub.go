package hub

import (
	"embed"
	"fmt"
	"net/url"

	"github.com/merliot/thing2"
)

//go:embed css images template
var fs embed.FS

type Hub struct {
	Demo bool `json:"-"`
}

func NewModel() thing2.Modeler {
	return &Hub{}
}

func (h *Hub) Config(cfg url.Values)        { fmt.Printf("%#v\n", cfg) }
func (h *Hub) GetModel() string             { return "hub" }
func (h *Hub) GetState() any                { return h }
func (h *Hub) GetFS() *embed.FS             { return &fs }
func (h *Hub) GetTargets() []string         { return []string{"demo", "x86-64", "rpi"} }
func (h *Hub) GetHandlers() thing2.Handlers { return thing2.Handlers{} }
