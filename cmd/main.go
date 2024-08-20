package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/models"
)

func main() {
	thing2.Models = models.Models
	thing2.Run()
}
