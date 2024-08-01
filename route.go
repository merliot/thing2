package thing2

import (
	_ "embed"
	"fmt"
	"net/http"
	"sync"
)

//go:embed template/routes.tmpl
var routesTemplate string

type route struct {
	nextHop Devicer
}

type routeMap map[string]*route // keyed by dst Id

var routes routeMap
var routesMu sync.RWMutex

func newRoute(nextHop Devicer) *route {
	return &route{nextHop}
}

func (r routeMap) String() string {
	m := make(map[string]string)
	for id, route := range r {
		m[id] = route.nextHop.GetId()
	}
	return fmt.Sprintf("%s", m)
}

func buildChildRoutes(parent Devicer, base Devicer) {
	for id, child := range parent.GetChildren() {
		routes[id] = newRoute(base)
		buildChildRoutes(child, base)
	}
}

func BuildRoutes(root Devicer) {
	routesMu.Lock()
	defer routesMu.Unlock()
	routes = make(routeMap)
	routes[root.GetId()] = newRoute(root)
	for id, child := range root.GetChildren() {
		routes[id] = newRoute(child)
		buildChildRoutes(child, child)
	}
	fmt.Println("Built Routes:", routes)
}

func SendTo(d Devicer, msg any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkt, err := NewPacketFromURL(r.URL, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.GetId()).RouteDown()
	})
}

func routesShow(w http.ResponseWriter, r *http.Request) {
	routesMu.RLock()
	defer routesMu.RUnlock()
	templateShow(w, routesTemplate, routes)
}
