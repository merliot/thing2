//go:build !rpi && !tinygo

package thing2

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (d *Device) run() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		var pkt = Packet{Dst: d.Id}
		select {
		case <-c:
			return
		case <-ticker.C:
			d.Lock()
			d.Poll(&pkt)
			d.Unlock()
		}
	}
}
