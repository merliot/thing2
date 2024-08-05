package thing2

import (
	_ "embed"
	"fmt"
	"net/http"
	"sync"
)

//go:embed template/routes.tmpl
var routesTemplate string

type routeMap map[string]string // key: dst dev Id; value: nextHop dev Id

var routes routeMap
var routesMu sync.RWMutex

func buildChildRoutes(parent Devicer, base Devicer) {
	for id, child := range parent.GetChildren() {
		routes[id] = base.GetId()
		buildChildRoutes(child, base)
	}
}

func BuildRoutes(root Devicer) {
	routesMu.Lock()
	defer routesMu.Unlock()
	routes = make(routeMap)
	routes[root.GetId()] = root.GetId()
	for id, child := range root.GetChildren() {
		routes[id] = child.GetId()
		buildChildRoutes(child, child)
	}
	fmt.Println("Built Routes:", routes)
}

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

func routesShow(w http.ResponseWriter, r *http.Request) {
	routesMu.RLock()
	defer routesMu.RUnlock()
	templateShow(w, routesTemplate, routes)
}
