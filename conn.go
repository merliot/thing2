package thing2

import "sync"

type Conn interface {
	String() string
}

var conns = make(map[Conn]bool)
var connsMu sync.RWMutex

func Plugin(c Conn) {
	println("Plugin", c.String())
	connsMu.Lock()
	defer connsMu.Unlock()
	conns[c] = true
}

func Unplug(c Conn) {
	println("Unplug", c.String())
	connsMu.Lock()
	defer connsMu.Unlock()
	delete(conns, c)
}
