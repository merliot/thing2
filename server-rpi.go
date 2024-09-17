//go:build rpi

package thing2

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

func (d *Device) run() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(d.PollFreq)
	defer ticker.Stop()

	for {
		var pkt = Packet{Dst: d.Id}
		select {
		case <-c:
			failSafe()
			return
		case <-ticker.C:
			d.Lock()
			d.Poll(&pkt)
			d.Unlock()
		}
	}
}
