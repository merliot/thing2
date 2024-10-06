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

	// Poll right away and then on ticker
	d.Lock()
	d.Poll(&Packet{Dst: d.Id})
	d.Unlock()

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c:
			return
		case <-ticker.C:
			d.Lock()
			d.Poll(&Packet{Dst: d.Id})
			d.Unlock()
		}
	}
}
