//go:build rpi

package thing2

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/merliot/thing2/target"
	"gobot.io/x/gobot/v2/drivers/gpio"
)

// failSafe by turning off all gpios
func failSafe() {
	for _, pin := range target.AllTargets["rpi"].GpioPins {
		rpin := strconv.Itoa(int(pin))
		driver := gpio.NewDirectPinDriver(target.GetAdaptor(), rpin)
		driver.Start()
		driver.Off()
	}

}

func run() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-c:
		failSafe()
	}
}
