package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/models"
)

var device = `{
	"hub": {
		"Id": "hub,
		"Model": "hub",
		"Name": "Hub",
		"Children": [],
		"DeployParams": "target=x86-64&port=8000"
	}
}`

func main() {
	thing2.Setenv("DEVICES", thing2.Getenv("DEVICES", device))
	thing2.Models = models.AllModels
	thing2.Run()
}
