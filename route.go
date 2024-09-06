package thing2

import (
	"fmt"
	"sync"
)

var routes map[string]string // key: dst id, value: nexthop id
var routesMu sync.RWMutex

func _routesBuild(parent, base *Device) {
	for _, childId := range parent.Children {
		// children point to base
		routes[childId] = base.Id
		child := devices[childId]
		_routesBuild(child, base)
	}
}

func routesBuild(root *Device) {

	routesMu.RLock()
	defer routesMu.RUnlock()

	routes = make(map[string]string)

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	// root points to self
	routes[root.Id] = root.Id

	for _, childId := range root.Children {
		// children of root point to self
		routes[childId] = childId
		child := devices[childId]
		_routesBuild(child, child)
	}

	fmt.Printf("Routes %#v\n", routes)
}
