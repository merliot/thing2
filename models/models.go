// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/examples/gadget"
	"github.com/merliot/thing2/examples/gps"
	"github.com/merliot/thing2/hub"
	"github.com/merliot/thing2/examples/relays"
)

var AllModels = thing2.ModelMap{
	"gadget": Gadget,
	"gps": Gps,
	"hub": Hub,
	"relays": Relays,
}
var Gadget = thing2.Model{
	Package: "github.com/merliot/thing2/examples/gadget",
	Source: "https://github.com/merliot/thing2/tree/main/examples/gadget",
	Maker: gadget.NewModel,
}
var Gps = thing2.Model{
	Package: "github.com/merliot/thing2/examples/gps",
	Source: "https://github.com/merliot/thing2/tree/main/examples/gps",
	Maker: gps.NewModel,
}
var Hub = thing2.Model{
	Package: "github.com/merliot/thing2/hub",
	Source: "https://github.com/merliot/thing2/tree/main/hub",
	Maker: hub.NewModel,
}
var Relays = thing2.Model{
	Package: "github.com/merliot/thing2/examples/relays",
	Source: "https://github.com/merliot/thing2/tree/main/examples/relays",
	Maker: relays.NewModel,
}
