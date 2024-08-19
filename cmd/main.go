package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/examples/gadget"
	"github.com/merliot/thing2/examples/relays"
	"github.com/merliot/thing2/hub"
)

func main() {

	thing2.Makers{
		hub.NewModel,
		gadget.NewModel,
		relays.NewModel,
	}.Register()

	thing2.Run()
}
