package thing2

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-playground/form"
)

var decoder = form.NewDecoder()

type NextHop Devicer
type Routes map[string]NextHop // dst ID: next hop to get dst

var routes Routes
var routesMu sync.RWMutex

func (r Routes) String() string {
	m := make(map[string]string)
	for id, nextHop := range r {
		m[id] = nextHop.GetId()
	}
	return fmt.Sprintf("%s", m)
}

func buildChildRoutes(parent Devicer, path Devicer) {
	for id, child := range parent.GetChildren() {
		routes[id] = path
		buildChildRoutes(child, path)
	}
}

func BuildRoutes(root Devicer) {
	routesMu.Lock()
	defer routesMu.Unlock()
	routes = make(Routes)
	routes[root.GetId()] = root
	for id, child := range root.GetChildren() {
		routes[id] = child
		buildChildRoutes(child, child)
	}
	fmt.Println("Built Routes:", routes)
}

type Packet struct {
	Dst  string
	Path string
	Msg  any
}

func NewPacketFromURL(url *url.URL, msg any) (*Packet, error) {
	var pkt = &Packet{
		Path: url.Path,
		Msg:  msg,
	}
	if err := decoder.Decode(pkt.Msg, url.Query()); err != nil {
		return nil, err
	}
	return pkt, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("%#v", p)
}

func (p *Packet) SetDst(dst string) *Packet {
	p.Dst = dst
	return p
}

func (p *Packet) Route() {
	routesMu.RLock()
	nextHop := routes[p.Dst]
	routesMu.RUnlock()
	nextHop.Route(p)
}

type PacketHandler func(pkt *Packet)
type PacketHandlers map[string]PacketHandler // keyed by path (/takeone)

func SendTo(d Devicer, msg any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pkt, err := NewPacketFromURL(r.URL, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pkt.SetDst(d.GetId()).Route()
	})
}
