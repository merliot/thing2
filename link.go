package thing2

import "sync"

type linker interface {
	Send(pkt *Packet)
}

type links map[string]linker // keyed by device Id

var uplinks = new(links)
var uplinksMu sync.RWMutex

var downlinks = new(links)
var downlinksMu sync.RWMutex
