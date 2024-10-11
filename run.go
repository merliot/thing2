package thing2

import "time"

func (d *Device) runDemo() {

	// Poll right away and then on ticker
	d.Lock()
	d.DemoPoll(&Packet{Dst: d.Id})
	d.Unlock()

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.Lock()
			d.DemoPoll(&Packet{Dst: d.Id})
			d.Unlock()
		case <-d.stopChan:
			return
		}
	}
}

// In demo mode, start a go func for each child device
func (d *Device) startDemoChildren() {
	for _, childId := range d.Children {
		child := devices[childId]
		go child.runDemo()
		child.startDemoChildren()
	}
}

func (d *Device) stopDemoChildren() {
	for _, childId := range d.Children {
		child := devices[childId]
		close(child.stopChan)
		child.stopDemoChildren()
	}
}

func (d *Device) run() {
	if runningDemo {
		d.startDemoChildren()
		d.runPolling(d.DemoPoll)
		d.stopDemoChildren()
	} else {
		d.runPolling(d.Poll)
	}
}
