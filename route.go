package thing2

import (
	_ "embed"
	"net/http"
)

func RouteDown(deviceId string, msg any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkt, err := NewPacketFromURL(r.URL, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(deviceId).RouteDown()
	})
}
