package thing2

import (
	"fmt"
	"net/url"

	"github.com/go-playground/form"
)

var decoder = form.NewDecoder()

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
	if msg == nil {
		return pkt, nil
	}
	if err := decoder.Decode(pkt.Msg, url.Query()); err != nil {
		return nil, err
	}
	return pkt, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("%#v", p)
}

func (p *Packet) GetMsg() any {
	return p.Msg
}

func (p *Packet) SetMsg(msg any) *Packet {
	p.Msg = msg
	return p
}

func (p *Packet) SetDst(dst string) *Packet {
	p.Dst = dst
	return p
}

func (p *Packet) SetPath(path string) *Packet {
	p.Path = path
	return p
}

func (p *Packet) RouteDown() {
	routesMu.RLock()
	route := routes[p.Dst]
	routesMu.RUnlock()
	route.nextHop.Route(p)
}

func (p *Packet) RouteUp() {
	println("RouteUp", p.String())
	sessionsRoute(p)
}

type PacketHandler func(pkt *Packet)
type PacketHandlers map[string]PacketHandler // keyed by path (/takeone)
