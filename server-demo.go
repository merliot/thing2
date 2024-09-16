//go:build !rpi && !tinygo

package thing2

import (
	"time"
)

func (d *Device) run() {
	ticker := time.NewTicker(d.PollFreq)
	for {
		var pkt = Packet{Dst: d.Id}
		select {
		case <-ticker.C:
			d.Lock()
			d.Poll(&pkt)
			d.Unlock()
		}
	}
}
