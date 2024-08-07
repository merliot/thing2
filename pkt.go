package thing2

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-playground/form"
)

var decoder = form.NewDecoder()

type Packet struct {
	Dst  string
	Path string
	Msg  json.RawMessage
}

func NewPacketFromURL(url *url.URL, v any) (*Packet, error) {
	var pkt = &Packet{
		Path: url.Path,
	}
	if v == nil {
		return pkt, nil
	}
	err := decoder.Decode(v, url.Query())
	if err != nil {
		return nil, err
	}
	pkt.Msg, err = json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return pkt, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("%#v", p)
}

// Marshal the packet message payload as JSON from v
func (p *Packet) Marshal(v any) *Packet {
	var err error
	p.Msg, err = json.Marshal(v)
	if err != nil {
		fmt.Printf("JSON marshal error %s\r\n", err.Error())
	}
	return p
}

// Unmarshal the packet message payload as JSON into v
func (p *Packet) Unmarshal(v any) *Packet {
	if err := json.Unmarshal(p.Msg, v); err != nil {
		fmt.Printf("JSON unmarshal error %s\r\n", err.Error())
	}
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
	/*
		routesMu.RLock()
		nexthop := routes[p.Dst]
		routesMu.RUnlock()
		deviceRouteDown(nexthop, p)
	*/
}

func (p *Packet) RouteUp() {
	println("RouteUp", p.String())
	sessionsRoute(p.Dst)
	uplinksRoute(p)
}

type PacketHandler func(pkt *Packet)
type PacketHandlers map[string]PacketHandler // keyed by path
