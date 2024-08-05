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

type session struct {
	sessionId  string
	conn       *websocket.Conn
	LastUpdate time.Time
	LastView   map[string]string // key: device Id; value: view
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
		LastView:   make(map[string]string),
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

func sessionDeviceSaveView(sessionId, deviceId, view string) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[sessionId]; ok {
		println("sessionDeviceView", sessionId, deviceId, view)
		session.LastUpdate = time.Now()
		session.LastView[deviceId] = view
	}
}

func (s session) _render(deviceId, view string) {
	var buf bytes.Buffer
	if err := _deviceRender(deviceId, view, &buf); err != nil {
		println("session.render error", err.Error())
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func (s session) render(deviceId, view string) {
	var buf bytes.Buffer
	if err := deviceRender(deviceId, view, &buf); err != nil {
		println("session.render error", err.Error())
		return
	}
	websocket.Message.Send(s.conn, string(buf.Bytes()))
}

func sessionDeviceRender(sessionId, deviceId string) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[sessionId]; ok {
		if session.conn == nil {
			return
		}
		if view, ok := session.LastView[deviceId]; ok {
			session.render(deviceId, view)
		}
	}
}

func sessionsRoute(deviceId string) {

	println("sessionsRoute", deviceId)
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	for sessionId, session := range sessions {
		println("sessionsRoute", sessionId)
		if session.conn == nil {
			println("sessionsRoute", sessionId, "skipping")
			continue
		}
		if view, ok := session.LastView[deviceId]; ok {
			println("sessionsRoute", sessionId, "render", deviceId, view)
			session._render(deviceId, view)
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
