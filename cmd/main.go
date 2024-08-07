package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/gadget"
	"github.com/merliot/thing2/hub"
)

func main() {

	thing2.Makers{
		hub.NewModel,
		gadget.NewModel,
	}.Register()

	thing2.Run()
}
