package thing2

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

//go:embed template/sessions.tmpl
var sessionTemplate string

type session struct {
	id         string
	conn       *websocket.Conn
	LastUpdate time.Time
	LastView   map[Devicer]string
}

var sessions = make(map[string]*session)
var sessionsMu sync.RWMutex

func init() {
	go gcSessions()
}

func _newSession(id string, conn *websocket.Conn) *session {
	return &session{
		id:         id,
		conn:       conn,
		LastUpdate: time.Now(),
		LastView:   make(map[Devicer]string),
	}
}

func newSession() string {
	// TODO check and limit size of sessions

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	id := uuid.New().String()
	sessions[id] = _newSession(id, nil)
	println("newSession", id)
	return id
}

func (s *session) Age() string {
	return time.Since(s.LastUpdate).String()
}

func sessionConn(id string, conn *websocket.Conn) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[id]; ok {
		println("sessionConn", id, conn)
		session.conn = conn
		session.LastUpdate = time.Now()
	} else {
		sessions[id] = _newSession(id, conn)
	}
}

func sessionUpdate(id string) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[id]; ok {
		println("sessionUpdate", id)
		session.LastUpdate = time.Now()
	}
}

func sessionDeviceView(id, view string, device Devicer) {

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if session, ok := sessions[id]; ok {
		println("sessionDeviceView", id, view, device)
		session.LastUpdate = time.Now()
		session.LastView[device] = view
	}
}

func sessionDeviceRender(id string, device Devicer) {

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	if session, ok := sessions[id]; ok {
		if session.conn == nil {
			return
		}
		if view, ok := session.LastView[device]; ok {
			println("sessionDeviceRender", id, view, device)
			var buf bytes.Buffer
			if err := device.Render(&buf, view); err != nil {
				println("device.Render error", err.Error())
				return
			}
			websocket.Message.Send(session.conn, string(buf.Bytes()))
		}
	}
}

func sessionsShow(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("sessions").Parse(sessionTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	err = tmpl.Execute(w, sessions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func gcSessions() {
	dur := 1 * time.Minute
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for range ticker.C {
		println("gcSessions ticked")
		sessionsMu.Lock()
		for id, session := range sessions {
			println("gcSessions considering", id)
			if time.Since(session.LastUpdate) > dur {
				println("gcSessions", id)
				delete(sessions, id)
			}
		}
		sessionsMu.Unlock()
	}
}
