package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/gadget"
)

func main() {
	g := gadget.New("g1", "gadget01")
	g.SetDeployParams(thing2.GetEnv("DEPLOY_PARAMS", ""))
	thing2.Run(g, ":8000")
}
