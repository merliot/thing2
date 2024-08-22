package main

import (
	"fmt"
	"os"

	"github.com/merliot/thing2"
	"github.com/merliot/thing2/models"
)

var device = `{
	"%s": {
		"Id": "%s",
		"Model": "hub",
		"Name": "Hub",
		"Children": [],
		"DeployParams": "target=x86-64&port=8000"
	}
}`

func main() {
	id := thing2.GenerateRandomId()
	devices := fmt.Sprintf(device, id, id)
	os.Setenv("DEVICES", devices)
	thing2.Models = models.AllModels
	thing2.Run()
}
