package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/examples/gadget"
	"github.com/merliot/thing2/examples/hub"
	"github.com/merliot/thing2/examples/relays"
)

func main() {

	thing2.Makers{
		hub.NewModel,
		gadget.NewModel,
		relays.NewModel,
	}.Register()

	thing2.Run()
}
