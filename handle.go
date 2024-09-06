package thing2

import (
	"fmt"
)

// TODO come up with better name than Generator, because it's used for callback also
type Generator interface {
	Gen() any
	Cb(pkt *Packet)
}

type Handler[T any] struct {
	Callback func(pkt *Packet)
}

func (h *Handler[T]) Gen() any {
	var v T
	return &v
}

func (h *Handler[T]) Cb(pkt *Packet) {
	h.Callback(pkt)
}

type Handlers map[string]Generator // key: path

func (d *Device) handle(pkt *Packet) {
	d.Lock()
	defer d.Unlock()
	if d.Flags.IsSet(flagOnline) {
		if handler, ok := d.Handlers[pkt.Path]; ok {
			fmt.Println("Handling", pkt.String())
			handler.Cb(pkt)
		}
	}
}
