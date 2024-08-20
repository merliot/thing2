// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/examples/gadget"
	"github.com/merliot/thing2/examples/relays"
	"github.com/merliot/thing2/hub"
)

var AllModels = thing2.ModelMap{
	"gadget": Gadget,
	"hub":    Hub,
	"relays": Relays,
}
var Gadget = thing2.Model{
	Package: "github.com/merliot/thing2/examples/gadget",
	Maker:   gadget.NewModel,
}
var Hub = thing2.Model{
	Package: "github.com/merliot/thing2/hub",
	Maker:   hub.NewModel,
}
var Relays = thing2.Model{
	Package: "github.com/merliot/thing2/examples/relays",
	Maker:   relays.NewModel,
}
