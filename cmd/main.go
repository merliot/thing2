package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/models"
)

func main() {
	thing2.Models = models.AllModels
	thing2.Run()
}
