package thing2

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-playground/form"
)

type NoMsg struct{}

var decoder = form.NewDecoder()

type Packet struct {
	Dst  string
	Path string
	Msg  json.RawMessage
}

func newPacketFromURL(url *url.URL, v any) (*Packet, error) {
	var pkt = &Packet{
		Path: url.Path,
	}
	if _, ok := v.(*NoMsg); ok {
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
	var msg any
	json.Unmarshal(p.Msg, &msg)
	return fmt.Sprintf("[%s%s] %v", p.Dst, p.Path, msg)
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
	if len(p.Msg) > 0 {
		if err := json.Unmarshal(p.Msg, v); err != nil {
			fmt.Printf("JSON unmarshal error %s\r\n", err.Error())
		}
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
	fmt.Println("RouteDown", p)
	downlinksRoute(p)
}

func (p *Packet) RouteUp() {
	fmt.Println("RouteUp", p)
	sessionsRoute(p)
	uplinksRoute(p)
}
