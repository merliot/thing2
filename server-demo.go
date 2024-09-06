//go:build !rpi && !tinygo

package thing2

import (
	"time"
)

func (d *Device) run() {
	ticker := time.NewTicker(d.PollFreq)
	for {
		select {
		case <-ticker.C:
			d.Lock()
			d.Poll()
			d.Unlock()
		}
	}
}
