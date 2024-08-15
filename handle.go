package thing2

import (
	"fmt"
	"net/http"
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

type NoMsgType struct{}

func (d *Device) handlersInstall() {
	for path, handler := range d.Handlers {
		if path == "/state" {
			// Special case /state to return a state page
			d.HandleFunc("GET "+path, d.showState)
			continue
		}
		d.Handle("POST "+path, d.newPacketRoute(handler))
	}
}

func (d *Device) newPacketRoute(h Generator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := h.Gen()
		pkt, err := NewPacketFromURL(r.URL, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.Id).RouteDown()
	})
}

func (d *Device) handle(pkt *Packet) {
	d.Lock()
	defer d.Unlock()
	if handler, ok := d.Handlers[pkt.Path]; ok {
		fmt.Println("Handling", pkt.String())
		handler.Cb(pkt)
	}
}
