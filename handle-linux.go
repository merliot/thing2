//go:build !tinygo

package thing2

import "net/http"

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
		pkt, err := newPacketFromURL(r.URL, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.Id).RouteDown()
	})
}
