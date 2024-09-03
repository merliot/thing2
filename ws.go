package thing2

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"
)

type wsLink struct {
	conn     *websocket.Conn
	lastRecv time.Time
	lastSend time.Time
}

type announcement struct {
	Id           string
	Model        string
	Name         string
	DeployParams template.HTML
}

// ws handles /ws requests
func wsHandle(w http.ResponseWriter, r *http.Request) {
	serv := websocket.Server{Handler: websocket.Handler(wsServer)}
	serv.ServeHTTP(w, r)
}

func (l *wsLink) Send(pkt *Packet) error {
	data, err := json.Marshal(pkt)
	if err != nil {
		return fmt.Errorf("Marshal error: %w", err)
	}
	if err := websocket.Message.Send(l.conn, string(data)); err != nil {
		return fmt.Errorf("Send error: %w", err)
	}
	l.lastSend = time.Now()
	return nil
}

func (l *wsLink) Close() {
	l.conn.Close()
}

func (l *wsLink) receive() (*Packet, error) {
	var data []byte
	var pkt Packet

	if err := websocket.Message.Receive(l.conn, &data); err != nil {
		return nil, err
	}
	l.lastRecv = time.Now()
	if err := json.Unmarshal(data, &pkt); err != nil {
		return nil, fmt.Errorf("Unmarshalling error: %w", err)
	}
	return &pkt, nil
}

func (l *wsLink) receiveTimeout(timeout time.Duration) (*Packet, error) {
	l.conn.SetReadDeadline(time.Now().Add(timeout))
	pkt, err := l.receive()
	l.conn.SetReadDeadline(time.Time{})
	return pkt, err
}

var pingDuration = 4 * time.Second
var pingTimeout = 2*pingDuration + time.Second

func (l *wsLink) receivePoll() (*Packet, error) {
	for {
		if time.Since(l.lastSend) >= pingDuration {
			if err := l.Send(&Packet{Path: "/ping"}); err != nil {
				return nil, err
			}
		}
		pkt, err := l.receiveTimeout(time.Second)
		if err == nil {
			if pkt.Path == "/ping" {
				continue
			}
			return pkt, nil
		}
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			if time.Since(l.lastRecv) > pingTimeout {
				return nil, err
			}
			continue
		}
		return nil, err
	}
	return nil, nil
}

func newConfig(url *url.URL, user, passwd string) (*websocket.Config, error) {
	var surl = url.String()
	var origin = "http://localhost/"

	// Configure the websocket
	config, err := websocket.NewConfig(surl, origin)
	if err != nil {
		return nil, err
	}

	// If valid user, set the basic auth header for the request
	if user != "" {
		req, err := http.NewRequest("GET", surl, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, passwd)
		config.Header = req.Header
	}

	return config, nil
}

func wsDial(url *url.URL) {
	var user = getenv("USER", "")
	var passwd = getenv("PASSWD", "")

	cfg, err := newConfig(url, user, passwd)
	if err != nil {
		fmt.Println("Error configuring websocket:", err)
		return
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(cfg)
		if err == nil {
			// Service the client websocket
			wsClient(conn)
		} else {
			fmt.Println("Dial error", url, err)
		}

		// Try again in a second
		time.Sleep(time.Second)
	}
}

func wsClient(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{conn: conn}
	var ann = announcement{
		Id:           root.Id,
		Model:        root.Model,
		Name:         root.Name,
		DeployParams: root.DeployParams,
	}
	var pkt = &Packet{
		Dst:  ann.Id,
		Path: "/announce",
	}

	// Send announcement
	fmt.Println("Sending announment:", pkt)
	err := link.Send(pkt.Marshal(&ann))
	if err != nil {
		fmt.Println("Send error:", err)
		return
	}

	// Receive welcome within 1 sec
	pkt, err = link.receiveTimeout(time.Second)
	if err != nil {
		fmt.Println("Receive error:", err)
		return
	}

	fmt.Println("Reply from announcement:", pkt)
	if pkt.Path != "/welcome" {
		fmt.Println("Not welcomed, got:", pkt.Path)
		return
	}

	fmt.Println("Adding Uplink")
	uplinksAdd(link)

	// Send /state packets to all devices

	fmt.Println("Sending /state to all devices")
	devicesMu.RLock()
	for id, d := range devices {
		var pkt = &Packet{
			Dst:  id,
			Path: "/state",
		}
		d.RLock()
		pkt.Marshal(d.State)
		d.RUnlock()
		fmt.Println("Sending:", pkt)
		link.Send(pkt)
	}
	devicesMu.RUnlock()

	// Route incoming packets down to the destination device.  Stop and
	// disconnect on EOF.

	fmt.Println("Receiving packets...")
	for {
		pkt, err := link.receivePoll()
		if err != nil {
			fmt.Println("Error receiving packet:", err)
			break
		}
		fmt.Println("Route packet DOWN:", pkt)
		deviceRouteDown(pkt.Dst, pkt)
	}

	fmt.Println("Removing Uplink")
	uplinksRemove(link)
}

func wsServer(conn *websocket.Conn) {

	defer conn.Close()

	var link = &wsLink{conn: conn}

	// First receive should be an /announce packet

	pkt, err := link.receive()
	if err != nil {
		fmt.Println("Error receiving first packet:", err)
		return
	}

	if pkt.Path != "/announce" {
		fmt.Println("Not Announcement, got:", pkt.Path)
		return
	}
	fmt.Println("Announcement", pkt)

	var ann announcement
	pkt.Unmarshal(&ann)

	if ann.Id != pkt.Dst {
		fmt.Println("Error: id mismatch", ann.Id, pkt.Dst)
		return
	}

	if ann.Id == root.Id {
		fmt.Println("Error: can't dial into self")
		return
	}

	if err := deviceOnline(ann); err != nil {
		fmt.Println("Device online error:", err)
		return
	}

	// Announcement is good, reply with /welcome packet

	link.Send(pkt.SetPath("/welcome"))

	// Add as active download link

	fmt.Println("Adding Downlink")
	id := ann.Id
	downlinksAdd(id, link)

	// Route incoming packets up to the destination device.  Stop and
	// disconnect on EOF.

	for {
		pkt, err := link.receivePoll()
		if err != nil {
			fmt.Println("Error receiving packet:", err)
			break
		}
		fmt.Println("Route packet UP:", pkt)
		deviceRouteUp(pkt.Dst, pkt)
	}

	fmt.Println("Removing Downlink")
	downlinksRemove(id)

	deviceOffline(id)
}
