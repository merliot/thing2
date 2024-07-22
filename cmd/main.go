package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/gadget"
	"github.com/merliot/thing2/hub"
)

func main() {
	thing2.User = "user"
	thing2.Passwd = "passwd"

	/*
		thing2.SupportedModels(thing2.Models{
			"gadget": gadget.New,
			"hub":    hub.New,
		})
	*/

	hub1 := hub.New("h1", "hub01")

	g1 := gadget.New("g1", "gadget01")
	g2 := gadget.New("g2", "gadget02")
	g3 := gadget.New("g3", "gadget03")
	g4 := gadget.New("g4", "gadget04")
	g5 := gadget.New("g5", "gadget05")

	hub1.AddChild(g1)
	hub1.AddChild(g2)
	hub1.AddChild(g3)

	g4.AddChild(g5)
	hub1.AddChild(g4)

	thing2.Run(hub1, ":8080")
}
