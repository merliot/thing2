// This file auto-generated from ./cmd/gen-models

package models

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/examples/gadget"
	"github.com/merliot/thing2/hub"
	"github.com/merliot/thing2/examples/relays"
)

var Models = map[string]thing2.Model{
	"gadget": {
		Package: "github.com/merliot/thing2/examples/gadget",
		Maker: gadget.NewModel,
	},
	"hub": {
		Package: "github.com/merliot/thing2/hub",
		Maker: hub.NewModel,
	},
	"relays": {
		Package: "github.com/merliot/thing2/examples/relays",
		Maker: relays.NewModel,
	},
}
