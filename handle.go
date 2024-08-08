package thing2

import "net/http"

type Handler struct {
	MsgScheme any
	Callback  func(pkt *Packet)
}

type Handlers map[string]Handler // key: path

func (d *Device) handlersInstall() {
	for path, handler := range d.Handlers {
		d.Handle("POST "+path, d.newPacketRoute(handler.MsgScheme))
	}
}

func (d *Device) newPacketRoute(msgScheme any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkt, err := NewPacketFromURL(r.URL, msgScheme)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.Id).RouteDown()
	})
}
