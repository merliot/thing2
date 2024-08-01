package thing2

import (
	"bytes"
	_ "embed"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

//go:embed template/sessions.tmpl
var sessionsTemplate string

type lastView struct {
	view   string
	device Devicer
}

type session struct {
	sessionId  string
	conn       *websocket.Conn
	LastUpdate time.Time
	LastViews  map[string]lastView // keyed by device Id
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex

func init() {
	go gcSessions()
}

func _newSession(sessionId string, conn *websocket.Conn) *session {
	return &session{
		sessionId:  sessionId,
		conn:       conn,
		LastUpdate: time.Now(),
		LastViews:  make(map[string]lastView),
	}
}

func newSession() string {
	// TODO check and limit size of sessions

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	sessionId := uuid.New().String()
	sessions[sessionId] = _newSession(sessionId, nil)
	println("newSession", sessionId)
	return sessionId
}

func (s session) Age() string {
	return time.Since(s.LastUpdate).String()
}

func sessionConn(sessionId string, conn *websocket.Conn) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		println("sessionConn", sessionId, conn)
		session.conn = conn
		session.LastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, conn)
	}
}

func sessionUpdate(sessionId string) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		println("sessionUpdate", sessionId)
		session.LastUpdate = time.Now()
	} else {
		sessions[sessionId] = _newSession(sessionId, nil)
	}
}

func sessionDeviceSaveView(sessionId string, device Devicer, view string) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		println("sessionDeviceView", sessionId, view, device)
		session.LastUpdate = time.Now()
		session.LastViews[device.GetId()] = lastView{
			view:   view,
			device: device,
		}
	}
}

func (s session) render(device Devicer, view string) {
	var buf bytes.Buffer
	if err := device.Render(&buf, view); err != nil {
		println("session.render error", err.Error())
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionDeviceRender(sessionId string, device Devicer) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[sessionId]; ok {
		if session.conn == nil {
			return
		}
		if last, ok := session.LastViews[device.GetId()]; ok {
			session.render(device, last.view)
		}
	}
}

func sessionsRoute(pkt *Packet) {

	println("sessionsRoute", pkt.String())
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for sessionId, session := range sessions {
		println("sessionsRoute", sessionId)
		if session.conn == nil {
			println("sessionsRoute", sessionId, "skipping")
			continue
		}
		if last, ok := session.LastViews[pkt.Dst]; ok {
			println("sessionsRoute", sessionId, "render", pkt.Dst, last.view, last.device.String())
			session.render(last.device, last.view)
		}

	}
}

func sessionsShow(w http.ResponseWriter, r *http.Request) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	templateShow(w, sessionsTemplate, sessions)
}

func gcSessions() {
	dur := 1 * time.Minute
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for range ticker.C {
		println("gcSessions ticked")
		sessionsMu.Lock()
		for sessionId, session := range sessions {
			println("gcSessions considering", sessionId)
			if time.Since(session.LastUpdate) > dur {
				println("gcSessions", sessionId)
				delete(sessions, sessionId)
			}
		}
		sessionsMu.Unlock()
	}
}
