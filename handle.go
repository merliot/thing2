package thing2

import (
	"fmt"
)

// TODO come up with better name than generator, because it's used for callback also
type generator interface {
	gen() any
	cb(pkt *Packet)
}

type Handler[T any] struct {
	Callback func(pkt *Packet)
}

func (h *Handler[T]) gen() any {
	var v T
	return &v
}

func (h *Handler[T]) cb(pkt *Packet) {
	h.Callback(pkt)
}

type Handlers map[string]generator // key: path

func (d *Device) handle(pkt *Packet) {
	d.Lock()
	defer d.Unlock()
	if d.IsSet(flagOnline) {
		if handler, ok := d.Handlers[pkt.Path]; ok {
			fmt.Println("Handling", pkt.String())
			handler.cb(pkt)
		}
	}
}
