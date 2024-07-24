package thing2

import (
	"net/http"
	"net/url"
	"sync"
)

var bus = make(map[string]Devicer)
var busMu sync.RWMutex

func registerDevice(d Devicer) {
	busMu.Lock()
	defer busMu.Unlock()
	bus[d.GetId()] = d
}

type Msg struct {
	Dst  string
	Path string
	url.Values
}

func Sink(d Devicer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := &Msg{
			Dst:    d.GetId(),
			Path:   r.URL.Path,
			Values: r.URL.Query(),
		}
		if d.Parent() == nil {
			d.Dispatch(msg)
		} else {
		}
	})
}
